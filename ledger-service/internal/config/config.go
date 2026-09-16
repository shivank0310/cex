package config

type Config struct {
	HTTPAddr string
}

func Default() Config {
	return Config{HTTPAddr: ":8083"}
}
