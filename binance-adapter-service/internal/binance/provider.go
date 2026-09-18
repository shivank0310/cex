package binance

import (
	"context"
	"time"
)

// OrderRequest is the normalized input for placing an order on Binance Spot.
type OrderRequest struct {
	ClientOrderID string
	Symbol        string // internal format: BTC/USDT
	Side          string // BUY or SELL
	Type          string // LIMIT or MARKET
	Price         int64
	Quantity      int64
}

// OrderResult is the normalized Binance order response.
type OrderResult struct {
	ClientOrderID   string
	BinanceOrderID  int64
	Symbol          string
	Side            string
	Type            string
	Status          string
	Price           int64
	Quantity        int64
	ExecutedQty     int64
	RemainingQty    int64
	CummulativeQuote int64
	TransactTime    time.Time
}

// AccountBalance mirrors Binance free/locked per asset.
type AccountBalance struct {
	Asset  string
	Free   int64
	Locked int64
}

// TickerPrice is the latest price for a symbol.
type TickerPrice struct {
	Symbol string
	Price  int64
}

// SpotProvider abstracts Binance Spot REST API operations.
type SpotProvider interface {
	PlaceOrder(ctx context.Context, req OrderRequest) (OrderResult, error)
	CancelOrder(ctx context.Context, symbol, clientOrderID string) (OrderResult, error)
	GetOrder(ctx context.Context, symbol, clientOrderID string) (OrderResult, error)
	GetAccount(ctx context.Context) ([]AccountBalance, error)
	GetTickerPrice(ctx context.Context, symbol string) (TickerPrice, error)
	Ping(ctx context.Context) error
}
