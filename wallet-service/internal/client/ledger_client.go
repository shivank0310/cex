package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shivank0310/cex.git/wallet-service/internal/model"
)

// LedgerClient talks to ledger-service for balance operations.
type LedgerClient interface {
	GetBalance(ctx context.Context, userID, asset string) (model.LedgerBalance, error)
	GetAccount(ctx context.Context, userID string) (AccountResponse, error)
	Deposit(ctx context.Context, userID, asset string, amount int64, ref string) (model.LedgerBalance, error)
	Reserve(ctx context.Context, userID, asset string, amount int64) (model.LedgerBalance, error)
	Release(ctx context.Context, userID, asset string, amount int64) (model.LedgerBalance, error)
	DebitLocked(ctx context.Context, userID, asset string, amount int64) (model.LedgerBalance, error)
}

type AccountResponse struct {
	CanTrade    bool           `json:"canTrade"`
	CanWithdraw bool           `json:"canWithdraw"`
	CanDeposit  bool           `json:"canDeposit"`
	UpdateTime  int64          `json:"updateTime"`
	AccountType string         `json:"accountType"`
	Balances    []AssetBalance `json:"balances"`
	Permissions []string       `json:"permissions"`
}

type AssetBalance struct {
	Asset  string `json:"asset"`
	Free   int64  `json:"free"`
	Locked int64  `json:"locked"`
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

func (c *HTTPLedgerClient) GetAccount(ctx context.Context, userID string) (AccountResponse, error) {
	url := fmt.Sprintf("%s/api/v3/account?user_id=%s", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return AccountResponse{}, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return AccountResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return AccountResponse{}, fmt.Errorf("ledger error: status %d", resp.StatusCode)
	}

	var account AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return AccountResponse{}, err
	}
	return account, nil
}

func (c *HTTPLedgerClient) GetBalance(ctx context.Context, userID, asset string) (model.LedgerBalance, error) {
	url := fmt.Sprintf("%s/api/v1/ledger/balance/%s/%s", c.baseURL, userID, asset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.LedgerBalance{}, err
	}
	return c.doBalance(req)
}

func (c *HTTPLedgerClient) Deposit(ctx context.Context, userID, asset string, amount int64, ref string) (model.LedgerBalance, error) {
	return c.postMutation(ctx, "/api/v1/ledger/deposit", mutationRequest{
		UserID: userID, Asset: asset, Amount: amount, Ref: ref,
	})
}

func (c *HTTPLedgerClient) Reserve(ctx context.Context, userID, asset string, amount int64) (model.LedgerBalance, error) {
	return c.postMutation(ctx, "/api/v1/ledger/reserve", mutationRequest{
		UserID: userID, Asset: asset, Amount: amount,
	})
}

func (c *HTTPLedgerClient) Release(ctx context.Context, userID, asset string, amount int64) (model.LedgerBalance, error) {
	return c.postMutation(ctx, "/api/v1/ledger/release", mutationRequest{
		UserID: userID, Asset: asset, Amount: amount,
	})
}

func (c *HTTPLedgerClient) DebitLocked(ctx context.Context, userID, asset string, amount int64) (model.LedgerBalance, error) {
	return c.postMutation(ctx, "/api/v1/ledger/debit-locked", mutationRequest{
		UserID: userID, Asset: asset, Amount: amount,
	})
}

func (c *HTTPLedgerClient) postMutation(ctx context.Context, path string, body mutationRequest) (model.LedgerBalance, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return model.LedgerBalance{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return model.LedgerBalance{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doBalance(req)
}

func (c *HTTPLedgerClient) doBalance(req *http.Request) (model.LedgerBalance, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.LedgerBalance{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return model.LedgerBalance{}, fmt.Errorf("ledger error: status %d", resp.StatusCode)
	}

	var br balanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&br); err != nil {
		return model.LedgerBalance{}, err
	}
	return model.LedgerBalance{
		UserID: br.UserID, Asset: br.Asset,
		Available: br.Available, Locked: br.Locked,
	}, nil
}

// InMemoryLedgerClient is used in tests and local monolith wiring.
type InMemoryLedgerClient struct {
	balances map[string]map[string]*model.LedgerBalance
}

func NewInMemoryLedgerClient() *InMemoryLedgerClient {
	return &InMemoryLedgerClient{balances: make(map[string]map[string]*model.LedgerBalance)}
}

func (c *InMemoryLedgerClient) key(userID, asset string) *model.LedgerBalance {
	if c.balances[userID] == nil {
		c.balances[userID] = make(map[string]*model.LedgerBalance)
	}
	if c.balances[userID][asset] == nil {
		c.balances[userID][asset] = &model.LedgerBalance{UserID: userID, Asset: asset}
	}
	return c.balances[userID][asset]
}

func (c *InMemoryLedgerClient) GetBalance(_ context.Context, userID, asset string) (model.LedgerBalance, error) {
	b := c.key(userID, asset)
	return *b, nil
}

func (c *InMemoryLedgerClient) GetAccount(_ context.Context, userID string) (AccountResponse, error) {
	balances := make([]AssetBalance, 0)
	if assets, ok := c.balances[userID]; ok {
		for asset, bal := range assets {
			if bal.Available == 0 && bal.Locked == 0 {
				continue
			}
			balances = append(balances, AssetBalance{
				Asset: asset, Free: bal.Available, Locked: bal.Locked,
			})
		}
	}
	return AccountResponse{
		CanTrade: true, CanWithdraw: true, CanDeposit: true,
		AccountType: "SPOT", Balances: balances, Permissions: []string{"SPOT"},
	}, nil
}

func (c *InMemoryLedgerClient) Deposit(_ context.Context, userID, asset string, amount int64, _ string) (model.LedgerBalance, error) {
	b := c.key(userID, asset)
	b.Available += amount
	return *b, nil
}

func (c *InMemoryLedgerClient) Reserve(_ context.Context, userID, asset string, amount int64) (model.LedgerBalance, error) {
	b := c.key(userID, asset)
	if b.Available < amount {
		return *b, fmt.Errorf("insufficient available")
	}
	b.Available -= amount
	b.Locked += amount
	return *b, nil
}

func (c *InMemoryLedgerClient) Release(_ context.Context, userID, asset string, amount int64) (model.LedgerBalance, error) {
	b := c.key(userID, asset)
	if b.Locked < amount {
		return *b, fmt.Errorf("insufficient locked")
	}
	b.Locked -= amount
	b.Available += amount
	return *b, nil
}

func (c *InMemoryLedgerClient) DebitLocked(_ context.Context, userID, asset string, amount int64) (model.LedgerBalance, error) {
	b := c.key(userID, asset)
	if b.Locked < amount {
		return *b, fmt.Errorf("insufficient locked")
	}
	b.Locked -= amount
	return *b, nil
}
