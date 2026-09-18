package binance

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/apperrors"
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
