package settlement

import (
	"github.com/shivank0310/cex.git/matching-engine/internal/order"
	"github.com/shivank0310/cex.git/matching-engine/internal/trade"
)

// PermissiveLedger is a no-op ledger used when an external ledger-service
// manages fund holds and trade settlement (Binance-style free/locked model).
type PermissiveLedger struct{}

func NewPermissiveLedger() *PermissiveLedger {
	return &PermissiveLedger{}
}

func (l *PermissiveLedger) Deposit(_ string, _ string, _ int64) {}

func (l *PermissiveLedger) Get(_ string, _ string) Balance {
	return Balance{}
}

func (l *PermissiveLedger) LockForOrder(_ *order.Order, _ string, _ string) error {
	return nil
}

func (l *PermissiveLedger) UnlockOrder(_ *order.Order, _ string, _ string) {}

func (l *PermissiveLedger) SettleTrade(_ *trade.Trade, _ string, _ string, _ int64, _ order.Side) error {
	return nil
}
