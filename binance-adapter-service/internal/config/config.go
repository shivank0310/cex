package config

import "os"

type Config struct {
	HTTPAddr    string
	APIKey      string
	APISecret   string
	BaseURL     string
	RecvWindow  int64
	UseMock     bool
}

func Default() Config {
	apiKey := os.Getenv("BINANCE_API_KEY")
	apiSecret := os.Getenv("BINANCE_API_SECRET")
	baseURL := os.Getenv("BINANCE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://testnet.binance.vision"
	}

	return Config{
		HTTPAddr:   ":8086",
		APIKey:     apiKey,
		APISecret:  apiSecret,
		BaseURL:    baseURL,
		RecvWindow: 5000,
		UseMock:    apiKey == "" || apiSecret == "",
	}
}
