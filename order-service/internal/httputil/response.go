package httputil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/dto"
)

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, err error) {
	appErr := toAppError(err)
	WriteJSON(w, statusForCode(appErr.Code), dto.ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Message,
	})
}

func toAppError(err error) *apperrors.AppError {
	if err == nil {
		return apperrors.New(apperrors.CodeInternal, "unknown error")
	}
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperrors.Wrap(apperrors.CodeInternal, err.Error(), err)
}

func statusForCode(code apperrors.Code) int {
	switch code {
	case apperrors.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperrors.CodeInvalidRequest, apperrors.CodeInvalidSymbol, apperrors.CodeInvalidPrice, apperrors.CodeInvalidQuantity:
		return http.StatusBadRequest
	case apperrors.CodeTradingRuleViolation:
		return http.StatusUnprocessableEntity
	case apperrors.CodeInsufficientBalance:
		return http.StatusConflict
	case apperrors.CodeOrderNotFound:
		return http.StatusNotFound
	case apperrors.CodeEngineError:
		return http.StatusBadGateway
	case apperrors.CodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
