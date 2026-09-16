package service

import (
	"context"
	"fmt"
	"log"

	"github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/settlement-service/internal/client"
	"github.com/shivank0310/cex.git/settlement-service/internal/engine"
	"github.com/shivank0310/cex.git/settlement-service/internal/model"
	"github.com/shivank0310/cex.git/settlement-service/internal/repository"
)

// SettlementService finalizes trades after matching.
type SettlementService struct {
	repo      *repository.SettlementRepository
	ledger    client.LedgerClient
	publisher events.Publisher
}

func NewSettlementService(
	repo *repository.SettlementRepository,
	ledger client.LedgerClient,
	publisher events.Publisher,
) *SettlementService {
	return &SettlementService{
		repo:      repo,
		ledger:    ledger,
		publisher: publisher,
	}
}

func (s *SettlementService) Handle(ctx context.Context, env events.Envelope) error {
	if env.EventType != events.TypeTradeExecuted {
		return nil
	}

	trade, err := events.DecodePayload[events.TradePayload](env)
	if err != nil {
		return err
	}

	if s.repo.ExistsByTradeID(trade.ID) {
		log.Printf("[settlement] trade %s already settled, skipping", trade.ID)
		return nil
	}

	settlement, err := engine.Compute(trade)
	if err != nil {
		return err
	}

	journalID, err := s.ledger.SettleTrade(ctx, trade)
	if err != nil {
		settlement.Status = model.StatusFailed
		return fmt.Errorf("ledger settlement failed for trade %s: %w", trade.ID, err)
	}

	settlement.ID = s.repo.NextID()
	settlement.JournalID = journalID
	s.repo.Save(settlement)

	if s.publisher != nil {
		if err := api.PublishSettlement(ctx, s.publisher, toPayload(settlement)); err != nil {
			return err
		}
	}

	log.Printf("[settlement] %s trade=%s buyer=%s +%d %s seller=%s +%d %s fees=%d/%d",
		settlement.ID, trade.ID,
		settlement.BuyerID, settlement.BuyerBaseCredit, settlement.BaseAsset,
		settlement.SellerID, settlement.SellerQuoteCredit, settlement.QuoteAsset,
		settlement.BuyerFee, settlement.SellerFee)
	return nil
}

func (s *SettlementService) GetByTradeID(tradeID string) (model.Settlement, bool) {
	return s.repo.GetByTradeID(tradeID)
}

func toPayload(s model.Settlement) events.SettlementPayload {
	return events.SettlementPayload{
		SettlementID:      s.ID,
		TradeID:           s.TradeID,
		JournalID:         s.JournalID,
		Symbol:            s.Symbol,
		BaseAsset:         s.BaseAsset,
		QuoteAsset:        s.QuoteAsset,
		BuyerID:           s.BuyerID,
		SellerID:          s.SellerID,
		Price:             s.Price,
		Quantity:          s.Quantity,
		Notional:          s.Notional,
		BuyerFee:          s.BuyerFee,
		SellerFee:         s.SellerFee,
		BuyerBaseCredit:   s.BuyerBaseCredit,
		SellerQuoteCredit: s.SellerQuoteCredit,
		Status:            string(s.Status),
		Timestamp:         s.SettledAt,
	}
}
