package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/mapping"
)

func (c *RESTClient) GetTicker24h(ctx context.Context, symbol string) (Ticker24h, error) {
	params := url.Values{}
	params.Set("symbol", mapping.ToBinanceSymbol(symbol))

	body, err := c.doPublic(ctx, http.MethodGet, "/api/v3/ticker/24hr", params)
	if err != nil {
		return Ticker24h{}, err
	}

	var resp ticker24hResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return Ticker24h{}, err
	}

	last, _ := mapping.PriceFromBinance(resp.LastPrice)
	bid, _ := mapping.PriceFromBinance(resp.BidPrice)
	ask, _ := mapping.PriceFromBinance(resp.AskPrice)
	high, _ := mapping.PriceFromBinance(resp.HighPrice)
	low, _ := mapping.PriceFromBinance(resp.LowPrice)
	change, _ := mapping.PriceFromBinance(resp.PriceChange)
	vol, _ := mapping.QuantityFromBinance(resp.Volume)
	quoteVol, _ := mapping.QuantityFromBinance(resp.QuoteVolume)
	pct, _ := strconv.ParseFloat(resp.PriceChangePercent, 64)

	return Ticker24h{
		Symbol:            symbol,
		LastPrice:         last,
		BestBid:           bid,
		BestAsk:           ask,
		High24h:           high,
		Low24h:            low,
		Volume24h:         vol,
		QuoteVolume24h:    quoteVol,
		PriceChange24h:    change,
		PriceChangePct24h: int64(pct * 100),
		TradeCount24h:     resp.Count,
	}, nil
}

func (c *RESTClient) GetDepth(ctx context.Context, symbol string, limit int) (DepthSnapshot, error) {
	if limit <= 0 {
		limit = 20
	}
	params := url.Values{}
	params.Set("symbol", mapping.ToBinanceSymbol(symbol))
	params.Set("limit", strconv.Itoa(limit))

	body, err := c.doPublic(ctx, http.MethodGet, "/api/v3/depth", params)
	if err != nil {
		return DepthSnapshot{}, err
	}

	var resp depthResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return DepthSnapshot{}, err
	}

	snap := DepthSnapshot{Symbol: symbol}
	for _, row := range resp.Bids {
		if len(row) < 2 {
			continue
		}
		price, _ := mapping.PriceFromBinance(row[0])
		qty, _ := mapping.QuantityFromBinance(row[1])
		snap.Bids = append(snap.Bids, DepthLevel{Price: price, Quantity: qty})
	}
	for _, row := range resp.Asks {
		if len(row) < 2 {
			continue
		}
		price, _ := mapping.PriceFromBinance(row[0])
		qty, _ := mapping.QuantityFromBinance(row[1])
		snap.Asks = append(snap.Asks, DepthLevel{Price: price, Quantity: qty})
	}
	return snap, nil
}

func (c *RESTClient) GetRecentTrades(ctx context.Context, symbol string, limit int) ([]MarketTrade, error) {
	if limit <= 0 {
		limit = 30
	}
	params := url.Values{}
	params.Set("symbol", mapping.ToBinanceSymbol(symbol))
	params.Set("limit", strconv.Itoa(limit))

	body, err := c.doPublic(ctx, http.MethodGet, "/api/v3/trades", params)
	if err != nil {
		return nil, err
	}

	var resp []tradeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	out := make([]MarketTrade, 0, len(resp))
	for _, t := range resp {
		price, _ := mapping.PriceFromBinance(t.Price)
		qty, _ := mapping.QuantityFromBinance(t.Qty)
		out = append(out, MarketTrade{
			ID:        strconv.FormatInt(t.ID, 10),
			Symbol:    symbol,
			Price:     price,
			Quantity:  qty,
			Timestamp: time.UnixMilli(t.Time),
		})
	}
	return out, nil
}

func (c *RESTClient) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	if limit <= 0 {
		limit = 60
	}
	plan := mapping.BinanceFetchPlanFor(interval)
	fetchLimit := limit
	if plan.MergeCount > 1 {
		fetchLimit = limit * plan.MergeCount
	}

	params := url.Values{}
	params.Set("symbol", mapping.ToBinanceSymbol(symbol))
	params.Set("interval", plan.FetchInterval)
	params.Set("limit", strconv.Itoa(fetchLimit))

	body, err := c.doPublic(ctx, http.MethodGet, "/api/v3/klines", params)
	if err != nil {
		return nil, err
	}

	var raw [][]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	out := make([]Kline, 0, len(raw))
	for _, row := range raw {
		if len(row) < 11 {
			continue
		}
		openTime := time.UnixMilli(toInt64(row[0]))
		open, _ := mapping.PriceFromBinance(fmt.Sprint(row[1]))
		high, _ := mapping.PriceFromBinance(fmt.Sprint(row[2]))
		low, _ := mapping.PriceFromBinance(fmt.Sprint(row[3]))
		close, _ := mapping.PriceFromBinance(fmt.Sprint(row[4]))
		vol, _ := mapping.QuantityFromBinance(fmt.Sprint(row[5]))
		closeTime := time.UnixMilli(toInt64(row[6]))
		quoteVol, _ := mapping.QuantityFromBinance(fmt.Sprint(row[7]))
		tradeCount := int(toInt64(row[8]))

		out = append(out, Kline{
			OpenTime: openTime, CloseTime: closeTime,
			Open: open, High: high, Low: low, Close: close,
			Volume: vol, QuoteVolume: quoteVol, TradeCount: tradeCount,
		})
	}

	if plan.MergeCount > 1 {
		out = mergeKlines(out, plan.MergeCount)
		if len(out) > limit {
			out = out[len(out)-limit:]
		}
	}
	return out, nil
}

func mergeKlines(klines []Kline, n int) []Kline {
	if n <= 1 || len(klines) == 0 {
		return klines
	}
	out := make([]Kline, 0, len(klines)/n)
	for i := 0; i+n <= len(klines); i += n {
		chunk := klines[i : i+n]
		merged := Kline{
			OpenTime:  chunk[0].OpenTime,
			CloseTime: chunk[len(chunk)-1].CloseTime,
			Open:      chunk[0].Open,
			High:      chunk[0].High,
			Low:       chunk[0].Low,
			Close:     chunk[len(chunk)-1].Close,
		}
		for _, k := range chunk {
			if k.High > merged.High {
				merged.High = k.High
			}
			if k.Low < merged.Low {
				merged.Low = k.Low
			}
			merged.Volume += k.Volume
			merged.QuoteVolume += k.QuoteVolume
			merged.TradeCount += k.TradeCount
		}
		out = append(out, merged)
	}
	return out
}

type ticker24hResponse struct {
	LastPrice          string `json:"lastPrice"`
	BidPrice           string `json:"bidPrice"`
	AskPrice           string `json:"askPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	OpenPrice          string `json:"openPrice"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	Count              int    `json:"count"`
}

type depthResponse struct {
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
}

type tradeResponse struct {
	ID    int64  `json:"id"`
	Price string `json:"price"`
	Qty   string `json:"qty"`
	Time  int64  `json:"time"`
}

func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}
