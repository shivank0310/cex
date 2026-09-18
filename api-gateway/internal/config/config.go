package config

import "time"

type Config struct {
	HTTPAddr string
	TLSCert  string
	TLSKey   string

	CORSAllowedOrigins []string
	RateLimitPerMinute int64

	AuthServiceURL       string
	UserServiceURL       string
	OrderServiceURL      string
	MarketDataURL        string
	LedgerServiceURL     string
	WalletServiceURL     string
	BlockchainServiceURL string
	AdminServiceURL      string
	MatchingServiceURL   string
}

func Default() Config {
	return Config{
		HTTPAddr:           ":8080",
		CORSAllowedOrigins: []string{"*"},
		RateLimitPerMinute: 300,

		AuthServiceURL:       "http://localhost:8090",
		UserServiceURL:       "http://localhost:8091",
		OrderServiceURL:      "http://localhost:8081",
		MarketDataURL:        "http://localhost:8082",
		LedgerServiceURL:     "http://localhost:8083",
		WalletServiceURL:     "http://localhost:8084",
		BlockchainServiceURL: "http://localhost:8085",
		AdminServiceURL:      "http://localhost:8087",
		MatchingServiceURL:   "http://localhost:8092",
	}
}

func (c Config) RequestTimeout() time.Duration {
	return 30 * time.Second
}
