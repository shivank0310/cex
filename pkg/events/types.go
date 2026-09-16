package events

import (
	"encoding/json"
	"time"
)

// Event type constants.
const (
	TypeOrderSubmitted    = "order.submitted"
	TypeOrderCancelled    = "order.cancelled"
	TypeOrderUpdated      = "order.updated"
	TypeTradeExecuted     = "trade.executed"
	TypeOrderBookUpdated  = "orderbook.updated"
	TypeLedgerEntry       = "ledger.entry"
	TypeLedgerJournal     = "ledger.journal"
	TypeSettlementDone    = "settlement.completed"
	TypeNotificationSend  = "notification.send"
)

// Envelope is the standard Kafka message wrapper.
type Envelope struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	Topic     string          `json:"topic"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload"`
}

// OrderPayload is published on the orders topic.
type OrderPayload struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Symbol    string `json:"symbol"`
	Side      string `json:"side"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Remaining int64  `json:"remaining"`
	Filled    int64  `json:"filled"`
	Sequence  uint64 `json:"sequence"`
}

// TradePayload is published on the trades topic.
type TradePayload struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	Sequence     uint64    `json:"sequence"`
	BuyOrderID   string    `json:"buy_order_id"`
	SellOrderID  string    `json:"sell_order_id"`
	BuyerID      string    `json:"buyer_id"`
	SellerID     string    `json:"seller_id"`
	Price        int64     `json:"price"`
	Quantity     int64     `json:"quantity"`
	MakerOrderID string    `json:"maker_order_id"`
	TakerOrderID string    `json:"taker_order_id"`
	MakerFee     int64     `json:"maker_fee"`
	TakerFee     int64     `json:"taker_fee"`
	Timestamp    time.Time `json:"timestamp"`
}

// DepthLevel is a single price level in the order book.
type DepthLevel struct {
	Price    int64 `json:"price"`
	Quantity int64 `json:"quantity"`
}

// OrderBookPayload is published on the orderbook topic.
type OrderBookPayload struct {
	Symbol string       `json:"symbol"`
	Bids   []DepthLevel `json:"bids"`
	Asks   []DepthLevel `json:"asks"`
}

// LedgerEntryPayload is a single ledger line (legacy).
type LedgerEntryPayload struct {
	EntryID   string `json:"entry_id"`
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Amount    int64  `json:"amount"` // positive = credit, negative = debit
	Reference string `json:"reference"`
	TradeID   string `json:"trade_id"`
}

// LedgerLegPayload is one leg of a double-entry journal.
type LedgerLegPayload struct {
	LegID     string `json:"leg_id"`
	UserID    string `json:"user_id"`
	Asset     string `json:"asset"`
	Type      string `json:"type"` // DEBIT or CREDIT
	Amount    int64  `json:"amount"`
	Reference string `json:"reference"`
}

// LedgerJournalPayload is a complete balanced journal published to the ledger topic.
type LedgerJournalPayload struct {
	JournalID string             `json:"journal_id"`
	Type      string             `json:"type"`
	Reference string             `json:"reference"`
	Symbol    string             `json:"symbol"`
	Legs      []LedgerLegPayload `json:"legs"`
	PostedAt  time.Time          `json:"posted_at"`
}

// SettlementPayload is published on the settlement topic.
type SettlementPayload struct {
	SettlementID      string    `json:"settlement_id"`
	TradeID           string    `json:"trade_id"`
	JournalID         string    `json:"journal_id"`
	Symbol            string    `json:"symbol"`
	BaseAsset         string    `json:"base_asset"`
	QuoteAsset        string    `json:"quote_asset"`
	BuyerID           string    `json:"buyer_id"`
	SellerID          string    `json:"seller_id"`
	Price             int64     `json:"price"`
	Quantity          int64     `json:"quantity"`
	Notional          int64     `json:"notional"`
	BuyerFee          int64     `json:"buyer_fee"`
	SellerFee         int64     `json:"seller_fee"`
	BuyerBaseCredit   int64     `json:"buyer_base_credit"`
	SellerQuoteCredit int64     `json:"seller_quote_credit"`
	Status            string    `json:"status"`
	Timestamp         time.Time `json:"timestamp"`
}

// NotificationPayload is published on the notifications topic.
type NotificationPayload struct {
	UserID  string `json:"user_id"`
	Channel string `json:"channel"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewEnvelope(topic, eventType, eventID string, payload interface{}) (Envelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		EventID:   eventID,
		EventType: eventType,
		Topic:     topic,
		Timestamp: time.Now().UTC(),
		Payload:   raw,
	}, nil
}

func DecodePayload[T any](env Envelope) (T, error) {
	var out T
	err := json.Unmarshal(env.Payload, &out)
	return out, err
}
