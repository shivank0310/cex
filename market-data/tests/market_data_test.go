package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shivank0310/cex.git/market-data/internal/dto"
	"github.com/shivank0310/cex.git/market-data/internal/handler"
	"github.com/shivank0310/cex.git/market-data/internal/service"
	"github.com/shivank0310/cex.git/market-data/internal/store"
	"github.com/shivank0310/cex.git/pkg/events"
)

func setup() *service.MarketDataService {
	st := store.New()
	return service.NewMarketDataService(st)
}

func applyTrade(svc *service.MarketDataService, id string, price, qty int64, ts time.Time) {
	env, _ := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, id, events.TradePayload{
		ID: id, Symbol: "BTC/USDT", Price: price, Quantity: qty, Timestamp: ts,
	})
	_ = svc.Handle(context.Background(), env)
}

func applyBook(svc *service.MarketDataService) {
	env, _ := events.NewEnvelope(events.TopicOrderBook, events.TypeOrderBookUpdated, "BTC/USDT", events.OrderBookPayload{
		Symbol: "BTC/USDT",
		Bids:   []events.DepthLevel{{Price: 101000, Quantity: 50}},
		Asks:   []events.DepthLevel{{Price: 101200, Quantity: 30}},
	})
	_ = svc.Handle(context.Background(), env)
}

func TestTickerWith24hStats(t *testing.T) {
	svc := setup()
	now := time.Now().UTC()

	applyBook(svc)
	applyTrade(svc, "T-1", 101100, 30, now.Add(-2*time.Hour))
	applyTrade(svc, "T-2", 101500, 20, now.Add(-1*time.Hour))
	applyTrade(svc, "T-3", 100800, 40, now)

	ticker, err := svc.GetTicker(context.Background(), "BTC/USDT")
	if err != nil {
		t.Fatal(err)
	}
	if ticker.LastPrice != 100800 {
		t.Fatalf("last price: %d", ticker.LastPrice)
	}
	if ticker.High24h != 101500 {
		t.Fatalf("high24h: %d", ticker.High24h)
	}
	if ticker.Low24h != 100800 {
		t.Fatalf("low24h: %d", ticker.Low24h)
	}
	if ticker.Volume24h != 90 {
		t.Fatalf("volume24h: %d", ticker.Volume24h)
	}
	if ticker.BestBid != 101000 || ticker.BestAsk != 101200 {
		t.Fatalf("bid/ask: %d/%d", ticker.BestBid, ticker.BestAsk)
	}
}

func TestCandles(t *testing.T) {
	svc := setup()
	now := time.Now().UTC().Truncate(time.Minute)

	applyTrade(svc, "T-1", 101000, 10, now)
	applyTrade(svc, "T-2", 101200, 20, now.Add(10*time.Second))
	applyTrade(svc, "T-3", 100900, 15, now.Add(20*time.Second))

	candles, err := svc.GetCandles("BTC/USDT", "1m", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(candles) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(candles))
	}
	c := candles[0]
	if c.Open != 101000 || c.Close != 100900 || c.High != 101200 || c.Low != 100900 {
		t.Fatalf("ohlc wrong: %+v", c)
	}
	if c.Volume != 45 {
		t.Fatalf("volume: %d", c.Volume)
	}
}

func TestOrderBookAndTrades(t *testing.T) {
	svc := setup()
	applyBook(svc)
	applyTrade(svc, "T-1", 101100, 30, time.Now().UTC())

	book, err := svc.GetOrderBook(context.Background(), "BTC/USDT")
	if err != nil {
		t.Fatal(err)
	}
	if len(book.Bids) != 1 || len(book.Asks) != 1 {
		t.Fatalf("book depth wrong: %+v", book)
	}

	trades := svc.GetTrades(context.Background(), "BTC/USDT", 10)
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade")
	}
}

func TestStats24h(t *testing.T) {
	svc := setup()
	now := time.Now().UTC()
	applyTrade(svc, "T-1", 101000, 10, now)
	applyTrade(svc, "T-2", 102000, 20, now)

	stats, err := svc.GetStats24h(context.Background(), "BTC/USDT")
	if err != nil {
		t.Fatal(err)
	}
	if stats.High != 102000 || stats.Low != 101000 || stats.Volume != 30 {
		t.Fatalf("stats wrong: %+v", stats)
	}
}

func TestHTTPTickerEndpoint(t *testing.T) {
	svc := setup()
	applyBook(svc)
	applyTrade(svc, "T-1", 101100, 30, time.Now().UTC())

	h := handler.NewMarketHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/market/ticker/BTC/USDT", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}

	var resp dto.TickerResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Symbol != "BTC/USDT" || resp.LastPrice != 101100 {
		t.Fatalf("unexpected ticker: %+v", resp)
	}
}

func TestHTTPCandlesEndpoint(t *testing.T) {
	svc := setup()
	applyTrade(svc, "T-1", 101100, 30, time.Now().UTC())

	h := handler.NewMarketHandler(svc)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/market/candles/BTC/USDT?interval=1m&limit=10", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
