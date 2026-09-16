package middleware

import (
	"net/http"
	"strings"

	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/auth"
	"github.com/shivank0310/cex.git/order-service/internal/httputil"
)

// AuthMiddleware authenticates requests and injects user ID into context.
func AuthMiddleware(authenticator auth.Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			userID, err := authenticator.Authenticate(r.Context(), token)
			if err != nil {
				httputil.WriteError(w, err)
				return
			}
			ctx := auth.WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// RecoverMiddleware catches panics and returns a standardized error response.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				httputil.WriteError(w, apperrors.New(apperrors.CodeInternal, "internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
