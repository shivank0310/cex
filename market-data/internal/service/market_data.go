package service

import (
	"context"
	"fmt"

	"github.com/shivank0310/cex.git/market-data/internal/cache"
	"github.com/shivank0310/cex.git/market-data/internal/client"
	"github.com/shivank0310/cex.git/market-data/internal/dto"
	"github.com/shivank0310/cex.git/market-data/internal/model"
	"github.com/shivank0310/cex.git/market-data/internal/store"
	"github.com/shivank0310/cex.git/pkg/events"
)

// MarketDataService consumes Kafka events and serves market data queries.
// Matching engine stays in RAM; Redis is used only as a read cache layer here.
type MarketDataService struct {
	store   *store.Store
	cache   *cache.MarketCache // optional Redis cache
	binance *client.BinanceAdapterClient
}

func NewMarketDataService(st *store.Store) *MarketDataService {
	return &MarketDataService{store: st}
}

func NewMarketDataServiceWithCache(st *store.Store, mc *cache.MarketCache) *MarketDataService {
	return &MarketDataService{store: st, cache: mc}
}

func NewMarketDataServiceWithBinance(st *store.Store, mc *cache.MarketCache, binance *client.BinanceAdapterClient) *MarketDataService {
	return &MarketDataService{store: st, cache: mc, binance: binance}
}

// Handle processes incoming Kafka events and updates RAM store + Redis cache.
func (s *MarketDataService) Handle(ctx context.Context, env events.Envelope) error {
	switch env.EventType {
	case events.TypeTradeExecuted:
		trade, err := events.DecodePayload[events.TradePayload](env)
		if err != nil {
			return err
		}
		s.store.ApplyTrade(trade)
		return s.refreshCache(ctx, trade.Symbol)

	case events.TypeOrderBookUpdated:
		book, err := events.DecodePayload[events.OrderBookPayload](env)
		if err != nil {
			return err
		}
		s.store.ApplyOrderBook(book)
		return s.refreshCache(ctx, book.Symbol)
	}
	return nil
}

func (s *MarketDataService) refreshCache(ctx context.Context, symbol string) error {
	if s.cache == nil {
		return nil
	}

	if t, ok := s.store.Ticker(symbol); ok && t.Symbol != "" {
		if err := s.cache.CacheTicker(ctx, t); err != nil {
			return err
		}
	}
	if book, ok := s.store.OrderBook(symbol); ok && book.Symbol != "" {
		if err := s.cache.CacheOrderBook(ctx, book); err != nil {
			return err
		}
	}
	if stats, ok := s.store.Stats24h(symbol); ok {
		if err := s.cache.CacheStats(ctx, symbol, stats); err != nil {
			return err
		}
	}
	trades := s.store.Trades(symbol, 50)
	return s.cache.CacheTrades(ctx, symbol, trades)
}

func (s *MarketDataService) GetTicker(ctx context.Context, symbol string) (model.Ticker, error) {
	if s.cache != nil {
		if cached, found, err := s.cache.GetTicker(ctx, symbol); err == nil && found {
			return model.Ticker{
				Symbol:            cached.Symbol,
				LastPrice:         cached.LastPrice,
				BestBid:           cached.BestBid,
				BestAsk:           cached.BestAsk,
				High24h:           cached.High24h,
				Low24h:            cached.Low24h,
				Volume24h:         cached.Volume24h,
				QuoteVolume24h:    cached.QuoteVolume24h,
				PriceChange24h:    cached.PriceChange24h,
				PriceChangePct24h: cached.PriceChangePct24h,
				TradeCount24h:     cached.TradeCount24h,
			}, nil
		}
	}

	t, ok := s.store.Ticker(symbol)
	if !ok || t.Symbol == "" {
		if s.binance != nil && s.binance.IsExternal(symbol) {
			return s.binance.GetTicker(ctx, symbol)
		}
		return model.Ticker{}, fmt.Errorf("ticker not found for symbol: %s", symbol)
	}
	if s.cache != nil {
		_ = s.cache.CacheTicker(ctx, t)
	}
	return t, nil
}

