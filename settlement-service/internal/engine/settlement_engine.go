package engine

import (
	"time"

	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/pkg/events"
	"github.com/shivank0310/cex.git/settlement-service/internal/model"
)

// Compute builds the settlement transfer amounts from a trade event.
//
// Buyer gets base asset (minus fee), seller gets quote currency (minus fee).
func Compute(trade events.TradePayload) (model.Settlement, error) {
	pair, err := model.ParseSymbol(trade.Symbol)
	if err != nil {
		return model.Settlement{}, err
	}

	notional := decimal.Notional(trade.Price, trade.Quantity)
	buyerFee, sellerFee := assignFees(trade)

	settledAt := trade.Timestamp.UTC()
	if settledAt.IsZero() {
		settledAt = time.Now().UTC()
	}

	return model.Settlement{
		TradeID:           trade.ID,
		Symbol:            trade.Symbol,
		BaseAsset:         pair.Base,
		QuoteAsset:        pair.Quote,
		BuyerID:           trade.BuyerID,
		SellerID:          trade.SellerID,
		Price:             trade.Price,
		Quantity:          trade.Quantity,
		Notional:          notional,
		BuyerFee:          buyerFee,
		SellerFee:         sellerFee,
		BuyerBaseCredit:   trade.Quantity - buyerFee,
		SellerQuoteCredit: notional - sellerFee,
		Status:            model.StatusCompleted,
		SettledAt:         settledAt,
	}, nil
}

func assignFees(trade events.TradePayload) (buyerFee, sellerFee int64) {
	if trade.MakerOrderID == trade.BuyOrderID {
		buyerFee = trade.MakerFee
		sellerFee = trade.TakerFee
	} else {
		buyerFee = trade.TakerFee
		sellerFee = trade.MakerFee
	}
	return buyerFee, sellerFee
}
