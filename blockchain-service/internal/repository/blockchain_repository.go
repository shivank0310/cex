package repository

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/shivank0310/cex.git/blockchain-service/internal/model"
)

type Repository struct {
	mu            sync.RWMutex
	wallets       map[string]*model.CustodyWallet // key: userID:asset
	byAddress     map[string]*model.CustodyWallet
	watched       map[string]*model.WatchedAddress
	transactions  map[string]*model.Transaction
	processedDeps map[string]bool
	walletSeq     uint64
	txSeq         uint64
}

func NewRepository() *Repository {
	return &Repository{
		wallets:       make(map[string]*model.CustodyWallet),
		byAddress:     make(map[string]*model.CustodyWallet),
		watched:       make(map[string]*model.WatchedAddress),
		transactions:  make(map[string]*model.Transaction),
		processedDeps: make(map[string]bool),
	}
}

func (r *Repository) NextWalletID() string {
	return fmt.Sprintf("CW-%d", atomic.AddUint64(&r.walletSeq, 1))
}

func (r *Repository) NextTxID() string {
	return fmt.Sprintf("TX-%d", atomic.AddUint64(&r.txSeq, 1))
}

func (r *Repository) SaveWallet(w *model.CustodyWallet) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := w.UserID + ":" + w.Asset
	r.wallets[key] = w
	r.byAddress[w.Address] = w
}

func (r *Repository) GetWallet(userID, asset string) (*model.CustodyWallet, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.wallets[userID+":"+asset]
	return w, ok
}

func (r *Repository) GetWalletByAddress(address string) (*model.CustodyWallet, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.byAddress[address]
	return w, ok
}

func (r *Repository) WatchAddress(w *model.WatchedAddress) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.watched[w.Address] = w
}

func (r *Repository) ListWatchedAddresses() []model.WatchedAddress {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.WatchedAddress, 0, len(r.watched))
	for _, w := range r.watched {
		out = append(out, *w)
	}
	return out
}

func (r *Repository) UpdateWatchedBlock(address string, block uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if w, ok := r.watched[address]; ok {
		w.LastBlock = block
	}
}

func (r *Repository) SaveTransaction(tx *model.Transaction) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transactions[tx.TxHash] = tx
}

func (r *Repository) GetTransaction(txHash string) (*model.Transaction, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tx, ok := r.transactions[txHash]
	return tx, ok
}

func (r *Repository) IsDepositProcessed(txHash string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.processedDeps[txHash]
}

func (r *Repository) MarkDepositProcessed(txHash string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processedDeps[txHash] = true
}

func (r *Repository) ListPendingTransactions() []*model.Transaction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*model.Transaction, 0)
	for _, tx := range r.transactions {
		if tx.Status == model.TxPending {
			out = append(out, tx)
		}
	}
	return out
}
