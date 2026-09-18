package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/binance"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/dto"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/handler"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/mapping"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/repository"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/service"
)

func setupHandler() *handler.AdapterHandler {
	provider := binance.NewMockProvider()
	repo := repository.NewOrderRepository()
	svc := service.NewAdapterService(provider, repo)
	return handler.NewAdapterHandler(svc)
}

func TestSymbolMapping(t *testing.T) {
	if mapping.ToBinanceSymbol("BTC/USDT") != "BTCUSDT" {
		t.Fatal("ToBinanceSymbol failed")
	}
	if mapping.FromBinanceSymbol("BTCUSDT") != "BTC/USDT" {
		t.Fatal("FromBinanceSymbol failed")
	}
	if mapping.QuantityToBinance(30) != "0.3" {
		t.Fatalf("QuantityToBinance: got %s", mapping.QuantityToBinance(30))
	}
	qty, err := mapping.QuantityFromBinance("0.3")
	if err != nil || qty != 30 {
		t.Fatalf("QuantityFromBinance: got %d", qty)
	}
}

func TestMockPlaceLimitOrder(t *testing.T) {
	h := setupHandler()
	body, _ := json.Marshal(dto.PlaceOrderRequest{
		ClientOrderID: "ORD-1",
		UserID:        "alice",
		Symbol:        "BTC/USDT",
		Side:          "BUY",
		Type:          "LIMIT",
		Price:         101100,
		Quantity:      30,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/binance/orders", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.Register(mux)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status: %d body: %s", w.Code, w.Body.String())
	}

	var resp dto.PlaceOrderResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Order.Status != "NEW" {
		t.Fatalf("expected NEW, got %s", resp.Order.Status)
	}
	if resp.Order.Venue != "binance" {
		t.Fatalf("expected binance venue, got %s", resp.Order.Venue)
	}
}

func TestMockPlaceMarketOrder(t *testing.T) {
	h := setupHandler()
	body, _ := json.Marshal(dto.PlaceOrderRequest{
		ClientOrderID: "ORD-2",
		UserID:        "alice",
		Symbol:        "BTC/USDT",
		Side:          "BUY",
		Type:          "MARKET",
		Price:         101100,
		Quantity:      30,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/binance/orders", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.Register(mux)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status: %d body: %s", w.Code, w.Body.String())
	}

	var resp dto.PlaceOrderResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Order.Status != "FILLED" {
		t.Fatalf("expected FILLED, got %s", resp.Order.Status)
	}
	if len(resp.Trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(resp.Trades))
	}
}

func TestMockCancelOrder(t *testing.T) {
	provider := binance.NewMockProvider()
	repo := repository.NewOrderRepository()
	svc := service.NewAdapterService(provider, repo)

	_, err := svc.PlaceOrder(context.Background(), dto.PlaceOrderRequest{
		ClientOrderID: "ORD-CANCEL",
		UserID:        "alice",
		Symbol:        "BTC/USDT",
		Side:          "SELL",
		Type:          "LIMIT",
		Price:         101100,
		Quantity:      30,
	})
	if err != nil {
		t.Fatal(err)
	}

	cancelled, err := svc.CancelOrder(context.Background(), "BTC/USDT", "ORD-CANCEL")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != "CANCELLED" {
		t.Fatalf("expected CANCELLED, got %s", cancelled.Status)
	}
}

func TestGetAccount(t *testing.T) {
	h := setupHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/binance/account", nil)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.Register(mux)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}

	var resp dto.AccountResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Venue != "binance" || len(resp.Balances) == 0 {
		t.Fatalf("unexpected account: %+v", resp)
	}
}

func TestGetTicker(t *testing.T) {
	h := setupHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/binance/ticker/BTC/USDT", nil)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	h.Register(mux)
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}
}
