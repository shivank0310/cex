package config

import "time"

type Config struct {
	HTTPAddr              string
	WalletServiceURL      string
	RPCURL                string
	RequiredConfirmations int
	DepositPollInterval   time.Duration
	TxTrackPollInterval   time.Duration
}

func Default() Config {
	return Config{
		HTTPAddr:              ":8085",
		WalletServiceURL:      "http://localhost:8084",
		RPCURL:                "",
		RequiredConfirmations: 12,
		DepositPollInterval:   5 * time.Second,
		TxTrackPollInterval:   3 * time.Second,
	}
}
