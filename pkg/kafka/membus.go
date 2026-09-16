package kafka

import (
	"context"
	"sync"

	"github.com/shivank0310/cex.git/pkg/events"
)

// MemBus is an in-memory pub/sub bus for tests and local dev without Kafka.
type MemBus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
}

func NewMemBus() *MemBus {
	return &MemBus{subscribers: make(map[string][]Handler)}
}

func (b *MemBus) Publish(ctx context.Context, topic string, _ string, env events.Envelope) error {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.subscribers[topic]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		if err := h(ctx, env); err != nil {
			return err
		}
	}
	return nil
}

func (b *MemBus) Subscribe(topic string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], handler)
}

func (b *MemBus) Close() error { return nil }
