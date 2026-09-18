package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/shivank0310/cex.git/pkg/health"
	"github.com/shivank0310/cex.git/user-service/internal/apperrors"
	"github.com/shivank0310/cex.git/user-service/internal/dto"
	"github.com/shivank0310/cex.git/user-service/internal/httputil"
	"github.com/shivank0310/cex.git/user-service/internal/service"
)

const apiPrefix = "/api/v1/users/"

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Register(mux *http.ServeMux) {
	health.Register(mux)
	mux.HandleFunc("/api/v1/users", h.handleCreate)
	mux.HandleFunc("/api/v1/users/", h.handleUsers)
}

func (h *UserHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/users" || r.Method != http.MethodPost {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
		return
	}
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
		return
	}
	user, err := h.svc.Create(r.Context(), req.Email, req.Username)
	if err != nil {
		httputil.WriteError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, dto.ToUser(user))
}

func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, apiPrefix)
	if path == "" {
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "not found"))
		return
	}

	if strings.HasPrefix(path, "by-email/") && r.Method == http.MethodGet {
		email := strings.TrimPrefix(path, "by-email/")
		email = strings.ReplaceAll(email, "%40", "@")
		user, err := h.svc.GetByEmail(r.Context(), email)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, dto.ToUser(user))
		return
	}

	parts := strings.Split(path, "/")
	id := parts[0]

	switch r.Method {
	case http.MethodGet:
		user, err := h.svc.GetByID(r.Context(), id)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, dto.ToUser(user))
	case http.MethodPatch:
		var req dto.UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "invalid JSON"))
			return
		}
		user, err := h.svc.Update(r.Context(), id, req.Username, req.Status, req.KYCStatus)
		if err != nil {
			httputil.WriteError(w, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, dto.ToUser(user))
	default:
		httputil.WriteError(w, apperrors.New(apperrors.CodeInvalidRequest, "method not allowed"))
	}
}
