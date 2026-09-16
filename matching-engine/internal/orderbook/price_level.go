package orderbook

import "github.com/shivank0310/cex.git/matching-engine/internal/order"

// PriceLevel holds all resting orders at a single price (FIFO = price-time priority).
type PriceLevel struct {
	Price  int64
	Orders []*order.Order
}

func newPriceLevel(price int64) *PriceLevel {
	return &PriceLevel{
		Price:  price,
		Orders: make([]*order.Order, 0),
	}
}

func (pl *PriceLevel) Add(o *order.Order) {
	pl.Orders = append(pl.Orders, o)
}

func (pl *PriceLevel) RemoveAt(index int) {
	pl.Orders = append(pl.Orders[:index], pl.Orders[index+1:]...)
}

func (pl *PriceLevel) IsEmpty() bool {
	return len(pl.Orders) == 0
}

func (pl *PriceLevel) TotalQuantity() int64 {
	var total int64
	for _, o := range pl.Orders {
		total += o.Remaining
	}
	return total
}
