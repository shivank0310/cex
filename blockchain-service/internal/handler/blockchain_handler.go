package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shivank0310/cex.git/blockchain-service/internal/apperrors"
	"github.com/shivank0310/cex.git/blockchain-service/internal/dto"
	"github.com/shivank0310/cex.git/blockchain-service/internal/httputil"
	"github.com/shivank0310/cex.git/blockchain-service/internal/service"
)

const apiPrefix = "/api/v1/blockchain/"

type BlockchainHandler struct {
	svc *service.BlockchainService
}

func NewBlockchainHandler(svc *service.BlockchainService) *BlockchainHandler {
	return &BlockchainHandler{svc: svc}
}

func (h *BlockchainHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/v1/blockchain/address", h.handleCreateAddress)
	mux.HandleFunc("/api/v1/blockchain/address/validate", h.handleValidateAddress)
	mux.HandleFunc("/api/v1/blockchain/withdraw", h.handleWithdraw)
	mux.HandleFunc("/api/v1/blockchain/", h.handleBlockchain)
}

func (h *BlockchainHandler) handleCreateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.CreateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	wallet, err := h.svc.CreateAddress(r.Context(), req.UserID, req.Asset, req.Chain)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, dto.ToWalletResponse(wallet))
}

func (h *BlockchainHandler) handleValidateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.ValidateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	if err := h.svc.ValidateAddress(r.Context(), req.Address); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"valid": true})
}

func (h *BlockchainHandler) handleWithdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	tx, err := h.svc.BroadcastWithdrawal(r.Context(), req.Asset, req.ToAddress, req.Amount)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, dto.ToTransactionResponse(tx))
}

func (h *BlockchainHandler) handleBlockchain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}

	path := strings.TrimPrefix(r.URL.Path, apiPrefix)
	parts := strings.Split(path, "/")

	if parts[0] == "transaction" && len(parts) >= 2 {
		tx, err := h.svc.GetTransaction(r.Context(), parts[1])
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, dto.ToTransactionResponse(tx))
		return
	}

	httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "not found"))
}
