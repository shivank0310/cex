package repository

import (
	"sync"

	"github.com/shivank0310/cex.git/wallet-service/internal/apperrors"
	"github.com/shivank0310/cex.git/wallet-service/internal/model"
)

type WalletRepository struct {
	mu       sync.RWMutex
	wallets  map[string]*model.Wallet       // id -> wallet
	byAddr   map[string]*model.Wallet       // address -> wallet
	deposits map[string]*model.Deposit      // id -> deposit
	byTxHash map[string]*model.Deposit
	withdraw map[string]*model.Withdrawal
}

func NewWalletRepository() *WalletRepository {
	return &WalletRepository{
		wallets:  make(map[string]*model.Wallet),
		byAddr:   make(map[string]*model.Wallet),
		deposits: make(map[string]*model.Deposit),
		byTxHash: make(map[string]*model.Deposit),
		withdraw: make(map[string]*model.Withdrawal),
	}
}

func (r *WalletRepository) SaveWallet(w *model.Wallet) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wallets[w.ID] = w
	r.byAddr[w.Address] = w
}

func (r *WalletRepository) GetWalletByAddress(address string) (*model.Wallet, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.byAddr[address]
	return w, ok
}

func (r *WalletRepository) GetWallet(userID, asset string) (*model.Wallet, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, w := range r.wallets {
		if w.UserID == userID && w.Asset == asset {
			return w, true
		}
	}
	return nil, false
}

func (r *WalletRepository) SaveDeposit(d *model.Deposit) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byTxHash[d.TxHash] != nil {
		return apperrors.New(apperrors.CodeDepositDuplicate, "deposit already processed")
	}
	r.deposits[d.ID] = d
	r.byTxHash[d.TxHash] = d
	return nil
}

func (r *WalletRepository) GetDeposit(id string) (*model.Deposit, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.deposits[id]
	return d, ok
}

func (r *WalletRepository) SaveWithdrawal(w *model.Withdrawal) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.withdraw[w.ID] = w
}

func (r *WalletRepository) GetWithdrawal(id string) (*model.Withdrawal, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.withdraw[id]
	return w, ok
}

func (r *WalletRepository) ListWithdrawals(userID string) []*model.Withdrawal {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*model.Withdrawal, 0)
	for _, w := range r.withdraw {
		if w.UserID == userID {
			out = append(out, w)
		}
	}
	return out
}
