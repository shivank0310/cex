package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/shivank0310/cex.git/ledger-service/internal/dto"
)

func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, dto.ErrorResponse{Code: code, Message: message})
}
