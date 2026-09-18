package tests

import (
	"testing"

	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
	"github.com/shivank0310/cex.git/order-service/internal/client"
	"github.com/shivank0310/cex.git/order-service/internal/model"
)

type stubEngine struct {
	venue string
}

func (s *stubEngine) Submit(o *meapi.Order) meapi.SubmitResult {
	o.Status = meapi.StatusNew
	return meapi.SubmitResult{Order: o}
}

func (s *stubEngine) Cancel(_ string, orderID string) (*meapi.Order, error) {
	return &meapi.Order{ID: orderID, Status: meapi.StatusCancelled}, nil
}

func TestVenueRouterInternal(t *testing.T) {
	registry := model.NewRegistry()
	registry.Register(model.Symbol{Name: "BTC/USDT", Venue: model.VenueInternal, Active: true})

	internal := &stubEngine{venue: "internal"}
	binance := &stubEngine{venue: "binance"}
	router := client.NewVenueRouter(registry, internal, binance)

	o := meapi.NewLimitOrder("ORD-1", "user-a", "BTC/USDT", meapi.Buy, 101100, 30)
	result := router.Submit(o)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}

func TestVenueRouterBinance(t *testing.T) {
	registry := model.NewRegistry()
	registry.Register(model.Symbol{Name: "BTC/USDT", Venue: model.VenueBinance, Active: true})

	internal := &stubEngine{venue: "internal"}
	binance := &stubEngine{venue: "binance"}
	router := client.NewVenueRouter(registry, internal, binance)

	o := meapi.NewLimitOrder("ORD-2", "user-a", "BTC/USDT", meapi.Buy, 101100, 30)
	result := router.Submit(o)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
}
