package binance

import "context"

// CompositeProvider routes market-data calls to a live REST client and trading
// calls to either REST (signed) or a mock provider.
type CompositeProvider struct {
	market  SpotProvider
	trading SpotProvider
}

func NewCompositeProvider(market, trading SpotProvider) *CompositeProvider {
	return &CompositeProvider{market: market, trading: trading}
}

func (c *CompositeProvider) Ping(ctx context.Context) error {
	if err := c.market.Ping(ctx); err != nil {
		return err
	}
	return c.trading.Ping(ctx)
}

func (c *CompositeProvider) GetTickerPrice(ctx context.Context, symbol string) (TickerPrice, error) {
	return c.market.GetTickerPrice(ctx, symbol)
}

func (c *CompositeProvider) GetTicker24h(ctx context.Context, symbol string) (Ticker24h, error) {
	return c.market.GetTicker24h(ctx, symbol)
}

func (c *CompositeProvider) GetDepth(ctx context.Context, symbol string, limit int) (DepthSnapshot, error) {
	return c.market.GetDepth(ctx, symbol, limit)
}

func (c *CompositeProvider) GetRecentTrades(ctx context.Context, symbol string, limit int) ([]MarketTrade, error) {
	return c.market.GetRecentTrades(ctx, symbol, limit)
}

func (c *CompositeProvider) GetKlines(ctx context.Context, symbol, interval string, limit int) ([]Kline, error) {
	return c.market.GetKlines(ctx, symbol, interval, limit)
}

func (c *CompositeProvider) PlaceOrder(ctx context.Context, req OrderRequest) (OrderResult, error) {
	return c.trading.PlaceOrder(ctx, req)
}

func (c *CompositeProvider) CancelOrder(ctx context.Context, symbol, clientOrderID string) (OrderResult, error) {
	return c.trading.CancelOrder(ctx, symbol, clientOrderID)
}

func (c *CompositeProvider) GetOrder(ctx context.Context, symbol, clientOrderID string) (OrderResult, error) {
	return c.trading.GetOrder(ctx, symbol, clientOrderID)
}

func (c *CompositeProvider) GetAccount(ctx context.Context) ([]AccountBalance, error) {
	return c.trading.GetAccount(ctx)
}
