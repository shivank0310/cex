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
	jwtpkg "github.com/shivank0310/cex.git/pkg/jwt"
	"github.com/shivank0310/cex.git/pkg/kafka"
	"github.com/shivank0310/cex.git/pkg/redis"
)

func main() {
	cfg := config.Default()

	var ledger *meapi.Ledger
	var ledgerClient client.LedgerClient
	if cfg.LedgerURL != "" {
		log.Printf("ledger-service enabled at %s (external fund holds)", cfg.LedgerURL)
		ledger = meapi.NewPermissiveLedger()
		ledgerClient = client.NewHTTPLedgerClient(cfg.LedgerURL)
	} else {
		log.Println("LEDGER_URL not set — using in-memory ledger (demo mode)")
		ledger = meapi.NewLedger()
	}

	eng := newEngine(ledger)

	registry := model.NewRegistry()
	symbols := model.DefaultSymbols()
	if cfg.RouteBinance {
		symbols = model.BinanceSymbols()
		log.Println("ROUTE_BINANCE=1 — routing orders to Binance adapter")
	}
	for _, sym := range symbols {
		registry.Register(sym)
		if sym.Venue == model.VenueInternal {
			eng.RegisterSymbol(sym.Name, meapi.SymbolConfig{
				BaseAsset:  sym.BaseAsset,
				QuoteAsset: sym.QuoteAsset,
			})
		}
	}

	if ledgerClient == nil {
		seedDemoBalances(ledger)
	}

	var walletClient client.WalletClient
	if ledgerClient != nil {
		walletClient = client.NewHTTPLedgerWallet(ledgerClient)
	} else {
		walletClient = client.NewLedgerWallet(ledger)
	}

	pipeline := validator.NewPipeline(registry, walletClient)
	funds := service.NewFundsHoldManager(ledgerClient)
	orderSvc := newOrderService(cfg, pipeline, registry, eng, funds)

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

func newOrderService(cfg config.Config, pipeline *validator.Pipeline, registry *model.Registry, eng *meapi.Engine, funds *service.FundsHoldManager) *service.OrderService {
	repo := repository.NewInMemoryOrderRepository()
	internalClient := client.NewEngineClient(eng)

	var engineClient client.MatchingEngineClient = internalClient
	if cfg.BinanceURL != "" {
		binanceClient := client.NewBinanceEngineClient(cfg.BinanceURL)
		engineClient = client.NewVenueRouter(registry, internalClient, binanceClient)
		log.Printf("Binance adapter enabled at %s", cfg.BinanceURL)
	}

	if os.Getenv("REDIS_ADDR") == "" {
		log.Println("REDIS_ADDR not set — Redis features disabled for order-service")
		return service.NewOrderService(pipeline, engineClient, repo, funds)
	}

	rcfg := redis.ConfigFromEnv()
	client, err := redis.NewClient(rcfg)
	if err != nil {
		log.Fatalf("redis connect failed: %v", err)
	}
	log.Printf("Redis enabled for order-service (addr: %s)", rcfg.Addr)

	orderState := redis.NewOrderStateStore(redis.NewCache(client), 10*time.Minute)
	return service.NewOrderServiceWithRedis(pipeline, engineClient, repo, funds, orderState)
}

func newAuthenticator() auth.Authenticator {
	var inner auth.Authenticator

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		issuer := os.Getenv("JWT_ISSUER")
		if issuer == "" {
			issuer = "cex-auth"
		}
		mgr := jwtpkg.NewManager(secret, issuer, 15*time.Minute)
		inner = auth.NewJWTAuthenticator(mgr)
		log.Println("JWT authentication enabled (auth-service tokens)")
	} else {
		inner = auth.NewStaticAuthenticator(map[string]string{
			"token-user-a": "user-a",
			"token-seller": "seller-1",
		})
		log.Println("static token authentication enabled (dev mode)")
	}

	if os.Getenv("REDIS_ADDR") == "" {
		return inner
	}

	rcfg := redis.ConfigFromEnv()
	client, err := redis.NewClient(rcfg)
	if err != nil {
		log.Printf("redis session cache unavailable: %v", err)
		return inner
	}
	sessions := redis.NewSessionStore(redis.NewCache(client), 24*time.Hour)
	return auth.NewRedisAuthenticator(inner, sessions)
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
