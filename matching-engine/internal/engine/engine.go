package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/shivank0310/cex.git/matching-engine/internal/order"
	"github.com/shivank0310/cex.git/matching-engine/internal/orderbook"
	"github.com/shivank0310/cex.git/matching-engine/internal/settlement"
	"github.com/shivank0310/cex.git/matching-engine/internal/trade"
)

// SymbolConfig maps a trading pair to its base/quote assets.
type SymbolConfig struct {
	BaseAsset  string
	QuoteAsset string
}

// SubmitResult contains the outcome of processing an order.
type SubmitResult struct {
	Order   *order.Order
	Trades  []*trade.Trade
	Error   error
}

// Engine is the core matching engine for one or more symbols.
type Engine struct {
	mu            sync.Mutex
	books         map[string]*orderbook.OrderBook
	symbols       map[string]SymbolConfig
	ledger        *settlement.Ledger
	fees          FeeConfig
	orderSequence uint64
	tradeSequence uint64
}

func New(ledger *settlement.Ledger, fees FeeConfig) *Engine {
	return &Engine{
		books:   make(map[string]*orderbook.OrderBook),
		symbols: make(map[string]SymbolConfig),
		ledger:  ledger,
		fees:    fees,
	}
}

func (e *Engine) RegisterSymbol(symbol string, cfg SymbolConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.symbols[symbol] = cfg
	e.books[symbol] = orderbook.New(symbol)
}

func (e *Engine) GetBook(symbol string) *orderbook.OrderBook {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.books[symbol]
}

func (e *Engine) Ledger() *settlement.Ledger {
	return e.ledger
}

// SubmitOrder validates funds, matches, settles trades, and rests any remainder.
func (e *Engine) SubmitOrder(o *order.Order) SubmitResult {
	e.mu.Lock()
	defer e.mu.Unlock()

	cfg, ok := e.symbols[o.Symbol]
	if !ok {
		return SubmitResult{Order: o, Error: fmt.Errorf("unknown symbol: %s", o.Symbol)}
	}

	e.orderSequence++
	o.Sequence = e.orderSequence

	if err := e.ledger.LockForOrder(o, cfg.BaseAsset, cfg.QuoteAsset); err != nil {
		return SubmitResult{Order: o, Error: err}
	}

	book := e.books[o.Symbol]
	trades := e.match(o, book, cfg)

	for _, t := range trades {
		makerSide := order.Sell
		if t.MakerOrderID == t.BuyOrderID {
			makerSide = order.Buy
		}
		if err := e.ledger.SettleTrade(t, cfg.BaseAsset, cfg.QuoteAsset, t.BuyLimitPrice, makerSide); err != nil {
			return SubmitResult{Order: o, Trades: trades, Error: err}
		}
	}

	if o.IsActive() && o.Type == order.Limit {
		book.Add(o)
	} else if o.IsActive() && o.Type == order.Market {
		// Unfilled market order remainder is cancelled; release locks.
		o.Cancel()
		e.ledger.UnlockOrder(o, cfg.BaseAsset, cfg.QuoteAsset)
	} else if !o.IsActive() && o.Remaining > 0 {
		e.ledger.UnlockOrder(o, cfg.BaseAsset, cfg.QuoteAsset)
	}

	return SubmitResult{Order: o, Trades: trades}
}

// CancelOrder removes a resting order from the book and unlocks funds.
func (e *Engine) CancelOrder(symbol, orderID string) (*order.Order, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cfg, ok := e.symbols[symbol]
	if !ok {
		return nil, fmt.Errorf("unknown symbol: %s", symbol)
	}

	book := e.books[symbol]
	removed := book.RemoveOrder(orderID)
	if removed == nil {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	removed.Cancel()
	e.ledger.UnlockOrder(removed, cfg.BaseAsset, cfg.QuoteAsset)
	return removed, nil
}

func (e *Engine) nextTradeID() string {
	e.tradeSequence++
	return fmt.Sprintf("T-%d", e.tradeSequence)
}

func (e *Engine) newTrade(taker, maker *order.Order, price, qty int64, _ SymbolConfig) *trade.Trade {
	e.tradeSequence++
	fees := computeFees(price, qty, maker.Side, e.fees)

	var buyerID, sellerID, buyOrderID, sellOrderID string
	var buyLimitPrice int64
	if taker.Side == order.Buy {
		buyerID, sellerID = taker.UserID, maker.UserID
		buyOrderID, sellOrderID = taker.ID, maker.ID
		if taker.Type == order.Limit {
			buyLimitPrice = taker.Price
		}
	} else {
		buyerID, sellerID = maker.UserID, taker.UserID
		buyOrderID, sellOrderID = maker.ID, taker.ID
		if maker.Type == order.Limit {
			buyLimitPrice = maker.Price
		}
	}

	return &trade.Trade{
		ID:           fmt.Sprintf("T-%d", e.tradeSequence),
		Symbol:       taker.Symbol,
		Sequence:     e.tradeSequence,
		BuyOrderID:   buyOrderID,
		SellOrderID:  sellOrderID,
		BuyerID:      buyerID,
		SellerID:     sellerID,
		Price:         price,
		Quantity:      qty,
		BuyLimitPrice: buyLimitPrice,
		MakerOrderID: maker.ID,
		TakerOrderID: taker.ID,
		MakerFee:     fees.makerFee,
		TakerFee:     fees.takerFee,
		Timestamp:    time.Now().UTC(),
	}
}
