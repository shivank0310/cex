package cache

import (
	"context"
	"time"

	"github.com/shivank0310/cex.git/market-data/internal/dto"
	"github.com/shivank0310/cex.git/market-data/internal/model"
	"github.com/shivank0310/cex.git/pkg/redis"
)

const defaultTTL = 5 * time.Second

// MarketCache is a Redis write-through / read-through cache for market data.
// In-memory store remains the source during event ingestion; Redis accelerates HTTP reads.
type MarketCache struct {
	cache  *redis.Cache
	pubsub *redis.PubSub
	ttl    time.Duration
}

func NewMarketCache(client *redis.Client, ttl time.Duration) *MarketCache {
	if ttl == 0 {
		ttl = defaultTTL
	}
	return &MarketCache{
		cache:  redis.NewCache(client),
		pubsub: redis.NewPubSub(client),
		ttl:    ttl,
	}
}

func (c *MarketCache) CacheTicker(ctx context.Context, ticker model.Ticker) error {
	resp := dto.ToTickerResponse(ticker)
	if err := c.cache.Set(ctx, redis.TickerKey(ticker.Symbol), resp, c.ttl); err != nil {
		return err
	}
	return c.pubsub.Publish(ctx, "ticker:"+redis.NormalizeSymbol(ticker.Symbol), resp)
}

func (c *MarketCache) GetTicker(ctx context.Context, symbol string) (dto.TickerResponse, bool, error) {
	var resp dto.TickerResponse
	found, err := c.cache.Get(ctx, redis.TickerKey(symbol), &resp)
	return resp, found, err
}

func (c *MarketCache) CacheOrderBook(ctx context.Context, book model.OrderBook) error {
	resp := dto.ToOrderBookResponse(book)
	if err := c.cache.Set(ctx, redis.OrderBookKey(book.Symbol), resp, c.ttl); err != nil {
		return err
	}
	return c.pubsub.Publish(ctx, "orderbook:"+redis.NormalizeSymbol(book.Symbol), resp)
}

func (c *MarketCache) GetOrderBook(ctx context.Context, symbol string) (dto.OrderBookResponse, bool, error) {
	var resp dto.OrderBookResponse
	found, err := c.cache.Get(ctx, redis.OrderBookKey(symbol), &resp)
	return resp, found, err
}

func (c *MarketCache) CacheStats(ctx context.Context, symbol string, stats model.Stats24h) error {
	resp := dto.ToStats24hResponse(symbol, stats)
	return c.cache.Set(ctx, redis.StatsKey(symbol), resp, c.ttl)
}

func (c *MarketCache) GetStats(ctx context.Context, symbol string) (dto.Stats24hResponse, bool, error) {
	var resp dto.Stats24hResponse
	found, err := c.cache.Get(ctx, redis.StatsKey(symbol), &resp)
	return resp, found, err
}

func (c *MarketCache) CacheTrades(ctx context.Context, symbol string, trades []model.Trade) error {
	resp := dto.ToTradeResponses(trades)
	return c.cache.Set(ctx, redis.TradesKey(symbol), resp, c.ttl)
}

func (c *MarketCache) GetTrades(ctx context.Context, symbol string) ([]dto.TradeResponse, bool, error) {
	var resp []dto.TradeResponse
	found, err := c.cache.Get(ctx, redis.TradesKey(symbol), &resp)
	return resp, found, err
}
