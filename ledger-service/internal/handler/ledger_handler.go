package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shivank0310/cex.git/ledger-service/internal/dto"
	"github.com/shivank0310/cex.git/ledger-service/internal/httputil"
	"github.com/shivank0310/cex.git/ledger-service/internal/service"
)

const apiPrefix = "/api/v1/ledger/"

type LedgerHandler struct {
	svc *service.LedgerService
}

func NewLedgerHandler(svc *service.LedgerService) *LedgerHandler {
	return &LedgerHandler{svc: svc}
}

func (h *LedgerHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/ledger/deposit", h.handleDeposit)
	mux.HandleFunc("/api/v1/ledger/settle-trade", h.handleSettleTrade)
	mux.HandleFunc("/api/v1/ledger/reserve", h.handleReserve)
	mux.HandleFunc("/api/v1/ledger/release", h.handleRelease)
	mux.HandleFunc("/api/v1/ledger/debit-locked", h.handleDebitLocked)
	mux.HandleFunc("/api/v1/ledger/", h.handleLedger)
}

func (h *LedgerHandler) handleSettleTrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	var req dto.SettleTradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.ID == "" || req.Symbol == "" || req.BuyerID == "" || req.SellerID == "" || req.Quantity <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "trade id, symbol, buyer, seller, and quantity required")
		return
	}

	journal, err := h.svc.SettleTrade(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "SETTLE_FAILED", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, dto.SettleTradeResponse{
		JournalID: journal.ID,
		TradeID:   req.ID,
	})
}

func (h *LedgerHandler) handleDeposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	var req dto.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.UserID == "" || req.Asset == "" || req.Amount <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id, asset, and positive amount required")
		return
	}
	if req.Ref == "" {
		req.Ref = "deposit-" + req.UserID + "-" + req.Asset
	}

	if err := h.svc.Deposit(req.UserID, req.Asset, req.Amount, req.Ref); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "DEPOSIT_FAILED", err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, dto.ToBalanceResponse(h.svc.GetBalance(req.UserID, req.Asset)))
}

func (h *LedgerHandler) handleReserve(w http.ResponseWriter, r *http.Request) {
	h.handleBalanceMutation(w, r, func(req dto.ReserveRequest) error {
		return h.svc.Reserve(req.UserID, req.Asset, req.Amount)
	})
}

func (h *LedgerHandler) handleRelease(w http.ResponseWriter, r *http.Request) {
	h.handleBalanceMutation(w, r, func(req dto.ReserveRequest) error {
		return h.svc.Release(req.UserID, req.Asset, req.Amount)
	})
}

func (h *LedgerHandler) handleDebitLocked(w http.ResponseWriter, r *http.Request) {
	h.handleBalanceMutation(w, r, func(req dto.ReserveRequest) error {
		return h.svc.DebitLocked(req.UserID, req.Asset, req.Amount)
	})
}

func (h *LedgerHandler) handleBalanceMutation(w http.ResponseWriter, r *http.Request, fn func(dto.ReserveRequest) error) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	var req dto.ReserveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}
	if req.UserID == "" || req.Asset == "" || req.Amount <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id, asset, and positive amount required")
		return
	}
	if err := fn(req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "OPERATION_FAILED", err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToBalanceResponse(h.svc.GetBalance(req.UserID, req.Asset)))
}

func (h *LedgerHandler) handleLedger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, apiPrefix)
	parts := strings.Split(path, "/")

	switch parts[0] {
	case "balance":
		if len(parts) < 3 {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id and asset required")
			return
		}
		userID, asset := parts[1], parts[2]
		httputil.WriteJSON(w, http.StatusOK, dto.ToBalanceResponse(h.svc.GetBalance(userID, asset)))

	case "balances":
		if len(parts) < 2 {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id required")
			return
		}
		accounts := h.svc.GetBalances(parts[1])
		resp := make([]dto.BalanceResponse, len(accounts))
		for i, a := range accounts {
			resp[i] = dto.ToBalanceResponse(a)
		}
		httputil.WriteJSON(w, http.StatusOK, resp)

	case "journal":
		if len(parts) < 2 {
			httputil.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "trade_id required")
			return
		}
		journal, ok := h.svc.GetJournalByTrade(parts[1])
		if !ok {
			httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "journal not found")
			return
		}
		httputil.WriteJSON(w, http.StatusOK, dto.ToJournalResponse(journal))

	default:
		httputil.WriteError(w, http.StatusNotFound, "NOT_FOUND", "not found")
	}
}
