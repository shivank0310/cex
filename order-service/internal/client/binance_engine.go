package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
)

// BinanceEngineClient routes orders to binance-adapter-service.
type BinanceEngineClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewBinanceEngineClient(baseURL string) *BinanceEngineClient {
	return &BinanceEngineClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type binancePlaceRequest struct {
	ClientOrderID string `json:"client_order_id"`
	UserID        string `json:"user_id"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	Type          string `json:"type"`
	Price         int64  `json:"price"`
	Quantity      int64  `json:"quantity"`
}

type binanceOrderResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Symbol    string `json:"symbol"`
	Side      string `json:"side"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Remaining int64  `json:"remaining"`
	Filled    int64  `json:"filled"`
	Venue     string `json:"venue"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type binanceTradeResponse struct {
	ID       string `json:"id"`
	Symbol   string `json:"symbol"`
	Price    int64  `json:"price"`
	Quantity int64  `json:"quantity"`
	Side     string `json:"side"`
	Venue    string `json:"venue"`
}

type binancePlaceResponse struct {
	Order  binanceOrderResponse   `json:"order"`
	Trades []binanceTradeResponse `json:"trades"`
}

func (c *BinanceEngineClient) Submit(o *meapi.Order) meapi.SubmitResult {
	reqBody := binancePlaceRequest{
		ClientOrderID: o.ID,
		UserID:        o.UserID,
		Symbol:        o.Symbol,
		Side:          string(o.Side),
		Type:          string(o.Type),
		Price:         o.Price,
		Quantity:      o.Quantity,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return meapi.SubmitResult{Order: o, Error: err}
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/v1/binance/orders", "application/json", bytes.NewReader(data))
	if err != nil {
		return meapi.SubmitResult{Order: o, Error: fmt.Errorf("binance adapter: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return meapi.SubmitResult{Order: o, Error: fmt.Errorf("binance adapter: status %d", resp.StatusCode)}
	}

	var result binancePlaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return meapi.SubmitResult{Order: o, Error: err}
	}

	order := toMEOrder(result.Order)
	trades := make([]*meapi.Trade, 0, len(result.Trades))
	for _, tr := range result.Trades {
		trades = append(trades, toMETrade(tr, o))
	}

	return meapi.SubmitResult{Order: order, Trades: trades}
}

func (c *BinanceEngineClient) Cancel(symbol, orderID string) (*meapi.Order, error) {
	url := fmt.Sprintf("%s/api/v1/binance/orders/%s?symbol=%s", c.baseURL, orderID, symbol)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("binance adapter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("binance adapter: status %d", resp.StatusCode)
	}

	var orderResp binanceOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return nil, err
	}
	return toMEOrder(orderResp), nil
}

func toMEOrder(r binanceOrderResponse) *meapi.Order {
	return &meapi.Order{
		ID:        r.ID,
		UserID:    r.UserID,
		Symbol:    r.Symbol,
		Side:      meapi.Side(r.Side),
		Type:      meapi.OrderType(r.Type),
		Status:    meapi.Status(r.Status),
		Price:     r.Price,
		Quantity:  r.Quantity,
		Remaining: r.Remaining,
		Filled:    r.Filled,
	}
}

func toMETrade(tr binanceTradeResponse, o *meapi.Order) *meapi.Trade {
	t := &meapi.Trade{
		ID:       tr.ID,
		Symbol:   tr.Symbol,
		Price:    tr.Price,
		Quantity: tr.Quantity,
	}
	if o.Side == meapi.Buy {
		t.BuyOrderID = o.ID
		t.BuyerID = o.UserID
	} else {
		t.SellOrderID = o.ID
		t.SellerID = o.UserID
	}
	return t
}
