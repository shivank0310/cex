package binance

import "time"

// Ticker24h is a normalized 24h market summary.
type Ticker24h struct {
	Symbol            string
	LastPrice         int64
	BestBid           int64
	BestAsk           int64
	High24h           int64
	Low24h            int64
	Volume24h         int64
	QuoteVolume24h    int64
	PriceChange24h    int64
	PriceChangePct24h int64 // hundredths of a percent (150 = 1.50%)
	TradeCount24h     int
}

// DepthLevel is a single order book level.
type DepthLevel struct {
	Price    int64
	Quantity int64
}

// DepthSnapshot is an order book depth snapshot.
type DepthSnapshot struct {
	Symbol string
	Bids   []DepthLevel
	Asks   []DepthLevel
}

// MarketTrade is a public market trade.
type MarketTrade struct {
	ID        string
	Symbol    string
	Price     int64
	Quantity  int64
	Timestamp time.Time
}

// Kline is an OHLCV candle.
type Kline struct {
	OpenTime    time.Time
	CloseTime   time.Time
	Open        int64
	High        int64
	Low         int64
	Close       int64
	Volume      int64
	QuoteVolume int64
	TradeCount  int
}
