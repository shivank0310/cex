package config

import "os"

// Config holds runtime configuration for the order service.
type Config struct {
	HTTPAddr  string
	LedgerURL string
}

func Default() Config {
	return Config{
		HTTPAddr:  ":8081",
		LedgerURL: os.Getenv("LEDGER_URL"),
	}
}
