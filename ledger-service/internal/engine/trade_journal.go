package engine

import (
	"fmt"
	"time"

	"github.com/shivank0310/cex.git/ledger-service/internal/model"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
	"github.com/shivank0310/cex.git/pkg/events"
)

const exchangeFeeAccount = "exchange:fees"

// BuildTradeJournal creates a balanced double-entry journal from a trade event.
//
// Example: Alice buys 0.1 BTC @ 10000 USDT (qty=10, scale=100)
//
//	BTC legs:   DEBIT alice BTC 10  |  CREDIT bob BTC 10
//	USDT legs:  DEBIT bob USDT 1000 |  CREDIT alice USDT 1000
func BuildTradeJournal(trade events.TradePayload, journalID string) (model.Journal, error) {
	pair, err := model.ParseSymbol(trade.Symbol)
	if err != nil {
		return model.Journal{}, err
	}

	notional := decimal.Notional(trade.Price, trade.Quantity)
	buyerFee, sellerFee := assignFees(trade)

	legs := []model.Leg{
		// BTC: buyer receives (minus fee), seller gives
		leg(journalID, trade.BuyerID, pair.Base, model.Debit, trade.Quantity-buyerFee, "trade-receive-base"),
		leg(journalID, trade.SellerID, pair.Base, model.Credit, trade.Quantity, "trade-send-base"),

		// USDT: seller receives (minus fee), buyer pays
		leg(journalID, trade.SellerID, pair.Quote, model.Debit, notional-sellerFee, "trade-receive-quote"),
		leg(journalID, trade.BuyerID, pair.Quote, model.Credit, notional, "trade-send-quote"),
	}

	if buyerFee > 0 {
		legs = append(legs,
			leg(journalID, exchangeFeeAccount, pair.Base, model.Debit, buyerFee, "trade-fee-base"),
		)
	}
	if sellerFee > 0 {
		legs = append(legs,
			leg(journalID, exchangeFeeAccount, pair.Quote, model.Debit, sellerFee, "trade-fee-quote"),
		)
	}

	journal := model.Journal{
		ID:        journalID,
		Type:      model.JournalTrade,
		Reference: trade.ID,
		Symbol:    trade.Symbol,
		Legs:      legs,
		PostedAt:  trade.Timestamp.UTC(),
	}
	if journal.PostedAt.IsZero() {
		journal.PostedAt = time.Now().UTC()
	}

	if err := journal.Validate(); err != nil {
		return model.Journal{}, fmt.Errorf("trade journal unbalanced: %w", err)
	}
	return journal, nil
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

func leg(journalID, userID, asset string, t model.EntryType, amount int64, ref string) model.Leg {
	return model.Leg{
		JournalID: journalID,
		UserID:    userID,
		Asset:     asset,
		Type:      t,
		Amount:    amount,
		Reference: ref,
	}
}

// BuildDepositJournal creates a journal for crediting a user account.
func BuildDepositJournal(journalID, userID, asset string, amount int64, ref string) model.Journal {
	return model.Journal{
		ID:        journalID,
		Type:      model.JournalDeposit,
		Reference: ref,
		Legs: []model.Leg{
			leg(journalID, userID, asset, model.Debit, amount, "deposit"),
			leg(journalID, "exchange:treasury", asset, model.Credit, amount, "deposit-source"),
		},
		PostedAt: time.Now().UTC(),
	}
}
