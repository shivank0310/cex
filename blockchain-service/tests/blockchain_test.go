package tests

import (
	"context"
	"testing"
	"time"

	"github.com/shivank0310/cex.git/blockchain-service/internal/client"
	"github.com/shivank0310/cex.git/blockchain-service/internal/config"
	"github.com/shivank0310/cex.git/blockchain-service/internal/evm"
	"github.com/shivank0310/cex.git/blockchain-service/internal/repository"
	"github.com/shivank0310/cex.git/blockchain-service/internal/service"
)

func setup(t *testing.T) (*service.BlockchainService, *evm.MockProvider, *client.InMemoryWalletClient) {
	cfg := config.Default()
	cfg.RequiredConfirmations = 1
	cfg.DepositPollInterval = 50 * time.Millisecond
	cfg.TxTrackPollInterval = 50 * time.Millisecond

	repo := repository.NewRepository()
	provider := evm.NewMockProvider("ethereum")
	wallet := client.NewInMemoryWalletClient()
	svc := service.NewBlockchainService(cfg, repo, provider, wallet)
	return svc, provider, wallet
}

func TestCreateAddress(t *testing.T) {
	svc, _, _ := setup(t)
	ctx := context.Background()

	wallet, err := svc.CreateAddress(ctx, "alice", "USDT", "ethereum")
	if err != nil {
		t.Fatal(err)
	}
	if wallet.Address == "" {
		t.Fatal("expected address")
	}
	if wallet.Address[:2] != "0x" {
		t.Fatalf("expected 0x address, got %s", wallet.Address)
	}

	// Idempotent: same user/asset returns existing
	wallet2, err := svc.CreateAddress(ctx, "alice", "USDT", "ethereum")
	if err != nil {
		t.Fatal(err)
	}
	if wallet2.Address != wallet.Address {
		t.Fatal("expected same address on repeat request")
	}
}

func TestValidateAddress(t *testing.T) {
	svc, _, _ := setup(t)
	ctx := context.Background()

	if err := svc.ValidateAddress(ctx, "0xabc1234567"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ValidateAddress(ctx, "bad"); err == nil {
		t.Fatal("expected invalid address error")
	}
}

func TestBroadcastWithdrawal(t *testing.T) {
	svc, _, _ := setup(t)
	ctx := context.Background()

	tx, err := svc.BroadcastWithdrawal(ctx, "USDT", "0xrecipient123456", 500)
	if err != nil {
		t.Fatal(err)
	}
	if tx.TxHash == "" {
		t.Fatal("expected tx hash")
	}
	if tx.Status != "PENDING" && tx.Status != "CONFIRMED" {
		t.Fatalf("unexpected status: %s", tx.Status)
	}

	got, err := svc.GetTransaction(ctx, tx.TxHash)
	if err != nil {
		t.Fatal(err)
	}
	if got.TxHash != tx.TxHash {
		t.Fatal("transaction not tracked")
	}
}

func TestDepositMonitorCreditsWallet(t *testing.T) {
	svc, provider, wallet := setup(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go svc.RunDepositMonitor(ctx)

	custody, err := svc.CreateAddress(ctx, "alice", "USDT", "ethereum")
	if err != nil {
		t.Fatal(err)
	}

	provider.SimulateDeposit(custody.Address, "USDT", 1000)
	provider.AdvanceBlocks(1)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(wallet.Confirmed) > 0 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if len(wallet.Confirmed) != 1 {
		t.Fatalf("expected wallet confirm, got %d", len(wallet.Confirmed))
	}
	if wallet.Confirmed[0].Amount != 1000 {
		t.Fatalf("expected amount 1000, got %d", wallet.Confirmed[0].Amount)
	}
}

func TestWithdrawalRejectsInvalidAddress(t *testing.T) {
	svc, _, _ := setup(t)
	_, err := svc.BroadcastWithdrawal(context.Background(), "USDT", "not-an-address", 100)
	if err == nil {
		t.Fatal("expected invalid address error")
	}
}
