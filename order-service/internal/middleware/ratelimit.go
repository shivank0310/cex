package middleware

import (
	"net/http"

	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/auth"
	"github.com/shivank0310/cex.git/order-service/internal/httputil"
	"github.com/shivank0310/cex.git/pkg/redis"
)

// RateLimitMiddleware enforces per-user request limits using Redis.
func RateLimitMiddleware(limiter *redis.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := auth.UserIDFromContext(r.Context())
			if !ok {
				userID = r.RemoteAddr
			}

			allowed, err := limiter.Allow(r.Context(), userID, r.URL.Path)
			if err != nil {
				httputil.WriteError(w, apperrors.Wrap(apperrors.CodeInternal, "rate limit check failed", err))
				return
			}
			if !allowed {
				httputil.WriteError(w, apperrors.New(apperrors.CodeRateLimited, "rate limit exceeded"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
