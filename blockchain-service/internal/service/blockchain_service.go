package service

import (
	"context"
	"log"
	"time"

	"github.com/shivank0310/cex.git/blockchain-service/internal/apperrors"
	"github.com/shivank0310/cex.git/blockchain-service/internal/client"
	"github.com/shivank0310/cex.git/blockchain-service/internal/config"
	"github.com/shivank0310/cex.git/blockchain-service/internal/evm"
	"github.com/shivank0310/cex.git/blockchain-service/internal/model"
	"github.com/shivank0310/cex.git/blockchain-service/internal/repository"
)

// BlockchainService isolates EVM infrastructure from core exchange services.
type BlockchainService struct {
	cfg    config.Config
	repo   *repository.Repository
	chain  evm.Provider
	wallet client.WalletClient
}

func NewBlockchainService(
	cfg config.Config,
	repo *repository.Repository,
	chain evm.Provider,
	wallet client.WalletClient,
) *BlockchainService {
	return &BlockchainService{cfg: cfg, repo: repo, chain: chain, wallet: wallet}
}

// CreateAddress generates a custody wallet address and registers it for deposit monitoring.
func (s *BlockchainService) CreateAddress(ctx context.Context, userID, asset, chain string) (model.CustodyWallet, error) {
	if userID == "" || asset == "" {
		return model.CustodyWallet{}, apperrors.New(apperrors.CodeInvalidRequest, "user_id and asset required")
	}
	if chain == "" {
		chain = s.chain.Chain()
	}

	if existing, ok := s.repo.GetWallet(userID, asset); ok {
		return *existing, nil
	}

	address, err := s.chain.GenerateAddress(ctx, userID, asset)
	if err != nil {
		return model.CustodyWallet{}, apperrors.Wrap(apperrors.CodeEVMError, "address generation failed", err)
	}

	wallet := model.CustodyWallet{
		ID:        s.repo.NextWalletID(),
		UserID:    userID,
		Asset:     asset,
		Chain:     chain,
		Address:   address,
		CreatedAt: time.Now().UTC(),
	}
	s.repo.SaveWallet(&wallet)
	s.repo.WatchAddress(&model.WatchedAddress{
		Address: address,
		UserID:  userID,
		Asset:   asset,
		Chain:   chain,
		AddedAt: wallet.CreatedAt,
	})

	log.Printf("[blockchain] address %s for %s:%s on %s", address, userID, asset, chain)
	return wallet, nil
}

func (s *BlockchainService) ValidateAddress(_ context.Context, address string) error {
	if !s.chain.ValidateAddress(address) {
		return apperrors.New(apperrors.CodeInvalidRequest, "invalid blockchain address")
	}
	return nil
}

// BroadcastWithdrawal sends funds on-chain and tracks the transaction.
func (s *BlockchainService) BroadcastWithdrawal(ctx context.Context, asset, toAddress string, amount int64) (model.Transaction, error) {
	if err := s.ValidateAddress(ctx, toAddress); err != nil {
		return model.Transaction{}, err
	}
	if amount <= 0 {
		return model.Transaction{}, apperrors.New(apperrors.CodeInvalidRequest, "amount must be positive")
	}

	txHash, err := s.chain.BroadcastWithdrawal(ctx, asset, toAddress, amount)
	if err != nil {
		return model.Transaction{}, apperrors.Wrap(apperrors.CodeEVMError, "broadcast failed", err)
	}

	now := time.Now().UTC()
	tx := model.Transaction{
		ID:          s.repo.NextTxID(),
		Type:        model.TxTypeWithdrawal,
		TxHash:      txHash,
		Asset:       asset,
		Chain:       s.chain.Chain(),
		ToAddress:   toAddress,
		Amount:      amount,
		Status:      model.TxPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.repo.SaveTransaction(&tx)

	if receipt, err := s.chain.GetReceipt(ctx, txHash); err == nil {
		s.applyReceipt(&tx, receipt)
		s.repo.SaveTransaction(&tx)
	}

	log.Printf("[blockchain] withdrawal broadcast %s → %s amount=%d", txHash, toAddress, amount)
	return tx, nil
}

func (s *BlockchainService) GetTransaction(ctx context.Context, txHash string) (model.Transaction, error) {
	tx, ok := s.repo.GetTransaction(txHash)
	if !ok {
		return model.Transaction{}, apperrors.New(apperrors.CodeTransactionNotFound, "transaction not found")
	}

	receipt, err := s.chain.GetReceipt(ctx, txHash)
	if err == nil {
		s.applyReceipt(tx, receipt)
		s.repo.SaveTransaction(tx)
	}
	return *tx, nil
}

// RunDepositMonitor polls watched addresses and credits wallet-service on confirmation.
func (s *BlockchainService) RunDepositMonitor(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.DepositPollInterval)
	defer ticker.Stop()

	log.Printf("[blockchain] deposit monitor started (confirmations=%d)", s.cfg.RequiredConfirmations)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.scanDeposits(ctx)
		}
	}
}

