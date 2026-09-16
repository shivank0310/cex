package config

type Config struct {
	HTTPAddr    string
	LedgerURL   string
	MinDeposit  int64
	MinWithdraw int64
	MaxWithdraw int64
}

func Default() Config {
	return Config{
		HTTPAddr:    ":8084",
		LedgerURL:   "http://localhost:8083",
		MinDeposit:  1,
		MinWithdraw: 1,
		MaxWithdraw: 1_000_000,
	}
}
