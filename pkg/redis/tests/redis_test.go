package tests

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/shivank0310/cex.git/pkg/redis"
)

func setupRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client, err := redis.NewClient(redis.Config{Addr: mr.Addr()})
	if err != nil {
		t.Fatal(err)
	}
	return client, mr
}

func TestCacheGetSet(t *testing.T) {
	client, _ := setupRedis(t)
	cache := redis.NewCache(client)
	ctx := context.Background()

	type payload struct{ Price int64 }
	if err := cache.Set(ctx, "test:key", payload{Price: 101100}, time.Minute); err != nil {
		t.Fatal(err)
	}

	var out payload
	found, err := cache.Get(ctx, "test:key", &out)
	if err != nil || !found || out.Price != 101100 {
		t.Fatalf("cache miss: found=%v out=%+v err=%v", found, out, err)
	}
}

func TestSessionStore(t *testing.T) {
	client, _ := setupRedis(t)
	sessions := redis.NewSessionStore(redis.NewCache(client), time.Hour)
	ctx := context.Background()

	if err := sessions.Set(ctx, "token-abc", "user-a"); err != nil {
		t.Fatal(err)
	}
	userID, found, err := sessions.Get(ctx, "token-abc")
	if err != nil || !found || userID != "user-a" {
		t.Fatalf("session: user=%s found=%v err=%v", userID, found, err)
	}
}

func TestRateLimiter(t *testing.T) {
	client, _ := setupRedis(t)
	limiter := redis.NewRateLimiter(client, 2, time.Minute)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		allowed, err := limiter.Allow(ctx, "user-a", "/orders")
		if err != nil || !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	allowed, err := limiter.Allow(ctx, "user-a", "/orders")
	if err != nil || allowed {
		t.Fatal("third request should be rate limited")
	}
}

func TestDistributedLock(t *testing.T) {
	client, _ := setupRedis(t)
	lock := redis.NewDistributedLock(client)
	ctx := context.Background()

	token, ok, err := lock.Acquire(ctx, "settlement-batch", time.Minute)
	if err != nil || !ok {
		t.Fatalf("acquire failed: ok=%v err=%v", ok, err)
	}

	_, ok2, _ := lock.Acquire(ctx, "settlement-batch", time.Minute)
	if ok2 {
		t.Fatal("second acquire should fail")
	}

	if err := lock.Release(ctx, "settlement-batch", token); err != nil {
		t.Fatal(err)
	}

	_, ok3, err := lock.Acquire(ctx, "settlement-batch", time.Minute)
	if err != nil || !ok3 {
		t.Fatal("acquire after release should succeed")
	}
}

func TestOrderStateStore(t *testing.T) {
	client, _ := setupRedis(t)
	store := redis.NewOrderStateStore(redis.NewCache(client), 5*time.Minute)
	ctx := context.Background()

	type order struct{ ID string; Status string }
	if err := store.Save(ctx, "ORD-1", order{ID: "ORD-1", Status: "FILLED"}); err != nil {
		t.Fatal(err)
	}
	var out order
	found, err := store.Get(ctx, "ORD-1", &out)
	if err != nil || !found || out.Status != "FILLED" {
		t.Fatalf("order state: %+v found=%v err=%v", out, found, err)
	}
}

func TestNormalizeSymbol(t *testing.T) {
	if redis.NormalizeSymbol("BTC/USDT") != "BTC-USDT" {
		t.Fatal("normalize failed")
	}
	if redis.DisplaySymbol("BTC-USDT") != "BTC/USDT" {
		t.Fatal("display failed")
	}
}
