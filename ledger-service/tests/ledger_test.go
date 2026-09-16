package tests

import (
	"context"
	"testing"
	"time"

	"github.com/shivank0310/cex.git/ledger-service/internal/engine"
	"github.com/shivank0310/cex.git/ledger-service/internal/repository"
	"github.com/shivank0310/cex.git/ledger-service/internal/service"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/pkg/kafka"
)

func setup() *service.LedgerService {
	repo := repository.NewLedgerRepository()
	eng := engine.NewDoubleEntryEngine(repo)
	return service.NewLedgerService(repo, eng, nil)
}

func TestAliceBobTrade(t *testing.T) {
	svc := setup()

	// Alice: 1000 USDT, Bob: 1 BTC (quantity scale 100 → 100 units)
	if err := svc.Deposit("alice", "USDT", 1000, "dep-alice-usdt"); err != nil {
		t.Fatal(err)
	}
	if err := svc.Deposit("bob", "BTC", 100, "dep-bob-btc"); err != nil {
		t.Fatal(err)
	}

	// Alice buys 0.1 BTC @ 10000 USDT
	env, err := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-1", events.TradePayload{
		ID: "T-1", Symbol: "BTC/USDT",
		BuyerID: "alice", SellerID: "bob",
		Price: 10000, Quantity: 10, // 0.1 BTC
		Timestamp: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Handle(context.Background(), env); err != nil {
		t.Fatal(err)
	}

	aliceBTC := svc.GetBalance("alice", "BTC")
	aliceUSDT := svc.GetBalance("alice", "USDT")
	bobBTC := svc.GetBalance("bob", "BTC")
	bobUSDT := svc.GetBalance("bob", "USDT")

	if aliceBTC.Available != 10 {
		t.Fatalf("alice BTC: expected 10 (0.1), got %d", aliceBTC.Available)
	}
	if aliceUSDT.Available != 0 {
		t.Fatalf("alice USDT: expected 0, got %d", aliceUSDT.Available)
	}
	if bobBTC.Available != 90 {
		t.Fatalf("bob BTC: expected 90 (0.9), got %d", bobBTC.Available)
	}
	if bobUSDT.Available != 1000 {
		t.Fatalf("bob USDT: expected 1000, got %d", bobUSDT.Available)
	}
}

func TestDoubleEntryJournalIsBalanced(t *testing.T) {
	journal, err := engine.BuildTradeJournal(events.TradePayload{
		ID: "T-1", Symbol: "BTC/USDT",
		BuyerID: "alice", SellerID: "bob",
		Price: 10000, Quantity: 10,
	}, "J-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := journal.Validate(); err != nil {
		t.Fatalf("journal not balanced: %v", err)
	}
	if len(journal.Legs) != 4 {
		t.Fatalf("expected 4 legs, got %d", len(journal.Legs))
	}
}

func TestTradeIdempotency(t *testing.T) {
	svc := setup()
	_ = svc.Deposit("alice", "USDT", 1000, "dep-1")
	_ = svc.Deposit("bob", "BTC", 100, "dep-2")

	trade := events.TradePayload{
		ID: "T-dup", Symbol: "BTC/USDT",
		BuyerID: "alice", SellerID: "bob",
		Price: 10000, Quantity: 10,
	}
	env, _ := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-dup", trade)

	if err := svc.Handle(context.Background(), env); err != nil {
		t.Fatal(err)
	}
	if err := svc.Handle(context.Background(), env); err == nil {
		t.Fatal("expected error on duplicate trade")
	}
}

func TestInsufficientBalanceRejected(t *testing.T) {
	svc := setup()
	_ = svc.Deposit("alice", "USDT", 500, "dep-small") // not enough
	_ = svc.Deposit("bob", "BTC", 100, "dep-bob")

	env, _ := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-fail", events.TradePayload{
		ID: "T-fail", Symbol: "BTC/USDT",
		BuyerID: "alice", SellerID: "bob",
		Price: 10000, Quantity: 10,
	})
	if err := svc.Handle(context.Background(), env); err == nil {
		t.Fatal("expected insufficient balance error")
	}
}

func TestPublishesJournalToKafka(t *testing.T) {
	bus := kafka.NewMemBus()
	var journalEvents int
	bus.Subscribe(events.TopicLedger, func(_ context.Context, env events.Envelope) error {
		if env.EventType == events.TypeLedgerJournal {
			journalEvents++
		}
		return nil
	})

	repo := repository.NewLedgerRepository()
	eng := engine.NewDoubleEntryEngine(repo)
	svc := service.NewLedgerService(repo, eng, bus)

	_ = svc.Deposit("alice", "USDT", 1000, "dep")
	_ = svc.Deposit("bob", "BTC", 100, "dep")

	env, _ := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-1", events.TradePayload{
		ID: "T-1", Symbol: "BTC/USDT",
		BuyerID: "alice", SellerID: "bob",
		Price: 10000, Quantity: 10,
	})
	if err := svc.Handle(context.Background(), env); err != nil {
		t.Fatal(err)
	}
	if journalEvents != 1 {
		t.Fatalf("expected 1 journal event, got %d", journalEvents)
	}
}

func TestJournalRetrieval(t *testing.T) {
	svc := setup()
	_ = svc.Deposit("alice", "USDT", 1000, "dep")
	_ = svc.Deposit("bob", "BTC", 100, "dep")

	env, _ := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-99", events.TradePayload{
		ID: "T-99", Symbol: "BTC/USDT",
		BuyerID: "alice", SellerID: "bob",
		Price: 10000, Quantity: 10,
	})
	_ = svc.Handle(context.Background(), env)

	journal, ok := svc.GetJournalByTrade("T-99")
	if !ok {
		t.Fatal("journal not found")
	}
	if journal.Reference != "T-99" || len(journal.Legs) != 4 {
		t.Fatalf("unexpected journal: %+v", journal)
	}
}
