package redis

import (
	"context"
	"fmt"
	"time"
)

// RateLimiter implements a fixed-window counter per user+endpoint.
type RateLimiter struct {
	client *Client
	limit  int64
	window time.Duration
}

func NewRateLimiter(client *Client, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{client: client, limit: limit, window: window}
}

// Allow returns true if the request is within the rate limit.
func (r *RateLimiter) Allow(ctx context.Context, userID, endpoint string) (bool, error) {
	key := RateLimitKey(userID, endpoint)
	pipe := r.client.rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, r.window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}
	count, err := incr.Result()
	if err != nil {
		return false, err
	}
	return count <= r.limit, nil
}

// Remaining returns how many requests are left in the current window.
func (r *RateLimiter) Remaining(ctx context.Context, userID, endpoint string) (int64, error) {
	key := RateLimitKey(userID, endpoint)
	count, err := r.client.rdb.Get(ctx, key).Int64()
	if err != nil {
		return r.limit, nil
	}
	remaining := r.limit - count
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

func (r *RateLimiter) Reset(ctx context.Context, userID, endpoint string) error {
	return r.client.rdb.Del(ctx, RateLimitKey(userID, endpoint)).Err()
}

// RateLimitError carries retry information for HTTP 429 responses.
type RateLimitError struct {
	Limit     int64
	Remaining int64
	Window    time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %d requests per %s", e.Limit, e.Window)
}
