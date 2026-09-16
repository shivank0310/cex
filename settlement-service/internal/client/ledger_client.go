package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shivank0310/cex.git/pkg/events"
)

// LedgerClient posts trade settlement to ledger-service.
type LedgerClient interface {
	SettleTrade(ctx context.Context, trade events.TradePayload) (string, error)
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

type settleTradeResponse struct {
	JournalID string `json:"journal_id"`
	TradeID   string `json:"trade_id"`
}

func (c *HTTPLedgerClient) SettleTrade(ctx context.Context, trade events.TradePayload) (string, error) {
	body, err := json.Marshal(trade)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/ledger/settle-trade", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ledger request failed: %w", err)
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
		return "", fmt.Errorf("ledger settle failed: %s", errBody.Message)
	}

	var out settleTradeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.JournalID, nil
}
