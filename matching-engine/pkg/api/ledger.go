package api

import "github.com/shivank0310/cex.git/matching-engine/internal/settlement"

// Balance tracks available and locked funds per user per asset.
type Balance struct {
	Available int64
	Locked    int64
}

// Ledger is the public ledger facade for balance operations.
type Ledger struct {
	inner settlement.Store
}

func NewLedger() *Ledger {
	return &Ledger{inner: settlement.NewLedger()}
}

// NewPermissiveLedger returns a ledger facade that skips internal balance
// mutations. Use when ledger-service owns fund holds and settlement.
func NewPermissiveLedger() *Ledger {
	return &Ledger{inner: settlement.NewPermissiveLedger()}
}

func (l *Ledger) Deposit(userID, asset string, amount int64) {
	l.inner.Deposit(userID, asset, amount)
}

func (l *Ledger) Get(userID, asset string) Balance {
	b := l.inner.Get(userID, asset)
	return Balance{Available: b.Available, Locked: b.Locked}
}
