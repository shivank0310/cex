package redis

import "os"

// Config holds Redis connection settings.
type Config struct {
	Addr     string
	Password string
	DB       int
}

func DefaultConfig() Config {
	return Config{Addr: "localhost:6379", DB: 0}
}

func ConfigFromEnv() Config {
	cfg := DefaultConfig()
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		cfg.Addr = addr
	}
	if pw := os.Getenv("REDIS_PASSWORD"); pw != "" {
		cfg.Password = pw
	}
	return cfg
}
