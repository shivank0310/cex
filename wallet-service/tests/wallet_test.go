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

func setup() *service.WalletService {
	cfg := config.Default()
	repo := repository.NewWalletRepository()
	ledger := client.NewInMemoryLedgerClient()
	blockchain := client.NewMockBlockchainClient()
	return service.NewWalletService(cfg, repo, ledger, blockchain)
}

func TestDepositFlow(t *testing.T) {
	svc := setup()
	ctx := context.Background()

	// Create blockchain deposit address for alice
	wallet, err := svc.CreateDepositAddress(ctx, "alice", "USDT", "ethereum")
	if err != nil {
		t.Fatal(err)
	}

	// Blockchain service confirms on-chain deposit
	deposit, err := svc.ConfirmDeposit(ctx, "0xabc123", wallet.Address, 1000, 12)
	if err != nil {
		t.Fatal(err)
	}
	if deposit.Status != model.DepositConfirmed {
		t.Fatalf("deposit status: %s", deposit.Status)
	}

	// Alice ledger balance updated
	bal, err := svc.GetLedgerBalance(ctx, "alice", "USDT")
	if err != nil {
		t.Fatal(err)
	}
	if bal.Available != 1000 {
		t.Fatalf("alice USDT: expected 1000, got %d", bal.Available)
	}
}

func TestWithdrawalFlow(t *testing.T) {
	svc := setup()
	ctx := context.Background()

	// Seed alice with 1000 USDT via deposit
	wallet, _ := svc.CreateDepositAddress(ctx, "alice", "USDT", "ethereum")
	_, _ = svc.ConfirmDeposit(ctx, "0xdep1", wallet.Address, 1000, 6)

	// Alice withdraws 500 USDT
	wd, err := svc.RequestWithdrawal(ctx, "alice", "USDT", 500, "0xExternalAddress123456")
	if err != nil {
		t.Fatal(err)
	}
	if wd.Status != model.WithdrawalCompleted {
		t.Fatalf("withdrawal status: %s", wd.Status)
	}
	if wd.TxHash == "" {
		t.Fatal("expected tx hash")
	}

	bal, _ := svc.GetLedgerBalance(ctx, "alice", "USDT")
	if bal.Available != 500 {
		t.Fatalf("alice USDT after withdraw: expected 500, got %d", bal.Available)
	}
	if bal.Locked != 0 {
		t.Fatalf("locked should be 0, got %d", bal.Locked)
	}
}

func TestWithdrawalInsufficientBalance(t *testing.T) {
	svc := setup()
	ctx := context.Background()

	_, err := svc.RequestWithdrawal(ctx, "bob", "BTC", 100, "0xBobAddress12345678")
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

	// Blockchain wallet exists
	wallet, err := svc.CreateDepositAddress(ctx, "alice", "USDT", "ethereum")
	if err != nil || wallet.Address == "" {
		t.Fatal("wallet address required")
	}

	// Ledger balance is zero until deposit confirmed
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

	wallet, _ := svc.CreateDepositAddress(ctx, "alice", "USDT", "ethereum")
	_, _ = svc.ConfirmDeposit(ctx, "0xdep2", wallet.Address, 1000, 6)

	_, err := svc.RequestWithdrawal(ctx, "alice", "USDT", 0, "0xExternalAddress123456")
	if err == nil {
		t.Fatal("expected risk rejection for zero amount")
	}
}
