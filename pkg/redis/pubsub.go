package redis

import (
	"context"
	"encoding/json"

	goredis "github.com/redis/go-redis/v9"
)

// PubSub wraps Redis pub/sub for real-time market data broadcasting.
type PubSub struct {
	client *Client
}

func NewPubSub(client *Client) *PubSub {
	return &PubSub{client: client}
}

func (p *PubSub) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return p.client.rdb.Publish(ctx, PubSubChannel(channel), data).Err()
}

func (p *PubSub) Subscribe(ctx context.Context, channels ...string) *goredis.PubSub {
	names := make([]string, len(channels))
	for i, ch := range channels {
		names[i] = PubSubChannel(ch)
	}
	return p.client.rdb.Subscribe(ctx, names...)
}
