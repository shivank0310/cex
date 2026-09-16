package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
	"github.com/shivank0310/cex.git/pkg/events"
)

// Producer publishes envelopes to Kafka topics.
type Producer struct {
	cfg     Config
	writers map[string]*kafka.Writer
	mu      sync.Mutex
}

func NewProducer(cfg Config) *Producer {
	return &Producer{
		cfg:     cfg,
		writers: make(map[string]*kafka.Writer),
	}
}

func (p *Producer) Publish(ctx context.Context, topic string, key string, env events.Envelope) error {
	w, err := p.writer(topic)
	if err != nil {
		return err
	}

	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	return w.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: body,
	})
}

func (p *Producer) writer(topic string) (*kafka.Writer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if w, ok := p.writers[topic]; ok {
		return w, nil
	}

	w := &kafka.Writer{
		Addr:     kafka.TCP(p.cfg.Brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	p.writers[topic] = w
	return w, nil
}

func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var first error
	for _, w := range p.writers {
		if err := w.Close(); err != nil && first == nil {
			first = err
		}
	}
	p.writers = make(map[string]*kafka.Writer)
	return first
}
