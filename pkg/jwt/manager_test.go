package jwt

import (
	"testing"
	"time"
)

func TestSignAndValidateAccess(t *testing.T) {
	mgr := NewManager("test-secret", "cex-auth", 15*time.Minute)
	token, _, err := mgr.SignAccess("user-123", "TRADER", "sess-1")
	if err != nil {
		t.Fatal(err)
	}

	claims, err := mgr.ValidateAccess(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "user-123" {
		t.Fatalf("sub: %s", claims.Subject)
	}
	if claims.Role != "TRADER" {
		t.Fatalf("role: %s", claims.Role)
	}
}

func TestValidateAccessRejectsBadToken(t *testing.T) {
	mgr := NewManager("test-secret", "cex-auth", 15*time.Minute)
	_, err := mgr.ValidateAccess("not-a-jwt")
	if err == nil {
		t.Fatal("expected error")
	}
}
