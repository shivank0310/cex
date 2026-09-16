package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/pkg/kafka"
	"github.com/shivank0310/cex.git/settlement-service/internal/client"
	"github.com/shivank0310/cex.git/settlement-service/internal/config"
	"github.com/shivank0310/cex.git/settlement-service/internal/repository"
	"github.com/shivank0310/cex.git/settlement-service/internal/service"
)

func main() {
	cfg := config.Default()
	if url := os.Getenv("LEDGER_URL"); url != "" {
		cfg.LedgerURL = url
	}

	kcfg := kafka.DefaultConfig()
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		kcfg.Brokers = []string{brokers}
	}

	repo := repository.NewSettlementRepository()
	ledger := client.NewHTTPLedgerClient(cfg.LedgerURL)
	producer := kafka.NewProducer(kcfg)
	svc := service.NewSettlementService(repo, ledger, producer)

	consumer := kafka.NewConsumer(kcfg, "settlement-service", []string{
		events.TopicTrades,
	}, svc.Handle)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("settlement-service starting (trades → ledger @ %s → settlement)", cfg.LedgerURL)
	if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}
