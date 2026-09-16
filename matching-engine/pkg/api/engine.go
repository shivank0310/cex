package api

import (
	"context"

	"github.com/shivank0310/cex.git/matching-engine/internal/engine"
	"github.com/shivank0310/cex.git/pkg/events"
)

// SymbolConfig maps a trading pair to its base/quote assets.
type SymbolConfig struct {
	BaseAsset  string
	QuoteAsset string
}

// FeeConfig defines maker/taker fees in basis points.
type FeeConfig struct {
	MakerBasisPoints int64
	TakerBasisPoints int64
}

func DefaultFeeConfig() FeeConfig {
	cfg := engine.DefaultFeeConfig()
	return FeeConfig{
		MakerBasisPoints: cfg.MakerBasisPoints,
		TakerBasisPoints: cfg.TakerBasisPoints,
	}
}

// SubmitResult is the outcome of submitting an order to the engine.
type SubmitResult struct {
	Order  *Order
	Trades []*Trade
	Error  error
}

// Engine is the public matching-engine facade for other services.
type Engine struct {
	inner   *engine.Engine
	ledger  *Ledger
	emitter *EventEmitter
}

func NewEngine(ledger *Ledger, fees FeeConfig) *Engine {
	inner := engine.New(ledger.inner, engine.FeeConfig{
		MakerBasisPoints: fees.MakerBasisPoints,
		TakerBasisPoints: fees.TakerBasisPoints,
	})
	e := &Engine{inner: inner, ledger: ledger}
	SetBookProvider(inner.GetBook)
	return e
}

// NewEngineWithPublisher creates an engine that publishes events to Kafka.
func NewEngineWithPublisher(ledger *Ledger, fees FeeConfig, publisher events.Publisher) *Engine {
	e := NewEngine(ledger, fees)
	e.emitter = NewEventEmitter(publisher)
	return e
}

func (e *Engine) RegisterSymbol(symbol string, cfg SymbolConfig) {
	e.inner.RegisterSymbol(symbol, engine.SymbolConfig{
		BaseAsset:  cfg.BaseAsset,
		QuoteAsset: cfg.QuoteAsset,
	})
}

func (e *Engine) SubmitOrder(o *Order) SubmitResult {
	res := e.inner.SubmitOrder(ToOrder(o))
	trades := make([]*Trade, 0, len(res.Trades))
	for _, t := range res.Trades {
		trades = append(trades, FromTrade(t))
	}
	result := SubmitResult{
		Order:  FromOrder(res.Order),
		Trades: trades,
		Error:  res.Error,
	}
	if e.emitter != nil && result.Error == nil {
		_ = e.emitter.EmitSubmitResult(context.Background(), result)
	}
	return result
}

func (e *Engine) CancelOrder(symbol, orderID string) (*Order, error) {
	o, err := e.inner.CancelOrder(symbol, orderID)
	if err != nil {
		return nil, err
	}
	order := FromOrder(o)
	if e.emitter != nil {
		_ = e.emitter.EmitCancel(context.Background(), order)
	}
	return order, nil
}

func (e *Engine) Ledger() *Ledger {
	return e.ledger
}

func (e *Engine) SetPublisher(publisher events.Publisher) {
	e.emitter = NewEventEmitter(publisher)
}
