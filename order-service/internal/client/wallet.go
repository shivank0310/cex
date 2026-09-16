package client

import (
	"context"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
)

// WalletClient abstracts balance lookups (ledger-service in production).
type WalletClient interface {
	GetAvailable(userID, asset string) (int64, error)
}

// LedgerWallet reads balances from the matching-engine in-memory ledger (demo mode).
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

// HTTPLedgerWallet reads free balances from ledger-service (Binance-style).
type HTTPLedgerWallet struct {
	ledger LedgerClient
}

func NewHTTPLedgerWallet(ledger LedgerClient) *HTTPLedgerWallet {
	return &HTTPLedgerWallet{ledger: ledger}
}

func (w *HTTPLedgerWallet) GetAvailable(userID, asset string) (int64, error) {
	bal, err := w.ledger.GetBalance(context.Background(), userID, asset)
	if err != nil {
		return 0, err
	}
	return bal.Available, nil
}
