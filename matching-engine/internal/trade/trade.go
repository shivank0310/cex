package trade

import "time"

type Trade struct {
	ID       string
	Symbol   string
	Sequence uint64

	BuyOrderID  string
	SellOrderID string
	BuyerID     string
	SellerID    string

	Price         int64
	Quantity      int64
	BuyLimitPrice int64 // buyer's limit price for settlement (0 = market buy)

	// Maker is the resting order; taker is the incoming aggressor.
	MakerOrderID string
	TakerOrderID string
	MakerFee     int64 // in quote currency for buyer-maker, base for seller-maker
	TakerFee     int64

	Timestamp time.Time
}
