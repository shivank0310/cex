package tests

import (
	"context"
	"testing"

	"github.com/shivank0310/cex.git/wallet-service/internal/client"
	"github.com/shivank0310/cex.git/wallet-service/internal/config"
	"github.com/shivank0310/cex.git/wallet-service/internal/model"
	"github.com/shivank0310/cex.git/wallet-service/internal/repository"
	"github.com/shivank0310/cex.git/wallet-service/internal/service"
)

const testAddress = "0xExternalAddress123456"

func setup() *service.WalletService {
	cfg := config.Default()
	repo := repository.NewWalletRepository()
	ledger := client.NewInMemoryLedgerClient()
	blockchain := client.NewMockBlockchainClient()
	return service.NewWalletService(cfg, repo, ledger, blockchain)
}

func seedAlice(svc *service.WalletService, ctx context.Context, amount int64) {
	wallet, _ := svc.CreateDepositAddress(ctx, "alice", "USDT", "ethereum")
	_, _ = svc.ConfirmDeposit(ctx, "0xdep-seed", wallet.Address, amount, 6)
	svc.WhitelistAddress("alice", testAddress)
}

func TestDepositFlow(t *testing.T) {
	svc := setup()
	ctx := context.Background()

	wallet, err := svc.CreateDepositAddress(ctx, "alice", "USDT", "ethereum")
	if err != nil {
		t.Fatal(err)
	}

	deposit, err := svc.ConfirmDeposit(ctx, "0xabc123", wallet.Address, 1000, 12)
	if err != nil {
		t.Fatal(err)
	}
	if deposit.Status != model.DepositConfirmed {
		t.Fatalf("deposit status: %s", deposit.Status)
	}

	bal, err := svc.GetLedgerBalance(ctx, "alice", "USDT")
	if err != nil {
		t.Fatal(err)
	}
	if bal.Available != 1000 {
		t.Fatalf("alice USDT: expected 1000, got %d", bal.Available)
	}
}

func TestWithdrawalFlow_AutoTier(t *testing.T) {
	svc := setup()
	ctx := context.Background()
	seedAlice(svc, ctx, 1000)

	// ₹10,000 tier → AUTO with whitelisted address → immediate execution
	wd, err := svc.RequestWithdrawal(ctx, "alice", "USDT", 500, testAddress)
	if err != nil {
		t.Fatal(err)
	}
	if wd.Status != model.WithdrawalCompleted {
		t.Fatalf("expected COMPLETED, got %s", wd.Status)
	}
	if wd.TxHash == "" {
		t.Fatal("expected tx hash")
	}
	if wd.HSMKeyID == "" {
		t.Fatal("expected HSM key ID")
	}

	bal, _ := svc.GetLedgerBalance(ctx, "alice", "USDT")
	if bal.Available != 500 {
		t.Fatalf("alice USDT after withdraw: expected 500, got %d", bal.Available)
	}
}

