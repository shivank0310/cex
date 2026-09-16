package evm

import (
	"context"

	"github.com/shivank0310/cex.git/blockchain-service/internal/model"
)

// Receipt is an on-chain transaction receipt.
type Receipt struct {
	TxHash        string
	BlockNumber   uint64
	Confirmations int
	Status        model.TxStatus
}

// Provider abstracts EVM chain access so core services never call RPC directly.
type Provider interface {
	Chain() string
	ValidateAddress(address string) bool
	GenerateAddress(ctx context.Context, userID, asset string) (string, error)
	BroadcastWithdrawal(ctx context.Context, asset, toAddress string, amount int64) (string, error)
	GetReceipt(ctx context.Context, txHash string) (*Receipt, error)
	GetBlockNumber(ctx context.Context) (uint64, error)
	ScanDeposits(ctx context.Context, addresses []string, fromBlock uint64) ([]model.DepositEvent, error)
}
