package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/shivank0310/cex.git/wallet-service/internal/apperrors"
	"github.com/shivank0310/cex.git/wallet-service/internal/client"
	"github.com/shivank0310/cex.git/wallet-service/internal/config"
	"github.com/shivank0310/cex.git/wallet-service/internal/model"
	"github.com/shivank0310/cex.git/wallet-service/internal/repository"
	"github.com/shivank0310/cex.git/wallet-service/internal/risk"
	"github.com/shivank0310/cex.git/wallet-service/internal/security"
)

// WalletService orchestrates blockchain wallets, deposits, and secure withdrawals.
type WalletService struct {
	cfg        config.Config
	repo       *repository.WalletRepository
	ledger     client.LedgerClient
	blockchain client.BlockchainClient
	pipeline   *WithdrawalPipeline
	mfa        *security.TOTPVerifier
	walletSeq  uint64
	depositSeq uint64
	withdrawSeq uint64
}

func NewWalletService(
	cfg config.Config,
	repo *repository.WalletRepository,
	ledger client.LedgerClient,
	blockchain client.BlockchainClient,
) *WalletService {
	riskEngine := risk.NewEngine(cfg)
	mfa := security.NewTOTPVerifier()
	approval := security.NewApprovalPolicy(security.NewApprovalStore())
	hsm := security.NewHSMSignerPool()
	multisig := security.NewMultisigWallet()

	return &WalletService{
		cfg:        cfg,
		repo:       repo,
		ledger:     ledger,
		blockchain: blockchain,
		pipeline:   NewWithdrawalPipeline(riskEngine, mfa, approval, hsm, multisig),
		mfa:        mfa,
	}
}

func (s *WalletService) CreateDepositAddress(ctx context.Context, userID, asset, chain string) (model.Wallet, error) {
	if userID == "" || asset == "" {
		return model.Wallet{}, apperrors.New(apperrors.CodeInvalidRequest, "user_id and asset required")
	}
	if chain == "" {
		chain = defaultChain(asset)
	}

	if existing, ok := s.repo.GetWallet(userID, asset); ok {
		return *existing, nil
	}

	address, err := s.blockchain.GenerateAddress(ctx, userID, asset, chain)
	if err != nil {
		return model.Wallet{}, apperrors.Wrap(apperrors.CodeBlockchainError, "failed to generate address", err)
	}

	wallet := model.Wallet{
		ID: s.nextWalletID(), UserID: userID, Asset: asset,
		Chain: chain, Address: address, CreatedAt: time.Now().UTC(),
	}
	s.repo.SaveWallet(&wallet)
	return wallet, nil
}

func (s *WalletService) GetAccount(ctx context.Context, userID string) (client.AccountResponse, error) {
	account, err := s.ledger.GetAccount(ctx, userID)
	if err != nil {
		return client.AccountResponse{}, apperrors.Wrap(apperrors.CodeLedgerError, "failed to fetch account", err)
	}
	return account, nil
}

func (s *WalletService) GetLedgerBalance(ctx context.Context, userID, asset string) (model.LedgerBalance, error) {
	bal, err := s.ledger.GetBalance(ctx, userID, asset)
	if err != nil {
		return model.LedgerBalance{}, apperrors.Wrap(apperrors.CodeLedgerError, "failed to fetch balance", err)
	}
	return bal, nil
}

