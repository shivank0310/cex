package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/shivank0310/cex.git/api-gateway/internal/config"
	"github.com/shivank0310/cex.git/api-gateway/internal/middleware"
	"github.com/shivank0310/cex.git/api-gateway/internal/proxy"
	"github.com/shivank0310/cex.git/pkg/redis"
)

func main() {
	cfg := loadConfig()

	gateway, err := proxy.NewGateway(cfg)
	if err != nil {
		log.Fatalf("gateway init failed: %v", err)
	}

	var handler http.Handler = gateway
	handler = middleware.RateLimit(newLimiter(cfg))(handler)
	handler = middleware.CORS(cfg.CORSAllowedOrigins)(handler)
	handler = middleware.Logging(handler)
	handler = middleware.RequestID(handler)

	addr := cfg.HTTPAddr
	if cfg.TLSCert != "" && cfg.TLSKey != "" {
		log.Printf("api-gateway listening on %s (TLS enabled)", addr)
		log.Fatal(http.ListenAndServeTLS(addr, cfg.TLSCert, cfg.TLSKey, handler))
	}

	log.Printf("api-gateway listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

func loadConfig() config.Config {
	cfg := config.Default()
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		cfg.HTTPAddr = v
	}
	cfg.TLSCert = os.Getenv("TLS_CERT")
	cfg.TLSKey = os.Getenv("TLS_KEY")

	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		cfg.CORSAllowedOrigins = splitCSV(v)
	}
	if v := os.Getenv("RATE_LIMIT_PER_MINUTE"); v != "" {
		if n, err := parseInt64(v); err == nil {
			cfg.RateLimitPerMinute = n
		}
	}

	cfg.AuthServiceURL = envOr(cfg.AuthServiceURL, "AUTH_SERVICE_URL")
	cfg.UserServiceURL = envOr(cfg.UserServiceURL, "USER_SERVICE_URL")
	cfg.OrderServiceURL = envOr(cfg.OrderServiceURL, "ORDER_SERVICE_URL")
	cfg.MarketDataURL = envOr(cfg.MarketDataURL, "MARKET_DATA_URL")
	cfg.LedgerServiceURL = envOr(cfg.LedgerServiceURL, "LEDGER_SERVICE_URL")
	cfg.WalletServiceURL = envOr(cfg.WalletServiceURL, "WALLET_SERVICE_URL")
	cfg.BlockchainServiceURL = envOr(cfg.BlockchainServiceURL, "BLOCKCHAIN_SERVICE_URL")
	cfg.AdminServiceURL = envOr(cfg.AdminServiceURL, "ADMIN_SERVICE_URL")
	cfg.MatchingServiceURL = envOr(cfg.MatchingServiceURL, "MATCHING_SERVICE_URL")
	return cfg
}

func newLimiter(cfg config.Config) middleware.Limiter {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		client, err := redis.NewClient(redis.ConfigFromEnv())
		if err != nil {
			log.Printf("redis rate limit unavailable: %v — using in-memory limiter", err)
		} else {
			log.Printf("redis rate limiting enabled (%d req/min)", cfg.RateLimitPerMinute)
			return middleware.NewRedisLimiter(client, cfg.RateLimitPerMinute, time.Minute)
		}
	}
	log.Printf("in-memory rate limiting enabled (%d req/min)", cfg.RateLimitPerMinute)
	return middleware.NewMemoryLimiter(cfg.RateLimitPerMinute, time.Minute)
}

func envOr(defaultValue, key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseInt64(value string) (int64, error) {
	var n int64
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return 0, os.ErrInvalid
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}
