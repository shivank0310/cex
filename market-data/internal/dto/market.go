package dto

import (
	"time"

	"github.com/shivank0310/cex.git/market-data/internal/model"
)

type TickerResponse struct {
	Symbol            string `json:"symbol"`
	LastPrice         int64  `json:"last_price"`
	BestBid           int64  `json:"best_bid"`
	BestAsk           int64  `json:"best_ask"`
	High24h           int64  `json:"high_24h"`
	Low24h            int64  `json:"low_24h"`
	Volume24h         int64  `json:"volume_24h"`
	QuoteVolume24h    int64  `json:"quote_volume_24h"`
	PriceChange24h    int64  `json:"price_change_24h"`
	PriceChangePct24h int64  `json:"price_change_pct_24h"`
	TradeCount24h     int    `json:"trade_count_24h"`
	UpdatedAt         string `json:"updated_at"`
}

type OrderBookResponse struct {
	Symbol    string             `json:"symbol"`
	Bids      []DepthLevelResponse `json:"bids"`
	Asks      []DepthLevelResponse `json:"asks"`
	UpdatedAt string             `json:"updated_at"`
}

type DepthLevelResponse struct {
	Price    int64 `json:"price"`
	Quantity int64 `json:"quantity"`
}

type TradeResponse struct {
	ID        string `json:"id"`
	Symbol    string `json:"symbol"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Notional  int64  `json:"notional"`
	Timestamp string `json:"timestamp"`
}

type CandleResponse struct {
	Symbol     string `json:"symbol"`
	Interval   string `json:"interval"`
	OpenTime   string `json:"open_time"`
	CloseTime  string `json:"close_time"`
	Open       int64  `json:"open"`
	High       int64  `json:"high"`
	Low        int64  `json:"low"`
	Close      int64  `json:"close"`
	Volume     int64  `json:"volume"`
	QuoteVol   int64  `json:"quote_volume"`
	TradeCount int    `json:"trade_count"`
}

type Stats24hResponse struct {
	Symbol      string `json:"symbol"`
	High        int64  `json:"high_24h"`
	Low         int64  `json:"low_24h"`
	Volume      int64  `json:"volume_24h"`
	QuoteVolume int64  `json:"quote_volume_24h"`
	TradeCount  int    `json:"trade_count_24h"`
	OpenPrice   int64  `json:"open_price_24h"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ToTickerResponse(t model.Ticker) TickerResponse {
	return TickerResponse{
		Symbol:            t.Symbol,
		LastPrice:         t.LastPrice,
		BestBid:           t.BestBid,
		BestAsk:           t.BestAsk,
		High24h:           t.High24h,
		Low24h:            t.Low24h,
		Volume24h:         t.Volume24h,
		QuoteVolume24h:    t.QuoteVolume24h,
		PriceChange24h:    t.PriceChange24h,
		PriceChangePct24h: t.PriceChangePct24h,
		TradeCount24h:     t.TradeCount24h,
		UpdatedAt:         t.UpdatedAt.Format(time.RFC3339),
	}
}

func ToOrderBookResponse(b model.OrderBook) OrderBookResponse {
	resp := OrderBookResponse{
		Symbol:    b.Symbol,
		UpdatedAt: b.UpdatedAt.Format(time.RFC3339),
	}
	for _, bid := range b.Bids {
		resp.Bids = append(resp.Bids, DepthLevelResponse{Price: bid.Price, Quantity: bid.Quantity})
	}
	for _, ask := range b.Asks {
		resp.Asks = append(resp.Asks, DepthLevelResponse{Price: ask.Price, Quantity: ask.Quantity})
	}
	return resp
}

func ToTradeResponses(trades []model.Trade) []TradeResponse {
	out := make([]TradeResponse, len(trades))
	for i, t := range trades {
		out[i] = TradeResponse{
			ID:        t.ID,
			Symbol:    t.Symbol,
			Price:     t.Price,
			Quantity:  t.Quantity,
			Notional:  t.Notional,
			Timestamp: t.Timestamp.Format(time.RFC3339),
		}
	}
	return out
}

func ToCandleResponses(candles []model.Candle) []CandleResponse {
	out := make([]CandleResponse, len(candles))
	for i, c := range candles {
		out[i] = CandleResponse{
			Symbol:     c.Symbol,
			Interval:   c.Interval,
			OpenTime:   c.OpenTime.Format(time.RFC3339),
			CloseTime:  c.CloseTime.Format(time.RFC3339),
			Open:       c.Open,
			High:       c.High,
			Low:        c.Low,
			Close:      c.Close,
			Volume:     c.Volume,
			QuoteVol:   c.QuoteVol,
			TradeCount: c.TradeCount,
		}
	}
	return out
}

func ToStats24hResponse(symbol string, s model.Stats24h) Stats24hResponse {
	return Stats24hResponse{
		Symbol:      symbol,
		High:        s.High,
		Low:         s.Low,
		Volume:      s.Volume,
		QuoteVolume: s.QuoteVolume,
		TradeCount:  s.TradeCount,
		OpenPrice:   s.OpenPrice,
	}
}
