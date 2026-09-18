package config

type Config struct {
	HTTPAddr string
	DatabaseURL string
}

func Default() Config {
	return Config{
		HTTPAddr:    ":8091",
		DatabaseURL: "postgres://cex:cex@localhost:5432/cex?sslmode=disable",
	}
}
