package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/binance"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/config"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/handler"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/repository"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/service"
)

func main() {
	cfg := config.Default()
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}

	var provider binance.SpotProvider
	if cfg.UseMock {
		log.Println("BINANCE_API_KEY not set — using mock Binance provider")
		provider = binance.NewMockProvider()
	} else {
		log.Printf("Binance Spot API enabled (base: %s)", cfg.BaseURL)
		provider = binance.NewRESTClient(cfg.BaseURL, cfg.APIKey, cfg.APISecret, cfg.RecvWindow)
	}

	repo := repository.NewOrderRepository()
	svc := service.NewAdapterService(provider, repo)
	adapterHandler := handler.NewAdapterHandler(svc)

	mux := http.NewServeMux()
	adapterHandler.Register(mux)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("binance-adapter-service listening on %s", cfg.HTTPAddr)
		if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	log.Println("binance-adapter-service running")
	<-ctx.Done()
}
