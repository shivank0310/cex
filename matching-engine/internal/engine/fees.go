package engine

import (
	"github.com/shivank0310/cex.git/matching-engine/internal/order"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
)

// FeeConfig defines maker/taker fees in basis points per symbol.
type FeeConfig struct {
	MakerBasisPoints int64
	TakerBasisPoints int64
}

func DefaultFeeConfig() FeeConfig {
	return FeeConfig{
		MakerBasisPoints: 10,  // 0.10%
		TakerBasisPoints: 20,  // 0.20%
	}
}

type tradeFees struct {
	makerFee int64
	takerFee int64
}

func computeFees(price, quantity int64, makerSide order.Side, cfg FeeConfig) tradeFees {
	notional := decimal.Notional(price, quantity)

	// Buyer pays fee in base (BTC); seller pays fee in quote (USDT).
	var makerFee, takerFee int64
	if makerSide == order.Sell {
		makerFee = decimal.Fee(notional, cfg.MakerBasisPoints)
		takerFee = decimal.Fee(notional, cfg.TakerBasisPoints)
	} else {
		makerFee = decimal.Fee(quantity, cfg.MakerBasisPoints)
		takerFee = decimal.Fee(quantity, cfg.TakerBasisPoints)
	}

	return tradeFees{makerFee: makerFee, takerFee: takerFee}
}
