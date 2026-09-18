package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/apperrors"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/dto"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/httputil"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/service"
)

const ordersPrefix = "/api/v1/binance/orders/"
const marketPrefix = "/api/v1/binance/market/"

type AdapterHandler struct {
	svc *service.AdapterService
}

func NewAdapterHandler(svc *service.AdapterService) *AdapterHandler {
	return &AdapterHandler{svc: svc}
}

func (h *AdapterHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/v1/binance/ping", h.handlePing)
	mux.HandleFunc("/api/v1/binance/account", h.handleAccount)
	mux.HandleFunc("/api/v1/binance/ticker/", h.handleTicker)
	mux.HandleFunc("/api/v1/binance/market/", h.handleMarket)
	mux.HandleFunc("/api/v1/binance/orders", h.handleOrders)
	mux.HandleFunc("/api/v1/binance/orders/", h.handleOrderByID)
}

func (h *AdapterHandler) handlePing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	if err := h.svc.Ping(r.Context()); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AdapterHandler) handleAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	account, err := h.svc.GetAccount(r.Context())
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, account)
}

func (h *AdapterHandler) handleMarket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}

	path := strings.TrimPrefix(r.URL.Path, marketPrefix)
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "not found"))
		return
	}

	resource := parts[0]
	symbol := strings.Join(parts[1:], "/")
	if symbol == "" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "symbol required"))
		return
	}

	switch resource {
	case "ticker":
		ticker, err := h.svc.GetMarketTicker(r.Context(), symbol)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, ticker)
	case "orderbook":
		limit := parseLimit(r, 20)
		book, err := h.svc.GetMarketOrderBook(r.Context(), symbol, limit)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, book)
	case "trades":
		limit := parseLimit(r, 30)
		trades, err := h.svc.GetMarketTrades(r.Context(), symbol, limit)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, trades)
	case "candles":
		interval := r.URL.Query().Get("interval")
		if interval == "" {
			interval = "1m"
		}
		limit := parseLimit(r, 60)
		candles, err := h.svc.GetMarketCandles(r.Context(), symbol, interval, limit)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, candles)
	default:
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "not found"))
	}
}

func parseLimit(r *http.Request, defaultLimit int) int {
	raw := r.URL.Query().Get("limit")
	if raw == "" {
		return defaultLimit
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return defaultLimit
	}
	return n
}

func (h *AdapterHandler) handleTicker(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	symbol := strings.TrimPrefix(r.URL.Path, "/api/v1/binance/ticker/")
	if symbol == "" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "symbol required"))
		return
	}
	ticker, err := h.svc.GetTicker(r.Context(), symbol)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, ticker)
}

func (h *AdapterHandler) handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/binance/orders" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "not found"))
		return
	}
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}

	var req dto.PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}

	resp, err := h.svc.PlaceOrder(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, resp)
}

func (h *AdapterHandler) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, ordersPrefix)
	if orderID == "" || strings.Contains(orderID, "/") {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "order id required"))
		return
	}
	symbol := r.URL.Query().Get("symbol")

	switch r.Method {
	case http.MethodGet:
		if symbol == "" {
			httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "symbol query param required"))
			return
		}
		resp, err := h.svc.GetOrder(r.Context(), symbol, orderID)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, resp)

	case http.MethodDelete:
		if symbol == "" {
			httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "symbol query param required"))
			return
		}
		resp, err := h.svc.CancelOrder(r.Context(), symbol, orderID)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, resp)

	default:
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
	}
}
