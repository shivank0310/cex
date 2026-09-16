package redis

import (
	"context"
	"time"
)

// SessionStore caches auth token → userID mappings.
// PostgreSQL remains the source of truth; Redis is a fast lookup layer.
type SessionStore struct {
	cache *Cache
	ttl   time.Duration
}

func NewSessionStore(cache *Cache, ttl time.Duration) *SessionStore {
	return &SessionStore{cache: cache, ttl: ttl}
}

func (s *SessionStore) Set(ctx context.Context, token, userID string) error {
	return s.cache.Set(ctx, SessionKey(token), userID, s.ttl)
}

func (s *SessionStore) Get(ctx context.Context, token string) (string, bool, error) {
	var userID string
	found, err := s.cache.Get(ctx, SessionKey(token), &userID)
	return userID, found, err
}

func (s *SessionStore) Delete(ctx context.Context, token string) error {
	return s.cache.Delete(ctx, SessionKey(token))
}
