package tests

import (
	"context"
	"testing"
	"time"

	"github.com/shivank0310/cex.git/ledger-service/pkg/testserver"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/pkg/kafka"
	"github.com/shivank0310/cex.git/settlement-service/internal/client"
	"github.com/shivank0310/cex.git/settlement-service/internal/engine"
	"github.com/shivank0310/cex.git/settlement-service/internal/repository"
	"github.com/shivank0310/cex.git/settlement-service/internal/service"
)

func setup(t *testing.T) (*service.SettlementService, *testserver.LedgerService) {
	srv, ledgerSvc := testserver.New()
	t.Cleanup(srv.Close)

	repo := repository.NewSettlementRepository()
	ledger := client.NewHTTPLedgerClient(srv.URL)
	bus := kafka.NewMemBus()
	svc := service.NewSettlementService(repo, ledger, bus)
	return svc, ledgerSvc
}

func tradePayload() events.TradePayload {
	return events.TradePayload{
		ID:           "T-1",
		Symbol:       "BTC/USDT",
		BuyOrderID:   "B-1",
		SellOrderID:  "S-1",
		BuyerID:      "alice",
		SellerID:     "bob",
		Price:        10000,
		Quantity:     10,
		MakerOrderID: "S-1",
		TakerOrderID: "B-1",
		MakerFee:     0,
		TakerFee:     0,
		Timestamp:    time.Now().UTC(),
	}
}

func TestSettlementComputesTransfers(t *testing.T) {
	settlement, err := engine.Compute(tradePayload())
	if err != nil {
		t.Fatal(err)
	}
	if settlement.BuyerBaseCredit != 10 {
		t.Fatalf("buyer base credit: expected 10, got %d", settlement.BuyerBaseCredit)
	}
	if settlement.SellerQuoteCredit != 1000 {
		t.Fatalf("seller quote credit: expected 1000, got %d", settlement.SellerQuoteCredit)
	}
	if settlement.Notional != 1000 {
		t.Fatalf("notional: expected 1000, got %d", settlement.Notional)
	}
}

func TestSettlementWithFees(t *testing.T) {
	trade := tradePayload()
	trade.MakerFee = 1
	trade.TakerFee = 2

	settlement, err := engine.Compute(trade)
	if err != nil {
		t.Fatal(err)
	}
	if settlement.BuyerFee != 2 {
		t.Fatalf("buyer fee: expected 2, got %d", settlement.BuyerFee)
	}
	if settlement.SellerFee != 1 {
		t.Fatalf("seller fee: expected 1, got %d", settlement.SellerFee)
	}
	if settlement.BuyerBaseCredit != 8 {
		t.Fatalf("buyer base credit: expected 8, got %d", settlement.BuyerBaseCredit)
	}
	if settlement.SellerQuoteCredit != 999 {
		t.Fatalf("seller quote credit: expected 999, got %d", settlement.SellerQuoteCredit)
	}
}

func TestAliceBobSettlementFlow(t *testing.T) {
	svc, ledgerSvc := setup(t)
	ctx := context.Background()

	_ = ledgerSvc.Deposit("alice", "USDT", 1000, "dep-alice")
	_ = ledgerSvc.Deposit("bob", "BTC", 100, "dep-bob")
	_ = ledgerSvc.Reserve("alice", "USDT", 1000)
	_ = ledgerSvc.Reserve("bob", "BTC", 10)

	env, err := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-1", tradePayload())
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Handle(ctx, env); err != nil {
		t.Fatal(err)
	}

	settlement, ok := svc.GetByTradeID("T-1")
	if !ok {
		t.Fatal("settlement not found")
	}
	if settlement.JournalID == "" {
		t.Fatal("expected journal id")
	}

	aliceBTC := ledgerSvc.GetBalance("alice", "BTC")
	aliceUSDT := ledgerSvc.GetBalance("alice", "USDT")
	bobBTC := ledgerSvc.GetBalance("bob", "BTC")
	bobUSDT := ledgerSvc.GetBalance("bob", "USDT")

	if aliceBTC.Available != 10 {
		t.Fatalf("alice BTC: expected 10, got %d", aliceBTC.Available)
	}
	if aliceUSDT.Available != 0 {
		t.Fatalf("alice USDT: expected 0, got %d", aliceUSDT.Available)
	}
	if bobBTC.Available != 90 {
		t.Fatalf("bob BTC: expected 90, got %d", bobBTC.Available)
	}
	if bobUSDT.Available != 1000 {
		t.Fatalf("bob USDT: expected 1000, got %d", bobUSDT.Available)
	}
}

func TestSettlementPublishesEvent(t *testing.T) {
	srv, ledgerSvc := testserver.New()
	t.Cleanup(srv.Close)

	bus := kafka.NewMemBus()
	var settlementEvents int
	bus.Subscribe(events.TopicSettlement, func(_ context.Context, env events.Envelope) error {
		settlementEvents++
		payload, err := events.DecodePayload[events.SettlementPayload](env)
		if err != nil {
			return err
		}
		if payload.BuyerBaseCredit != 10 {
			t.Errorf("payload buyer base credit: expected 10, got %d", payload.BuyerBaseCredit)
		}
		if payload.SellerQuoteCredit != 1000 {
			t.Errorf("payload seller quote credit: expected 1000, got %d", payload.SellerQuoteCredit)
		}
		return nil
	})

	repo := repository.NewSettlementRepository()
	ledger := client.NewHTTPLedgerClient(srv.URL)
	_ = ledgerSvc.Deposit("alice", "USDT", 1000, "dep-1")
	_ = ledgerSvc.Deposit("bob", "BTC", 100, "dep-2")
	_ = ledgerSvc.Reserve("alice", "USDT", 1000)
	_ = ledgerSvc.Reserve("bob", "BTC", 10)

	svc := service.NewSettlementService(repo, ledger, bus)
	env, _ := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-1", tradePayload())

	if err := svc.Handle(context.Background(), env); err != nil {
		t.Fatal(err)
	}
	if settlementEvents != 1 {
		t.Fatalf("expected 1 settlement event, got %d", settlementEvents)
	}
}

func TestSettlementIdempotency(t *testing.T) {
	svc, ledgerSvc := setup(t)
	ctx := context.Background()

	_ = ledgerSvc.Deposit("alice", "USDT", 1000, "dep-1")
	_ = ledgerSvc.Deposit("bob", "BTC", 100, "dep-2")
	_ = ledgerSvc.Reserve("alice", "USDT", 1000)
	_ = ledgerSvc.Reserve("bob", "BTC", 10)

	env, _ := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-dup", tradePayload())
	env.EventID = "T-dup"

	if err := svc.Handle(ctx, env); err != nil {
		t.Fatal(err)
	}
	if err := svc.Handle(ctx, env); err != nil {
		t.Fatal(err)
	}

	aliceBTC := ledgerSvc.GetBalance("alice", "BTC")
	if aliceBTC.Available != 10 {
		t.Fatalf("expected single settlement, alice BTC=%d", aliceBTC.Available)
	}
}

func TestSettlementInsufficientBalance(t *testing.T) {
	svc, ledgerSvc := setup(t)
	ctx := context.Background()

	_ = ledgerSvc.Deposit("alice", "USDT", 500, "dep-small")
	_ = ledgerSvc.Deposit("bob", "BTC", 100, "dep-bob")

	env, _ := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, "T-fail", tradePayload())
	if err := svc.Handle(ctx, env); err == nil {
		t.Fatal("expected insufficient balance error")
	}
}
