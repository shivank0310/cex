package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/shivank0310/cex.git/pkg/events"
)

// Handler processes a single event envelope.
type Handler func(ctx context.Context, env events.Envelope) error

// Consumer reads from one or more Kafka topics.
type Consumer struct {
	cfg     Config
	groupID string
	topics  []string
	handler Handler
}

func NewConsumer(cfg Config, groupID string, topics []string, handler Handler) *Consumer {
	return &Consumer{
		cfg:     cfg,
		groupID: groupID,
		topics:  topics,
		handler: handler,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.cfg.Brokers,
		GroupID:  c.groupID,
		GroupTopics: c.topics,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	defer r.Close()

	for {
		msg, err := r.FetchMessage(ctx)
		if err != nil {
			return fmt.Errorf("fetch message: %w", err)
		}

		var env events.Envelope
		if err := json.Unmarshal(msg.Value, &env); err != nil {
			log.Printf("kafka: skip invalid message on %s: %v", msg.Topic, err)
			_ = r.CommitMessages(ctx, msg)
			continue
		}

		if err := c.handler(ctx, env); err != nil {
			log.Printf("kafka: handler error topic=%s type=%s: %v", env.Topic, env.EventType, err)
			continue
		}

		if err := r.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit message: %w", err)
		}
	}
}
