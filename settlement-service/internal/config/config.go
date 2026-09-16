package config

type Config struct {
	LedgerURL string
}

func Default() Config {
	return Config{
		LedgerURL: "http://localhost:8083",
	}
}
