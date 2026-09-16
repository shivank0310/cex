package events

// Kafka topic names for the CEX event bus.
const (
	TopicOrders        = "orders"
	TopicTrades        = "trades"
	TopicOrderBook     = "orderbook"
	TopicLedger        = "ledger"
	TopicSettlement    = "settlement"
	TopicNotifications = "notifications"
)

// AllTopics returns every topic used by the platform.
func AllTopics() []string {
	return []string{
		TopicOrders,
		TopicTrades,
		TopicOrderBook,
		TopicLedger,
		TopicSettlement,
		TopicNotifications,
	}
}
