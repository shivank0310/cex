package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shivank0310/cex.git/wallet-service/internal/apperrors"
	"github.com/shivank0310/cex.git/wallet-service/internal/dto"
	"github.com/shivank0310/cex.git/wallet-service/internal/httputil"
	"github.com/shivank0310/cex.git/wallet-service/internal/service"
)

const apiPrefix = "/api/v1/wallet/"

type WalletHandler struct {
	svc *service.WalletService
}

func NewWalletHandler(svc *service.WalletService) *WalletHandler {
	return &WalletHandler{svc: svc}
}

func (h *WalletHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v3/account", h.handleAccount)
	mux.HandleFunc("/api/v1/wallet/address", h.handleCreateAddress)
	mux.HandleFunc("/api/v1/wallet/deposit/confirm", h.handleConfirmDeposit)
	mux.HandleFunc("/api/v1/wallet/withdraw", h.handleWithdraw)
	mux.HandleFunc("/api/v1/wallet/whitelist", h.handleWhitelist)
	mux.HandleFunc("/api/v1/wallet/", h.handleWallet)
}

func (h *WalletHandler) handleAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "user_id query param required"))
		return
	}
	account, err := h.svc.GetAccount(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	resp := dto.AccountResponse{
		CanTrade: account.CanTrade, CanWithdraw: account.CanWithdraw, CanDeposit: account.CanDeposit,
		UpdateTime: account.UpdateTime, AccountType: account.AccountType, Permissions: account.Permissions,
	}
	for _, b := range account.Balances {
		resp.Balances = append(resp.Balances, dto.AssetBalance{
			Asset: b.Asset, Free: b.Free, Locked: b.Locked,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *WalletHandler) handleCreateAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.CreateAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	wallet, err := h.svc.CreateDepositAddress(r.Context(), req.UserID, req.Asset, req.Chain)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, dto.ToWalletResponse(wallet))
}

func (h *WalletHandler) handleConfirmDeposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.ConfirmDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	deposit, err := h.svc.ConfirmDeposit(r.Context(), req.TxHash, req.ToAddress, req.Amount, req.Confirmations)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, dto.ToDepositResponse(deposit))
}

func (h *WalletHandler) handleWithdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	withdrawal, err := h.svc.RequestWithdrawal(r.Context(), req.UserID, req.Asset, req.Amount, req.ToAddress)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, dto.ToWithdrawalResponse(withdrawal))
}

func (h *WalletHandler) handleWhitelist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.WhitelistAddressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	h.svc.WhitelistAddress(req.UserID, req.Address)
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "whitelisted"})
}

func (h *WalletHandler) handleWallet(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, apiPrefix)
	parts := strings.Split(path, "/")

	switch parts[0] {
	case "balance":
		h.handleBalance(w, r, parts)
	case "withdrawal":
		h.handleWithdrawalActions(w, r, parts)
	default:
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "not found"))
	}
}

func (h *WalletHandler) handleBalance(w http.ResponseWriter, r *http.Request, parts []string) {
	if r.Method != http.MethodGet || len(parts) < 3 {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "user_id and asset required"))
		return
	}
	bal, err := h.svc.GetLedgerBalance(r.Context(), parts[1], parts[2])
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToBalanceResponse(bal))
}

func (h *WalletHandler) handleWithdrawalActions(w http.ResponseWriter, r *http.Request, parts []string) {
	if len(parts) < 2 {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "withdrawal id required"))
		return
	}
	withdrawalID := parts[1]

	// Sub-actions: /withdrawal/{id}/verify-mfa, /approve, /multisig-sign
	if len(parts) >= 3 {
		switch parts[2] {
		case "verify-mfa":
			h.handleVerifyMFA(w, r, withdrawalID)
			return
		case "approve":
			h.handleApprove(w, r, withdrawalID)
			return
		case "multisig-sign":
			h.handleMultisigSign(w, r, withdrawalID)
			return
		}
	}

	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	wd, err := h.svc.GetWithdrawal(withdrawalID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToWithdrawalResponse(wd))
}

func (h *WalletHandler) handleVerifyMFA(w http.ResponseWriter, r *http.Request, withdrawalID string) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.MFAVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	wd, err := h.svc.VerifyMFA(r.Context(), withdrawalID, req.UserID, req.Code)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToWithdrawalResponse(wd))
}

func (h *WalletHandler) handleApprove(w http.ResponseWriter, r *http.Request, withdrawalID string) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.ApproveWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	wd, err := h.svc.ApproveWithdrawal(r.Context(), withdrawalID, req.ApproverID, req.Note)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToWithdrawalResponse(wd))
}

func (h *WalletHandler) handleMultisigSign(w http.ResponseWriter, r *http.Request, withdrawalID string) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.MultisigSignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON body"))
		return
	}
	wd, err := h.svc.MultisigSign(r.Context(), withdrawalID, req.KeyID, req.SignerID)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToWithdrawalResponse(wd))
}
