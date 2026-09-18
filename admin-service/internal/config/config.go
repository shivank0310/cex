package config

type Config struct {
	HTTPAddr       string
	AdminAPIKey    string
	OrderServiceURL      string
	MarketDataURL        string
	LedgerServiceURL     string
	WalletServiceURL     string
	BlockchainServiceURL string
}

func Default() Config {
	return Config{
		HTTPAddr:             ":8087",
		AdminAPIKey:          "admin-dev-key",
		OrderServiceURL:      "http://localhost:8081",
		MarketDataURL:        "http://localhost:8082",
		LedgerServiceURL:     "http://localhost:8083",
		WalletServiceURL:     "http://localhost:8084",
		BlockchainServiceURL: "http://localhost:8085",
	}
}
