package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LedgerClient talks to ledger-service for Binance-style fund holds (free/locked).
type LedgerClient interface {
	GetBalance(ctx context.Context, userID, asset string) (Balance, error)
	GetBalances(ctx context.Context, userID string) ([]Balance, error)
	Deposit(ctx context.Context, userID, asset string, amount int64, ref string) (Balance, error)
	Reserve(ctx context.Context, userID, asset string, amount int64) error
	Release(ctx context.Context, userID, asset string, amount int64) error
}

// Balance mirrors Binance free/locked fields using fixed-point int64 amounts.
type Balance struct {
	UserID    string
	Asset     string
	Available int64 // free
	Locked    int64
}

func (b Balance) Total() int64 {
	return b.Available + b.Locked
}

type HTTPLedgerClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPLedgerClient(baseURL string) *HTTPLedgerClient {
	return &HTTPLedgerClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type balanceResponse struct {
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Available int64  `json:"available"`
	Locked    int64  `json:"locked"`
	Total     int64  `json:"total"`
}

type mutationRequest struct {
	UserID string `json:"user_id"`
	Asset  string `json:"asset"`
	Amount int64  `json:"amount"`
	Ref    string `json:"ref,omitempty"`
}

func (c *HTTPLedgerClient) GetBalance(ctx context.Context, userID, asset string) (Balance, error) {
	url := fmt.Sprintf("%s/api/v1/ledger/balance/%s/%s", c.baseURL, userID, asset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Balance{}, err
	}
	return c.doBalance(req)
}

func (c *HTTPLedgerClient) GetBalances(ctx context.Context, userID string) ([]Balance, error) {
	url := fmt.Sprintf("%s/api/v1/ledger/balances/%s", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ledger error: status %d", resp.StatusCode)
	}

	var rows []balanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}

	out := make([]Balance, len(rows))
	for i, row := range rows {
		out[i] = toBalance(row)
	}
	return out, nil
}

func (c *HTTPLedgerClient) Deposit(ctx context.Context, userID, asset string, amount int64, ref string) (Balance, error) {
	return c.postMutation(ctx, "/api/v1/ledger/deposit", mutationRequest{
		UserID: userID, Asset: asset, Amount: amount, Ref: ref,
	})
}

func (c *HTTPLedgerClient) Reserve(ctx context.Context, userID, asset string, amount int64) error {
	_, err := c.postMutation(ctx, "/api/v1/ledger/reserve", mutationRequest{
		UserID: userID, Asset: asset, Amount: amount,
	})
	return err
}

func (c *HTTPLedgerClient) Release(ctx context.Context, userID, asset string, amount int64) error {
	_, err := c.postMutation(ctx, "/api/v1/ledger/release", mutationRequest{
		UserID: userID, Asset: asset, Amount: amount,
	})
	return err
}

func (c *HTTPLedgerClient) postMutation(ctx context.Context, path string, body mutationRequest) (Balance, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return Balance{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return Balance{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doBalance(req)
}

func (c *HTTPLedgerClient) doBalance(req *http.Request) (Balance, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Balance{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return Balance{}, fmt.Errorf("ledger error: status %d", resp.StatusCode)
	}

	var br balanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&br); err != nil {
		return Balance{}, err
	}
	return toBalance(br), nil
}

func toBalance(br balanceResponse) Balance {
	return Balance{
		UserID: br.UserID, Asset: br.Asset,
		Available: br.Available, Locked: br.Locked,
	}
}
