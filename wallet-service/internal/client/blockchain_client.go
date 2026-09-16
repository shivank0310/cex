package client

import (
	"context"
	"fmt"
	"sync/atomic"
)

// BlockchainClient talks to blockchain-service for on-chain operations.
type BlockchainClient interface {
	GenerateAddress(ctx context.Context, userID, asset, chain string) (string, error)
	BroadcastWithdrawal(ctx context.Context, asset, toAddress string, amount int64) (string, error)
	ValidateAddress(ctx context.Context, asset, address string) error
}

// MockBlockchainClient simulates blockchain-service for development and tests.
type MockBlockchainClient struct {
	seq uint64
}

func NewMockBlockchainClient() *MockBlockchainClient {
	return &MockBlockchainClient{}
}

func (c *MockBlockchainClient) GenerateAddress(_ context.Context, userID, asset, chain string) (string, error) {
	id := atomic.AddUint64(&c.seq, 1)
	return fmt.Sprintf("%s-%s-addr-%s-%d", chain, asset, userID, id), nil
}

func (c *MockBlockchainClient) BroadcastWithdrawal(_ context.Context, asset, toAddress string, amount int64) (string, error) {
	id := atomic.AddUint64(&c.seq, 1)
	return fmt.Sprintf("0xtx-%s-%d-%d", asset, amount, id), nil
}

func (c *MockBlockchainClient) ValidateAddress(_ context.Context, _, address string) error {
	if address == "" || len(address) < 10 {
		return fmt.Errorf("invalid blockchain address")
	}
	return nil
}
