package repository

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/shivank0310/cex.git/ledger-service/internal/model"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
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

// RecordJournal stores a journal for audit without applying balance changes.
func (r *LedgerRepository) RecordJournal(journal model.Journal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, leg := range journal.Legs {
		if leg.ID == "" {
			journal.Legs[i].ID = r.nextLegID()
		}
		journal.Legs[i].JournalID = journal.ID
	}

	r.journals = append(r.journals, journal)
	if journal.Type == model.JournalTrade {
		r.processed[journal.Reference] = true
	}
	return nil
}

func (r *LedgerRepository) ListJournals() []model.Journal {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.Journal(nil), r.journals...)
}

// SettleTradeLocked settles a trade using locked funds (Binance-style free/locked model).
// Seller base asset is debited from locked; buyer quote asset is debited from locked
// (with price-improvement refund when buyLimitPrice > execution price).
func (r *LedgerRepository) SettleTradeLocked(
	buyerID, sellerID, baseAsset, quoteAsset string,
	price, quantity, buyerFee, sellerFee, buyLimitPrice int64,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	notional := decimal.Notional(price, quantity)

	sellerBase := r.getOrCreate(sellerID, baseAsset)
	if sellerBase.Locked < quantity {
		return fmt.Errorf("seller insufficient locked %s: have %d need %d", baseAsset, sellerBase.Locked, quantity)
	}
	sellerBase.Locked -= quantity

	buyerQuote := r.getOrCreate(buyerID, quoteAsset)
	if buyLimitPrice > 0 {
		lockedSlice := decimal.Notional(buyLimitPrice, quantity)
		if buyerQuote.Locked < lockedSlice {
			return fmt.Errorf("buyer insufficient locked %s: have %d need %d", quoteAsset, buyerQuote.Locked, lockedSlice)
		}
		buyerQuote.Locked -= lockedSlice
		buyerQuote.Available += lockedSlice - notional
	} else {
		if buyerQuote.Locked < notional {
			return fmt.Errorf("buyer insufficient locked %s: have %d need %d", quoteAsset, buyerQuote.Locked, notional)
		}
		buyerQuote.Locked -= notional
	}

	buyerBase := r.getOrCreate(buyerID, baseAsset)
	buyerBase.Available += quantity - buyerFee

	sellerQuote := r.getOrCreate(sellerID, quoteAsset)
	sellerQuote.Available += notional - sellerFee

	if buyerFee > 0 {
		r.getOrCreate("exchange:fees", baseAsset).Available += buyerFee
	}
	if sellerFee > 0 {
		r.getOrCreate("exchange:fees", quoteAsset).Available += sellerFee
	}

	return nil
}
