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
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

func (c *Client) Raw() *goredis.Client { return c.rdb }

func (c *Client) Close() error { return c.rdb.Close() }

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}
