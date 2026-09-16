package config

// Config holds runtime settings for the market-data service.
type Config struct {
	HTTPAddr string
}

func Default() Config {
	return Config{HTTPAddr: ":8082"}
}
