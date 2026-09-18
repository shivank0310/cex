package tests

import (
	"testing"

	"github.com/shivank0310/cex.git/wallet-service/internal/security"
)

func TestMFAVerification(t *testing.T) {
	mfa := security.NewTOTPVerifier()
	mfa.Enable("alice")

	if err := mfa.Verify("alice", "123456"); err != nil {
		t.Fatalf("valid code should pass: %v", err)
	}
	if err := mfa.Verify("alice", "000000"); err == nil {
		t.Fatal("invalid code should fail")
	}
	if err := mfa.Verify("bob", "123456"); err == nil {
		t.Fatal("MFA not enabled for bob")
	}
}

func TestMultisigQuorum(t *testing.T) {
	ms := security.NewMultisigWallet()

	if ms.IsFullySigned("wd-1") {
		t.Fatal("should not be signed yet")
	}

	_ = ms.Sign("wd-1", "key-a", "ops")
	if ms.IsFullySigned("wd-1") {
		t.Fatal("1-of-3 should not be enough")
	}

	_ = ms.Sign("wd-1", "key-b", "compliance")
	if !ms.IsFullySigned("wd-1") {
		t.Fatal("2-of-3 should be enough")
	}

	signed, err := ms.SignAndBroadcast("wd-1", "USDT", testAddress, 15_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if signed.KeyID != "hsm-key-c-cold" {
		t.Fatalf("expected cold storage key, got %s", signed.KeyID)
	}
}

func TestHSMSigning(t *testing.T) {
	pool := security.NewHSMSignerPool()

	hot, err := pool.SignForAmount("USDT", testAddress, 5000, "wd-hot", "AUTO")
	if err != nil {
		t.Fatal(err)
	}
	if hot.KeyID != "hsm-key-a-hot" {
		t.Fatalf("expected hot key, got %s", hot.KeyID)
	}

	warm, err := pool.SignForAmount("USDT", testAddress, 200_000, "wd-warm", "MFA")
	if err != nil {
		t.Fatal(err)
	}
	if warm.KeyID != "hsm-key-b-warm" {
		t.Fatalf("expected warm key, got %s", warm.KeyID)
	}
}

func TestApprovalPolicy(t *testing.T) {
	store := security.NewApprovalStore()
	policy := security.NewApprovalPolicy(store)

	if err := policy.RequireApproval("wd-1"); err == nil {
		t.Fatal("should require approval")
	}

	policy.Approve("wd-1", "admin", "ok")
	if err := policy.RequireApproval("wd-1"); err != nil {
		t.Fatal("should be approved")
	}
}
