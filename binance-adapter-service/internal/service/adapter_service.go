package service

import (
	"context"
	"fmt"
	"time"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/apperrors"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/binance"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/dto"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/mapping"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/repository"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
)

// AdapterService orchestrates Binance Spot API calls for the CEX.
type AdapterService struct {
	provider binance.SpotProvider
	repo     *repository.OrderRepository
}

func NewAdapterService(provider binance.SpotProvider, repo *repository.OrderRepository) *AdapterService {
	return &AdapterService{provider: provider, repo: repo}
}

func (s *AdapterService) PlaceOrder(ctx context.Context, req dto.PlaceOrderRequest) (dto.PlaceOrderResponse, error) {
	if req.Symbol == "" || req.Side == "" || req.Type == "" || req.Quantity <= 0 {
		return dto.PlaceOrderResponse{}, apperrors.New(apperrors.CodeInvalidRequest, "symbol, side, type, and quantity required")
	}
	if req.Type == "LIMIT" && req.Price <= 0 {
		return dto.PlaceOrderResponse{}, apperrors.New(apperrors.CodeInvalidRequest, "price required for limit orders")
	}

	clientID := req.ClientOrderID
	if clientID == "" {
		clientID = fmt.Sprintf("CEX-%s", req.UserID)
	}

	result, err := s.provider.PlaceOrder(ctx, binance.OrderRequest{
		ClientOrderID: clientID,
		Symbol:        req.Symbol,
		Side:          req.Side,
		Type:          req.Type,
		Price:         req.Price,
		Quantity:      req.Quantity,
	})
	if err != nil {
		return dto.PlaceOrderResponse{}, err
	}

	s.repo.Save(req.UserID, result)
	orderResp := dto.ToOrderResponse(req.UserID, result)

	var trades []dto.TradeResponse
	if result.ExecutedQty > 0 {
		trades = append(trades, dto.TradeResponse{
			ID:       fmt.Sprintf("BN-T-%d", result.BinanceOrderID),
			Symbol:   result.Symbol,
			Price:    result.Price,
			Quantity: result.ExecutedQty,
			Side:     result.Side,
			Venue:    "binance",
		})
	}

	return dto.PlaceOrderResponse{Order: orderResp, Trades: trades}, nil
}

func (s *AdapterService) CancelOrder(ctx context.Context, symbol, orderID string) (dto.OrderResponse, error) {
	rec, ok := s.repo.Get(orderID)
	if !ok {
		return dto.OrderResponse{}, apperrors.New(apperrors.CodeOrderNotFound, "order not found")
	}

	result, err := s.provider.CancelOrder(ctx, symbol, orderID)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	s.repo.Update(result)
	return dto.ToOrderResponse(rec.UserID, result), nil
}

func (s *AdapterService) GetOrder(ctx context.Context, symbol, orderID string) (dto.OrderResponse, error) {
	rec, ok := s.repo.Get(orderID)
	if !ok {
		return dto.OrderResponse{}, apperrors.New(apperrors.CodeOrderNotFound, "order not found")
	}

	result, err := s.provider.GetOrder(ctx, symbol, orderID)
	if err != nil {
		return dto.OrderResponse{}, err
	}

	s.repo.Update(result)
	return dto.ToOrderResponse(rec.UserID, result), nil
}

func (s *AdapterService) GetAccount(ctx context.Context) (dto.AccountResponse, error) {
	balances, err := s.provider.GetAccount(ctx)
	if err != nil {
		return dto.AccountResponse{}, err
	}

	resp := dto.AccountResponse{
		CanTrade: true, CanWithdraw: true, CanDeposit: true,
		AccountType: "SPOT", Permissions: []string{"SPOT"}, Venue: "binance",
	}
	for _, b := range balances {
		resp.Balances = append(resp.Balances, dto.AssetBalance{
			Asset: b.Asset, Free: b.Free, Locked: b.Locked,
		})
	}
	return resp, nil
}

func (s *AdapterService) GetTicker(ctx context.Context, symbol string) (dto.TickerResponse, error) {
	ticker, err := s.provider.GetTickerPrice(ctx, symbol)
	if err != nil {
		return dto.TickerResponse{}, err
	}
	return dto.TickerResponse{Symbol: ticker.Symbol, Price: ticker.Price, Venue: "binance"}, nil
}

func (s *AdapterService) GetMarketTicker(ctx context.Context, symbol string) (dto.MarketTickerResponse, error) {
	t, err := s.provider.GetTicker24h(ctx, symbol)
	if err != nil {
		return dto.MarketTickerResponse{}, err
	}
	return dto.MarketTickerResponse{
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
		UpdatedAt:         time.Now().UTC().Format(time.RFC3339),
		Venue:             "binance",
	}, nil
}

func (s *AdapterService) GetMarketOrderBook(ctx context.Context, symbol string, limit int) (dto.MarketOrderBookResponse, error) {
	book, err := s.provider.GetDepth(ctx, symbol, limit)
	if err != nil {
		return dto.MarketOrderBookResponse{}, err
	}
	resp := dto.MarketOrderBookResponse{
		Symbol:    book.Symbol,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Venue:     "binance",
	}
	for _, b := range book.Bids {
		resp.Bids = append(resp.Bids, dto.DepthLevelResponse{Price: b.Price, Quantity: b.Quantity})
	}
	for _, a := range book.Asks {
		resp.Asks = append(resp.Asks, dto.DepthLevelResponse{Price: a.Price, Quantity: a.Quantity})
	}
	return resp, nil
}

func (s *AdapterService) GetMarketTrades(ctx context.Context, symbol string, limit int) ([]dto.MarketTradeResponse, error) {
	trades, err := s.provider.GetRecentTrades(ctx, symbol, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.MarketTradeResponse, len(trades))
	for i, t := range trades {
		out[i] = dto.MarketTradeResponse{
			ID:        t.ID,
			Symbol:    t.Symbol,
			Price:     t.Price,
			Quantity:  t.Quantity,
			Notional:  decimal.Notional(t.Price, t.Quantity),
			Timestamp: t.Timestamp.UTC().Format(time.RFC3339),
			Venue:     "binance",
		}
	}
	return out, nil
}

func (s *AdapterService) GetMarketCandles(ctx context.Context, symbol, interval string, limit int) ([]dto.MarketCandleResponse, error) {
	interval = mapping.NormalizeInterval(interval)
	klines, err := s.provider.GetKlines(ctx, symbol, interval, limit)
	if err != nil {
		return nil, err
	}
	out := make([]dto.MarketCandleResponse, len(klines))
	for i, k := range klines {
		out[i] = dto.MarketCandleResponse{
			Symbol:     symbol,
			Interval:   interval,
			OpenTime:   k.OpenTime.UTC().Format(time.RFC3339),
			CloseTime:  k.CloseTime.UTC().Format(time.RFC3339),
			Open:       k.Open,
			High:       k.High,
			Low:        k.Low,
			Close:      k.Close,
			Volume:     k.Volume,
			QuoteVol:   k.QuoteVolume,
			TradeCount: k.TradeCount,
			Venue:      "binance",
		}
	}
	return out, nil
}

func (s *AdapterService) Ping(ctx context.Context) error {
	return s.provider.Ping(ctx)
}
