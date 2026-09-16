package redis

import (
	"context"
	"encoding/json"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Cache provides JSON get/set with TTL. Used for market-data, sessions, etc.
type Cache struct {
	client *Client
}

func NewCache(client *Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.rdb.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	data, err := c.client.rdb.Get(ctx, key).Bytes()
	if err == goredis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.rdb.Del(ctx, keys...).Err()
}
