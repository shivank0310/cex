package api

import (
	"time"

	"github.com/shivank0310/cex.git/matching-engine/internal/order"
	"github.com/shivank0310/cex.git/matching-engine/internal/trade"
)

type Side = order.Side
type OrderType = order.OrderType
type Status = order.Status

const (
	Buy  = order.Buy
	Sell = order.Sell
)

const (
	Limit  = order.Limit
	Market = order.Market
)

const (
	StatusNew             = order.StatusNew
	StatusPartiallyFilled = order.StatusPartiallyFilled
	StatusFilled          = order.StatusFilled
	StatusCancelled       = order.StatusCancelled
)

// Order is the public order view exposed to other services.
type Order struct {
	ID        string
	UserID    string
	Symbol    string
	Side      Side
	Type      OrderType
	Status    Status
	Price     int64
	Quantity  int64
	Remaining int64
	Filled    int64
	Sequence  uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Trade is the public trade view exposed to other services.
type Trade struct {
	ID           string
	Symbol       string
	Sequence     uint64
	BuyOrderID   string
	SellOrderID  string
	BuyerID      string
	SellerID     string
	Price        int64
	Quantity     int64
	MakerOrderID string
	TakerOrderID string
	MakerFee     int64
	TakerFee     int64
	Timestamp    time.Time
}

func FromOrder(o *order.Order) *Order {
	if o == nil {
		return nil
	}
	return &Order{
		ID:        o.ID,
		UserID:    o.UserID,
		Symbol:    o.Symbol,
		Side:      o.Side,
		Type:      o.Type,
		Status:    o.Status,
		Price:     o.Price,
		Quantity:  o.Quantity,
		Remaining: o.Remaining,
		Filled:    o.Filled,
		Sequence:  o.Sequence,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func ToOrder(o *Order) *order.Order {
	if o == nil {
		return nil
	}
	return &order.Order{
		ID:        o.ID,
		UserID:    o.UserID,
		Symbol:    o.Symbol,
		Side:      o.Side,
		Type:      o.Type,
		Status:    o.Status,
		Price:     o.Price,
		Quantity:  o.Quantity,
		Remaining: o.Remaining,
		Filled:    o.Filled,
		Sequence:  o.Sequence,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func FromTrade(t *trade.Trade) *Trade {
	if t == nil {
		return nil
	}
	return &Trade{
		ID:           t.ID,
		Symbol:       t.Symbol,
		Sequence:     t.Sequence,
		BuyOrderID:   t.BuyOrderID,
		SellOrderID:  t.SellOrderID,
		BuyerID:      t.BuyerID,
		SellerID:     t.SellerID,
		Price:        t.Price,
		Quantity:     t.Quantity,
		MakerOrderID: t.MakerOrderID,
		TakerOrderID: t.TakerOrderID,
		MakerFee:     t.MakerFee,
		TakerFee:     t.TakerFee,
		Timestamp:    t.Timestamp,
	}
}

func NewLimitOrder(id, userID, symbol string, side Side, price, quantity int64) *Order {
	return FromOrder(order.NewLimit(id, userID, symbol, side, price, quantity, 0))
}

func NewMarketOrder(id, userID, symbol string, side Side, quantity int64) *Order {
	return FromOrder(order.NewMarket(id, userID, symbol, side, quantity, 0))
}
