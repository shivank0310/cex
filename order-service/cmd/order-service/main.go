package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/order-service/internal/auth"
	"github.com/shivank0310/cex.git/order-service/internal/client"
	"github.com/shivank0310/cex.git/order-service/internal/config"
	"github.com/shivank0310/cex.git/order-service/internal/handler"
	"github.com/shivank0310/cex.git/order-service/internal/middleware"
	"github.com/shivank0310/cex.git/order-service/internal/model"
	"github.com/shivank0310/cex.git/order-service/internal/repository"
	"github.com/shivank0310/cex.git/order-service/internal/service"
	"github.com/shivank0310/cex.git/order-service/internal/validator"
	"github.com/shivank0310/cex.git/pkg/health"
	"github.com/shivank0310/cex.git/pkg/kafka"
	"github.com/shivank0310/cex.git/pkg/redis"
)

func main() {
	cfg := config.Default()

	ledger := meapi.NewLedger()
	eng := newEngine(ledger)

	registry := model.NewRegistry()
	for _, sym := range model.DefaultSymbols() {
		registry.Register(sym)
		eng.RegisterSymbol(sym.Name, meapi.SymbolConfig{
			BaseAsset:  sym.BaseAsset,
			QuoteAsset: sym.QuoteAsset,
		})
	}

	seedDemoBalances(ledger)

	pipeline := validator.NewPipeline(registry, client.NewLedgerWallet(ledger))
	orderSvc := newOrderService(pipeline, eng)

	authenticator := newAuthenticator()
	orderHandler := handler.NewOrderHandler(orderSvc)

	mux := http.NewServeMux()
	health.Register(mux)
	orderHandler.Register(mux)

	var root http.Handler = mux
	root = middleware.AuthMiddleware(authenticator)(root)
	root = rootWithRateLimit(root)
	root = middleware.RecoverMiddleware(root)

	log.Printf("order-service listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, root))
}

func newEngine(ledger *meapi.Ledger) *meapi.Engine {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		log.Println("KAFKA_BROKERS not set — events disabled")
		return meapi.NewEngine(ledger, meapi.DefaultFeeConfig())
	}

	kcfg := kafka.Config{Brokers: strings.Split(brokers, ",")}
	producer := kafka.NewProducer(kcfg)
	log.Printf("Kafka enabled — publishing to brokers: %v", kcfg.Brokers)
	return meapi.NewEngineWithPublisher(ledger, meapi.DefaultFeeConfig(), producer)
}

func newOrderService(pipeline *validator.Pipeline, eng *meapi.Engine) *service.OrderService {
	repo := repository.NewInMemoryOrderRepository()
	engineClient := client.NewEngineClient(eng)

	if os.Getenv("REDIS_ADDR") == "" {
		log.Println("REDIS_ADDR not set — Redis features disabled for order-service")
		return service.NewOrderService(pipeline, engineClient, repo)
	}

	rcfg := redis.ConfigFromEnv()
	client, err := redis.NewClient(rcfg)
	if err != nil {
		log.Fatalf("redis connect failed: %v", err)
	}
	log.Printf("Redis enabled for order-service (addr: %s)", rcfg.Addr)

	orderState := redis.NewOrderStateStore(redis.NewCache(client), 10*time.Minute)
	return service.NewOrderServiceWithRedis(pipeline, engineClient, repo, orderState)
}

func newAuthenticator() auth.Authenticator {
	static := auth.NewStaticAuthenticator(map[string]string{
		"token-user-a": "user-a",
		"token-seller": "seller-1",
	})

	if os.Getenv("REDIS_ADDR") == "" {
		return static
	}

	rcfg := redis.ConfigFromEnv()
	client, err := redis.NewClient(rcfg)
	if err != nil {
		log.Printf("redis session cache unavailable: %v", err)
		return static
	}
	sessions := redis.NewSessionStore(redis.NewCache(client), 24*time.Hour)
	return auth.NewRedisAuthenticator(static, sessions)
}

func rootWithRateLimit(next http.Handler) http.Handler {
	if os.Getenv("REDIS_ADDR") == "" {
		return next
	}

	rcfg := redis.ConfigFromEnv()
	client, err := redis.NewClient(rcfg)
	if err != nil {
		log.Printf("redis rate limit unavailable: %v", err)
		return next
	}

	limiter := redis.NewRateLimiter(client, 100, time.Minute)
	log.Println("Redis rate limiting enabled (100 req/min per user)")
	return middleware.RateLimitMiddleware(limiter)(next)
}

func seedDemoBalances(ledger *meapi.Ledger) {
	ledger.Deposit("seller-1", "BTC", 30)
	ledger.Deposit("seller-2", "BTC", 100)
	ledger.Deposit("seller-3", "BTC", 250)

	ledger.Deposit("bidder-1", "USDT", decimal.Notional(101000, 50))
	ledger.Deposit("bidder-2", "USDT", decimal.Notional(100900, 120))
	ledger.Deposit("bidder-3", "USDT", decimal.Notional(100800, 200))

	ledger.Deposit("user-a", "USDT", decimal.Notional(101100, 30))
}
