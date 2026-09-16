package model

import "time"

// Trade is a normalized market trade.
type Trade struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Price     int64     `json:"price"`
	Quantity  int64     `json:"quantity"`
	Notional  int64     `json:"notional"`
	Timestamp time.Time `json:"timestamp"`
}

// DepthLevel is a single order book price level.
type DepthLevel struct {
	Price    int64 `json:"price"`
	Quantity int64 `json:"quantity"`
}

// OrderBook is the current depth snapshot for a symbol.
type OrderBook struct {
	Symbol    string       `json:"symbol"`
	Bids      []DepthLevel `json:"bids"`
	Asks      []DepthLevel `json:"asks"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// Ticker is a 24h rolling market summary.
type Ticker struct {
	Symbol             string    `json:"symbol"`
	LastPrice          int64     `json:"last_price"`
	BestBid            int64     `json:"best_bid"`
	BestAsk            int64     `json:"best_ask"`
	High24h            int64     `json:"high_24h"`
	Low24h             int64     `json:"low_24h"`
	Volume24h          int64     `json:"volume_24h"`
	QuoteVolume24h     int64     `json:"quote_volume_24h"`
	PriceChange24h     int64     `json:"price_change_24h"`
	PriceChangePct24h  int64     `json:"price_change_pct_24h"` // basis points * 100 for precision (e.g. 150 = 1.50%)
	TradeCount24h      int       `json:"trade_count_24h"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Stats24h holds rolling 24-hour statistics.
type Stats24h struct {
	High        int64 `json:"high"`
	Low         int64 `json:"low"`
	Volume      int64 `json:"volume"`
	QuoteVolume int64 `json:"quote_volume"`
	TradeCount  int   `json:"trade_count"`
	OpenPrice   int64 `json:"open_price"` // first trade price in window
}

// Candle is an OHLCV bar for a given interval.
type Candle struct {
	Symbol    string    `json:"symbol"`
	Interval  string    `json:"interval"`
	OpenTime  time.Time `json:"open_time"`
	CloseTime time.Time `json:"close_time"`
	Open      int64     `json:"open"`
	High      int64     `json:"high"`
	Low       int64     `json:"low"`
	Close     int64     `json:"close"`
	Volume    int64     `json:"volume"`
	QuoteVol  int64     `json:"quote_volume"`
	TradeCount int      `json:"trade_count"`
}
