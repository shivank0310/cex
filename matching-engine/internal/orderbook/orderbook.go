package orderbook

import (
	"fmt"
	"sort"

	"github.com/shivank0310/cex.git/matching-engine/internal/order"
)

type orderLocation struct {
	side  order.Side
	price int64
	index int
}

// OrderBook maintains resting limit orders with price-time priority.
type OrderBook struct {
	Symbol string

	bidLevels []*PriceLevel // sorted descending by price
	askLevels []*PriceLevel // sorted ascending by price
	index     map[string]*orderLocation
}

func New(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:    symbol,
		bidLevels: make([]*PriceLevel, 0),
		askLevels: make([]*PriceLevel, 0),
		index:     make(map[string]*orderLocation),
	}
}

func (ob *OrderBook) Add(o *order.Order) {
	if o.Type != order.Limit {
		panic("only limit orders can be added to the book")
	}

	switch o.Side {
	case order.Buy:
		ob.addToSide(o, &ob.bidLevels, false)
	case order.Sell:
		ob.addToSide(o, &ob.askLevels, true)
	default:
		panic(fmt.Sprintf("unknown side: %s", o.Side))
	}
}

func (ob *OrderBook) addToSide(o *order.Order, levels *[]*PriceLevel, ascending bool) {
	idx := ob.findLevel(*levels, o.Price)
	if idx < 0 {
		pl := newPriceLevel(o.Price)
		pl.Add(o)
		*levels = append(*levels, pl)
		ob.sortLevels(*levels, ascending)
		idx = ob.findLevel(*levels, o.Price)
	} else {
		(*levels)[idx].Add(o)
	}

	orderIdx := len((*levels)[idx].Orders) - 1
	ob.index[o.ID] = &orderLocation{side: o.Side, price: o.Price, index: orderIdx}
}

func (ob *OrderBook) findLevel(levels []*PriceLevel, price int64) int {
	for i, pl := range levels {
		if pl.Price == price {
			return i
		}
	}
	return -1
}

func (ob *OrderBook) sortLevels(levels []*PriceLevel, ascending bool) {
	sort.Slice(levels, func(i, j int) bool {
		if ascending {
			return levels[i].Price < levels[j].Price
		}
		return levels[i].Price > levels[j].Price
	})
}

// BestAsk returns the lowest sell price level, or nil if empty.
func (ob *OrderBook) BestAsk() *PriceLevel {
	if len(ob.askLevels) == 0 {
		return nil
	}
	return ob.askLevels[0]
}

// BestBid returns the highest buy price level, or nil if empty.
func (ob *OrderBook) BestBid() *PriceLevel {
	if len(ob.bidLevels) == 0 {
		return nil
	}
	return ob.bidLevels[0]
}

// RemoveOrder removes a resting order by ID. Returns the removed order or nil.
func (ob *OrderBook) RemoveOrder(orderID string) *order.Order {
	loc, ok := ob.index[orderID]
	if !ok {
		return nil
	}

	levels := ob.bidLevels
	if loc.side == order.Sell {
		levels = ob.askLevels
	}

	levelIdx := ob.findLevel(levels, loc.price)
	if levelIdx < 0 {
		delete(ob.index, orderID)
		return nil
	}

	pl := levels[levelIdx]
	if loc.index >= len(pl.Orders) || pl.Orders[loc.index].ID != orderID {
		ob.rebuildIndex(loc.side, levels)
		return ob.RemoveOrder(orderID)
	}

	removed := pl.Orders[loc.index]
	pl.RemoveAt(loc.index)
	delete(ob.index, orderID)

	if pl.IsEmpty() {
		levels = append(levels[:levelIdx], levels[levelIdx+1:]...)
		if loc.side == order.Buy {
			ob.bidLevels = levels
		} else {
			ob.askLevels = levels
		}
	} else {
		ob.reindexLevel(loc.side, loc.price, pl)
	}

	return removed
}

func (ob *OrderBook) rebuildIndex(side order.Side, levels []*PriceLevel) {
	for id := range ob.index {
		if ob.index[id].side == side {
			delete(ob.index, id)
		}
	}
	for _, pl := range levels {
		ob.reindexLevel(side, pl.Price, pl)
	}
}

func (ob *OrderBook) reindexLevel(side order.Side, price int64, pl *PriceLevel) {
	for i, o := range pl.Orders {
		ob.index[o.ID] = &orderLocation{side: side, price: price, index: i}
	}
}

// RemoveFilledAt removes the front order from a price level after a full fill.
func (ob *OrderBook) RemoveFilledAt(side order.Side, price int64) {
	levels := ob.bidLevels
	if side == order.Sell {
		levels = ob.askLevels
	}

	levelIdx := ob.findLevel(levels, price)
	if levelIdx < 0 {
		return
	}

	pl := levels[levelIdx]
	if len(pl.Orders) == 0 {
		return
	}

	front := pl.Orders[0]
	delete(ob.index, front.ID)
	pl.RemoveAt(0)

	if pl.IsEmpty() {
		levels = append(levels[:levelIdx], levels[levelIdx+1:]...)
		if side == order.Buy {
			ob.bidLevels = levels
		} else {
			ob.askLevels = levels
		}
	} else {
		ob.reindexLevel(side, price, pl)
	}
}

// Snapshot returns aggregated depth for display.
type DepthLevel struct {
	Price    int64
	Quantity int64
}

func (ob *OrderBook) Bids() []DepthLevel {
	out := make([]DepthLevel, 0, len(ob.bidLevels))
	for _, pl := range ob.bidLevels {
		out = append(out, DepthLevel{Price: pl.Price, Quantity: pl.TotalQuantity()})
	}
	return out
}

func (ob *OrderBook) Asks() []DepthLevel {
	out := make([]DepthLevel, 0, len(ob.askLevels))
	for _, pl := range ob.askLevels {
		out = append(out, DepthLevel{Price: pl.Price, Quantity: pl.TotalQuantity()})
	}
	return out
}