// RunTransactionTracker updates pending transaction statuses.
func (s *BlockchainService) RunTransactionTracker(ctx context.Context) {
	ticker := time.NewTicker(s.cfg.TxTrackPollInterval)
	defer ticker.Stop()

	log.Println("[blockchain] transaction tracker started")
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.trackTransactions(ctx)
		}
	}
}

func (s *BlockchainService) scanDeposits(ctx context.Context) {
	watched := s.repo.ListWatchedAddresses()
	if len(watched) == 0 {
		return
	}

	addresses := make([]string, len(watched))
	minBlock := uint64(0)
	for i, w := range watched {
		addresses[i] = w.Address
		if w.LastBlock > minBlock {
			minBlock = w.LastBlock
		}
	}

	deposits, err := s.chain.ScanDeposits(ctx, addresses, minBlock)
	if err != nil {
		log.Printf("[blockchain] deposit scan error: %v", err)
		return
	}

	for _, dep := range deposits {
		if s.repo.IsDepositProcessed(dep.TxHash) {
			continue
		}
		if dep.Confirmations < s.cfg.RequiredConfirmations {
			continue
		}

		wallet, ok := s.repo.GetWalletByAddress(dep.ToAddress)
		if !ok {
			continue
		}

		if err := s.wallet.ConfirmDeposit(ctx, dep.TxHash, dep.ToAddress, dep.Amount, dep.Confirmations); err != nil {
			log.Printf("[blockchain] wallet confirm failed tx=%s: %v", dep.TxHash, err)
			continue
		}

		now := time.Now().UTC()
		s.repo.SaveTransaction(&model.Transaction{
			ID:            s.repo.NextTxID(),
			Type:          model.TxTypeDeposit,
			TxHash:        dep.TxHash,
			Asset:         dep.Asset,
			Chain:         wallet.Chain,
			ToAddress:     dep.ToAddress,
			Amount:        dep.Amount,
			Status:        model.TxConfirmed,
			Confirmations: dep.Confirmations,
			BlockNumber:   dep.BlockNumber,
			UserID:        wallet.UserID,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
		s.repo.MarkDepositProcessed(dep.TxHash)
		s.repo.UpdateWatchedBlock(dep.ToAddress, dep.BlockNumber)

		if mock, ok := s.chain.(*evm.MockProvider); ok {
			mock.MarkProcessed(dep.TxHash)
		}

		log.Printf("[blockchain] deposit confirmed tx=%s user=%s amount=%d %s",
			dep.TxHash, wallet.UserID, dep.Amount, dep.Asset)
	}
}

func (s *BlockchainService) trackTransactions(ctx context.Context) {
	watched := s.repo.ListWatchedAddresses()
	_ = watched // tracker scans stored txs via GetTransaction on demand; batch scan below

	// Re-check all pending withdrawal transactions.
	for _, w := range s.listPendingTransactions() {
		receipt, err := s.chain.GetReceipt(ctx, w.TxHash)
		if err != nil {
			continue
		}
		s.applyReceipt(w, receipt)
		s.repo.SaveTransaction(w)
	}
}

func (s *BlockchainService) listPendingTransactions() []*model.Transaction {
	// Simple scan: repository doesn't expose list; track via watched isn't ideal.
	// For production, add ListPending to repository. For now use internal access pattern.
	return s.repo.ListPendingTransactions()
}

func (s *BlockchainService) applyReceipt(tx *model.Transaction, receipt *evm.Receipt) {
	tx.Confirmations = receipt.Confirmations
	tx.BlockNumber = receipt.BlockNumber
	tx.Status = receipt.Status
	tx.UpdatedAt = time.Now().UTC()
}

// Provider exposes the underlying EVM provider for tests.
func (s *BlockchainService) Provider() evm.Provider {
	return s.chain
}
