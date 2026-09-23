package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/client"
	"github.com/shivank0310/cex.git/auth-service/internal/config"
	"github.com/shivank0310/cex.git/auth-service/internal/database"
	"github.com/shivank0310/cex.git/auth-service/internal/handler"
	"github.com/shivank0310/cex.git/auth-service/internal/repository"
	"github.com/shivank0310/cex.git/auth-service/internal/service"
	cexredis "github.com/shivank0310/cex.git/pkg/redis"
)

func main() {
	cfg := config.Default()
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWTSecret = secret
	}
	if issuer := os.Getenv("JWT_ISSUER"); issuer != "" {
		cfg.JWTIssuer = issuer
	}
	if ttl := os.Getenv("ACCESS_TOKEN_TTL"); ttl != "" {
		if d, err := time.ParseDuration(ttl); err == nil {
			cfg.AccessTokenTTL = d
		}
	}
	if ttl := os.Getenv("REFRESH_TOKEN_TTL"); ttl != "" {
		if d, err := time.ParseDuration(ttl); err == nil {
			cfg.RefreshTokenTTL = d
		}
	}
	cfg.RedisAddr = os.Getenv("REDIS_ADDR")
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if url := os.Getenv("USER_SERVICE_URL"); url != "" {
		cfg.UserServiceURL = url
	}

	var profiles client.UserClient
	if cfg.UserServiceURL != "" {
		profiles = client.NewHTTPUserClient(cfg.UserServiceURL)
		log.Printf("user-service enabled at %s", cfg.UserServiceURL)
	} else {
		profiles = client.NewInMemoryUserClient()
		log.Println("USER_SERVICE_URL not set — using in-memory user profiles (dev mode)")
	}

	var credentials repository.CredentialStore
	if cfg.DatabaseURL != "" {
		db, err := database.Open(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("database connect failed: %v", err)
		}
		defer db.Close()
		credentials = repository.NewPostgresCredentialStore(db)
		log.Println("PostgreSQL credential store enabled")
	} else {
		credentials = repository.NewMemoryCredentialStore()
		log.Println("DATABASE_URL not set — credentials in-memory only (lost on restart)")
	}

	var sessions repository.SessionStore
	if cfg.RedisAddr != "" {
		rcfg := cexredis.ConfigFromEnv()
		redisClient, err := cexredis.NewClientWithRetry(rcfg, 60*time.Second)
		if err != nil {
			log.Fatalf("redis connect failed: %v", err)
		}
		defer redisClient.Close()
		sessions = repository.NewRedisSessionStore(cexredis.NewCache(redisClient))
		log.Printf("Redis session store enabled (addr: %s)", rcfg.Addr)
	} else {
		sessions = repository.NewMemorySessionStore()
		log.Println("REDIS_ADDR not set — refresh sessions in-memory only (lost on restart)")
	}

	svc := service.NewAuthService(cfg, credentials, sessions, profiles)
	authHandler := handler.NewAuthHandler(svc)

	mux := http.NewServeMux()
	authHandler.Register(mux)

	log.Printf("auth-service listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}
