package tests

import (
	"testing"

	"github.com/shivank0310/cex.git/ledger-service/internal/engine"
	"github.com/shivank0310/cex.git/ledger-service/internal/model"
	"github.com/shivank0310/cex.git/pkg/events"
)

func TestBuildTradeJournalWithFees(t *testing.T) {
	journal, err := engine.BuildTradeJournal(events.TradePayload{
		ID: "T-1", Symbol: "BTC/USDT",
		BuyerID: "alice", SellerID: "bob",
		BuyOrderID: "buy-1", SellOrderID: "sell-1",
		MakerOrderID: "sell-1", // seller is maker
		Price: 10000, Quantity: 10,
		MakerFee: 1, TakerFee: 2,
	}, "J-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := journal.Validate(); err != nil {
		t.Fatalf("unbalanced with fees: %v", err)
	}
	if len(journal.Legs) != 6 { // 4 trade + 2 fee legs
		t.Fatalf("expected 6 legs, got %d", len(journal.Legs))
	}
}

func TestDepositJournalBalanced(t *testing.T) {
	journal := engine.BuildDepositJournal("J-1", "alice", "USDT", 1000, "dep-1")
	if err := journal.Validate(); err != nil {
		t.Fatalf("deposit journal unbalanced: %v", err)
	}
	if journal.Legs[0].Type != model.Debit {
		t.Fatal("user leg should be debit")
	}
}
