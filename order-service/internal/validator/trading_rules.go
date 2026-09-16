package validator

import (
	"fmt"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/model"
)

type TradingRulesValidator struct{}

func NewTradingRulesValidator() *TradingRulesValidator {
	return &TradingRulesValidator{}
}

func (v *TradingRulesValidator) Validate(sym model.Symbol, side meapi.Side, orderType meapi.OrderType, price, quantity int64) error {
	if orderType == meapi.Market && side == meapi.Buy {
		return apperrors.New(apperrors.CodeTradingRuleViolation, "market buy orders are not supported yet")
	}

	if orderType == meapi.Limit {
		notional := decimal.Notional(price, quantity)
		if notional < sym.MinNotional {
			return apperrors.New(
				apperrors.CodeTradingRuleViolation,
				fmt.Sprintf("order notional %d below minimum %d", notional, sym.MinNotional),
			)
		}
	}

	return nil
}
