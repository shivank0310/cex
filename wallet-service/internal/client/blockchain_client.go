package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// BlockchainClient talks to blockchain-service for on-chain operations.
// Core exchange services must not call EVM RPC directly.
type BlockchainClient interface {
	GenerateAddress(ctx context.Context, userID, asset, chain string) (string, error)
	BroadcastWithdrawal(ctx context.Context, asset, toAddress string, amount int64) (string, error)
	ValidateAddress(ctx context.Context, asset, address string) error
}

type HTTPBlockchainClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPBlockchainClient(baseURL string) *HTTPBlockchainClient {
	return &HTTPBlockchainClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type createAddressRequest struct {
	UserID string `json:"user_id"`
	Asset  string `json:"asset"`
	Chain  string `json:"chain"`
}

type createAddressResponse struct {
	Address string `json:"address"`
}

type validateAddressRequest struct {
	Address string `json:"address"`
}

type withdrawRequest struct {
	Asset     string `json:"asset"`
	ToAddress string `json:"to_address"`
	Amount    int64  `json:"amount"`
}

type withdrawResponse struct {
	TxHash string `json:"tx_hash"`
}

func (c *HTTPBlockchainClient) GenerateAddress(ctx context.Context, userID, asset, chain string) (string, error) {
	body, err := json.Marshal(createAddressRequest{UserID: userID, Asset: asset, Chain: chain})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/blockchain/address", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("blockchain request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("blockchain address error: status %d", resp.StatusCode)
	}

	var out createAddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Address, nil
}

func (c *HTTPBlockchainClient) BroadcastWithdrawal(ctx context.Context, asset, toAddress string, amount int64) (string, error) {
	body, err := json.Marshal(withdrawRequest{Asset: asset, ToAddress: toAddress, Amount: amount})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/blockchain/withdraw", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("blockchain withdraw failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("blockchain withdraw error: status %d", resp.StatusCode)
	}

	var out withdrawResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.TxHash, nil
}

func (c *HTTPBlockchainClient) ValidateAddress(ctx context.Context, _, address string) error {
	body, err := json.Marshal(validateAddressRequest{Address: address})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/blockchain/address/validate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("invalid blockchain address")
	}
	return nil
}

// MockBlockchainClient simulates blockchain-service for unit tests.
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
