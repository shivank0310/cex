package repository

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/shivank0310/cex.git/settlement-service/internal/model"
)

type SettlementRepository struct {
	mu          sync.RWMutex
	byTradeID   map[string]model.Settlement
	byID        map[string]model.Settlement
	settlementSeq uint64
}

func NewSettlementRepository() *SettlementRepository {
	return &SettlementRepository{
		byTradeID: make(map[string]model.Settlement),
		byID:      make(map[string]model.Settlement),
	}
}

func (r *SettlementRepository) NextID() string {
	id := atomic.AddUint64(&r.settlementSeq, 1)
	return fmt.Sprintf("ST-%d", id)
}

func (r *SettlementRepository) ExistsByTradeID(tradeID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.byTradeID[tradeID]
	return ok
}

func (r *SettlementRepository) GetByTradeID(tradeID string) (model.Settlement, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.byTradeID[tradeID]
	return s, ok
}

func (r *SettlementRepository) Save(s model.Settlement) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byTradeID[s.TradeID] = s
	r.byID[s.ID] = s
}
