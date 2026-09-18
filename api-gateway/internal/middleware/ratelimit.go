package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shivank0310/cex.git/pkg/redis"
)

type Limiter interface {
	Allow(r *http.Request) bool
}

func RateLimit(limiter Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/gateway/health" {
				next.ServeHTTP(w, r)
				return
			}
			if !limiter.Allow(r) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code": "RATE_LIMITED", "message": "rate limit exceeded",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type RedisLimiter struct {
	inner *redis.RateLimiter
}

func NewRedisLimiter(client *redis.Client, limit int64, window time.Duration) *RedisLimiter {
	return &RedisLimiter{inner: redis.NewRateLimiter(client, limit, window)}
}

func (l *RedisLimiter) Allow(r *http.Request) bool {
	key := rateLimitKey(r)
	allowed, err := l.inner.Allow(r.Context(), key, r.URL.Path)
	return err == nil && allowed
}

type MemoryLimiter struct {
	mu      sync.Mutex
	limit   int64
	window  time.Duration
	buckets map[string][]time.Time
}

func NewMemoryLimiter(limit int64, window time.Duration) *MemoryLimiter {
	return &MemoryLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string][]time.Time),
	}
}

func (l *MemoryLimiter) Allow(r *http.Request) bool {
	key := rateLimitKey(r)
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	events := l.buckets[key]
	filtered := make([]time.Time, 0, len(events))
	for _, ts := range events {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}
	if int64(len(filtered)) >= l.limit {
		l.buckets[key] = filtered
		return false
	}
	filtered = append(filtered, now)
	l.buckets[key] = filtered
	return true
}

func rateLimitKey(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth != "" {
		return "auth:" + auth
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return "ip:" + ip
	}
	host, _, found := strings.Cut(r.RemoteAddr, ":")
	if found {
		return "ip:" + host
	}
	return "ip:" + r.RemoteAddr
}
