package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/shivank0310/cex.git/market-data/internal/cache"
	"github.com/shivank0310/cex.git/market-data/internal/client"
	"github.com/shivank0310/cex.git/market-data/internal/config"
	"github.com/shivank0310/cex.git/market-data/internal/handler"
	"github.com/shivank0310/cex.git/market-data/internal/service"
	"github.com/shivank0310/cex.git/market-data/internal/store"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/pkg/health"
	"github.com/shivank0310/cex.git/pkg/kafka"
	"github.com/shivank0310/cex.git/pkg/redis"
)

func main() {
	cfg := config.Default()
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}

	st := store.New()
	svc := newMarketDataService(cfg, st)
	marketHandler := handler.NewMarketHandler(svc)

	mux := http.NewServeMux()
	health.Register(mux)
	marketHandler.Register(mux)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("market-data HTTP listening on %s", cfg.HTTPAddr)
		if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		kcfg := kafka.Config{Brokers: strings.Split(brokers, ",")}
		consumer := kafka.NewConsumer(kcfg, "market-data", []string{
			events.TopicTrades,
			events.TopicOrderBook,
		}, svc.Handle)

		log.Printf("market-data Kafka consumer starting (brokers: %v)", kcfg.Brokers)
		if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
			log.Fatal(err)
		}
	} else {
		log.Println("KAFKA_BROKERS not set — running HTTP API only (no event consumption)")
		<-ctx.Done()
	}
}

func newMarketDataService(cfg config.Config, st *store.Store) *service.MarketDataService {
	var binanceClient *client.BinanceAdapterClient
	if cfg.BinanceAdapterURL != "" && len(cfg.ExternalSymbols) > 0 {
		binanceClient = client.NewBinanceAdapterClient(cfg.BinanceAdapterURL, cfg.ExternalSymbols)
		log.Printf("Binance market fallback enabled for: %v", cfg.ExternalSymbols)
	}

	var marketCache *cache.MarketCache
	if os.Getenv("REDIS_ADDR") != "" {
		rcfg := redis.ConfigFromEnv()
		redisClient, err := redis.NewClientWithRetry(rcfg, 60*time.Second)
		if err != nil {
			log.Fatalf("redis connect failed: %v", err)
		}
		log.Printf("Redis cache enabled (addr: %s)", rcfg.Addr)
		marketCache = cache.NewMarketCache(redisClient, 5*time.Second)
	} else {
		log.Println("REDIS_ADDR not set — market-data cache disabled (in-memory only)")
	}

	if binanceClient != nil {
		return service.NewMarketDataServiceWithBinance(st, marketCache, binanceClient)
	}
	if marketCache != nil {
		return service.NewMarketDataServiceWithCache(st, marketCache)
	}
	return service.NewMarketDataService(st)
}
