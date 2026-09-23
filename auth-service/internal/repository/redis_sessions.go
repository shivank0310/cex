package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/shivank0310/cex.git/auth-service/internal/model"
	cexredis "github.com/shivank0310/cex.git/pkg/redis"
)

type redisSessionPayload struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	RefreshToken string     `json:"refresh_token"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `json:"expires_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

type RedisSessionStore struct {
	cache *cexredis.Cache
}

func NewRedisSessionStore(cache *cexredis.Cache) *RedisSessionStore {
	return &RedisSessionStore{cache: cache}
}

func (r *RedisSessionStore) Create(userID, refreshToken string, expiresAt time.Time) model.Session {
	now := time.Now().UTC()
	session := model.Session{
		ID:           newSessionID(),
		UserID:       userID,
		RefreshToken: refreshToken,
		CreatedAt:    now,
		ExpiresAt:    expiresAt,
	}

	payload := redisSessionPayload{
		ID:           session.ID,
		UserID:       session.UserID,
		RefreshToken: session.RefreshToken,
		CreatedAt:    session.CreatedAt,
		ExpiresAt:    session.ExpiresAt,
	}

	ttl := expiresAt.Sub(now)
	if ttl < time.Second {
		ttl = time.Second
	}
	ctx := context.Background()
	_ = r.cache.Set(ctx, cexredis.AuthRefreshKey(refreshToken), payload, ttl)

	return session
}

func (r *RedisSessionStore) GetByRefreshToken(token string) (*model.Session, bool) {
	ctx := context.Background()
	var payload redisSessionPayload
	found, err := r.cache.Get(ctx, cexredis.AuthRefreshKey(token), &payload)
	if err != nil || !found {
		return nil, false
	}
	if payload.RevokedAt != nil {
		return nil, false
	}
	session := model.Session{
		ID:           payload.ID,
		UserID:       payload.UserID,
		RefreshToken: payload.RefreshToken,
		CreatedAt:    payload.CreatedAt,
		ExpiresAt:    payload.ExpiresAt,
		RevokedAt:    payload.RevokedAt,
	}
	if !session.Active(time.Now().UTC()) {
		return nil, false
	}
	return &session, true
}

func (r *RedisSessionStore) Revoke(refreshToken string) bool {
	ctx := context.Background()
	key := cexredis.AuthRefreshKey(refreshToken)

	var payload redisSessionPayload
	found, err := r.cache.Get(ctx, key, &payload)
	if err != nil || !found || payload.RevokedAt != nil {
		return false
	}

	now := time.Now().UTC()
	payload.RevokedAt = &now
	ttl := payload.ExpiresAt.Sub(now)
	if ttl < time.Second {
		_ = r.cache.Delete(ctx, key)
		return true
	}
	_ = r.cache.Set(ctx, key, payload, ttl)
	return true
}

func newSessionID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "sess-fallback"
	}
	return "sess-" + hex.EncodeToString(buf)
}
