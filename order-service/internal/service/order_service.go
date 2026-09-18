package service

import (
	"context"
	"fmt"
	"sync/atomic"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/auth"
	"github.com/shivank0310/cex.git/order-service/internal/client"
	"github.com/shivank0310/cex.git/order-service/internal/dto"
	"github.com/shivank0310/cex.git/order-service/internal/repository"
	"github.com/shivank0310/cex.git/order-service/internal/validator"
	"github.com/shivank0310/cex.git/pkg/redis"
)

// OrderService orchestrates authentication, validation, and matching-engine submission.
// Order flow: validate → matching engine (RAM) → persist. Redis is NOT on the critical matching path.
type OrderService struct {
	validator  *validator.Pipeline
	engine     client.MatchingEngineClient
	repo       repository.OrderRepository
	funds      *FundsHoldManager
	orderState *redis.OrderStateStore // optional temporary order cache
	orderSeq   uint64
}

func NewOrderService(
	pipeline *validator.Pipeline,
	engine client.MatchingEngineClient,
	repo repository.OrderRepository,
	funds *FundsHoldManager,
) *OrderService {
	return &OrderService{
		validator: pipeline,
		engine:    engine,
		repo:      repo,
		funds:     funds,
	}
}

func NewOrderServiceWithRedis(
	pipeline *validator.Pipeline,
	engine client.MatchingEngineClient,
	repo repository.OrderRepository,
	funds *FundsHoldManager,
	orderState *redis.OrderStateStore,
) *OrderService {
	return &OrderService{
		validator:  pipeline,
		engine:     engine,
		repo:       repo,
		funds:      funds,
		orderState: orderState,
	}
}

// PlaceOrder runs the full order-service pipeline:
// authenticate → validate symbol → validate price → validate quantity →
// check trading rules → check balance → submit to matching engine.
func (s *OrderService) PlaceOrder(ctx context.Context, req dto.PlaceOrderRequest) (dto.PlaceOrderResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return dto.PlaceOrderResponse{}, apperrors.New(apperrors.CodeUnauthorized, "user not authenticated")
	}

	validated, err := s.validator.Validate(userID, req)
	if err != nil {
		return dto.PlaceOrderResponse{}, err
	}

	orderID := s.nextOrderID()
	if err := s.funds.LockForOrder(ctx, userID, validated.Symbol, validated.Side, validated.Type, validated.Price, validated.Quantity); err != nil {
		return dto.PlaceOrderResponse{}, apperrors.Wrap(apperrors.CodeInsufficientBalance, "failed to lock order funds", err)
	}

	var o *meapi.Order
	switch validated.Type {
	case meapi.Limit:
		o = meapi.NewLimitOrder(orderID, userID, validated.Symbol.Name, validated.Side, validated.Price, validated.Quantity)
	case meapi.Market:
		o = meapi.NewMarketOrder(orderID, userID, validated.Symbol.Name, validated.Side, validated.Quantity)
	default:
		_ = s.funds.UnlockOrder(ctx, userID, validated.Symbol, validated.Side, validated.Type, validated.Price, validated.Quantity)
		return dto.PlaceOrderResponse{}, apperrors.New(apperrors.CodeInvalidRequest, "unsupported order type")
	}

	result := s.engine.Submit(o)
	if result.Error != nil {
		_ = s.funds.UnlockOrder(ctx, userID, validated.Symbol, validated.Side, validated.Type, validated.Price, validated.Quantity)
		return dto.PlaceOrderResponse{}, apperrors.Wrap(apperrors.CodeEngineError, "matching engine rejected order", result.Error)
	}

	// Unfilled market-order remainder is cancelled by the engine; release the hold.
	if validated.Type == meapi.Market && result.Order.Remaining > 0 {
		_ = s.funds.UnlockOrder(ctx, userID, validated.Symbol, validated.Side, validated.Type, validated.Price, result.Order.Remaining)
	}

	if err := s.repo.Save(result.Order); err != nil {
		return dto.PlaceOrderResponse{}, apperrors.Wrap(apperrors.CodeInternal, "failed to persist order", err)
	}

	// Cache temporary order state in Redis for fast lookup (not on matching critical path).
	if s.orderState != nil {
		_ = s.orderState.Save(ctx, result.Order.ID, dto.ToOrderResponse(result.Order))
	}

	resp := dto.PlaceOrderResponse{Order: dto.ToOrderResponse(result.Order)}
	for _, tr := range result.Trades {
		resp.Trades = append(resp.Trades, dto.ToTradeResponse(tr))
	}
	return resp, nil
}

// CancelOrder cancels a resting order for the authenticated user.
func (s *OrderService) CancelOrder(ctx context.Context, req dto.CancelOrderRequest) (dto.OrderResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return dto.OrderResponse{}, apperrors.New(apperrors.CodeUnauthorized, "user not authenticated")
	}

	if req.Symbol == "" || req.OrderID == "" {
		return dto.OrderResponse{}, apperrors.New(apperrors.CodeInvalidRequest, "symbol and order_id are required")
	}

	existing, err := s.repo.Get(req.OrderID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	if existing.UserID != userID {
		return dto.OrderResponse{}, apperrors.New(apperrors.CodeOrderNotFound, "order not found")
	}

	cancelled, err := s.engine.Cancel(req.Symbol, req.OrderID)
	if err != nil {
		return dto.OrderResponse{}, apperrors.Wrap(apperrors.CodeEngineError, "failed to cancel order", err)
	}

	if sym, symErr := s.validator.SymbolFor(req.Symbol); symErr == nil {
		_ = s.funds.UnlockOrder(ctx, userID, sym, cancelled.Side, cancelled.Type, cancelled.Price, cancelled.Remaining)
	}

	if err := s.repo.Save(cancelled); err != nil {
		return dto.OrderResponse{}, apperrors.Wrap(apperrors.CodeInternal, "failed to persist cancellation", err)
	}

	return dto.ToOrderResponse(cancelled), nil
}

// GetOrder returns a stored order for the authenticated user.
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (dto.OrderResponse, error) {
	userID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return dto.OrderResponse{}, apperrors.New(apperrors.CodeUnauthorized, "user not authenticated")
	}

	// Fast path: check Redis temporary order state first.
	if s.orderState != nil {
		var cached dto.OrderResponse
		if found, _ := s.orderState.Get(ctx, orderID, &cached); found && cached.UserID == userID {
			return cached, nil
		}
	}

	o, err := s.repo.Get(orderID)
	if err != nil {
		return dto.OrderResponse{}, err
	}
	if o.UserID != userID {
		return dto.OrderResponse{}, apperrors.New(apperrors.CodeOrderNotFound, "order not found")
	}

	return dto.ToOrderResponse(o), nil
}

func (s *OrderService) nextOrderID() string {
	id := atomic.AddUint64(&s.orderSeq, 1)
	return fmt.Sprintf("ORD-%d", id)
}
