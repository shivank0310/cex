package client

import (
	"context"
	"net/http"
	"time"

	"github.com/shivank0310/cex.git/admin-service/internal/config"
	"github.com/shivank0310/cex.git/admin-service/internal/model"
)

// HealthChecker probes downstream CEX services for the admin dashboard.
type HealthChecker interface {
	CheckAll(ctx context.Context) []model.SystemComponent
}

type HTTPHealthChecker struct {
	httpClient *http.Client
	targets    []healthTarget
}

type healthTarget struct {
	name string
	url  string
}

func NewHTTPHealthChecker(cfg config.Config) *HTTPHealthChecker {
	return &HTTPHealthChecker{
		httpClient: &http.Client{Timeout: 3 * time.Second},
		targets: []healthTarget{
			{name: "Matching Engine", url: cfg.OrderServiceURL + "/health"},
			{name: "Market Data", url: cfg.MarketDataURL + "/health"},
			{name: "Ledger", url: cfg.LedgerServiceURL + "/health"},
			{name: "Database", url: cfg.LedgerServiceURL + "/health"}, // ledger proxies DB readiness
			{name: "Wallet Service", url: cfg.WalletServiceURL + "/health"},
			{name: "Blockchain RPC", url: cfg.BlockchainServiceURL + "/health"},
		},
	}
}

func (c *HTTPHealthChecker) CheckAll(ctx context.Context) []model.SystemComponent {
	out := make([]model.SystemComponent, 0, len(c.targets))
	seen := make(map[string]bool)
	for _, t := range c.targets {
		if seen[t.name] {
			continue
		}
		seen[t.name] = true
		out = append(out, c.check(ctx, t.name, t.url))
	}
	// Kafka inferred from market-data health (consumer dependency)
	kafkaStatus := model.HealthHealthy
	for _, comp := range out {
		if comp.Name == "Market Data" && comp.Status != model.HealthHealthy {
			kafkaStatus = model.HealthDegraded
		}
	}
	out = append(out, model.SystemComponent{
		Name: "Kafka", Status: kafkaStatus, LatencyMS: 0,
		Message: "inferred from event consumers", CheckedAt: time.Now().UTC(),
	})
	return out
}

func (c *HTTPHealthChecker) check(ctx context.Context, name, url string) model.SystemComponent {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return componentDown(name, err.Error(), start)
	}
	resp, err := c.httpClient.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return componentDown(name, err.Error(), start)
	}
	defer resp.Body.Close()
	status := model.HealthHealthy
	msg := "ok"
	if resp.StatusCode >= 500 {
		status = model.HealthDown
		msg = resp.Status
	} else if resp.StatusCode >= 400 {
		status = model.HealthDegraded
		msg = resp.Status
	}
	return model.SystemComponent{
		Name: name, Status: status, LatencyMS: latency,
		Message: msg, CheckedAt: time.Now().UTC(),
	}
}

func componentDown(name, msg string, start time.Time) model.SystemComponent {
	return model.SystemComponent{
		Name: name, Status: model.HealthDown,
		LatencyMS: time.Since(start).Milliseconds(),
		Message: msg, CheckedAt: time.Now().UTC(),
	}
}

// InMemoryHealthChecker returns healthy for all components (tests).
type InMemoryHealthChecker struct{}

func (InMemoryHealthChecker) CheckAll(_ context.Context) []model.SystemComponent {
	now := time.Now().UTC()
	names := []string{"Matching Engine", "Database", "Kafka", "Blockchain RPC", "Ledger", "Wallet Service"}
	out := make([]model.SystemComponent, len(names))
	for i, n := range names {
		out[i] = model.SystemComponent{Name: n, Status: model.HealthHealthy, LatencyMS: 1, Message: "ok", CheckedAt: now}
	}
	return out
}
