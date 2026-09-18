package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shivank0310/cex.git/binance-adapter-service/internal/apperrors"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/dto"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/httputil"
	"github.com/shivank0310/cex.git/binance-adapter-service/internal/service"
)

const ordersPrefix = "/api/v1/binance/orders/"

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
