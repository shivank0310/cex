package model

import "time"

type Role string

const (
	RoleTrader Role = "TRADER"
	RoleAdmin  Role = "ADMIN"
)

type UserStatus string

const (
	UserActive    UserStatus = "ACTIVE"
	UserSuspended UserStatus = "SUSPENDED"
)

// User is the auth identity record. Profile data lives in user-service.
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	Role         Role
	Status       UserStatus
	KYCStatus    string
	TwoFAEnabled bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Session tracks a refresh-token login session.
type Session struct {
	ID           string
	UserID       string
	RefreshToken string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	RevokedAt    *time.Time
}

func (s Session) Active(now time.Time) bool {
	return s.RevokedAt == nil && now.Before(s.ExpiresAt)
}

// TokenPair is returned after register/login/refresh.
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
	TokenType        string
}
