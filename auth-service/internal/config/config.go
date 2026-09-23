package config

import "time"

type Config struct {
	HTTPAddr         string
	JWTSecret        string
	JWTIssuer        string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	RedisAddr        string
	DatabaseURL      string
	UserServiceURL   string
	BcryptCost       int
}

func Default() Config {
	return Config{
		HTTPAddr:        ":8090",
		JWTSecret:       "cex-dev-jwt-secret-change-in-production",
		JWTIssuer:       "cex-auth",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		BcryptCost:      12,
	}
}
