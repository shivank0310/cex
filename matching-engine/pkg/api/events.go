package api

import (
	"context"
	"fmt"

	"github.com/shivank0310/cex.git/matching-engine/internal/orderbook"
	"github.com/shivank0310/cex.git/pkg/events"
)

// EventEmitter publishes matching-engine events to Kafka.
type EventEmitter struct {
	publisher events.Publisher
}

func NewEventEmitter(publisher events.Publisher) *EventEmitter {
	return &EventEmitter{publisher: publisher}
}

func (e *EventEmitter) EmitSubmitResult(ctx context.Context, result SubmitResult) error {
	if e == nil || e.publisher == nil || result.Error != nil {
		return nil
	}

	eventType := events.TypeOrderSubmitted
	switch result.Order.Status {
	case StatusFilled:
		eventType = events.TypeOrderUpdated
	case StatusPartiallyFilled:
		eventType = events.TypeOrderUpdated
	case StatusCancelled:
		eventType = events.TypeOrderCancelled
	}

	if err := e.publishOrder(ctx, eventType, result.Order); err != nil {
		return err
	}

	for _, tr := range result.Trades {
		if err := e.publishTrade(ctx, tr); err != nil {
			return err
		}
	}

	return e.publishOrderBook(ctx, result.Order.Symbol)
}

func (e *EventEmitter) EmitCancel(ctx context.Context, o *Order) error {
	if e == nil || e.publisher == nil || o == nil {
		return nil
	}
	if err := e.publishOrder(ctx, events.TypeOrderCancelled, o); err != nil {
		return err
	}
	return e.publishOrderBook(ctx, o.Symbol)
}

func (e *EventEmitter) publishOrder(ctx context.Context, eventType string, o *Order) error {
	payload := events.OrderPayload{
		ID:        o.ID,
		UserID:    o.UserID,
		Symbol:    o.Symbol,
		Side:      string(o.Side),
		Type:      string(o.Type),
		Status:    string(o.Status),
		Price:     o.Price,
		Quantity:  o.Quantity,
		Remaining: o.Remaining,
		Filled:    o.Filled,
		Sequence:  o.Sequence,
	}
	env, err := events.NewEnvelope(events.TopicOrders, eventType, o.ID, payload)
	if err != nil {
		return err
	}
	return e.publisher.Publish(ctx, events.TopicOrders, o.Symbol, env)
}

func (e *EventEmitter) publishTrade(ctx context.Context, t *Trade) error {
	payload := events.TradePayload{
		ID:           t.ID,
		Symbol:       t.Symbol,
		Sequence:     t.Sequence,
		BuyOrderID:   t.BuyOrderID,
		SellOrderID:  t.SellOrderID,
		BuyerID:      t.BuyerID,
		SellerID:     t.SellerID,
		Price:        t.Price,
		Quantity:     t.Quantity,
		MakerOrderID: t.MakerOrderID,
		TakerOrderID: t.TakerOrderID,
		MakerFee:     t.MakerFee,
		TakerFee:     t.TakerFee,
		Timestamp:    t.Timestamp,
	}
	env, err := events.NewEnvelope(events.TopicTrades, events.TypeTradeExecuted, t.ID, payload)
	if err != nil {
		return err
	}
	return e.publisher.Publish(ctx, events.TopicTrades, t.Symbol, env)
}

func (e *EventEmitter) publishOrderBook(ctx context.Context, symbol string) error {
	book := depthFromBook(symbol)
	if book == nil {
		return nil
	}
	env, err := events.NewEnvelope(events.TopicOrderBook, events.TypeOrderBookUpdated, symbol, book)
	if err != nil {
		return err
	}
	return e.publisher.Publish(ctx, events.TopicOrderBook, symbol, env)
}

// bookProvider is set by Engine to supply order book snapshots for events.
var bookProvider func(symbol string) *orderbook.OrderBook

func depthFromBook(symbol string) *events.OrderBookPayload {
	if bookProvider == nil {
		return &events.OrderBookPayload{Symbol: symbol}
	}
	book := bookProvider(symbol)
	if book == nil {
		return &events.OrderBookPayload{Symbol: symbol}
	}

	payload := &events.OrderBookPayload{Symbol: symbol}
	for _, b := range book.Bids() {
		payload.Bids = append(payload.Bids, events.DepthLevel{Price: b.Price, Quantity: b.Quantity})
	}
	for _, a := range book.Asks() {
		payload.Asks = append(payload.Asks, events.DepthLevel{Price: a.Price, Quantity: a.Quantity})
	}
	return payload
}

// SetBookProvider registers a callback used when publishing orderbook events.
func SetBookProvider(fn func(symbol string) *orderbook.OrderBook) {
	bookProvider = fn
}

// PublishLedger publishes a ledger entry (used by ledger-service or tests).
func PublishLedger(ctx context.Context, pub events.Publisher, entry events.LedgerEntryPayload) error {
	env, err := events.NewEnvelope(events.TopicLedger, events.TypeLedgerEntry, entry.EntryID, entry)
	if err != nil {
		return err
	}
	return pub.Publish(ctx, events.TopicLedger, entry.UserID, env)
}

// PublishSettlement publishes a settlement event.
func PublishSettlement(ctx context.Context, pub events.Publisher, s events.SettlementPayload) error {
	env, err := events.NewEnvelope(events.TopicSettlement, events.TypeSettlementDone, s.SettlementID, s)
	if err != nil {
		return err
	}
	return pub.Publish(ctx, events.TopicSettlement, s.TradeID, env)
}

// PublishNotification publishes a user notification event.
func PublishNotification(ctx context.Context, pub events.Publisher, n events.NotificationPayload) error {
	id := fmt.Sprintf("%s-%s", n.UserID, n.Subject)
	env, err := events.NewEnvelope(events.TopicNotifications, events.TypeNotificationSend, id, n)
	if err != nil {
		return err
	}
	return pub.Publish(ctx, events.TopicNotifications, n.UserID, env)
}
