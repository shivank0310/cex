package model

import "time"

// CustodyWallet is an on-chain deposit address managed by blockchain-service.
type CustodyWallet struct {
	ID        string
	UserID    string
	Asset     string
	Chain     string
	Address   string
	CreatedAt time.Time
}

// WatchedAddress is monitored for incoming deposits.
type WatchedAddress struct {
	Address   string
	UserID    string
	Asset     string
	Chain     string
	AddedAt   time.Time
	LastBlock uint64
}
