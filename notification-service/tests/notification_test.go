package tests

import (
	"context"
	"testing"

	"github.com/shivank0310/cex.git/notification-service/internal/service"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/pkg/kafka"
)

func TestNotificationServiceHandlesTrade(t *testing.T) {
	bus := kafka.NewMemBus()
	var notifications int
	bus.Subscribe(events.TopicNotifications, func(_ context.Context, _ events.Envelope) error {
		notifications++
		return nil
	})

	svc := service.NewNotificationService(bus)
	env, err := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-1", events.TradePayload{
		ID: "T-1", Symbol: "BTC/USDT", BuyerID: "buyer", SellerID: "seller",
		Price: 101100, Quantity: 30,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := svc.Handle(context.Background(), env); err != nil {
		t.Fatal(err)
	}
	if notifications != 2 {
		t.Fatalf("expected 2 notifications (buyer+seller), got %d", notifications)
	}
}
