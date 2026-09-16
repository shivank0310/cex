package tests

import (
	"testing"
	"time"

	"github.com/shivank0310/cex.git/market-data/internal/store"
	"github.com/shivank0310/cex.git/pkg/events"
)

func TestStorePrunes24hWindow(t *testing.T) {
	st := store.New()
	now := time.Now().UTC()

	st.ApplyTrade(events.TradePayload{
		ID: "old", Symbol: "BTC/USDT", Price: 100000, Quantity: 10,
		Timestamp: now.Add(-25 * time.Hour),
	})
	st.ApplyTrade(events.TradePayload{
		ID: "new", Symbol: "BTC/USDT", Price: 101000, Quantity: 20,
		Timestamp: now,
	})

	stats, ok := st.Stats24h("BTC/USDT")
	if !ok {
		t.Fatal("stats not found")
	}
	if stats.TradeCount != 1 || stats.Volume != 20 {
		t.Fatalf("expected only recent trade in window: %+v", stats)
	}
}
