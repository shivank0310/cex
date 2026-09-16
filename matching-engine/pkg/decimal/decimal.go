package decimal

// Fixed-point scales for deterministic arithmetic without floating point.
// Price is quoted in quote currency per 1 base unit (e.g. USDT per BTC).
// Quantity is base asset amount in smallest units.

const (
	QuantityScale = 100 // 2 decimal places: 0.30 BTC = 30
	PriceScale    = 1   // whole USDT per BTC: 101100 = 101100 USDT
	FeeBasisPoints = 10000 // 100% = 10000 bps
)

// Notional computes quote currency value: price * quantity / QuantityScale.
func Notional(price, quantity int64) int64 {
	return price * quantity / QuantityScale
}

// Fee computes fee in the same unit as amount using basis points.
func Fee(amount int64, basisPoints int64) int64 {
	return amount * basisPoints / FeeBasisPoints
}
