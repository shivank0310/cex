package repository

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/shivank0310/cex.git/ledger-service/internal/model"
)

// LedgerRepository is the source of truth for accounts and journals (in-memory; PostgreSQL later).
type LedgerRepository struct {
	mu         sync.RWMutex
	accounts   map[string]*model.Account // key: userID:asset
	journals   []model.Journal
	processed  map[string]bool // trade IDs for idempotency
	journalSeq uint64
	legSeq     uint64
}

func NewLedgerRepository() *LedgerRepository {
	return &LedgerRepository{
		accounts:  make(map[string]*model.Account),
		processed: make(map[string]bool),
	}
}

func (r *LedgerRepository) NextJournalID() string {
	id := atomic.AddUint64(&r.journalSeq, 1)
	return fmt.Sprintf("J-%d", id)
}

func (r *LedgerRepository) nextLegID() string {
	id := atomic.AddUint64(&r.legSeq, 1)
	return fmt.Sprintf("L-%d", id)
}

func (r *LedgerRepository) IsReferenceProcessed(ref string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.processed[ref]
}

func (r *LedgerRepository) ApplyJournal(journal model.Journal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, leg := range journal.Legs {
		if leg.ID == "" {
			journal.Legs[i].ID = r.nextLegID()
		}
		journal.Legs[i].JournalID = journal.ID

		acct := r.getOrCreate(leg.UserID, leg.Asset)
		delta := leg.SignedAmount()
		newBalance := acct.Available + delta
		if newBalance < 0 && !isSystemAccount(leg.UserID) {
			return fmt.Errorf("insufficient balance: %s:%s need %d have %d",
				leg.UserID, leg.Asset, leg.Amount, acct.Available)
		}
		acct.Available = newBalance
	}

	r.journals = append(r.journals, journal)
	if journal.Type == model.JournalTrade {
		r.processed[journal.Reference] = true
	}
	return nil
}

func (r *LedgerRepository) getOrCreate(userID, asset string) *model.Account {
	key := model.AccountKey(userID, asset)
	if r.accounts[key] == nil {
		r.accounts[key] = &model.Account{UserID: userID, Asset: asset}
	}
	return r.accounts[key]
}

func (r *LedgerRepository) GetAccount(userID, asset string) model.Account {
	r.mu.RLock()
	defer r.mu.RUnlock()
	acct := r.getOrCreate(userID, asset)
	return model.Account{
		UserID:    acct.UserID,
		Asset:     acct.Asset,
		Available: acct.Available,
		Locked:    acct.Locked,
	}
}

func (r *LedgerRepository) GetUserAccounts(userID string) []model.Account {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]model.Account, 0)
	for _, acct := range r.accounts {
		if acct.UserID == userID {
			out = append(out, *acct)
		}
	}
	return out
}

func (r *LedgerRepository) GetJournalByReference(ref string) (model.Journal, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, j := range r.journals {
		if j.Reference == ref {
			return j, true
		}
	}
	return model.Journal{}, false
}

func isSystemAccount(userID string) bool {
	return len(userID) > 9 && userID[:9] == "exchange:"
}

// Reserve moves funds from available to locked (withdrawal hold).
func (r *LedgerRepository) Reserve(userID, asset string, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	acct := r.getOrCreate(userID, asset)
	if acct.Available < amount {
		return fmt.Errorf("insufficient available: have %d need %d", acct.Available, amount)
	}
	acct.Available -= amount
	acct.Locked += amount
	return nil
}

// Release moves funds from locked back to available (cancelled withdrawal).
func (r *LedgerRepository) Release(userID, asset string, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	acct := r.getOrCreate(userID, asset)
	if acct.Locked < amount {
		return fmt.Errorf("insufficient locked: have %d need %d", acct.Locked, amount)
	}
	acct.Locked -= amount
	acct.Available += amount
	return nil
}

// DebitLocked removes locked funds after on-chain withdrawal completes.
func (r *LedgerRepository) DebitLocked(userID, asset string, amount int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	acct := r.getOrCreate(userID, asset)
	if acct.Locked < amount {
		return fmt.Errorf("insufficient locked: have %d need %d", acct.Locked, amount)
	}
	acct.Locked -= amount
	return nil
}

func (r *LedgerRepository) ListJournals() []model.Journal {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.Journal(nil), r.journals...)
}
