package service

import (
	"context"
	"fmt"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/order-service/internal/client"
	"github.com/shivank0310/cex.git/order-service/internal/model"
)

// FundsHoldManager locks and releases order funds via ledger-service
// following Binance's free → locked hold model.
type FundsHoldManager struct {
	ledger client.LedgerClient
}

func NewFundsHoldManager(ledger client.LedgerClient) *FundsHoldManager {
	return &FundsHoldManager{ledger: ledger}
}

func (m *FundsHoldManager) LockForOrder(ctx context.Context, userID string, sym model.Symbol, side meapi.Side, orderType meapi.OrderType, price, quantity int64) error {
	if m == nil || m.ledger == nil {
		return nil
	}

	asset, amount, err := lockRequirement(sym, side, orderType, price, quantity)
	if err != nil {
		return err
	}
	if amount <= 0 {
		return nil
	}
	return m.ledger.Reserve(ctx, userID, asset, amount)
}

func (m *FundsHoldManager) UnlockOrder(ctx context.Context, userID string, sym model.Symbol, side meapi.Side, orderType meapi.OrderType, price, remaining int64) error {
	if m == nil || m.ledger == nil || remaining <= 0 {
		return nil
	}

	asset, amount, err := lockRequirement(sym, side, orderType, price, remaining)
	if err != nil {
		return err
	}
	if amount <= 0 {
		return nil
	}
	return m.ledger.Release(ctx, userID, asset, amount)
}

func lockRequirement(sym model.Symbol, side meapi.Side, orderType meapi.OrderType, price, quantity int64) (asset string, amount int64, err error) {
	switch side {
	case meapi.Buy:
		if orderType == meapi.Market {
			return "", 0, fmt.Errorf("market buy fund hold requires quote amount")
		}
		return sym.QuoteAsset, decimal.Notional(price, quantity), nil
	case meapi.Sell:
		return sym.BaseAsset, quantity, nil
	default:
		return "", 0, fmt.Errorf("unknown side: %s", side)
	}
}
