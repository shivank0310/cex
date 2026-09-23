package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Client wraps go-redis with CEX-specific helpers.
type Client struct {
	rdb *goredis.Client
}

func NewClient(cfg Config) (*Client, error) {
	return NewClientWithRetry(cfg, 0)
}

// NewClientWithRetry connects to Redis, retrying until maxWait elapses (0 = single attempt).
func NewClientWithRetry(cfg Config, maxWait time.Duration) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	deadline := time.Now().Add(maxWait)
	var lastErr error
	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := rdb.Ping(ctx).Err()
		cancel()
		if err == nil {
			return &Client{rdb: rdb}, nil
		}
		lastErr = fmt.Errorf("redis ping failed: %w", err)
		if maxWait == 0 || time.Now().After(deadline) {
			_ = rdb.Close()
			return nil, lastErr
		}
		sleep := time.Duration(attempt) * time.Second
		if sleep > 5*time.Second {
			sleep = 5 * time.Second
		}
		time.Sleep(sleep)
	}
}

func (c *Client) Raw() *goredis.Client { return c.rdb }

func (c *Client) Close() error { return c.rdb.Close() }

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}
