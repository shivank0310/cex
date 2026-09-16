package validator

import (
	"fmt"

	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/model"
)

type QuantityValidator struct{}

func NewQuantityValidator() *QuantityValidator {
	return &QuantityValidator{}
}

func (v *QuantityValidator) Validate(sym model.Symbol, quantity int64) error {
	if quantity <= 0 {
		return apperrors.New(apperrors.CodeInvalidQuantity, "quantity must be greater than zero")
	}
	if quantity < sym.MinQuantity {
		return apperrors.New(apperrors.CodeInvalidQuantity, fmt.Sprintf("quantity below minimum %d", sym.MinQuantity))
	}
	if quantity > sym.MaxQuantity {
		return apperrors.New(apperrors.CodeInvalidQuantity, fmt.Sprintf("quantity above maximum %d", sym.MaxQuantity))
	}
	if sym.LotSize > 0 && quantity%sym.LotSize != 0 {
		return apperrors.New(apperrors.CodeInvalidQuantity, fmt.Sprintf("quantity must be a multiple of lot size %d", sym.LotSize))
	}
	return nil
}
