package tests

import (
	"context"
	"sync"
	"testing"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/pkg/kafka"
)

func TestEnginePublishesKafkaEvents(t *testing.T) {
	bus := kafka.NewMemBus()
	var mu sync.Mutex
	var orderEvents, tradeEvents, bookEvents int

	bus.Subscribe(events.TopicOrders, func(_ context.Context, env events.Envelope) error {
		mu.Lock()
		orderEvents++
		mu.Unlock()
		return nil
	})
	bus.Subscribe(events.TopicTrades, func(_ context.Context, env events.Envelope) error {
		mu.Lock()
		tradeEvents++
		mu.Unlock()
		return nil
	})
	bus.Subscribe(events.TopicOrderBook, func(_ context.Context, env events.Envelope) error {
		mu.Lock()
		bookEvents++
		mu.Unlock()
		return nil
	})

	ledger := meapi.NewLedger()
	eng := meapi.NewEngineWithPublisher(ledger, meapi.FeeConfig{MakerBasisPoints: 0, TakerBasisPoints: 0}, bus)
	eng.RegisterSymbol(symbol, meapi.SymbolConfig{BaseAsset: baseAsset, QuoteAsset: quoteAsset})

	ledger.Deposit("seller-1", baseAsset, 30)
	ledger.Deposit("user-a", quoteAsset, decimal.Notional(101100, 30))

	eng.SubmitOrder(meapi.NewLimitOrder("ask1", "seller-1", symbol, meapi.Sell, 101100, 30))
	res := eng.SubmitOrder(meapi.NewLimitOrder("buy-a", "user-a", symbol, meapi.Buy, 101100, 30))
	if res.Error != nil {
		t.Fatalf("submit: %v", res.Error)
	}

	mu.Lock()
	defer mu.Unlock()
	if orderEvents < 2 {
		t.Fatalf("expected >=2 order events, got %d", orderEvents)
	}
	if tradeEvents != 1 {
		t.Fatalf("expected 1 trade event, got %d", tradeEvents)
	}
	if bookEvents < 2 {
		t.Fatalf("expected >=2 orderbook events, got %d", bookEvents)
	}
}

func TestEnvelopeRoundTrip(t *testing.T) {
	payload := events.TradePayload{ID: "T-1", Symbol: "BTC/USDT", Price: 101100, Quantity: 30}
	env, err := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-1", payload)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := events.DecodePayload[events.TradePayload](env)
	if err != nil || decoded.ID != "T-1" {
		t.Fatalf("decode failed: %v", err)
	}
}
