package kafka

// Config holds Kafka connection settings.
type Config struct {
	Brokers []string
}

func DefaultConfig() Config {
	return Config{Brokers: []string{"localhost:9092"}}
}
