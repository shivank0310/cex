package risk

import (
	"fmt"
	"sync"
	"time"
)

// VelocityTracker enforces per-user withdrawal rate limits.
type VelocityTracker struct {
	mu      sync.RWMutex
	limits  VelocityLimits
	history map[string][]withdrawalRecord // userID -> recent withdrawals
}

type VelocityLimits struct {
	MaxPerHour  int64
	MaxPerDay   int64
	MaxCountDay int
}

type withdrawalRecord struct {
	Amount    int64
	Timestamp time.Time
}

func NewVelocityTracker(limits VelocityLimits) *VelocityTracker {
	return &VelocityTracker{
		limits:  limits,
		history: make(map[string][]withdrawalRecord),
	}
}

func DefaultVelocityLimits() VelocityLimits {
	return VelocityLimits{
		MaxPerHour:  500_000,
		MaxPerDay:   2_000_000,
		MaxCountDay: 10,
	}
}

// Check returns an error if the user exceeds velocity limits.
func (v *VelocityTracker) Check(userID string, amount int64) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now().UTC()
	hourAgo := now.Add(-1 * time.Hour)
	dayAgo := now.Add(-24 * time.Hour)

	records := v.filterRecent(v.history[userID], dayAgo)
	v.history[userID] = records

	var hourTotal, dayTotal int64
	dayCount := len(records)

	for _, r := range records {
		dayTotal += r.Amount
		if r.Timestamp.After(hourAgo) {
			hourTotal += r.Amount
		}
	}

	if hourTotal+amount > v.limits.MaxPerHour {
		return fmt.Errorf("hourly withdrawal limit exceeded: used %d, limit %d", hourTotal, v.limits.MaxPerHour)
	}
	if dayTotal+amount > v.limits.MaxPerDay {
		return fmt.Errorf("daily withdrawal limit exceeded: used %d, limit %d", dayTotal, v.limits.MaxPerDay)
	}
	if dayCount >= v.limits.MaxCountDay {
		return fmt.Errorf("daily withdrawal count limit exceeded: %d/%d", dayCount, v.limits.MaxCountDay)
	}
	return nil
}

// Record logs a completed withdrawal for velocity tracking.
func (v *VelocityTracker) Record(userID string, amount int64) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.history[userID] = append(v.history[userID], withdrawalRecord{
		Amount: amount, Timestamp: time.Now().UTC(),
	})
}

func (v *VelocityTracker) filterRecent(records []withdrawalRecord, since time.Time) []withdrawalRecord {
	out := make([]withdrawalRecord, 0, len(records))
	for _, r := range records {
		if r.Timestamp.After(since) {
			out = append(out, r)
		}
	}
	return out
}