func (s *MarketDataService) GetAllTickers() []model.Ticker {
	tickers := s.store.AllTickers()
	if s.binance == nil {
		return tickers
	}

	seen := make(map[string]bool, len(tickers))
	for _, t := range tickers {
		seen[t.Symbol] = true
	}
	for _, symbol := range s.binance.ExternalSymbols() {
		if seen[symbol] {
			continue
		}
		t, err := s.binance.GetTicker(context.Background(), symbol)
		if err == nil {
			tickers = append(tickers, t)
		}
	}
	return tickers
}

func (s *MarketDataService) GetOrderBook(ctx context.Context, symbol string) (model.OrderBook, error) {
	if s.cache != nil {
		if cached, found, err := s.cache.GetOrderBook(ctx, symbol); err == nil && found {
			return dtoToOrderBook(cached), nil
		}
	}

	book, ok := s.store.OrderBook(symbol)
	if !ok || book.Symbol == "" {
		if s.binance != nil && s.binance.IsExternal(symbol) {
			return s.binance.GetOrderBook(ctx, symbol, 20)
		}
		return model.OrderBook{}, fmt.Errorf("order book not found for symbol: %s", symbol)
	}
	if s.cache != nil {
		_ = s.cache.CacheOrderBook(ctx, book)
	}
	return book, nil
}

func (s *MarketDataService) GetTrades(ctx context.Context, symbol string, limit int) []model.Trade {
	if s.cache != nil && limit <= 50 {
		if cached, found, err := s.cache.GetTrades(ctx, symbol); err == nil && found && len(cached) > 0 {
			if limit > 0 && limit < len(cached) {
				cached = cached[:limit]
			}
			return dtoToTrades(cached)
		}
	}
	trades := s.store.Trades(symbol, limit)
	if len(trades) == 0 && s.binance != nil && s.binance.IsExternal(symbol) {
		remote, err := s.binance.GetTrades(ctx, symbol, limit)
		if err == nil {
			return remote
		}
	}
	return trades
}

func (s *MarketDataService) GetCandles(symbol, interval string, limit int) ([]model.Candle, error) {
	interval = model.NormalizeInterval(interval)
	if _, err := model.ParseInterval(interval); err != nil {
		return nil, err
	}
	candles, ok := s.store.Candles(symbol, interval, limit)
	if !ok || len(candles) == 0 {
		if s.binance != nil && s.binance.IsExternal(symbol) {
			return s.binance.GetCandles(context.Background(), symbol, interval, limit)
		}
		return nil, fmt.Errorf("candles not found for symbol: %s", symbol)
	}
	return candles, nil
}

func (s *MarketDataService) GetStats24h(ctx context.Context, symbol string) (model.Stats24h, error) {
	if s.cache != nil {
		if cached, found, err := s.cache.GetStats(ctx, symbol); err == nil && found {
			return model.Stats24h{
				High: cached.High, Low: cached.Low,
				Volume: cached.Volume, QuoteVolume: cached.QuoteVolume,
				TradeCount: cached.TradeCount, OpenPrice: cached.OpenPrice,
			}, nil
		}
	}

	stats, ok := s.store.Stats24h(symbol)
	if !ok {
		return model.Stats24h{}, fmt.Errorf("stats not found for symbol: %s", symbol)
	}
	if s.cache != nil {
		_ = s.cache.CacheStats(ctx, symbol, stats)
	}
	return stats, nil
}

// Backward-compatible methods without context (for existing tests).
func (s *MarketDataService) GetTickerSync(symbol string) (model.Ticker, error) {
	return s.GetTicker(context.Background(), symbol)
}

func dtoToOrderBook(d dto.OrderBookResponse) model.OrderBook {
	book := model.OrderBook{Symbol: d.Symbol}
	for _, b := range d.Bids {
		book.Bids = append(book.Bids, model.DepthLevel{Price: b.Price, Quantity: b.Quantity})
	}
	for _, a := range d.Asks {
		book.Asks = append(book.Asks, model.DepthLevel{Price: a.Price, Quantity: a.Quantity})
	}
	return book
}

func dtoToTrades(dtos []dto.TradeResponse) []model.Trade {
	out := make([]model.Trade, len(dtos))
	for i, d := range dtos {
		out[i] = model.Trade{ID: d.ID, Symbol: d.Symbol, Price: d.Price, Quantity: d.Quantity, Notional: d.Notional}
	}
	return out
}
