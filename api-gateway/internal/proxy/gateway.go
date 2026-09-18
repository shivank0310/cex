package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/shivank0310/cex.git/api-gateway/internal/config"
)

// Gateway reverse-proxies API traffic to internal CEX services.
type Gateway struct {
	routes  []Route
	timeout time.Duration
}

func NewGateway(cfg config.Config) (*Gateway, error) {
	routes := Routes(cfg)
	sort.Slice(routes, func(i, j int) bool {
		return len(routes[i].GatewayPrefix) > len(routes[j].GatewayPrefix)
	})
	return &Gateway{routes: routes, timeout: cfg.RequestTimeout()}, nil
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/health" || r.URL.Path == "/gateway/health" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
		return
	}

	route, rest, ok := g.match(r.URL.Path)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"code": "NOT_FOUND", "message": "route not found",
		})
		return
	}

	target, err := url.Parse(strings.TrimRight(route.ServiceURL, "/"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"code": "INTERNAL_ERROR", "message": "invalid upstream",
		})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"code": "UPSTREAM_ERROR", "message": "upstream unavailable",
		})
	}
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		req.URL.Path = joinPath(route.ServicePrefix, rest)
		req.URL.RawPath = ""
		forwardHeaders(r, req)
	}

	inner := proxy.Transport
	if inner == nil {
		inner = http.DefaultTransport
	}
	proxy.Transport = &timeoutTransport{inner: inner, timeout: g.timeout}
	proxy.ServeHTTP(w, r)
}

func (g *Gateway) match(path string) (Route, string, bool) {
	for _, route := range g.routes {
		if path == route.GatewayPrefix || strings.HasPrefix(path, route.GatewayPrefix+"/") {
			rest := strings.TrimPrefix(path, route.GatewayPrefix)
			rest = strings.TrimPrefix(rest, "/")
			return route, rest, true
		}
	}
	return Route{}, "", false
}

func joinPath(prefix, rest string) string {
	prefix = strings.TrimRight(prefix, "/")
	if rest == "" {
		return prefix
	}
	return prefix + "/" + rest
}

func forwardHeaders(in *http.Request, out *http.Request) {
	copyHeader(out.Header, in.Header, "Authorization")
	copyHeader(out.Header, in.Header, "X-Admin-API-Key")
	copyHeader(out.Header, in.Header, "X-Request-ID")
	copyHeader(out.Header, in.Header, "Content-Type")
	copyHeader(out.Header, in.Header, "Accept")

	if out.Header.Get("X-Request-ID") == "" {
		out.Header.Set("X-Request-ID", in.Header.Get("X-Request-Id"))
	}
	if clientIP := clientIP(in); clientIP != "" {
		out.Header.Set("X-Real-IP", clientIP)
		out.Header.Set("X-Forwarded-For", clientIP)
	}
}

func copyHeader(dst, src http.Header, key string) {
	if value := src.Get(key); value != "" {
		dst.Set(key, value)
	}
}

func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		return strings.TrimSpace(parts[0])
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	host, _, found := strings.Cut(r.RemoteAddr, ":")
	if !found {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type timeoutTransport struct {
	inner   http.RoundTripper
	timeout time.Duration
}

func (t *timeoutTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.timeout > 0 {
		if _, ok := req.Context().Deadline(); !ok {
			ctx, cancel := context.WithTimeout(req.Context(), t.timeout)
			defer cancel()
			req = req.WithContext(ctx)
		}
	}
	return t.inner.RoundTrip(req)
}
