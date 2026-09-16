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
)

// WalletService orchestrates blockchain wallets, deposits, and withdrawals.
// Wallet (on-chain address) is separate from LedgerBalance (internal accounting).
type WalletService struct {
	cfg        config.Config
	repo       *repository.WalletRepository
	ledger     client.LedgerClient
	blockchain client.BlockchainClient
	risk       *risk.Checker
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
	return &WalletService{
		cfg:        cfg,
		repo:       repo,
		ledger:     ledger,
		blockchain: blockchain,
		risk:       risk.NewChecker(cfg),
	}
}

// CreateDepositAddress assigns a blockchain wallet address to a user.
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
		ID:        s.nextWalletID(),
		UserID:    userID,
		Asset:     asset,
		Chain:     chain,
		Address:   address,
		CreatedAt: time.Now().UTC(),
	}
	s.repo.SaveWallet(&wallet)
	return wallet, nil
}

// GetLedgerBalance returns the user's internal ledger balance (not blockchain wallet).
func (s *WalletService) GetLedgerBalance(ctx context.Context, userID, asset string) (model.LedgerBalance, error) {
	bal, err := s.ledger.GetBalance(ctx, userID, asset)
	if err != nil {
		return model.LedgerBalance{}, apperrors.Wrap(apperrors.CodeLedgerError, "failed to fetch balance", err)
	}
	return bal, nil
}

// ConfirmDeposit processes an on-chain deposit confirmed by blockchain-service.
//
// Flow: Blockchain Wallet → Deposit → Blockchain Service → Ledger → User Available Balance
func (s *WalletService) ConfirmDeposit(ctx context.Context, txHash, toAddress string, amount int64, confirmations int) (model.Deposit, error) {
	if amount < s.cfg.MinDeposit {
		return model.Deposit{}, apperrors.New(apperrors.CodeInvalidRequest, "deposit below minimum")
	}

	wallet, ok := s.repo.GetWalletByAddress(toAddress)
	if !ok {
		return model.Deposit{}, apperrors.New(apperrors.CodeWalletNotFound, "unknown deposit address")
	}

	deposit := model.Deposit{
		ID:            s.nextDepositID(),
		UserID:        wallet.UserID,
		Asset:         wallet.Asset,
		Amount:        amount,
		TxHash:        txHash,
		ToAddress:     toAddress,
		Confirmations: confirmations,
		Status:        model.DepositConfirmed,
		LedgerRef:     "deposit-" + txHash,
		CreatedAt:     time.Now().UTC(),
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

// RequestWithdrawal initiates a withdrawal with risk checks and ledger reservation.
//
// Flow: User → Withdrawal Request → Risk checks → Ledger reservation → Blockchain broadcast
func (s *WalletService) RequestWithdrawal(ctx context.Context, userID, asset string, amount int64, toAddress string) (model.Withdrawal, error) {
	if err := s.blockchain.ValidateAddress(ctx, asset, toAddress); err != nil {
		return model.Withdrawal{}, apperrors.Wrap(apperrors.CodeInvalidRequest, "invalid address", err)
	}

	balance, err := s.ledger.GetBalance(ctx, userID, asset)
	if err != nil {
		return model.Withdrawal{}, apperrors.Wrap(apperrors.CodeLedgerError, "balance check failed", err)
	}

	if err := s.risk.Validate(risk.WithdrawalRequest{
		UserID: userID, Asset: asset, Amount: amount,
		ToAddress: toAddress, Balance: balance,
	}); err != nil {
		return model.Withdrawal{}, err
	}

	withdrawal := model.Withdrawal{
		ID:        s.nextWithdrawID(),
		UserID:    userID,
		Asset:     asset,
		Amount:    amount,
		ToAddress: toAddress,
		Status:    model.WithdrawalPending,
		LedgerRef: "withdraw-" + userID,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Ledger reservation: move available → locked
	_, err = s.ledger.Reserve(ctx, userID, asset, amount)
	if err != nil {
		return model.Withdrawal{}, apperrors.Wrap(apperrors.CodeInsufficientBalance, "reservation failed", err)
	}
	withdrawal.Status = model.WithdrawalReserved

	// Broadcast to blockchain-service
	txHash, err := s.blockchain.BroadcastWithdrawal(ctx, asset, toAddress, amount)
	if err != nil {
		_, _ = s.ledger.Release(ctx, userID, asset, amount)
		withdrawal.Status = model.WithdrawalFailed
		withdrawal.RiskNote = err.Error()
		s.repo.SaveWithdrawal(&withdrawal)
		return withdrawal, apperrors.Wrap(apperrors.CodeBlockchainError, "broadcast failed", err)
	}

	withdrawal.TxHash = txHash
	withdrawal.Status = model.WithdrawalBroadcast
	s.repo.SaveWithdrawal(&withdrawal)

	// Complete: debit locked funds (sent on-chain)
	_, err = s.ledger.DebitLocked(ctx, userID, asset, amount)
	if err != nil {
		return withdrawal, apperrors.Wrap(apperrors.CodeLedgerError, "debit locked failed", err)
	}
	withdrawal.Status = model.WithdrawalCompleted
	withdrawal.UpdatedAt = time.Now().UTC()
	s.repo.SaveWithdrawal(&withdrawal)

	return withdrawal, nil
}

func (s *WalletService) GetWithdrawal(id string) (model.Withdrawal, error) {
	w, ok := s.repo.GetWithdrawal(id)
	if !ok {
		return model.Withdrawal{}, apperrors.New(apperrors.CodeWithdrawalNotFound, "withdrawal not found")
	}
	return *w, nil
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
