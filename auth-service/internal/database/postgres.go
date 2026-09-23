package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseURL string) (*sql.DB, error) {
	return OpenWithRetry(databaseURL, 60*time.Second)
}

func OpenWithRetry(databaseURL string, maxWait time.Duration) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	deadline := time.Now().Add(maxWait)
	var lastErr error
	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := db.PingContext(ctx)
		cancel()
		if err == nil {
			return db, nil
		}
		lastErr = fmt.Errorf("ping database: %w", err)
		if time.Now().After(deadline) {
			db.Close()
			return nil, lastErr
		}
		sleep := time.Duration(attempt) * time.Second
		if sleep > 5*time.Second {
			sleep = 5 * time.Second
		}
		time.Sleep(sleep)
	}
}
