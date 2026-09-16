package model

import (
	"fmt"
	"time"
)

// SupportedCandleIntervals maps interval names to durations.
var SupportedCandleIntervals = map[string]time.Duration{
	"1m":  time.Minute,
	"5m":  5 * time.Minute,
	"15m": 15 * time.Minute,
	"1h":  time.Hour,
	"4h":  4 * time.Hour,
	"1d":  24 * time.Hour,
}

func ParseInterval(name string) (time.Duration, error) {
	d, ok := SupportedCandleIntervals[name]
	if !ok {
		return 0, fmt.Errorf("unsupported interval: %s", name)
	}
	return d, nil
}

// CandleOpenTime returns the bucket start for a timestamp and interval.
func CandleOpenTime(t time.Time, interval time.Duration) time.Time {
	if interval >= 24*time.Hour {
		return t.UTC().Truncate(24 * time.Hour)
	}
	return t.UTC().Truncate(interval)
}