func (s *WalletService) ConfirmDeposit(ctx context.Context, txHash, toAddress string, amount int64, confirmations int) (model.Deposit, error) {
	if amount < s.cfg.MinDeposit {
		return model.Deposit{}, apperrors.New(apperrors.CodeInvalidRequest, "deposit below minimum")
	}

	wallet, ok := s.repo.GetWalletByAddress(toAddress)
	if !ok {
		return model.Deposit{}, apperrors.New(apperrors.CodeWalletNotFound, "unknown deposit address")
	}

	deposit := model.Deposit{
		ID: s.nextDepositID(), UserID: wallet.UserID, Asset: wallet.Asset,
		Amount: amount, TxHash: txHash, ToAddress: toAddress,
		Confirmations: confirmations, Status: model.DepositConfirmed,
		LedgerRef: "deposit-" + txHash, CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.SaveDeposit(&deposit); err != nil {
		return model.Deposit{}, err
	}

	_, err := s.ledger.Deposit(ctx, wallet.UserID, wallet.Asset, amount, deposit.LedgerRef)
	if err != nil {
		deposit.Status = model.DepositFailed
		return deposit, apperrors.Wrap(apperrors.CodeLedgerError, "ledger credit failed", err)
	}
	return deposit, nil
}

// RequestWithdrawal initiates the secure withdrawal pipeline.
// Funds NEVER go directly to blockchain — they pass through risk, MFA, approval, HSM.
func (s *WalletService) RequestWithdrawal(ctx context.Context, userID, asset string, amount int64, toAddress string) (model.Withdrawal, error) {
	if err := s.blockchain.ValidateAddress(ctx, asset, toAddress); err != nil {
		return model.Withdrawal{}, apperrors.Wrap(apperrors.CodeInvalidRequest, "invalid address", err)
	}

	balance, err := s.ledger.GetBalance(ctx, userID, asset)
	if err != nil {
		return model.Withdrawal{}, apperrors.Wrap(apperrors.CodeLedgerError, "balance check failed", err)
	}

	riskResult, status, err := s.pipeline.EvaluateRisk(risk.WithdrawalRequest{
		UserID: userID, Asset: asset, Amount: amount,
		ToAddress: toAddress, Balance: balance,
	})
	if err != nil {
		return model.Withdrawal{}, err
	}

	withdrawal := model.Withdrawal{
		ID: s.nextWithdrawID(), UserID: userID, Asset: asset,
		Amount: amount, ToAddress: toAddress, Status: status,
		ApprovalTier: string(riskResult.Tier),
		LedgerRef: "withdraw-" + userID,
		CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
	s.repo.SaveWithdrawal(&withdrawal)

	// AUTO tier with whitelisted address → execute immediately.
	if status == model.WithdrawalApproved {
		return s.executeWithdrawal(ctx, &withdrawal)
	}

	return withdrawal, nil
}

// VerifyMFA completes MFA verification for a pending withdrawal.
func (s *WalletService) VerifyMFA(ctx context.Context, withdrawalID, userID, code string) (model.Withdrawal, error) {
	w, err := s.getUserWithdrawal(withdrawalID, userID)
	if err != nil {
		return model.Withdrawal{}, err
	}
	if w.Status != model.WithdrawalPendingMFA {
		return model.Withdrawal{}, apperrors.New(apperrors.CodeInvalidRequest, "withdrawal not awaiting MFA")
	}

	tier := risk.ApprovalTier(w.ApprovalTier)
	if err := s.pipeline.VerifyMFA(userID, code, tier); err != nil {
		return model.Withdrawal{}, err
	}

	w.MFAVerified = true
	w.Status = s.pipeline.NextStatusAfterMFA(tier)
	w.UpdatedAt = time.Now().UTC()
	s.repo.SaveWithdrawal(w)

	if w.Status == model.WithdrawalApproved {
		return s.executeWithdrawal(ctx, w)
	}
	return *w, nil
}

// ApproveWithdrawal records manual approval for high-value withdrawals.
func (s *WalletService) ApproveWithdrawal(ctx context.Context, withdrawalID, approverID, note string) (model.Withdrawal, error) {
	w, ok := s.repo.GetWithdrawal(withdrawalID)
	if !ok {
		return model.Withdrawal{}, apperrors.New(apperrors.CodeWithdrawalNotFound, "withdrawal not found")
	}
	if w.Status != model.WithdrawalPendingApproval && w.Status != model.WithdrawalPendingMultisig {
		return model.Withdrawal{}, apperrors.New(apperrors.CodeInvalidRequest, "withdrawal not awaiting approval")
	}

	tier := risk.ApprovalTier(w.ApprovalTier)
	newStatus, err := s.pipeline.Approve(withdrawalID, approverID, note, tier)
	if err != nil {
		return model.Withdrawal{}, err
	}

	w.Status = newStatus
	w.ApprovedBy = approverID
	w.UpdatedAt = time.Now().UTC()
	s.repo.SaveWithdrawal(w)

	if newStatus == model.WithdrawalApproved {
		return s.executeWithdrawal(ctx, w)
	}
	return *w, nil
}

// MultisigSign adds a co-signer signature for treasury withdrawals.
func (s *WalletService) MultisigSign(ctx context.Context, withdrawalID, keyID, signerID string) (model.Withdrawal, error) {
	w, ok := s.repo.GetWithdrawal(withdrawalID)
	if !ok {
		return model.Withdrawal{}, apperrors.New(apperrors.CodeWithdrawalNotFound, "withdrawal not found")
	}
	if w.Status != model.WithdrawalPendingMultisig {
		return model.Withdrawal{}, apperrors.New(apperrors.CodeInvalidRequest, "withdrawal not awaiting multisig")
	}

	count, err := s.pipeline.MultisigSign(withdrawalID, keyID, signerID)
	if err != nil {
		return model.Withdrawal{}, err
	}

	w.MultisigSigs = count
	w.UpdatedAt = time.Now().UTC()
	s.repo.SaveWithdrawal(w)

	if s.pipeline.multisig.IsFullySigned(withdrawalID) {
		w.Status = model.WithdrawalApproved
		s.repo.SaveWithdrawal(w)
		return s.executeWithdrawal(ctx, w)
	}
	return *w, nil
}

// WhitelistAddress marks a destination address as trusted (skips MFA tier upgrade).
func (s *WalletService) WhitelistAddress(userID, address string) {
	s.pipeline.WhitelistAddress(userID, address)
}

// EnableMFA enables 2FA for a user.
func (s *WalletService) EnableMFA(userID string) {
	s.mfa.Enable(userID)
}

func (s *WalletService) GetWithdrawal(id string) (model.Withdrawal, error) {
	w, ok := s.repo.GetWithdrawal(id)
	if !ok {
		return model.Withdrawal{}, apperrors.New(apperrors.CodeWithdrawalNotFound, "withdrawal not found")
	}
	return *w, nil
}

// executeWithdrawal runs the final steps: reserve → HSM sign → broadcast → debit.
func (s *WalletService) executeWithdrawal(ctx context.Context, w *model.Withdrawal) (model.Withdrawal, error) {
	if err := s.pipeline.CanExecute(w); err != nil {
		return *w, err
	}

	// 1. Reserve funds (available → locked)
	_, err := s.ledger.Reserve(ctx, w.UserID, w.Asset, w.Amount)
	if err != nil {
		w.Status = model.WithdrawalFailed
		w.RiskNote = "reservation failed"
		s.repo.SaveWithdrawal(w)
		return *w, apperrors.Wrap(apperrors.CodeInsufficientBalance, "reservation failed", err)
	}
	w.Status = model.WithdrawalReserved
	s.repo.SaveWithdrawal(w)

	// 2. HSM signing (private keys never leave HSM)
	w.Status = model.WithdrawalSigning
	s.repo.SaveWithdrawal(w)

	signed, err := s.pipeline.SignWithHSM(w)
	if err != nil {
		_, _ = s.ledger.Release(ctx, w.UserID, w.Asset, w.Amount)
		w.Status = model.WithdrawalFailed
		w.RiskNote = "HSM signing failed: " + err.Error()
		s.repo.SaveWithdrawal(w)
		return *w, apperrors.Wrap(apperrors.CodeInternal, "HSM signing failed", err)
	}
	w.HSMSignature = signed.Signature
	w.HSMKeyID = signed.KeyID

	// 3. Broadcast to blockchain
	txHash, err := s.blockchain.BroadcastWithdrawal(ctx, w.Asset, w.ToAddress, w.Amount)
	if err != nil {
		_, _ = s.ledger.Release(ctx, w.UserID, w.Asset, w.Amount)
		w.Status = model.WithdrawalFailed
		w.RiskNote = err.Error()
		s.repo.SaveWithdrawal(w)
		return *w, apperrors.Wrap(apperrors.CodeBlockchainError, "broadcast failed", err)
	}

	w.TxHash = txHash
	w.Status = model.WithdrawalBroadcast
	s.repo.SaveWithdrawal(w)

	// 4. Debit locked funds
	_, err = s.ledger.DebitLocked(ctx, w.UserID, w.Asset, w.Amount)
	if err != nil {
		return *w, apperrors.Wrap(apperrors.CodeLedgerError, "debit locked failed", err)
	}

	w.Status = model.WithdrawalCompleted
	w.UpdatedAt = time.Now().UTC()
	s.pipeline.RecordCompleted(w.UserID, w.Amount)
	s.repo.SaveWithdrawal(w)
	return *w, nil
}

func (s *WalletService) getUserWithdrawal(id, userID string) (*model.Withdrawal, error) {
	w, ok := s.repo.GetWithdrawal(id)
	if !ok {
		return nil, apperrors.New(apperrors.CodeWithdrawalNotFound, "withdrawal not found")
	}
	if w.UserID != userID {
		return nil, apperrors.New(apperrors.CodeWithdrawalNotFound, "withdrawal not found")
	}
	return w, nil
}

func defaultChain(asset string) string {
	switch asset {
	case "BTC":
		return "bitcoin"
	case "ETH":
		return "ethereum"
	default:
		return "erc20"
	}
}

func (s *WalletService) nextWalletID() string {
	return fmt.Sprintf("W-%d", atomic.AddUint64(&s.walletSeq, 1))
}
func (s *WalletService) nextDepositID() string {
	return fmt.Sprintf("D-%d", atomic.AddUint64(&s.depositSeq, 1))
}
func (s *WalletService) nextWithdrawID() string {
	return fmt.Sprintf("WD-%d", atomic.AddUint64(&s.withdrawSeq, 1))
}
