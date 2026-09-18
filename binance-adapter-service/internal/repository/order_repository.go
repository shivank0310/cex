package repository

import (
	"sync"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/binance"
)

// OrderRepository tracks client order ID → user ID mappings.
type OrderRepository struct {
	mu     sync.RWMutex
	orders map[string]orderRecord
}

type orderRecord struct {
	UserID        string
	ClientOrderID string
	BinanceResult binance.OrderResult
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{orders: make(map[string]orderRecord)}
}

func (r *OrderRepository) Save(userID string, result binance.OrderResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[result.ClientOrderID] = orderRecord{
		UserID: userID, ClientOrderID: result.ClientOrderID, BinanceResult: result,
	}
}

func (r *OrderRepository) Get(clientOrderID string) (orderRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	rec, ok := r.orders[clientOrderID]
	return rec, ok
}

func (r *OrderRepository) Update(result binance.OrderResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rec, ok := r.orders[result.ClientOrderID]; ok {
		rec.BinanceResult = result
		r.orders[result.ClientOrderID] = rec
	}
}
