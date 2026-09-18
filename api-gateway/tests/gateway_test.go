package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shivank0310/cex.git/api-gateway/internal/config"
	"github.com/shivank0310/cex.git/api-gateway/internal/middleware"
	"github.com/shivank0310/cex.git/api-gateway/internal/proxy"
)

func TestGatewayHealth(t *testing.T) {
	gw := newTestGateway(t, config.Default())
	req := httptest.NewRequest(http.MethodGet, "/gateway/health", nil)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestGatewayRoutesVersionedPath(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/orders" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing auth header")
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"ord-1"}`))
	}))
	defer backend.Close()

	cfg := config.Default()
	cfg.OrderServiceURL = backend.URL
	gw := newTestGateway(t, cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestGatewayRewritesShortPath(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/orders/ord-1" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := config.Default()
	cfg.OrderServiceURL = backend.URL
	gw := newTestGateway(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/orders/ord-1", nil)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestGatewayForwardsAdminKey(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/admin/dashboard" {
			t.Fatalf("path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Admin-API-Key") != "admin-dev-key" {
			t.Fatalf("admin key not forwarded")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := config.Default()
	cfg.AdminServiceURL = backend.URL
	gw := newTestGateway(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
	req.Header.Set("X-Admin-API-Key", "admin-dev-key")
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	gw := newTestGateway(t, config.Default())
	req := httptest.NewRequest(http.MethodOptions, "/api/orders", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("expected CORS header")
	}
}

func TestRateLimit(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := config.Default()
	cfg.OrderServiceURL = backend.URL
	cfg.RateLimitPerMinute = 2

	gateway, err := proxy.NewGateway(cfg)
	if err != nil {
		t.Fatal(err)
	}
	handler := middleware.RateLimit(middleware.NewMemoryLimiter(2, time.Minute))(gateway)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status: %d", i, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func newTestGateway(t *testing.T, cfg config.Config) http.Handler {
	gateway, err := proxy.NewGateway(cfg)
	if err != nil {
		t.Fatal(err)
	}
	handler := middleware.CORS(cfg.CORSAllowedOrigins)(gateway)
	handler = middleware.RequestID(handler)
	return handler
}
