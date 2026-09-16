package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shivank0310/cex.git/blockchain-service/internal/client"
	"github.com/shivank0310/cex.git/blockchain-service/internal/config"
	"github.com/shivank0310/cex.git/blockchain-service/internal/evm"
	"github.com/shivank0310/cex.git/blockchain-service/internal/handler"
	"github.com/shivank0310/cex.git/blockchain-service/internal/repository"
	"github.com/shivank0310/cex.git/blockchain-service/internal/service"
)

func main() {
	cfg := config.Default()
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}
	if url := os.Getenv("WALLET_SERVICE_URL"); url != "" {
		cfg.WalletServiceURL = url
	}
	if rpc := os.Getenv("EVM_RPC_URL"); rpc != "" {
		cfg.RPCURL = rpc
	}
	if vault := os.Getenv("VAULT_CONTRACT_ADDRESS"); vault != "" {
		cfg.VaultContractAddress = vault
	}
	if v := os.Getenv("REQUIRED_CONFIRMATIONS"); v != "" {
		if n, err := parseInt(v); err == nil && n > 0 {
			cfg.RequiredConfirmations = n
		}
	}

	repo := repository.NewRepository()
	chain := evm.NewProvider("ethereum", cfg.RPCURL)
	wallet := client.NewHTTPWalletClient(cfg.WalletServiceURL)
	svc := service.NewBlockchainService(cfg, repo, chain, wallet)
	blockchainHandler := handler.NewBlockchainHandler(svc)

	mux := http.NewServeMux()
	blockchainHandler.Register(mux)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go svc.RunDepositMonitor(ctx)
	go svc.RunTransactionTracker(ctx)

	go func() {
		log.Printf("blockchain-service listening on %s (chain=%s wallet=%s vault=%s)",
			cfg.HTTPAddr, chain.Chain(), cfg.WalletServiceURL, cfg.VaultContractAddress)
		if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("blockchain-service shutting down")
	time.Sleep(100 * time.Millisecond)
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, os.ErrInvalid
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
