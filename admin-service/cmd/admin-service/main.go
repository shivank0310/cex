package main

import (
	"log"
	"net/http"
	"os"

	"github.com/shivank0310/cex.git/admin-service/internal/client"
	"github.com/shivank0310/cex.git/admin-service/internal/config"
	"github.com/shivank0310/cex.git/admin-service/internal/handler"
	"github.com/shivank0310/cex.git/admin-service/internal/middleware"
	"github.com/shivank0310/cex.git/admin-service/internal/repository"
	"github.com/shivank0310/cex.git/admin-service/internal/service"
)

func main() {
	cfg := config.Default()
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}
	if key := os.Getenv("ADMIN_API_KEY"); key != "" {
		cfg.AdminAPIKey = key
	}
	if url := os.Getenv("ORDER_SERVICE_URL"); url != "" {
		cfg.OrderServiceURL = url
	}
	if url := os.Getenv("MARKET_DATA_URL"); url != "" {
		cfg.MarketDataURL = url
	}
	if url := os.Getenv("LEDGER_SERVICE_URL"); url != "" {
		cfg.LedgerServiceURL = url
	}
	if url := os.Getenv("WALLET_SERVICE_URL"); url != "" {
		cfg.WalletServiceURL = url
	}
	if url := os.Getenv("BLOCKCHAIN_SERVICE_URL"); url != "" {
		cfg.BlockchainServiceURL = url
	}

	repo := repository.NewRepository()
	health := client.NewHTTPHealthChecker(cfg)
	svc := service.NewAdminService(repo, health)
	adminHandler := handler.NewAdminHandler(svc)

	mux := http.NewServeMux()
	adminHandler.Register(mux)

	var root http.Handler = mux
	root = middleware.AdminAuth(cfg.AdminAPIKey, root)

	log.Printf("admin-service listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, root))
}
