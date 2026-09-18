package config

type Config struct {
	HTTPAddr      string
	LedgerURL     string
	BlockchainURL string
	MinDeposit    int64
	MinWithdraw   int64
	MaxWithdraw   int64
}

func Default() Config {
	return Config{
		HTTPAddr:      ":8084",
		LedgerURL:     "http://localhost:8083",
		BlockchainURL: "http://localhost:8085",
		MinDeposit:    1,
		MinWithdraw:   1,
		MaxWithdraw:   100_000_000, // must exceed highest approval tier
	}
}
