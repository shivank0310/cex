package main

import (
	"log"
	"net/http"
	"os"

	"github.com/shivank0310/cex.git/wallet-service/internal/client"
	"github.com/shivank0310/cex.git/wallet-service/internal/config"
	"github.com/shivank0310/cex.git/wallet-service/internal/handler"
	"github.com/shivank0310/cex.git/wallet-service/internal/repository"
	"github.com/shivank0310/cex.git/wallet-service/internal/service"
)

func main() {
	cfg := config.Default()
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}
	if url := os.Getenv("LEDGER_URL"); url != "" {
		cfg.LedgerURL = url
	}
	if url := os.Getenv("BLOCKCHAIN_URL"); url != "" {
		cfg.BlockchainURL = url
	}

	repo := repository.NewWalletRepository()
	ledger := client.NewHTTPLedgerClient(cfg.LedgerURL)
	blockchain := client.NewHTTPBlockchainClient(cfg.BlockchainURL)

	svc := service.NewWalletService(cfg, repo, ledger, blockchain)
	walletHandler := handler.NewWalletHandler(svc)

	mux := http.NewServeMux()
	walletHandler.Register(mux)

	log.Printf("wallet-service listening on %s (ledger: %s, blockchain: %s)", cfg.HTTPAddr, cfg.LedgerURL, cfg.BlockchainURL)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}
