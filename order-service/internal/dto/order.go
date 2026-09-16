package dto

import (
	"time"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
)

// PlaceOrderRequest is the inbound API payload for submitting an order.
type PlaceOrderRequest struct {
	Symbol   string `json:"symbol"`
	Side     string `json:"side"`
	Type     string `json:"type"`
	Price    int64  `json:"price"`
	Quantity int64  `json:"quantity"`
}

// CancelOrderRequest cancels a resting order.
type CancelOrderRequest struct {
	Symbol  string `json:"symbol"`
	OrderID string `json:"order_id"`
}

// OrderResponse is the outbound representation of an order.
type OrderResponse struct {
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
	Sequence  uint64 `json:"sequence"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// TradeResponse is the outbound representation of a trade.
type TradeResponse struct {
	ID           string `json:"id"`
	Symbol       string `json:"symbol"`
	Sequence     uint64 `json:"sequence"`
	BuyOrderID   string `json:"buy_order_id"`
	SellOrderID  string `json:"sell_order_id"`
	Price        int64  `json:"price"`
	Quantity     int64  `json:"quantity"`
	MakerOrderID string `json:"maker_order_id"`
	TakerOrderID string `json:"taker_order_id"`
	MakerFee     int64  `json:"maker_fee"`
	TakerFee     int64  `json:"taker_fee"`
	Timestamp    string `json:"timestamp"`
}

// PlaceOrderResponse wraps order + trades after submission.
type PlaceOrderResponse struct {
	Order  OrderResponse   `json:"order"`
	Trades []TradeResponse `json:"trades"`
}

// ErrorResponse is the standard API error envelope.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToOrderResponse(o *meapi.Order) OrderResponse {
	return OrderResponse{
		ID:        o.ID,
		UserID:    o.UserID,
		Symbol:    o.Symbol,
		Side:      string(o.Side),
		Type:      string(o.Type),
		Status:    string(o.Status),
		Price:     o.Price,
		Quantity:  o.Quantity,
		Remaining: o.Remaining,
		Filled:    o.Filled,
		Sequence:  o.Sequence,
		CreatedAt: o.CreatedAt.Format(time.RFC3339),
		UpdatedAt: o.UpdatedAt.Format(time.RFC3339),
	}
}

func ToTradeResponse(t *meapi.Trade) TradeResponse {
	return TradeResponse{
		ID:           t.ID,
		Symbol:       t.Symbol,
		Sequence:     t.Sequence,
		BuyOrderID:   t.BuyOrderID,
		SellOrderID:  t.SellOrderID,
		Price:        t.Price,
		Quantity:     t.Quantity,
		MakerOrderID: t.MakerOrderID,
		TakerOrderID: t.TakerOrderID,
		MakerFee:     t.MakerFee,
		TakerFee:     t.TakerFee,
		Timestamp:    t.Timestamp.Format(time.RFC3339),
	}
}
