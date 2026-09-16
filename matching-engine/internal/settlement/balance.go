package settlement

import (
	"fmt"
	"sync"

	"github.com/shivank0310/cex.git/matching-engine/internal/order"
	"github.com/shivank0310/cex.git/matching-engine/internal/trade"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
)

// Store performs balance mutations for order placement and trade settlement.
type Store interface {
	Deposit(userID, asset string, amount int64)
	Get(userID, asset string) Balance
	LockForOrder(o *order.Order, baseAsset, quoteAsset string) error
	UnlockOrder(o *order.Order, baseAsset, quoteAsset string)
	SettleTrade(t *trade.Trade, baseAsset, quoteAsset string, buyLimitPrice int64, makerSide order.Side) error
}

// Balance tracks available and locked funds per user per asset.
type Balance struct {
	Available int64
	Locked    int64
}

func (b *Balance) Total() int64 {
	return b.Available + b.Locked
}

// Ledger performs atomic balance updates for order placement and trade settlement.
type Ledger struct {
	mu       sync.RWMutex
	balances map[string]map[string]*Balance // userID -> asset -> balance
}

func NewLedger() *Ledger {
	return &Ledger{
		balances: make(map[string]map[string]*Balance),
	}
}

func (l *Ledger) Deposit(userID, asset string, amount int64) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b := l.getOrCreate(userID, asset)
	b.Available += amount
}

func (l *Ledger) getOrCreate(userID, asset string) *Balance {
	if l.balances[userID] == nil {
		l.balances[userID] = make(map[string]*Balance)
	}
	if l.balances[userID][asset] == nil {
		l.balances[userID][asset] = &Balance{}
	}
	return l.balances[userID][asset]
}

func (l *Ledger) Get(userID, asset string) Balance {
	l.mu.RLock()
	defer l.mu.RUnlock()
	b := l.getOrCreate(userID, asset)
	return Balance{Available: b.Available, Locked: b.Locked}
}

// LockForOrder reserves funds when an order is submitted.
func (l *Ledger) LockForOrder(o *order.Order, baseAsset, quoteAsset string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch o.Side {
	case order.Buy:
		var lockAmount int64
		if o.Type == order.Market {
			return fmt.Errorf("market buy requires quote amount lock; use LockQuote")
		}
		lockAmount = decimal.Notional(o.Price, o.Remaining)
		b := l.getOrCreate(o.UserID, quoteAsset)
		if b.Available < lockAmount {
			return fmt.Errorf("insufficient %s: need %d, have %d", quoteAsset, lockAmount, b.Available)
		}
		b.Available -= lockAmount
		b.Locked += lockAmount
	case order.Sell:
		b := l.getOrCreate(o.UserID, baseAsset)
		if b.Available < o.Remaining {
			return fmt.Errorf("insufficient %s: need %d, have %d", baseAsset, o.Remaining, b.Available)
		}
		b.Available -= o.Remaining
		b.Locked += o.Remaining
	default:
		return fmt.Errorf("unknown side: %s", o.Side)
	}
	return nil
}

// UnlockOrder releases unused locked funds after cancel or full fill adjustment.
func (l *Ledger) UnlockOrder(o *order.Order, baseAsset, quoteAsset string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch o.Side {
	case order.Buy:
		if o.Remaining > 0 && o.Type == order.Limit {
			amount := decimal.Notional(o.Price, o.Remaining)
			b := l.getOrCreate(o.UserID, quoteAsset)
			b.Locked -= amount
			b.Available += amount
		}
	case order.Sell:
		if o.Remaining > 0 {
			b := l.getOrCreate(o.UserID, baseAsset)
			b.Locked -= o.Remaining
			b.Available += o.Remaining
		}
	}
}

// SettleTrade atomically moves assets between buyer and seller after a match.
// buyLimitPrice is the buyer's limit price (0 for market buys); enables price-improvement refunds.
func (l *Ledger) SettleTrade(t *trade.Trade, baseAsset, quoteAsset string, buyLimitPrice int64, makerSide order.Side) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	notional := decimal.Notional(t.Price, t.Quantity)

	buyerBase := l.getOrCreate(t.BuyerID, baseAsset)
	buyerQuote := l.getOrCreate(t.BuyerID, quoteAsset)
	sellerBase := l.getOrCreate(t.SellerID, baseAsset)
	sellerQuote := l.getOrCreate(t.SellerID, quoteAsset)

	sellerBase.Locked -= t.Quantity

	if buyLimitPrice > 0 {
		lockedSlice := decimal.Notional(buyLimitPrice, t.Quantity)
		buyerQuote.Locked -= lockedSlice
		buyerQuote.Available += lockedSlice - notional
	} else {
		buyerQuote.Locked -= notional
	}

	buyerFee := t.TakerFee
	if makerSide == order.Buy {
		buyerFee = t.MakerFee
	}
	buyerBase.Available += t.Quantity - buyerFee

	sellerFee := t.TakerFee
	if makerSide == order.Sell {
		sellerFee = t.MakerFee
	}
	sellerQuote.Available += notional - sellerFee

	if buyerBase.Available < 0 || buyerQuote.Locked < 0 || sellerBase.Locked < 0 || sellerQuote.Available < 0 {
		return fmt.Errorf("settlement would cause negative balance")
	}

	return nil
}
