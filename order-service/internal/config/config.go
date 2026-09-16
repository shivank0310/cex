package config

// Config holds runtime configuration for the order service.
type Config struct {
	HTTPAddr string
}

func Default() Config {
	return Config{
		HTTPAddr: ":8081",
	}
}
