package model

import "time"

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusSuspended Status = "SUSPENDED"
)

type KYCStatus string

const (
	KYCNone     KYCStatus = "NONE"
	KYCPending  KYCStatus = "PENDING"
	KYCApproved KYCStatus = "APPROVED"
	KYCRejected KYCStatus = "REJECTED"
)

// User is the persisted user profile (no credentials).
type User struct {
	ID        string
	Email     string
	Username  string
	Status    Status
	KYCStatus KYCStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}
