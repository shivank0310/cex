package client

import (
	meapi "github.com/shivank0310/cex.git/matching-engine/pkg/api"
)

// MatchingEngineClient submits and cancels orders on the matching engine.
type MatchingEngineClient interface {
	Submit(o *meapi.Order) meapi.SubmitResult
	Cancel(symbol, orderID string) (*meapi.Order, error)
}

// EngineClient adapts the public matching-engine API for the order service.
type EngineClient struct {
	engine *meapi.Engine
}

func NewEngineClient(eng *meapi.Engine) *EngineClient {
	return &EngineClient{engine: eng}
}

func (c *EngineClient) Submit(o *meapi.Order) meapi.SubmitResult {
	return c.engine.SubmitOrder(o)
}

func (c *EngineClient) Cancel(symbol, orderID string) (*meapi.Order, error) {
	return c.engine.CancelOrder(symbol, orderID)
}
