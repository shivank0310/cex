package engine

import (
	"github.com/shivank0310/cex.git/matching-engine/internal/order"
	"github.com/shivank0310/cex.git/matching-engine/internal/orderbook"
	"github.com/shivank0310/cex.git/matching-engine/internal/trade"
)

// match executes the incoming order against the opposite side of the book.
// Trades execute at the maker's price (price-time priority).
func (e *Engine) match(incoming *order.Order, book *orderbook.OrderBook, cfg SymbolConfig) []*trade.Trade {
	trades := make([]*trade.Trade, 0)

	for incoming.Remaining > 0 {
		var level *orderbook.PriceLevel
		var makerSide order.Side

		switch incoming.Side {
		case order.Buy:
			level = book.BestAsk()
			makerSide = order.Sell
		case order.Sell:
			level = book.BestBid()
			makerSide = order.Buy
		default:
			return trades
		}

		if level == nil {
			break
		}

		if !incoming.CanMatch(level.Price) {
			break
		}

		maker := level.Orders[0]
		fillQty := min(incoming.Remaining, maker.Remaining)
		tradePrice := maker.Price // execute at maker price

		t := e.newTrade(incoming, maker, tradePrice, fillQty, cfg)
		trades = append(trades, t)

		incoming.Fill(fillQty)
		maker.Fill(fillQty)

		if !maker.IsActive() {
			book.RemoveFilledAt(makerSide, level.Price)
		}
	}

	return trades
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
