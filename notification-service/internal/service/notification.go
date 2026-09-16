package service

import (
	"context"
	"fmt"
	"log"

	"github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/pkg/events"
)

// NotificationService sends user notifications based on platform events.
type NotificationService struct {
	publisher events.Publisher
}

func NewNotificationService(publisher events.Publisher) *NotificationService {
	return &NotificationService{publisher: publisher}
}

func (s *NotificationService) Handle(ctx context.Context, env events.Envelope) error {
	switch env.EventType {
	case events.TypeOrderSubmitted, events.TypeOrderUpdated, events.TypeOrderCancelled:
		order, err := events.DecodePayload[events.OrderPayload](env)
		if err != nil {
			return err
		}
		return s.notify(ctx, order.UserID, "Order Update",
			fmt.Sprintf("Order %s %s %s status=%s", order.Side, order.Symbol, order.ID, order.Status))

	case events.TypeTradeExecuted:
		trade, err := events.DecodePayload[events.TradePayload](env)
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Trade executed: %.2f %s @ %d", qty(trade.Quantity), trade.Symbol, trade.Price)
		if err := s.notify(ctx, trade.BuyerID, "Trade Executed", msg); err != nil {
			return err
		}
		return s.notify(ctx, trade.SellerID, "Trade Executed", msg)

	case events.TypeSettlementDone:
		settlement, err := events.DecodePayload[events.SettlementPayload](env)
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("Settlement %s completed for trade %s", settlement.SettlementID, settlement.TradeID)
		if err := s.notify(ctx, settlement.BuyerID, "Settlement Complete", msg); err != nil {
			return err
		}
		return s.notify(ctx, settlement.SellerID, "Settlement Complete", msg)
	}
	return nil
}

func (s *NotificationService) notify(ctx context.Context, userID, subject, body string) error {
	n := events.NotificationPayload{
		UserID:  userID,
		Channel: "in-app",
		Subject: subject,
		Body:    body,
	}
	if err := api.PublishNotification(ctx, s.publisher, n); err != nil {
		return err
	}
	log.Printf("[notification] → %s: %s", userID, subject)
	return nil
}

func qty(units int64) float64 {
	return float64(units) / float64(decimal.QuantityScale)
}
