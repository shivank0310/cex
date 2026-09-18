package config

import (
	"os"
	"strings"
)

// Config holds runtime settings for the market-data service.
type Config struct {
	HTTPAddr            string
	BinanceAdapterURL   string
	ExternalSymbols     []string
}

func Default() Config {
	cfg := Config{HTTPAddr: ":8082"}
	if url := os.Getenv("BINANCE_ADAPTER_URL"); url != "" {
		cfg.BinanceAdapterURL = url
	}
	if raw := os.Getenv("EXTERNAL_SYMBOLS"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				cfg.ExternalSymbols = append(cfg.ExternalSymbols, s)
			}
		}
	}
	return cfg
}
