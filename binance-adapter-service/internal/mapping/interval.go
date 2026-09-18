package mapping

// NormalizeInterval maps user-facing interval names to canonical form.
func NormalizeInterval(interval string) string {
	switch interval {
	case "1d", "24h":
		return "24h"
	case "1hr":
		return "1h"
	default:
		return interval
	}
}

// ToBinanceInterval returns the Binance Spot kline interval parameter.
func ToBinanceInterval(interval string) string {
	switch NormalizeInterval(interval) {
	case "24h":
		return "1d"
	default:
		return NormalizeInterval(interval)
	}
}

// BinanceFetchPlan describes how to fetch klines for a requested interval.
// Some intervals (e.g. 10m) are aggregated from finer Binance candles.
type BinanceFetchPlan struct {
	FetchInterval string
	MergeCount    int
}

func BinanceFetchPlanFor(interval string) BinanceFetchPlan {
	norm := NormalizeInterval(interval)
	switch norm {
	case "10m":
		return BinanceFetchPlan{FetchInterval: "5m", MergeCount: 2}
	default:
		return BinanceFetchPlan{FetchInterval: ToBinanceInterval(norm), MergeCount: 1}
	}
}
