package config

import "os"

type Config struct {
	HTTPAddr      string
	APIKey        string
	APISecret     string
	BaseURL       string
	MarketDataURL string
	RecvWindow    int64
	UseMockOrders bool
}

func Default() Config {
	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")
	baseURL := os.Getenv("BINANCE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://testnet.binance.vision"
	}
	marketDataURL := os.Getenv("BINANCE_MARKET_DATA_URL")
	if marketDataURL == "" {
		marketDataURL = "https://api.binance.com"
	}

	return Config{
		HTTPAddr:      ":8086",
		APIKey:        apiKey,
		APISecret:     apiSecret,
		BaseURL:       baseURL,
		MarketDataURL: marketDataURL,
		RecvWindow:    5000,
		UseMockOrders: apiKey == "" || apiSecret == "",
	}
}
