package order

import "time"

type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

type OrderType string

const (
	Limit  OrderType = "LIMIT"
	Market OrderType = "MARKET"
)

type Status string

const (
	StatusNew             Status = "NEW"
	StatusPartiallyFilled Status = "PARTIALLY_FILLED"
	StatusFilled          Status = "FILLED"
	StatusCancelled       Status = "CANCELLED"
)

type Order struct {
	ID        string
	UserID    string
	Symbol    string
	Side      Side
	Type      OrderType
	Status    Status

	Price    int64 // quote per 1 base unit; 0 for market orders
	Quantity int64
	Remaining int64
	Filled   int64

	Sequence  uint64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewLimit(id, userID, symbol string, side Side, price, quantity int64, seq uint64) *Order {
	now := time.Now().UTC()
	return &Order{
		ID:        id,
		UserID:    userID,
		Symbol:    symbol,
		Side:      side,
		Type:      Limit,
		Status:    StatusNew,
		Price:     price,
		Quantity:  quantity,
		Remaining: quantity,
		Sequence:  seq,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewMarket(id, userID, symbol string, side Side, quantity int64, seq uint64) *Order {
	now := time.Now().UTC()
	return &Order{
		ID:        id,
		UserID:    userID,
		Symbol:    symbol,
		Side:      side,
		Type:      Market,
		Status:    StatusNew,
		Price:     0,
		Quantity:  quantity,
		Remaining: quantity,
		Sequence:  seq,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (o *Order) Fill(qty int64) {
	o.Filled += qty
	o.Remaining -= qty
	o.UpdatedAt = time.Now().UTC()
	switch {
	case o.Remaining == 0:
		o.Status = StatusFilled
	case o.Filled > 0:
		o.Status = StatusPartiallyFilled
	}
}

func (o *Order) Cancel() {
	o.Status = StatusCancelled
	o.UpdatedAt = time.Now().UTC()
}

func (o *Order) IsActive() bool {
	return o.Status == StatusNew || o.Status == StatusPartiallyFilled
}

// CanMatch returns whether this order can trade against a resting order at matchPrice.
func (o *Order) CanMatch(matchPrice int64) bool {
	if o.Type == Market {
		return true
	}
	if o.Side == Buy {
		return matchPrice <= o.Price
	}
	return matchPrice >= o.Price
}
