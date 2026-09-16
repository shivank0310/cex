package validator

import (
	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/model"
)

type SymbolValidator struct {
	registry *model.Registry
}

func NewSymbolValidator(registry *model.Registry) *SymbolValidator {
	return &SymbolValidator{registry: registry}
}

func (v *SymbolValidator) Validate(symbolName string) (model.Symbol, error) {
	if symbolName == "" {
		return model.Symbol{}, apperrors.New(apperrors.CodeInvalidSymbol, "symbol is required")
	}

	sym, ok := v.registry.Get(symbolName)
	if !ok {
		return model.Symbol{}, apperrors.New(apperrors.CodeInvalidSymbol, "symbol not supported")
	}
	if !sym.IsTradable() {
		return model.Symbol{}, apperrors.New(apperrors.CodeInvalidSymbol, "symbol is not tradable")
	}

	return sym, nil
}
