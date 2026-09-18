package dto

import (
	"time"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/binance"
)

type PlaceOrderRequest struct {
	ClientOrderID string `json:"client_order_id"`
	UserID        string `json:"user_id"`
	Symbol        string `json:"symbol"`
	Side          string `json:"side"`
	Type          string `json:"type"`
	Price         int64  `json:"price"`
	Quantity      int64  `json:"quantity"`
}

type CancelOrderRequest struct {
	Symbol  string `json:"symbol"`
	OrderID string `json:"order_id"`
}

type OrderResponse struct {
	ID              string `json:"id"`
	UserID          string `json:"user_id"`
	BinanceOrderID  int64  `json:"binance_order_id"`
	Symbol          string `json:"symbol"`
	Side            string `json:"side"`
	Type            string `json:"type"`
	Status          string `json:"status"`
	Price           int64  `json:"price"`
	Quantity        int64  `json:"quantity"`
	Remaining       int64  `json:"remaining"`
	Filled          int64  `json:"filled"`
	Venue           string `json:"venue"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type TradeResponse struct {
	ID       string `json:"id"`
	Symbol   string `json:"symbol"`
	Price    int64  `json:"price"`
	Quantity int64  `json:"quantity"`
	Side     string `json:"side"`
	Venue    string `json:"venue"`
}

type PlaceOrderResponse struct {
	Order  OrderResponse   `json:"order"`
	Trades []TradeResponse `json:"trades"`
}

type AccountResponse struct {
	CanTrade    bool           `json:"canTrade"`
	CanWithdraw bool           `json:"canWithdraw"`
	CanDeposit  bool           `json:"canDeposit"`
	UpdateTime  int64          `json:"updateTime"`
	AccountType string         `json:"accountType"`
	Balances    []AssetBalance `json:"balances"`
	Permissions []string       `json:"permissions"`
	Venue       string         `json:"venue"`
}

type AssetBalance struct {
	Asset  string `json:"asset"`
	Free   int64  `json:"free"`
	Locked int64  `json:"locked"`
}

type TickerResponse struct {
	Symbol string `json:"symbol"`
	Price  int64  `json:"price"`
	Venue  string `json:"venue"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToOrderResponse(userID string, r binance.OrderResult) OrderResponse {
	now := time.Now().UTC().Format(time.RFC3339)
	return OrderResponse{
		ID:             r.ClientOrderID,
		UserID:         userID,
		BinanceOrderID: r.BinanceOrderID,
		Symbol:         r.Symbol,
		Side:           r.Side,
		Type:           r.Type,
		Status:         mapBinanceStatus(r.Status),
		Price:          r.Price,
		Quantity:       r.Quantity,
		Remaining:      r.RemainingQty,
		Filled:         r.ExecutedQty,
		Venue:          "binance",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func mapBinanceStatus(s string) string {
	switch s {
	case "NEW":
		return "NEW"
	case "PARTIALLY_FILLED":
		return "PARTIALLY_FILLED"
	case "FILLED":
		return "FILLED"
	case "CANCELED", "CANCELLED", "EXPIRED", "REJECTED":
		return "CANCELLED"
	default:
		return s
	}
}
