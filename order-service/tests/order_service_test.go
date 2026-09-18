package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/auth"
	"github.com/shivank0310/cex.git/order-service/internal/client"
	"github.com/shivank0310/cex.git/order-service/internal/dto"
	"github.com/shivank0310/cex.git/order-service/internal/handler"
	"github.com/shivank0310/cex.git/order-service/internal/middleware"
	"github.com/shivank0310/cex.git/order-service/internal/model"
	"github.com/shivank0310/cex.git/order-service/internal/repository"
	"github.com/shivank0310/cex.git/order-service/internal/service"
	"github.com/shivank0310/cex.git/order-service/internal/validator"
)

func setup(t *testing.T) (*service.OrderService, *meapi.Ledger, *meapi.Engine) {
	ledger := meapi.NewLedger()
	eng := meapi.NewEngine(ledger, meapi.FeeConfig{MakerBasisPoints: 0, TakerBasisPoints: 0})

	registry := model.NewRegistry()
	for _, sym := range model.DefaultSymbols() {
		registry.Register(sym)
		eng.RegisterSymbol(sym.Name, meapi.SymbolConfig{
			BaseAsset:  sym.BaseAsset,
			QuoteAsset: sym.QuoteAsset,
		})
	}

	ledger.Deposit("seller-1", "BTC", 30)
	ledger.Deposit("user-a", "USDT", decimal.Notional(101100, 30))

	pipeline := validator.NewPipeline(registry, client.NewLedgerWallet(ledger))
	svc := service.NewOrderService(
		pipeline,
		client.NewEngineClient(eng),
		repository.NewInMemoryOrderRepository(),
		service.NewFundsHoldManager(nil),
	)

	seedSell(t, eng, ledger, "ask1", "seller-1", 101100, 30)

	return svc, ledger, eng
}

func seedSell(t *testing.T, eng *meapi.Engine, ledger *meapi.Ledger, id, user string, price, qty int64) {
	ledger.Deposit(user, "BTC", qty)
	o := meapi.NewLimitOrder(id, user, "BTC/USDT", meapi.Sell, price, qty)
	res := eng.SubmitOrder(o)
	if res.Error != nil {
		t.Fatalf("seed sell: %v", res.Error)
	}
}

func TestPlaceOrder_Success(t *testing.T) {
	svc, ledger, _ := setup(t)
	ctx := auth.WithUserID(context.Background(), "user-a")

	resp, err := svc.PlaceOrder(ctx, dto.PlaceOrderRequest{
		Symbol:   "BTC/USDT",
		Side:     "BUY",
		Type:     "LIMIT",
		Price:    101100,
		Quantity: 30,
	})
	if err != nil {
		t.Fatalf("place order: %v", err)
	}
	if len(resp.Trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(resp.Trades))
	}
	if resp.Order.Status != "FILLED" {
		t.Fatalf("order status: %s", resp.Order.Status)
	}

	btc := ledger.Get("user-a", "BTC")
	if btc.Available != 30 {
		t.Fatalf("buyer BTC: %d", btc.Available)
	}
}

func TestPlaceOrder_InsufficientBalance(t *testing.T) {
	svc, _, _ := setup(t)
	ctx := auth.WithUserID(context.Background(), "user-a")

	_, err := svc.PlaceOrder(ctx, dto.PlaceOrderRequest{
		Symbol:   "BTC/USDT",
		Side:     "BUY",
		Type:     "LIMIT",
		Price:    101100,
		Quantity: 100,
	})
	if err == nil {
		t.Fatal("expected insufficient balance error")
	}
	var appErr *apperrors.AppError
	if !asAppError(err, &appErr) || appErr.Code != apperrors.CodeInsufficientBalance {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlaceOrder_InvalidSymbol(t *testing.T) {
	svc, _, _ := setup(t)
	ctx := auth.WithUserID(context.Background(), "user-a")

	_, err := svc.PlaceOrder(ctx, dto.PlaceOrderRequest{
		Symbol:   "ETH/USDT",
		Side:     "BUY",
		Type:     "LIMIT",
		Price:    100,
		Quantity: 10,
	})
	if err == nil {
		t.Fatal("expected invalid symbol error")
	}
}

func TestPlaceOrder_InvalidPrice(t *testing.T) {
	svc, _, _ := setup(t)
	ctx := auth.WithUserID(context.Background(), "user-a")

	_, err := svc.PlaceOrder(ctx, dto.PlaceOrderRequest{
		Symbol:   "BTC/USDT",
		Side:     "BUY",
		Type:     "LIMIT",
		Price:    0,
		Quantity: 30,
	})
	if err == nil {
		t.Fatal("expected invalid price error")
	}
}

func TestPlaceOrder_MinNotionalRule(t *testing.T) {
	svc, ledger, _ := setup(t)
	ledger.Deposit("user-a", "USDT", 100)
	ctx := auth.WithUserID(context.Background(), "user-a")

	_, err := svc.PlaceOrder(ctx, dto.PlaceOrderRequest{
		Symbol:   "BTC/USDT",
		Side:     "BUY",
		Type:     "LIMIT",
		Price:    1,
		Quantity: 1,
	})
	if err == nil {
		t.Fatal("expected trading rule violation")
	}
}

func TestHTTPPlaceOrder(t *testing.T) {
	svc, _, _ := setup(t)
	orderHandler := handler.NewOrderHandler(svc)
	authenticator := auth.NewStaticAuthenticator(map[string]string{"token-user-a": "user-a"})

	mux := http.NewServeMux()
	orderHandler.Register(mux)
	var root http.Handler = mux
	root = middleware.AuthMiddleware(authenticator)(root)

	body, _ := json.Marshal(dto.PlaceOrderRequest{
		Symbol:   "BTC/USDT",
		Side:     "BUY",
		Type:     "LIMIT",
		Price:    101100,
		Quantity: 30,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer token-user-a")
	rec := httptest.NewRecorder()
	root.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status: %d body: %s", rec.Code, rec.Body.String())
	}
}

func asAppError(err error, target **apperrors.AppError) bool {
	if err == nil {
		return false
	}
	if ae, ok := err.(*apperrors.AppError); ok {
		*target = ae
		return true
	}
	return false
}
