package model

import (
	"strings"
	"sync"
)

// Registry holds all tradable symbols for the exchange.
type Registry struct {
	mu      sync.RWMutex
	symbols map[string]Symbol
}

func NewRegistry() *Registry {
	return &Registry{symbols: make(map[string]Symbol)}
}

func (r *Registry) Register(s Symbol) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.symbols[normalize(s.Name)] = s
}

func (r *Registry) Get(symbol string) (Symbol, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.symbols[normalize(symbol)]
	return s, ok
}

func (r *Registry) List() []Symbol {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Symbol, 0, len(r.symbols))
	for _, s := range r.symbols {
		out = append(out, s)
	}
	return out
}

func normalize(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

// DefaultSymbols returns production-like defaults for BTC/USDT.
func DefaultSymbols() []Symbol {
	return []Symbol{
		{
			Name:        "BTC/USDT",
			BaseAsset:   "BTC",
			QuoteAsset:  "USDT",
			Active:      true,
			MinPrice:    1,
			MaxPrice:    10_000_000,
			TickSize:    1,
			MinQuantity: 1,
			MaxQuantity: 1_000_000,
			LotSize:     1,
			MinNotional: 10,
		},
	}
}
