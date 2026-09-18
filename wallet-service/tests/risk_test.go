package tests

import (
	"testing"

	"github.com/shivank0310/cex.git/wallet-service/internal/config"
	"github.com/shivank0310/cex.git/wallet-service/internal/model"
	"github.com/shivank0310/cex.git/wallet-service/internal/risk"
)

func TestApprovalTiers(t *testing.T) {
	policy := risk.DefaultPolicy()

	if risk.TierForAmount(5_000, policy) != risk.TierAuto {
		t.Fatal("5k should be AUTO")
	}
	if risk.TierForAmount(50_000, policy) != risk.TierMFA {
		t.Fatal("50k should be MFA")
	}
	if risk.TierForAmount(500_000, policy) != risk.TierManualApproval {
		t.Fatal("500k should be MANUAL_APPROVAL")
	}
	if risk.TierForAmount(15_000_000, policy) != risk.TierMultisig {
		t.Fatal("15M should be MULTISIG")
	}
}

func TestRiskEngineAmountCheck(t *testing.T) {
	engine := risk.NewEngine(config.Default())

	_, err := engine.Evaluate(risk.WithdrawalRequest{
		UserID: "alice", Asset: "USDT", Amount: 0,
		ToAddress: testAddress,
		Balance:   model.LedgerBalance{Available: 1000},
	})
	if err == nil {
		t.Fatal("expected amount rejection")
	}
}

func TestRiskEngineVelocityCheck(t *testing.T) {
	engine := risk.NewEngine(config.Default())
	engine.WhitelistAddress("alice", testAddress)

	for i := 0; i < 10; i++ {
		_, err := engine.Evaluate(risk.WithdrawalRequest{
			UserID: "alice", Asset: "USDT", Amount: 1000,
			ToAddress: testAddress,
			Balance:   model.LedgerBalance{Available: 10_000_000},
		})
		if err != nil {
			t.Fatalf("withdrawal %d should pass: %v", i, err)
		}
		engine.RecordCompleted("alice", 1000)
	}

	_, err := engine.Evaluate(risk.WithdrawalRequest{
		UserID: "alice", Asset: "USDT", Amount: 1000,
		ToAddress: testAddress,
		Balance:   model.LedgerBalance{Available: 10_000_000},
	})
	if err == nil {
		t.Fatal("expected velocity limit rejection")
	}
}

func TestRiskEngineAddressBlacklist(t *testing.T) {
	engine := risk.NewEngine(config.Default())
	// Blacklist testing would require exposing blacklist method;
	// address check passes for valid non-empty addresses.
	_, err := engine.Evaluate(risk.WithdrawalRequest{
		UserID: "alice", Asset: "USDT", Amount: 100,
		ToAddress: "",
		Balance:   model.LedgerBalance{Available: 1000},
	})
	if err == nil {
		t.Fatal("expected empty address rejection")
	}
}
