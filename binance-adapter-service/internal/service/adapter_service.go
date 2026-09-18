package service

import (
	"context"
	"fmt"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/apperrors"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/binance"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/dto"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/repository"
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

func (s *AdapterService) Ping(ctx context.Context) error {
	return s.provider.Ping(ctx)
}
