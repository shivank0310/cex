package client

import (
	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
)

// WalletClient abstracts balance lookups (wallet-service in production).
type WalletClient interface {
	GetAvailable(userID, asset string) (int64, error)
}

// LedgerWallet reads balances from the matching-engine ledger facade.
type LedgerWallet struct {
	ledger *meapi.Ledger
}

func NewLedgerWallet(ledger *meapi.Ledger) *LedgerWallet {
	return &LedgerWallet{ledger: ledger}
}

func (w *LedgerWallet) GetAvailable(userID, asset string) (int64, error) {
	bal := w.ledger.Get(userID, asset)
	return bal.Available, nil
}
