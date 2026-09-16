package service

import (
	"context"
	"fmt"
	"log"

	"github.com/shivank0310/cex.git/ledger-service/internal/engine"
	"github.com/shivank0310/cex.git/ledger-service/internal/model"
	"github.com/shivank0310/cex.git/ledger-service/internal/repository"
	"github.com/shivank0310/cex.git/pkg/events"
)

// LedgerService consumes trade events and posts double-entry journals.
type LedgerService struct {
	engine    *engine.DoubleEntryEngine
	repo      *repository.LedgerRepository
	publisher events.Publisher
}

func NewLedgerService(repo *repository.LedgerRepository, eng *engine.DoubleEntryEngine, publisher events.Publisher) *LedgerService {
	return &LedgerService{repo: repo, engine: eng, publisher: publisher}
}

func (s *LedgerService) Handle(ctx context.Context, env events.Envelope) error {
	if env.EventType != events.TypeTradeExecuted {
		return nil
	}

	trade, err := events.DecodePayload[events.TradePayload](env)
	if err != nil {
		return err
	}

	_, err = s.SettleTrade(ctx, trade)
	return err
}

// SettleTrade posts a balanced trade journal to the ledger.
func (s *LedgerService) SettleTrade(ctx context.Context, trade events.TradePayload) (model.Journal, error) {
	journalID := s.repo.NextJournalID()
	journal, err := engine.BuildTradeJournal(trade, journalID)
	if err != nil {
		return model.Journal{}, err
	}

	if err := s.engine.PostJournal(journal); err != nil {
		return model.Journal{}, err
	}

	if s.publisher != nil {
		if err := publishJournal(ctx, s.publisher, journal); err != nil {
			return model.Journal{}, err
		}
	}

	log.Printf("[ledger] posted journal %s trade=%s legs=%d", journal.ID, trade.ID, len(journal.Legs))
	return journal, nil
}

func (s *LedgerService) Deposit(userID, asset string, amount int64, ref string) error {
	return s.engine.Deposit(userID, asset, amount, ref)
}

func (s *LedgerService) GetBalance(userID, asset string) model.Account {
	return s.repo.GetAccount(userID, asset)
}

func (s *LedgerService) GetBalances(userID string) []model.Account {
	return s.repo.GetUserAccounts(userID)
}

func (s *LedgerService) GetJournalByTrade(tradeID string) (model.Journal, bool) {
	return s.repo.GetJournalByReference(tradeID)
}

func (s *LedgerService) Reserve(userID, asset string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.repo.Reserve(userID, asset, amount)
}

func (s *LedgerService) Release(userID, asset string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.repo.Release(userID, asset, amount)
}

func (s *LedgerService) DebitLocked(userID, asset string, amount int64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.repo.DebitLocked(userID, asset, amount)
}

func publishJournal(ctx context.Context, pub events.Publisher, journal model.Journal) error {
	payload := events.LedgerJournalPayload{
		JournalID: journal.ID,
		Type:      string(journal.Type),
		Reference: journal.Reference,
		Symbol:    journal.Symbol,
		Legs:      make([]events.LedgerLegPayload, len(journal.Legs)),
		PostedAt:  journal.PostedAt,
	}
	for i, leg := range journal.Legs {
		payload.Legs[i] = events.LedgerLegPayload{
			LegID:     leg.ID,
			UserID:    leg.UserID,
			Asset:     leg.Asset,
			Type:      string(leg.Type),
			Amount:    leg.Amount,
			Reference: leg.Reference,
		}
	}
	env, err := events.NewEnvelope(events.TopicLedger, events.TypeLedgerJournal, journal.ID, payload)
	if err != nil {
		return err
	}
	return pub.Publish(ctx, events.TopicLedger, journal.Reference, env)
}