func TestWithdrawalFlow_MFARequired(t *testing.T) {
	svc := setup()
	ctx := context.Background()
	seedAlice(svc, ctx, 200_000)
	svc.EnableMFA("alice")

	// ₹1,00,000 tier → MFA required
	wd, err := svc.RequestWithdrawal(ctx, "alice", "USDT", 50_000, testAddress)
	if err != nil {
		t.Fatal(err)
	}
	if wd.Status != model.WithdrawalPendingMFA {
		t.Fatalf("expected PENDING_MFA, got %s", wd.Status)
	}

	// Wrong MFA code
	_, err = svc.VerifyMFA(ctx, wd.ID, "alice", "000000")
	if err == nil {
		t.Fatal("expected MFA error")
	}

	// Correct MFA → auto-execute (MFA tier)
	wd, err = svc.VerifyMFA(ctx, wd.ID, "alice", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if wd.Status != model.WithdrawalCompleted {
		t.Fatalf("expected COMPLETED after MFA, got %s", wd.Status)
	}
}

func TestWithdrawalFlow_ManualApproval(t *testing.T) {
	svc := setup()
	ctx := context.Background()
	seedAlice(svc, ctx, 2_000_000)
	svc.EnableMFA("alice")

	wd, err := svc.RequestWithdrawal(ctx, "alice", "USDT", 500_000, testAddress)
	if err != nil {
		t.Fatal(err)
	}
	if wd.ApprovalTier != "MANUAL_APPROVAL" {
		t.Fatalf("expected MANUAL_APPROVAL tier, got %s", wd.ApprovalTier)
	}

	wd, err = svc.VerifyMFA(ctx, wd.ID, "alice", "123456")
	if err != nil {
		t.Fatal(err)
	}
	if wd.Status != model.WithdrawalPendingApproval {
		t.Fatalf("expected PENDING_APPROVAL, got %s", wd.Status)
	}

	wd, err = svc.ApproveWithdrawal(ctx, wd.ID, "ops-admin", "verified KYC")
	if err != nil {
		t.Fatal(err)
	}
	if wd.Status != model.WithdrawalCompleted {
		t.Fatalf("expected COMPLETED after approval, got %s", wd.Status)
	}
	if wd.ApprovedBy != "ops-admin" {
		t.Fatal("expected approver recorded")
	}
}

func TestWithdrawalFlow_Multisig(t *testing.T) {
	svc := setup()
	ctx := context.Background()
	seedAlice(svc, ctx, 20_000_000)
	svc.EnableMFA("alice")

	wd, err := svc.RequestWithdrawal(ctx, "alice", "USDT", 15_000_000, testAddress)
	if err != nil {
		t.Fatal(err)
	}
	if wd.ApprovalTier != "MULTISIG" {
		t.Fatalf("expected MULTISIG tier, got %s", wd.ApprovalTier)
	}

	wd, _ = svc.VerifyMFA(ctx, wd.ID, "alice", "123456")
	wd, _ = svc.ApproveWithdrawal(ctx, wd.ID, "compliance-officer", "treasury withdrawal approved")
	if wd.Status != model.WithdrawalPendingMultisig {
		t.Fatalf("expected PENDING_MULTISIG, got %s", wd.Status)
	}

	// 2-of-3 multisig signatures required
	wd, err = svc.MultisigSign(ctx, wd.ID, "key-a", "signer-ops")
	if err != nil {
		t.Fatal(err)
	}
	if wd.MultisigSigs != 1 {
		t.Fatalf("expected 1 sig, got %d", wd.MultisigSigs)
	}

	wd, err = svc.MultisigSign(ctx, wd.ID, "key-b", "signer-compliance")
	if err != nil {
		t.Fatal(err)
	}
	if wd.Status != model.WithdrawalCompleted {
		t.Fatalf("expected COMPLETED after multisig, got %s", wd.Status)
	}
}

func TestWithdrawalInsufficientBalance(t *testing.T) {
	svc := setup()
	ctx := context.Background()
	_, err := svc.RequestWithdrawal(ctx, "bob", "BTC", 100, testAddress)
	if err == nil {
		t.Fatal("expected insufficient balance error")
	}
}

func TestDepositDuplicateRejected(t *testing.T) {
	svc := setup()
	ctx := context.Background()

	wallet, _ := svc.CreateDepositAddress(ctx, "alice", "BTC", "bitcoin")
	_, err := svc.ConfirmDeposit(ctx, "0xtx-dup", wallet.Address, 100, 6)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.ConfirmDeposit(ctx, "0xtx-dup", wallet.Address, 100, 6)
	if err == nil {
		t.Fatal("expected duplicate deposit error")
	}
}

func TestWalletSeparateFromLedger(t *testing.T) {
	svc := setup()
	ctx := context.Background()

	wallet, err := svc.CreateDepositAddress(ctx, "alice", "USDT", "ethereum")
	if err != nil || wallet.Address == "" {
		t.Fatal("wallet address required")
	}

	bal, err := svc.GetLedgerBalance(ctx, "alice", "USDT")
	if err != nil {
		t.Fatal(err)
	}
	if bal.Available != 0 {
		t.Fatal("ledger should be zero before deposit")
	}
}

func TestRiskRejectsSmallWithdrawal(t *testing.T) {
	svc := setup()
	ctx := context.Background()

	seedAlice(svc, ctx, 1000)
	_, err := svc.RequestWithdrawal(ctx, "alice", "USDT", 0, testAddress)
	if err == nil {
		t.Fatal("expected risk rejection for zero amount")
	}
}

func TestNonWhitelistedAddressRequiresMFA(t *testing.T) {
	svc := setup()
	ctx := context.Background()

	wallet, _ := svc.CreateDepositAddress(ctx, "alice", "USDT", "ethereum")
	_, _ = svc.ConfirmDeposit(ctx, "0xdep-nw", wallet.Address, 1000, 6)
	svc.EnableMFA("alice")

	// Small amount but non-whitelisted address → MFA tier
	wd, err := svc.RequestWithdrawal(ctx, "alice", "USDT", 500, "0xNewUnknownAddress123")
	if err != nil {
		t.Fatal(err)
	}
	if wd.Status != model.WithdrawalPendingMFA {
		t.Fatalf("non-whitelisted address should require MFA, got %s", wd.Status)
	}
}
