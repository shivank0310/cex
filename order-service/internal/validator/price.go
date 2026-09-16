package validator

import (
	"fmt"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/model"
)

type PriceValidator struct{}

func NewPriceValidator() *PriceValidator {
	return &PriceValidator{}
}

func (v *PriceValidator) Validate(sym model.Symbol, orderType meapi.OrderType, price int64) error {
	if orderType == meapi.Market {
		if price != 0 {
			return apperrors.New(apperrors.CodeInvalidPrice, "market orders must not include price")
		}
		return nil
	}

	if price <= 0 {
		return apperrors.New(apperrors.CodeInvalidPrice, "price must be greater than zero")
	}
	if price < sym.MinPrice {
		return apperrors.New(apperrors.CodeInvalidPrice, fmt.Sprintf("price below minimum %d", sym.MinPrice))
	}
	if price > sym.MaxPrice {
		return apperrors.New(apperrors.CodeInvalidPrice, fmt.Sprintf("price above maximum %d", sym.MaxPrice))
	}
	if sym.TickSize > 0 && price%sym.TickSize != 0 {
		return apperrors.New(apperrors.CodeInvalidPrice, fmt.Sprintf("price must be a multiple of tick size %d", sym.TickSize))
	}

	return nil
}
