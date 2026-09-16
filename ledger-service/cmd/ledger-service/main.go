package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/shivank0310/cex.git/ledger-service/internal/config"
	"github.com/shivank0310/cex.git/ledger-service/internal/engine"
	"github.com/shivank0310/cex.git/ledger-service/internal/handler"
	"github.com/shivank0310/cex.git/ledger-service/internal/repository"
	"github.com/shivank0310/cex.git/ledger-service/internal/service"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/pkg/kafka"
)

func main() {
	cfg := config.Default()
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		cfg.HTTPAddr = addr
	}

	repo := repository.NewLedgerRepository()
	eng := engine.NewDoubleEntryEngine(repo)

	var publisher events.Publisher
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		kcfg := kafka.Config{Brokers: strings.Split(brokers, ",")}
		producer := kafka.NewProducer(kcfg)
		publisher = producer
		log.Printf("Kafka publisher enabled (brokers: %v)", kcfg.Brokers)
	}

	svc := service.NewLedgerService(repo, eng, publisher)
	ledgerHandler := handler.NewLedgerHandler(svc)

	mux := http.NewServeMux()
	ledgerHandler.Register(mux)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("ledger-service HTTP listening on %s", cfg.HTTPAddr)
		if err := http.ListenAndServe(cfg.HTTPAddr, mux); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	log.Println("ledger-service running (HTTP API; trade settlement via settlement-service)")
	<-ctx.Done()
}
