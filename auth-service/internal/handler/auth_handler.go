package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shivank0310/cex.git/auth-service/internal/apperrors"
	"github.com/shivank0310/cex.git/auth-service/internal/dto"
	"github.com/shivank0310/cex.git/auth-service/internal/httputil"
	"github.com/shivank0310/cex.git/auth-service/internal/service"
	"github.com/shivank0310/cex.git/pkg/health"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(mux *http.ServeMux) {
	health.Register(mux)
	mux.HandleFunc("/api/v1/auth/register", h.handleRegister)
	mux.HandleFunc("/api/v1/auth/login", h.handleLogin)
	mux.HandleFunc("/api/v1/auth/refresh", h.handleRefresh)
	mux.HandleFunc("/api/v1/auth/logout", h.handleLogout)
	mux.HandleFunc("/api/v1/auth/me", h.handleMe)
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
		return
	}
	user, pair, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Username)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, dto.ToAuthResponse(user, pair))
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
		return
	}
	user, pair, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToAuthResponse(user, pair))
}

func (h *AuthHandler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
		return
	}
	user, pair, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToAuthResponse(user, pair))
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
		return
	}
	if err := h.svc.Logout(req.RefreshToken); err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h *AuthHandler) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	token := bearerToken(r)
	if token == "" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeUnauthorized, "missing bearer token"))
		return
	}
	user, err := h.svc.Me(r.Context(), token)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, dto.ToUser(user))
}

func bearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
