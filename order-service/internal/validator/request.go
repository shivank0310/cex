package validator

import (
	"strings"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/dto"
)

// RequestValidator normalizes and validates basic request fields.
type RequestValidator struct{}

func NewRequestValidator() *RequestValidator {
	return &RequestValidator{}
}

func (v *RequestValidator) ValidatePlace(req dto.PlaceOrderRequest) (meapi.Side, meapi.OrderType, error) {
	if strings.TrimSpace(req.Symbol) == "" {
		return "", "", apperrors.New(apperrors.CodeInvalidRequest, "symbol is required")
	}

	side, err := parseSide(req.Side)
	if err != nil {
		return "", "", err
	}

	orderType, err := parseOrderType(req.Type)
	if err != nil {
		return "", "", err
	}

	return side, orderType, nil
}

func parseSide(raw string) (meapi.Side, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "BUY":
		return meapi.Buy, nil
	case "SELL":
		return meapi.Sell, nil
	default:
		return "", apperrors.New(apperrors.CodeInvalidRequest, "side must be BUY or SELL")
	}
}

func parseOrderType(raw string) (meapi.OrderType, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "LIMIT":
		return meapi.Limit, nil
	case "MARKET":
		return meapi.Market, nil
	default:
		return "", apperrors.New(apperrors.CodeInvalidRequest, "type must be LIMIT or MARKET")
	}
}
