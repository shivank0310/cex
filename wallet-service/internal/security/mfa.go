package security

import (
	"fmt"
	"sync"
)

// MFAVerifier validates 2FA/MFA codes before withdrawal execution.
type MFAVerifier interface {
	Verify(userID, code string) error
	IsEnabled(userID string) bool
}

// TOTPVerifier is a mock TOTP verifier for development.
// In production, integrate with Google Authenticator / Authy / SMS OTP.
type TOTPVerifier struct {
	mu       sync.RWMutex
	enabled  map[string]bool
	validCode string // demo: "123456"
}

func NewTOTPVerifier() *TOTPVerifier {
	return &TOTPVerifier{
		enabled:   make(map[string]bool),
		validCode: "123456",
	}
}

func (v *TOTPVerifier) Enable(userID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.enabled[userID] = true
}

func (v *TOTPVerifier) Verify(userID, code string) error {
	v.mu.RLock()
	enabled := v.enabled[userID]
	v.mu.RUnlock()

	if !enabled {
		return fmt.Errorf("MFA not enabled for user")
	}
	if code != v.validCode {
		return fmt.Errorf("invalid MFA code")
	}
	return nil
}

func (v *TOTPVerifier) IsEnabled(userID string) bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.enabled[userID]
}
