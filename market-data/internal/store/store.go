package store

import (
	"sync"
	"time"

	"github.com/shivank0310/cex.git/market-data/internal/model"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/pkg/events"
)

const (
	maxRecentTrades = 500
	maxCandles      = 1000
	window24h       = 24 * time.Hour
)

// Store holds in-memory market state per symbol.
type Store struct {
	mu      sync.RWMutex
	symbols map[string]*symbolState
}

type symbolState struct {
	orderBook   model.OrderBook
	trades      []model.Trade
	stats24h    []tradePoint
	candles     map[string][]model.Candle // interval -> candles (oldest first)
	ticker      model.Ticker
}

type tradePoint struct {
	price     int64
	quantity  int64
	notional  int64
	timestamp time.Time
}

func New() *Store {
	return &Store{symbols: make(map[string]*symbolState)}
}

func (s *Store) getOrCreate(symbol string) *symbolState {
	if s.symbols[symbol] == nil {
		s.symbols[symbol] = &symbolState{
			candles: make(map[string][]model.Candle),
		}
	}
	return s.symbols[symbol]
}

// ApplyTrade ingests a trade event and updates all derived market data.
func (s *Store) ApplyTrade(trade events.TradePayload) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := s.getOrCreate(trade.Symbol)
	notional := decimal.Notional(trade.Price, trade.Quantity)
	now := trade.Timestamp.UTC()

	t := model.Trade{
		ID:        trade.ID,
		Symbol:    trade.Symbol,
		Price:     trade.Price,
		Quantity:  trade.Quantity,
		Notional:  notional,
		Timestamp: now,
	}

	st.trades = prependTrade(st.trades, t, maxRecentTrades)
	st.stats24h = append(st.stats24h, tradePoint{
		price: trade.Price, quantity: trade.Quantity, notional: notional, timestamp: now,
	})
	st.stats24h = pruneWindow(st.stats24h, now, window24h)

	for intervalName, dur := range model.SupportedCandleIntervals {
		st.candles[intervalName] = updateCandle(st.candles[intervalName], trade.Symbol, intervalName, dur, t, maxCandles)
	}

	st.ticker = buildTicker(st)
}

// ApplyOrderBook ingests an order book snapshot.
func (s *Store) ApplyOrderBook(book events.OrderBookPayload) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := s.getOrCreate(book.Symbol)
	st.orderBook = model.OrderBook{
		Symbol:    book.Symbol,
		Bids:      toDepth(book.Bids),
		Asks:      toDepth(book.Asks),
		UpdatedAt: time.Now().UTC(),
	}
	st.ticker = buildTicker(st)
}

func (s *Store) Ticker(symbol string) (model.Ticker, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.symbols[symbol]
	if !ok {
		return model.Ticker{}, false
	}
	return st.ticker, true
}

func (s *Store) AllTickers() []model.Ticker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]model.Ticker, 0, len(s.symbols))
	for _, st := range s.symbols {
		if st.ticker.Symbol != "" {
			out = append(out, st.ticker)
		}
	}
	return out
}

func (s *Store) OrderBook(symbol string) (model.OrderBook, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.symbols[symbol]
	if !ok {
		return model.OrderBook{}, false
	}
	return st.orderBook, st.orderBook.Symbol != ""
}

func (s *Store) Trades(symbol string, limit int) []model.Trade {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.symbols[symbol]
	if !ok {
		return nil
	}
	if limit <= 0 || limit > len(st.trades) {
		limit = len(st.trades)
	}
	return st.trades[:limit]
}

func (s *Store) Candles(symbol, interval string, limit int) ([]model.Candle, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.symbols[symbol]
	if !ok {
		return nil, false
	}
	candles := st.candles[interval]
	if limit <= 0 || limit > len(candles) {
		limit = len(candles)
	}
	// Return newest first
	out := make([]model.Candle, limit)
	for i := 0; i < limit; i++ {
		out[i] = candles[len(candles)-1-i]
	}
	return out, true
}

