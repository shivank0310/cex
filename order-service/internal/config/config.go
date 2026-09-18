package config

import "os"

// Config holds runtime configuration for the order service.
type Config struct {
	HTTPAddr         string
	LedgerURL        string
	BinanceURL       string
	RouteBinance     bool
}

func Default() Config {
	return Config{
		HTTPAddr:     ":8081",
		LedgerURL:    os.Getenv("LEDGER_URL"),
		BinanceURL:   os.Getenv("BINANCE_ADAPTER_URL"),
		RouteBinance: os.Getenv("ROUTE_BINANCE") == "1",
	}
}
