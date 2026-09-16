package redis

import (
	"context"
	"time"
)

// OrderStateStore holds temporary order state with TTL.
// Used after order submission for fast lookup before PostgreSQL persistence.
// The matching engine itself stays in RAM — this is surrounding infrastructure only.
type OrderStateStore struct {
	cache *Cache
	ttl   time.Duration
}

func NewOrderStateStore(cache *Cache, ttl time.Duration) *OrderStateStore {
	return &OrderStateStore{cache: cache, ttl: ttl}
}

func (s *OrderStateStore) Save(ctx context.Context, orderID string, state interface{}) error {
	return s.cache.Set(ctx, PendingOrderKey(orderID), state, s.ttl)
}

func (s *OrderStateStore) Get(ctx context.Context, orderID string, dest interface{}) (bool, error) {
	return s.cache.Get(ctx, PendingOrderKey(orderID), dest)
}

func (s *OrderStateStore) Delete(ctx context.Context, orderID string) error {
	return s.cache.Delete(ctx, PendingOrderKey(orderID))
}
