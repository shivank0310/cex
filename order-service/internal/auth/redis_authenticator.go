package auth

import (
	"context"

	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/pkg/redis"
)

// RedisAuthenticator caches successful auth lookups in Redis for fast session validation.
// Falls back to the underlying authenticator on cache miss.
type RedisAuthenticator struct {
	inner   Authenticator
	sessions *redis.SessionStore
}

func NewRedisAuthenticator(inner Authenticator, sessions *redis.SessionStore) *RedisAuthenticator {
	return &RedisAuthenticator{inner: inner, sessions: sessions}
}

func (a *RedisAuthenticator) Authenticate(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", apperrors.New(apperrors.CodeUnauthorized, "missing authentication token")
	}

	if a.sessions != nil {
		userID, found, err := a.sessions.Get(ctx, token)
		if err == nil && found {
			return userID, nil
		}
	}

	userID, err := a.inner.Authenticate(ctx, token)
	if err != nil {
		return "", err
	}

	if a.sessions != nil {
		_ = a.sessions.Set(ctx, token, userID)
	}
	return userID, nil
}
