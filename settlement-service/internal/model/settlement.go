package model

import (
	"strings"
	"time"
)

type Status string

const (
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

// Settlement is the post-trade transfer record.
type Settlement struct {
	ID                string
	TradeID           string
	JournalID         string
	Symbol            string
	BaseAsset         string
	QuoteAsset        string
	BuyerID           string
	SellerID          string
	Price             int64
	Quantity          int64
	Notional          int64
	BuyerFee          int64
	SellerFee         int64
	BuyerBaseCredit   int64
	SellerQuoteCredit int64
	Status            Status
	SettledAt         time.Time
}

type Pair struct {
	Base  string
	Quote string
}

func ParseSymbol(symbol string) (Pair, error) {
	parts := strings.Split(strings.TrimSpace(symbol), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Pair{}, ErrInvalidSymbol
	}
	return Pair{Base: parts[0], Quote: parts[1]}, nil
}

var ErrInvalidSymbol = &symbolError{}

type symbolError struct{}

func (e *symbolError) Error() string {
	return "invalid symbol format, expected BASE/QUOTE"
}
