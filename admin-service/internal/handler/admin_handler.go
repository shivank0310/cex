package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shivank0310/cex.git/admin-service/internal/apperrors"
	"github.com/shivank0310/cex.git/admin-service/internal/dto"
	"github.com/shivank0310/cex.git/admin-service/internal/httputil"
	"github.com/shivank0310/cex.git/admin-service/internal/model"
	"github.com/shivank0310/cex.git/admin-service/internal/service"
	"github.com/shivank0310/cex.git/pkg/health"
)

const apiPrefix = "/api/v1/admin/"

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) Register(mux *http.ServeMux) {
	health.Register(mux)
	mux.HandleFunc("/api/v1/admin/dashboard", h.handleDashboard)
	mux.HandleFunc("/api/v1/admin/system/status", h.handleSystemStatus)
	mux.HandleFunc("/api/v1/admin/users", h.handleUsers)
	mux.HandleFunc("/api/v1/admin/markets", h.handleMarkets)
	mux.HandleFunc("/api/v1/admin/fees/", h.handleFees)
	mux.HandleFunc("/api/v1/admin/deposits", h.handleDeposits)
	mux.HandleFunc("/api/v1/admin/withdrawals", h.handleWithdrawals)
	mux.HandleFunc("/api/v1/admin/withdrawals/", h.handleWithdrawalAction)
	mux.HandleFunc("/api/v1/admin/kyc", h.handleKYC)
	mux.HandleFunc("/api/v1/admin/kyc/", h.handleKYCAction)
	mux.HandleFunc("/api/v1/admin/risk/rules", h.handleRiskRules)
	mux.HandleFunc("/api/v1/admin/risk/rules/", h.handleRiskRuleUpdate)
	mux.HandleFunc("/api/v1/admin/", h.handleAdmin)
}

func (h *AdminHandler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToDashboard(h.svc.GetDashboard()))
}

func (h *AdminHandler) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	components, overall := h.svc.GetSystemStatus(r.Context())
	resp := dto.SystemStatusResponse{Overall: overall}
	for _, c := range components {
		resp.Components = append(resp.Components, dto.ToComponent(c))
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AdminHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	users := h.svc.ListUsers()
	resp := make([]dto.UserResponse, len(users))
	for i, u := range users {
		resp[i] = dto.ToUser(u)
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AdminHandler) handleMarkets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		pairs := h.svc.ListMarkets()
		resp := make([]dto.TradingPairResponse, len(pairs))
		for i, p := range pairs {
			resp[i] = dto.ToPair(p)
		}
		httputil.WriteJSON(w, http.StatusOK, resp)
	case http.MethodPost:
		var req dto.UpsertPairRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
			return
		}
		pair, err := h.svc.UpsertMarket(model.TradingPair{
			Symbol: req.Symbol, BaseAsset: req.BaseAsset, QuoteAsset: req.QuoteAsset,
			Active: req.Active, MakerFeeBPS: req.MakerFeeBPS, TakerFeeBPS: req.TakerFeeBPS,
			MinQuantity: req.MinQuantity, MaxQuantity: req.MaxQuantity, MinNotional: req.MinNotional,
		})
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusCreated, dto.ToPair(pair))
	default:
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
	}
}

func (h *AdminHandler) handleFees(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	symbol := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/fees/")
	if symbol == "" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "symbol required"))
		return
	}
	symbol = strings.ReplaceAll(symbol, "%2F", "/")
	var req dto.UpdateFeesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
		return
	}
	pair, err := h.svc.UpdateFees(symbol, req.MakerFeeBPS, req.TakerFeeBPS)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToPair(pair))
}

func (h *AdminHandler) handleDeposits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	deps := h.svc.ListDeposits()
	resp := make([]dto.DepositResponse, len(deps))
	for i, d := range deps {
		resp[i] = dto.ToDeposit(d)
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AdminHandler) handleWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/api/v1/admin/withdrawals" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	wds := h.svc.ListWithdrawals()
	resp := make([]dto.WithdrawalResponse, len(wds))
	for i, wd := range wds {
		resp[i] = dto.ToWithdrawal(wd)
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AdminHandler) handleWithdrawalAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/withdrawals/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "withdrawal id and action required"))
		return
	}
	id, action := parts[0], parts[1]
	var wd model.WithdrawalRecord
	var err error
	switch action {
	case "approve":
		wd, err = h.svc.ApproveWithdrawal(id)
	case "reject":
		wd, err = h.svc.RejectWithdrawal(id)
	default:
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "unknown action"))
		return
	}
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToWithdrawal(wd))
}

func (h *AdminHandler) handleKYC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/api/v1/admin/kyc" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	apps := h.svc.ListKYC()
	resp := make([]dto.KYCResponse, len(apps))
	for i, k := range apps {
		resp[i] = dto.ToKYC(k)
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AdminHandler) handleKYCAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/kyc/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "kyc id and action required"))
		return
	}
	var req dto.KYCReviewRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	id, action := parts[0], parts[1]
	var kyc model.KYCApplication
	var err error
	switch action {
	case "approve":
		kyc, err = h.svc.ApproveKYC(id, req.Note)
	case "reject":
		kyc, err = h.svc.RejectKYC(id, req.Note)
	default:
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "unknown action"))
		return
	}
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToKYC(kyc))
}

func (h *AdminHandler) handleRiskRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	rules := h.svc.ListRiskRules()
	resp := make([]dto.RiskRuleResponse, len(rules))
	for i, rule := range rules {
		resp[i] = dto.ToRiskRule(rule)
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AdminHandler) handleRiskRuleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/risk/rules/")
	if id == "" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "rule id required"))
		return
	}
	var req dto.UpdateRiskRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
		return
	}
	rule, err := h.svc.UpdateRiskRule(id, req.Threshold, req.Enabled)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToRiskRule(rule))
}

func (h *AdminHandler) handleAdmin(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, apiPrefix)
	parts := strings.Split(path, "/")

	if parts[0] == "users" && len(parts) >= 2 {
		if r.Method == http.MethodGet {
			u, err := h.svc.GetUser(parts[1])
			if err != nil {
				httputil.WriteError(w, err)
				return
			}
			httputil.WriteJSON(w, http.StatusOK, dto.ToUser(u))
			return
		}
		if r.Method == http.MethodPatch {
			var req dto.UpdateUserStatusRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
				return
			}
			if err := h.svc.UpdateUserStatus(parts[1], req.Status); err != nil {
				httputil.WriteError(w, err)
				return
			}
			u, _ := h.svc.GetUser(parts[1])
			httputil.WriteJSON(w, http.StatusOK, dto.ToUser(u))
			return
		}
	}

	httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "not found"))
}
