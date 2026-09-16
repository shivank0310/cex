package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/dto"
	"github.com/shivank0310/cex.git/order-service/internal/service"
)

const ordersPathPrefix = "/api/v1/orders/"

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{service: svc}
}

func (h *OrderHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/orders", h.handleOrders)
	mux.HandleFunc("/api/v1/orders/", h.handleOrderByID)
}

func (h *OrderHandler) handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/orders" {
		WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "not found"))
		return
	}
	if r.Method != http.MethodPost {
		WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}

	var req dto.PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}

	resp, err := h.service.PlaceOrder(r.Context(), req)
	if err != nil {
		WriteError(w, err)
		return
	}

	WriteJSON(w, http.StatusCreated, resp)
}

func (h *OrderHandler) handleOrderByID(w http.ResponseWriter, r *http.Request) {
	orderID := strings.TrimPrefix(r.URL.Path, ordersPathPrefix)
	if orderID == "" || strings.Contains(orderID, "/") {
		WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "order id is required"))
		return
	}

	switch r.Method {
	case http.MethodGet:
		resp, err := h.service.GetOrder(r.Context(), orderID)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, resp)
	case http.MethodDelete:
		req := dto.CancelOrderRequest{
			OrderID: orderID,
			Symbol:  r.URL.Query().Get("symbol"),
		}
		resp, err := h.service.CancelOrder(r.Context(), req)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteJSON(w, http.StatusOK, resp)
	default:
		WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
	}
}
