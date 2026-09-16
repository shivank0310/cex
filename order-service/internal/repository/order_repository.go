package repository

import (
	"sync"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
)

// OrderRepository persists orders for query/audit (in-memory for now).
type OrderRepository interface {
	Save(o *meapi.Order) error
	Get(orderID string) (*meapi.Order, error)
}

type InMemoryOrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*meapi.Order
}

func NewInMemoryOrderRepository() *InMemoryOrderRepository {
	return &InMemoryOrderRepository{orders: make(map[string]*meapi.Order)}
}

func (r *InMemoryOrderRepository) Save(o *meapi.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[o.ID] = o
	return nil
}

func (r *InMemoryOrderRepository) Get(orderID string) (*meapi.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.orders[orderID]
	if !ok {
		return nil, apperrors.New(apperrors.CodeOrderNotFound, "order not found")
	}
	return o, nil
}
