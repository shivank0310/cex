package validator

import (
	"fmt"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/order-service/internal/apperrors"
	"github.com/shivank0310/cex.git/order-service/internal/client"
	"github.com/shivank0310/cex.git/order-service/internal/model"
)

type BalanceValidator struct {
	wallet client.WalletClient
}

func NewBalanceValidator(wallet client.WalletClient) *BalanceValidator {
	return &BalanceValidator{wallet: wallet}
}

func (v *BalanceValidator) Validate(userID string, sym model.Symbol, side meapi.Side, orderType meapi.OrderType, price, quantity int64) error {
	var asset string
	var required int64

	switch side {
	case meapi.Buy:
		if orderType == meapi.Market {
			return apperrors.New(apperrors.CodeTradingRuleViolation, "market buy balance check not supported")
		}
		asset = sym.QuoteAsset
		required = decimal.Notional(price, quantity)
	case meapi.Sell:
		asset = sym.BaseAsset
		required = quantity
	default:
		return apperrors.New(apperrors.CodeInvalidRequest, fmt.Sprintf("unknown side: %s", side))
	}

	available, err := v.wallet.GetAvailable(userID, asset)
	if err != nil {
		return apperrors.Wrap(apperrors.CodeInternal, "failed to fetch balance", err)
	}
	if available < required {
		return apperrors.New(
			apperrors.CodeInsufficientBalance,
			fmt.Sprintf("insufficient %s: need %d, available %d", asset, required, available),
		)
	}

	return nil
}
