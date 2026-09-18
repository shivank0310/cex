package client

import (
	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/order-service/internal/model"
)

const (
	VenueInternal = "internal"
	VenueBinance  = "binance"
)

// VenueRouter dispatches orders to the internal matching engine or Binance adapter.
type VenueRouter struct {
	registry *model.Registry
	internal MatchingEngineClient
	binance  MatchingEngineClient
}

func NewVenueRouter(
	registry *model.Registry,
	internal MatchingEngineClient,
	binance MatchingEngineClient,
) *VenueRouter {
	return &VenueRouter{
		registry: registry,
		internal: internal,
		binance:  binance,
	}
}

func (r *VenueRouter) Submit(o *meapi.Order) meapi.SubmitResult {
	return r.route(o.Symbol).Submit(o)
}

func (r *VenueRouter) Cancel(symbol, orderID string) (*meapi.Order, error) {
	return r.route(symbol).Cancel(symbol, orderID)
}

func (r *VenueRouter) route(symbol string) MatchingEngineClient {
	sym, ok := r.registry.Get(symbol)
	if ok && sym.Venue == VenueBinance && r.binance != nil {
		return r.binance
	}
	return r.internal
}
