package redis

import "strings"

// Key prefixes — all Redis keys are namespaced under cex:
const (
	prefixSession   = "cex:session:"
	prefixRateLimit = "cex:ratelimit:"
	prefixTicker    = "cex:market:ticker:"
	prefixOrderBook = "cex:market:orderbook:"
	prefixStats     = "cex:market:stats:"
	prefixTrades    = "cex:market:trades:"
	prefixOrder     = "cex:order:pending:"
	prefixLock      = "cex:lock:"
	prefixPubSub    = "cex:pubsub:"
)

func SessionKey(token string) string       { return prefixSession + token }
func RateLimitKey(userID, endpoint string) string { return prefixRateLimit + userID + ":" + endpoint }
func TickerKey(symbol string) string       { return prefixTicker + NormalizeSymbol(symbol) }
func OrderBookKey(symbol string) string    { return prefixOrderBook + NormalizeSymbol(symbol) }
func StatsKey(symbol string) string        { return prefixStats + NormalizeSymbol(symbol) }
func TradesKey(symbol string) string       { return prefixTrades + NormalizeSymbol(symbol) }
func PendingOrderKey(orderID string) string { return prefixOrder + orderID }
func LockKey(resource string) string       { return prefixLock + resource }
func PubSubChannel(channel string) string    { return prefixPubSub + channel }

// NormalizeSymbol converts BTC/USDT ↔ BTC-USDT for consistent cache keys.
func NormalizeSymbol(symbol string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(symbol)), "/", "-")
}

func DisplaySymbol(normalized string) string {
	return strings.ReplaceAll(normalized, "-", "/")
}
