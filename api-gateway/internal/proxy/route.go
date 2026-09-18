package proxy

import "github.com/shivank0310/cex.git/api-gateway/internal/config"

// Route maps a gateway path prefix to a backend service path prefix.
type Route struct {
	Name          string
	GatewayPrefix string
	ServiceURL    string
	ServicePrefix string
}

func Routes(cfg config.Config) []Route {
	return []Route{
		{Name: "auth", GatewayPrefix: "/api/v1/auth", ServiceURL: cfg.AuthServiceURL, ServicePrefix: "/api/v1/auth"},
		{Name: "auth-short", GatewayPrefix: "/api/auth", ServiceURL: cfg.AuthServiceURL, ServicePrefix: "/api/v1/auth"},

		{Name: "users", GatewayPrefix: "/api/v1/users", ServiceURL: cfg.UserServiceURL, ServicePrefix: "/api/v1/users"},
		{Name: "users-short", GatewayPrefix: "/api/users", ServiceURL: cfg.UserServiceURL, ServicePrefix: "/api/v1/users"},

		{Name: "orders", GatewayPrefix: "/api/v1/orders", ServiceURL: cfg.OrderServiceURL, ServicePrefix: "/api/v1/orders"},
		{Name: "orders-short", GatewayPrefix: "/api/orders", ServiceURL: cfg.OrderServiceURL, ServicePrefix: "/api/v1/orders"},

		{Name: "market", GatewayPrefix: "/api/v1/market", ServiceURL: cfg.MarketDataURL, ServicePrefix: "/api/v1/market"},
		{Name: "market-short", GatewayPrefix: "/api/market", ServiceURL: cfg.MarketDataURL, ServicePrefix: "/api/v1/market"},

		{Name: "ledger", GatewayPrefix: "/api/v1/ledger", ServiceURL: cfg.LedgerServiceURL, ServicePrefix: "/api/v1/ledger"},
		{Name: "ledger-short", GatewayPrefix: "/api/ledger", ServiceURL: cfg.LedgerServiceURL, ServicePrefix: "/api/v1/ledger"},

		{Name: "wallet", GatewayPrefix: "/api/v1/wallet", ServiceURL: cfg.WalletServiceURL, ServicePrefix: "/api/v1/wallet"},
		{Name: "wallet-short", GatewayPrefix: "/api/wallet", ServiceURL: cfg.WalletServiceURL, ServicePrefix: "/api/v1/wallet"},

		{Name: "blockchain", GatewayPrefix: "/api/v1/blockchain", ServiceURL: cfg.BlockchainServiceURL, ServicePrefix: "/api/v1/blockchain"},
		{Name: "blockchain-short", GatewayPrefix: "/api/blockchain", ServiceURL: cfg.BlockchainServiceURL, ServicePrefix: "/api/v1/blockchain"},

		{Name: "admin", GatewayPrefix: "/api/v1/admin", ServiceURL: cfg.AdminServiceURL, ServicePrefix: "/api/v1/admin"},
		{Name: "admin-short", GatewayPrefix: "/api/admin", ServiceURL: cfg.AdminServiceURL, ServicePrefix: "/api/v1/admin"},

		{Name: "matching", GatewayPrefix: "/api/v1/matching", ServiceURL: cfg.MatchingServiceURL, ServicePrefix: "/api/v1/matching"},
		{Name: "matching-short", GatewayPrefix: "/api/matching", ServiceURL: cfg.MatchingServiceURL, ServicePrefix: "/api/v1/matching"},
	}
}
