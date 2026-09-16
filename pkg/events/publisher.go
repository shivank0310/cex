package events

import "context"

// Publisher sends events to the message bus.
type Publisher interface {
	Publish(ctx context.Context, topic string, key string, env Envelope) error
	Close() error
}

// NoOpPublisher discards all events (useful for tests and local dev without Kafka).
type NoOpPublisher struct{}

func (n *NoOpPublisher) Publish(_ context.Context, _ string, _ string, _ Envelope) error {
	return nil
}

func (n *NoOpPublisher) Close() error { return nil }
