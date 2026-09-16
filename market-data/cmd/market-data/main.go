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
	"github.com/shivank0310/cex.git/market-data/internal/config"
	"github.com/shivank0310/cex.git/market-data/internal/handler"
	"github.com/shivank0310/cex.git/market-data/internal/service"
	"github.com/shivank0310/cex.git/market-data/internal/store"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/pkg/kafka"
	"github.com/shivank0310/cex.git/pkg/redis"
)

func main() {
	cfg := config.Default()
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}

	st := store.New()
	svc := newMarketDataService(st)
	marketHandler := handler.NewMarketHandler(svc)

	mux := http.NewServeMux()
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

func newMarketDataService(st *store.Store) *service.MarketDataService {
	if os.Getenv("REDIS_ADDR") == "" {
		log.Println("REDIS_ADDR not set — market-data cache disabled (in-memory only)")
		return service.NewMarketDataService(st)
	}

	rcfg := redis.ConfigFromEnv()
	client, err := redis.NewClient(rcfg)
	if err != nil {
		log.Fatalf("redis connect failed: %v", err)
	}
	log.Printf("Redis cache enabled (addr: %s)", rcfg.Addr)
	return service.NewMarketDataServiceWithCache(st, cache.NewMarketCache(client, 5*time.Second))
}
