package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shivank0310/cex.git/notification-service/internal/service"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/pkg/kafka"
)

func main() {
	cfg := kafka.DefaultConfig()
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Brokers = []string{brokers}
	}

	producer := kafka.NewProducer(cfg)
	svc := service.NewNotificationService(producer)

	consumer := kafka.NewConsumer(cfg, "notification-service", []string{
		events.TopicOrders,
		events.TopicTrades,
		events.TopicSettlement,
	}, svc.Handle)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("notification-service consumer starting (topics: orders, trades, settlement)")
	if err := consumer.Run(ctx); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}
