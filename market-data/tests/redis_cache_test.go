package tests

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/shivank0310/cex.git/market-data/internal/cache"
	"github.com/shivank0310/cex.git/market-data/internal/service"
	"github.com/shivank0310/cex.git/market-data/internal/store"
	"github.com/shivank0310/cex.git/pkg/redis"
)

func TestMarketDataRedisCacheHit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client, err := redis.NewClient(redis.Config{Addr: mr.Addr()})
	if err != nil {
		t.Fatal(err)
	}

	st := store.New()
	mc := cache.NewMarketCache(client, time.Minute)
	svc := service.NewMarketDataServiceWithCache(st, mc)

	applyBook(svc)
	applyTrade(svc, "T-1", 101100, 30, time.Now().UTC())

	// Second read should hit Redis cache
	ticker, err := svc.GetTicker(context.Background(), "BTC/USDT")
	if err != nil {
		t.Fatal(err)
	}
	if ticker.LastPrice != 101100 {
		t.Fatalf("last price: %d", ticker.LastPrice)
	}

	// Verify key exists in Redis
	if !mr.Exists("cex:market:ticker:BTC-USDT") {
		t.Fatal("ticker not cached in redis")
	}
}

func TestTickerAliasFormat(t *testing.T) {
	if redis.NormalizeSymbol("BTC/USDT") != "BTC-USDT" {
		t.Fatal("expected BTC-USDT key format")
	}
}
