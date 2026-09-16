package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// DistributedLock provides Redis-based locking for non-critical coordination.
// NOT used on the matching engine critical path.
type DistributedLock struct {
	client *Client
}

func NewDistributedLock(client *Client) *DistributedLock {
	return &DistributedLock{client: client}
}

// Acquire attempts to acquire a lock. Returns a token on success.
func (l *DistributedLock) Acquire(ctx context.Context, resource string, ttl time.Duration) (string, bool, error) {
	token := fmt.Sprintf("%d", time.Now().UnixNano())
	key := LockKey(resource)
	ok, err := l.client.rdb.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", false, err
	}
	return token, ok, nil
}

// Release releases a lock only if the token matches (safe unlock).
func (l *DistributedLock) Release(ctx context.Context, resource, token string) error {
	key := LockKey(resource)
	script := goredis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		end
		return 0
	`)
	return script.Run(ctx, l.client.rdb, []string{key}, token).Err()
}
