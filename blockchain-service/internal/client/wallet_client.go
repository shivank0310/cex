package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WalletClient notifies wallet-service of confirmed on-chain deposits.
type WalletClient interface {
	ConfirmDeposit(ctx context.Context, txHash, toAddress string, amount int64, confirmations int) error
}

type HTTPWalletClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPWalletClient(baseURL string) *HTTPWalletClient {
	return &HTTPWalletClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

type confirmDepositRequest struct {
	TxHash        string `json:"tx_hash"`
	ToAddress     string `json:"to_address"`
	Amount        int64  `json:"amount"`
	Confirmations int    `json:"confirmations"`
}

func (c *HTTPWalletClient) ConfirmDeposit(ctx context.Context, txHash, toAddress string, amount int64, confirmations int) error {
	body, err := json.Marshal(confirmDepositRequest{
		TxHash: txHash, ToAddress: toAddress, Amount: amount, Confirmations: confirmations,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/wallet/deposit/confirm", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("wallet request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errBody struct {
			Message string `json:"message"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		if errBody.Message == "" {
			errBody.Message = resp.Status
		}
		return fmt.Errorf("wallet confirm failed: %s", errBody.Message)
	}
	return nil
}

// InMemoryWalletClient records deposit confirmations for tests.
type InMemoryWalletClient struct {
	Confirmed []confirmDepositRequest
}

func NewInMemoryWalletClient() *InMemoryWalletClient {
	return &InMemoryWalletClient{}
}

func (c *InMemoryWalletClient) ConfirmDeposit(_ context.Context, txHash, toAddress string, amount int64, confirmations int) error {
	c.Confirmed = append(c.Confirmed, confirmDepositRequest{
		TxHash: txHash, ToAddress: toAddress, Amount: amount, Confirmations: confirmations,
	})
	return nil
}