func (s *Store) Stats24h(symbol string) (model.Stats24h, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.symbols[symbol]
	if !ok || len(st.stats24h) == 0 {
		return model.Stats24h{}, ok
	}
	return computeStats(st.stats24h), true
}

func prependTrade(trades []model.Trade, t model.Trade, max int) []model.Trade {
	out := make([]model.Trade, 0, min(len(trades)+1, max))
	out = append(out, t)
	for i := 0; i < len(trades) && len(out) < max; i++ {
		out = append(out, trades[i])
	}
	return out
}

func pruneWindow(points []tradePoint, now time.Time, window time.Duration) []tradePoint {
	cutoff := now.Add(-window)
	out := make([]tradePoint, 0, len(points))
	for _, p := range points {
		if !p.timestamp.Before(cutoff) {
			out = append(out, p)
		}
	}
	return out
}

func computeStats(points []tradePoint) model.Stats24h {
	if len(points) == 0 {
		return model.Stats24h{}
	}
	stats := model.Stats24h{
		High:       points[0].price,
		Low:        points[0].price,
		OpenPrice:  points[0].price,
		TradeCount: len(points),
	}
	for _, p := range points {
		if p.price > stats.High {
			stats.High = p.price
		}
		if p.price < stats.Low {
			stats.Low = p.price
		}
		stats.Volume += p.quantity
		stats.QuoteVolume += p.notional
	}
	return stats
}

func buildTicker(st *symbolState) model.Ticker {
	ticker := model.Ticker{
		Symbol:    st.orderBook.Symbol,
		UpdatedAt: time.Now().UTC(),
	}
	if ticker.Symbol == "" && len(st.trades) > 0 {
		ticker.Symbol = st.trades[0].Symbol
	}

	if len(st.orderBook.Bids) > 0 {
		ticker.BestBid = st.orderBook.Bids[0].Price
	}
	if len(st.orderBook.Asks) > 0 {
		ticker.BestAsk = st.orderBook.Asks[0].Price
	}
	if len(st.trades) > 0 {
		ticker.LastPrice = st.trades[0].Price
	}

	stats := computeStats(st.stats24h)
	ticker.High24h = stats.High
	ticker.Low24h = stats.Low
	ticker.Volume24h = stats.Volume
	ticker.QuoteVolume24h = stats.QuoteVolume
	ticker.TradeCount24h = stats.TradeCount

	if stats.OpenPrice > 0 && ticker.LastPrice > 0 {
		ticker.PriceChange24h = ticker.LastPrice - stats.OpenPrice
		ticker.PriceChangePct24h = (ticker.PriceChange24h * 10000) / stats.OpenPrice
	}

	return ticker
}

func updateCandle(candles []model.Candle, symbol, intervalName string, dur time.Duration, t model.Trade, max int) []model.Candle {
	openTime := model.CandleOpenTime(t.Timestamp, dur)
	closeTime := openTime.Add(dur)

	if len(candles) > 0 {
		last := candles[len(candles)-1]
		if last.OpenTime.Equal(openTime) {
			last.High = max64(last.High, t.Price)
			last.Low = min64(last.Low, t.Price)
			last.Close = t.Price
			last.Volume += t.Quantity
			last.QuoteVol += t.Notional
			last.TradeCount++
			candles[len(candles)-1] = last
			return candles
		}
	}

	candle := model.Candle{
		Symbol:     symbol,
		Interval:   intervalName,
		OpenTime:   openTime,
		CloseTime:  closeTime,
		Open:       t.Price,
		High:       t.Price,
		Low:        t.Price,
		Close:      t.Price,
		Volume:     t.Quantity,
		QuoteVol:   t.Notional,
		TradeCount: 1,
	}
	candles = append(candles, candle)
	if len(candles) > max {
		candles = candles[len(candles)-max:]
	}
	return candles
}

func toDepth(levels []events.DepthLevel) []model.DepthLevel {
	out := make([]model.DepthLevel, len(levels))
	for i, l := range levels {
		out[i] = model.DepthLevel{Price: l.Price, Quantity: l.Quantity}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
