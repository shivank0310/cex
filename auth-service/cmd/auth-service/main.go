package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/client"
	"github.com/shivank0310/cex.git/auth-service/internal/config"
	"github.com/shivank0310/cex.git/auth-service/internal/handler"
	"github.com/shivank0310/cex.git/auth-service/internal/repository"
	"github.com/shivank0310/cex.git/auth-service/internal/service"
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

	users := repository.NewUserRepository()
	sessions := repository.NewSessionRepository()
	svc := service.NewAuthService(cfg, users, sessions, profiles)
	authHandler := handler.NewAuthHandler(svc)

	mux := http.NewServeMux()
	authHandler.Register(mux)

	log.Printf("auth-service listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}
