package middleware

import (
	"net/http"

	"github.com/shivank0310/cex.git/admin-service/internal/apperrors"
	"github.com/shivank0310/cex.git/admin-service/internal/httputil"
)

const adminKeyHeader = "X-Admin-API-Key"

// AdminAuth validates the admin API key on every request.
func AdminAuth(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		key := r.Header.Get(adminKeyHeader)
		if key == "" {
			key = r.Header.Get("Authorization")
		}
		if apiKey != "" && key != apiKey {
			httputil.WriteError(w, apperrors.New(apperrors.CodeUnauthorized, "invalid admin credentials"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
