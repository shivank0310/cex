package binance

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/apperrors"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/mapping"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
)

// MockProvider simulates Binance Spot API for local dev without API keys.
type MockProvider struct {
	mu      sync.RWMutex
	orders  map[string]OrderResult
	balances map[string]AccountBalance
	orderSeq uint64
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		orders: make(map[string]OrderResult),
		balances: map[string]AccountBalance{
			"BTC":  {Asset: "BTC", Free: 10000, Locked: 0},
			"USDT": {Asset: "USDT", Free: 100_000_000, Locked: 0},
		},
	}
}

func (m *MockProvider) Ping(_ context.Context) error { return nil }

func (m *MockProvider) PlaceOrder(_ context.Context, req OrderRequest) (OrderResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if req.Quantity <= 0 {
		return OrderResult{}, apperrors.New(apperrors.CodeInvalidRequest, "quantity must be positive")
	}

	orderID := atomic.AddUint64(&m.orderSeq, 1)
	clientID := req.ClientOrderID
	if clientID == "" {
		clientID = fmt.Sprintf("MOCK-%d", orderID)
	}

	result := OrderResult{
		ClientOrderID:  clientID,
		BinanceOrderID: int64(orderID),
		Symbol:         req.Symbol,
		Side:           req.Side,
		Type:           req.Type,
		Status:         "NEW",
		Price:          req.Price,
		Quantity:       req.Quantity,
		ExecutedQty:    0,
		RemainingQty:   req.Quantity,
		TransactTime:   time.Now().UTC(),
	}

	// Simulate immediate fill for market orders.
	if req.Type == "MARKET" {
		result.Status = "FILLED"
		result.ExecutedQty = req.Quantity
		result.RemainingQty = 0
		result.CummulativeQuote = decimal.Notional(req.Price, req.Quantity)
		m.applyFill(req, result.ExecutedQty)
	}

	m.orders[clientID] = result
	return result, nil
}

func (m *MockProvider) CancelOrder(_ context.Context, symbol, clientOrderID string) (OrderResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	o, ok := m.orders[clientOrderID]
	if !ok {
		return OrderResult{}, apperrors.New(apperrors.CodeOrderNotFound, "order not found")
	}
	o.Status = "CANCELED"
	o.RemainingQty = 0
	m.orders[clientOrderID] = o
	return o, nil
}

func (m *MockProvider) GetOrder(_ context.Context, _ string, clientOrderID string) (OrderResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.orders[clientOrderID]
	if !ok {
		return OrderResult{}, apperrors.New(apperrors.CodeOrderNotFound, "order not found")
	}
	return o, nil
}

func (m *MockProvider) GetAccount(_ context.Context) ([]AccountBalance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]AccountBalance, 0, len(m.balances))
	for _, b := range m.balances {
		out = append(out, b)
	}
	return out, nil
}

func (m *MockProvider) GetTickerPrice(_ context.Context, symbol string) (TickerPrice, error) {
	return TickerPrice{Symbol: symbol, Price: 101100}, nil
}

func (m *MockProvider) GetTicker24h(_ context.Context, symbol string) (Ticker24h, error) {
	return Ticker24h{
		Symbol:            symbol,
		LastPrice:         101100,
		BestBid:           101050,
		BestAsk:           101150,
		High24h:           102000,
		Low24h:            100500,
		Volume24h:         12500,
		QuoteVolume24h:    1_260_000_000,
		PriceChange24h:    800,
		PriceChangePct24h: 79,
		TradeCount24h:     8420,
	}, nil
}

func (m *MockProvider) GetDepth(_ context.Context, symbol string, limit int) (DepthSnapshot, error) {
	if limit <= 0 {
		limit = 20
	}
	snap := DepthSnapshot{Symbol: symbol}
	for i := 1; i <= limit; i++ {
		snap.Bids = append(snap.Bids, DepthLevel{
			Price:    101100 - int64(i*50),
			Quantity: int64(10 + i*5),
		})
		snap.Asks = append(snap.Asks, DepthLevel{
			Price:    101100 + int64(i*50),
			Quantity: int64(10 + i*5),
		})
	}
	return snap, nil
}

func (m *MockProvider) GetRecentTrades(_ context.Context, symbol string, limit int) ([]MarketTrade, error) {
	if limit <= 0 {
		limit = 30
	}
	now := time.Now().UTC()
	out := make([]MarketTrade, 0, limit)
	for i := 0; i < limit; i++ {
		price := 101100 + int64((i%5)-2)*25
		out = append(out, MarketTrade{
			ID:        fmt.Sprintf("MOCK-T-%d", i+1),
			Symbol:    symbol,
			Price:     price,
			Quantity:  20 + int64(i%3)*10,
			Timestamp: now.Add(-time.Duration(i) * time.Minute),
		})
	}
	return out, nil
}

func (m *MockProvider) GetKlines(_ context.Context, symbol, interval string, limit int) ([]Kline, error) {
	if limit <= 0 {
		limit = 60
	}
	dur, ok := intervalDuration(interval)
	if !ok {
		return nil, apperrors.New(apperrors.CodeInvalidRequest, "unsupported interval")
	}

	now := time.Now().UTC().Truncate(dur)
	out := make([]Kline, 0, limit)
	price := 101000
	for i := limit - 1; i >= 0; i-- {
		openTime := now.Add(-time.Duration(i) * dur)
		closeTime := openTime.Add(dur - time.Millisecond)
		open := int64(price)
		high := open + 150
		low := open - 120
		close := open + int64((i%7)-3)*20
		out = append(out, Kline{
			OpenTime: openTime, CloseTime: closeTime,
			Open: open, High: high, Low: low, Close: close,
			Volume: 500 + int64(i*3), QuoteVolume: decimal.Notional(close, 500),
			TradeCount: 12 + i%5,
		})
		price = int(close)
	}
	return out, nil
}

func intervalDuration(name string) (time.Duration, bool) {
	switch mapping.NormalizeInterval(name) {
	case "1m":
		return time.Minute, true
	case "5m":
		return 5 * time.Minute, true
	case "10m":
		return 10 * time.Minute, true
	case "15m":
		return 15 * time.Minute, true
	case "30m":
		return 30 * time.Minute, true
	case "1h":
		return time.Hour, true
	case "4h":
		return 4 * time.Hour, true
	case "24h":
		return 24 * time.Hour, true
	default:
		return 0, false
	}
}

func (m *MockProvider) applyFill(req OrderRequest, qty int64) {
	if req.Side == "BUY" {
		notional := decimal.Notional(req.Price, qty)
		usdt := m.balances["USDT"]
		usdt.Free -= notional
		m.balances["USDT"] = usdt
		btc := m.balances["BTC"]
		btc.Free += qty
		m.balances["BTC"] = btc
	} else {
		notional := decimal.Notional(req.Price, qty)
		btc := m.balances["BTC"]
		btc.Free -= qty
		m.balances["BTC"] = btc
		usdt := m.balances["USDT"]
		usdt.Free += notional
		m.balances["USDT"] = usdt
	}
}
