package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shivank0310/cex.git/market-data/internal/model"
)

// BinanceAdapterClient fetches market data from binance-adapter-service for external symbols.
type BinanceAdapterClient struct {
	baseURL    string
	httpClient *http.Client
	external   map[string]bool
}

func NewBinanceAdapterClient(baseURL string, symbols []string) *BinanceAdapterClient {
	external := make(map[string]bool, len(symbols))
	for _, s := range symbols {
		s = strings.TrimSpace(s)
		if s != "" {
			external[s] = true
		}
	}
	return &BinanceAdapterClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 8 * time.Second},
		external:   external,
	}
}

func (c *BinanceAdapterClient) IsExternal(symbol string) bool {
	return c != nil && c.external[symbol]
}

func (c *BinanceAdapterClient) ExternalSymbols() []string {
	if c == nil {
		return nil
	}
	out := make([]string, 0, len(c.external))
	for symbol := range c.external {
		out = append(out, symbol)
	}
	return out
}

func (c *BinanceAdapterClient) GetTicker(ctx context.Context, symbol string) (model.Ticker, error) {
	var resp tickerDTO
	if err := c.getJSON(ctx, fmt.Sprintf("/api/v1/binance/market/ticker/%s", encodeSymbol(symbol)), &resp); err != nil {
		return model.Ticker{}, err
	}
	updatedAt, _ := time.Parse(time.RFC3339, resp.UpdatedAt)
	return model.Ticker{
		Symbol:            resp.Symbol,
		LastPrice:         resp.LastPrice,
		BestBid:           resp.BestBid,
		BestAsk:           resp.BestAsk,
		High24h:           resp.High24h,
		Low24h:            resp.Low24h,
		Volume24h:         resp.Volume24h,
		QuoteVolume24h:    resp.QuoteVolume24h,
		PriceChange24h:    resp.PriceChange24h,
		PriceChangePct24h: resp.PriceChangePct24h,
		TradeCount24h:     resp.TradeCount24h,
		UpdatedAt:         updatedAt,
	}, nil
}

func (c *BinanceAdapterClient) GetOrderBook(ctx context.Context, symbol string, limit int) (model.OrderBook, error) {
	var resp orderBookDTO
	path := fmt.Sprintf("/api/v1/binance/market/orderbook/%s?limit=%d", encodeSymbol(symbol), limit)
	if err := c.getJSON(ctx, path, &resp); err != nil {
		return model.OrderBook{}, err
	}
	updatedAt, _ := time.Parse(time.RFC3339, resp.UpdatedAt)
	book := model.OrderBook{Symbol: resp.Symbol, UpdatedAt: updatedAt}
	for _, b := range resp.Bids {
		book.Bids = append(book.Bids, model.DepthLevel{Price: b.Price, Quantity: b.Quantity})
	}
	for _, a := range resp.Asks {
		book.Asks = append(book.Asks, model.DepthLevel{Price: a.Price, Quantity: a.Quantity})
	}
	return book, nil
}

func (c *BinanceAdapterClient) GetTrades(ctx context.Context, symbol string, limit int) ([]model.Trade, error) {
	var resp []tradeDTO
	path := fmt.Sprintf("/api/v1/binance/market/trades/%s?limit=%d", encodeSymbol(symbol), limit)
	if err := c.getJSON(ctx, path, &resp); err != nil {
		return nil, err
	}
	out := make([]model.Trade, len(resp))
	for i, t := range resp {
		ts, _ := time.Parse(time.RFC3339, t.Timestamp)
		out[i] = model.Trade{
			ID: t.ID, Symbol: t.Symbol, Price: t.Price,
			Quantity: t.Quantity, Notional: t.Notional, Timestamp: ts,
		}
	}
	return out, nil
}

func (c *BinanceAdapterClient) GetCandles(ctx context.Context, symbol, interval string, limit int) ([]model.Candle, error) {
	var resp []candleDTO
	path := fmt.Sprintf("/api/v1/binance/market/candles/%s?interval=%s&limit=%d",
		encodeSymbol(symbol), interval, limit)
	if err := c.getJSON(ctx, path, &resp); err != nil {
		return nil, err
	}
	out := make([]model.Candle, len(resp))
	for i, cnd := range resp {
		openTime, _ := time.Parse(time.RFC3339, cnd.OpenTime)
		closeTime, _ := time.Parse(time.RFC3339, cnd.CloseTime)
		out[i] = model.Candle{
			Symbol: symbol, Interval: interval,
			OpenTime: openTime, CloseTime: closeTime,
			Open: cnd.Open, High: cnd.High, Low: cnd.Low, Close: cnd.Close,
			Volume: cnd.Volume, QuoteVol: cnd.QuoteVol, TradeCount: cnd.TradeCount,
		}
	}
	return out, nil
}

func (c *BinanceAdapterClient) getJSON(ctx context.Context, path string, dest interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("binance adapter %s: %s", path, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, dest)
}

func encodeSymbol(symbol string) string {
	return strings.ReplaceAll(symbol, "/", "%2F")
}

type tickerDTO struct {
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

type orderBookDTO struct {
	Symbol    string          `json:"symbol"`
	Bids      []depthLevelDTO `json:"bids"`
	Asks      []depthLevelDTO `json:"asks"`
	UpdatedAt string          `json:"updated_at"`
}

type depthLevelDTO struct {
	Price    int64 `json:"price"`
	Quantity int64 `json:"quantity"`
}

type tradeDTO struct {
	ID        string `json:"id"`
	Symbol    string `json:"symbol"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Notional  int64  `json:"notional"`
	Timestamp string `json:"timestamp"`
}

type candleDTO struct {
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
