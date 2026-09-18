package mapping

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
)

// ToBinanceSymbol converts internal "BTC/USDT" to Binance "BTCUSDT".
func ToBinanceSymbol(symbol string) string {
	return strings.ReplaceAll(strings.ToUpper(strings.TrimSpace(symbol)), "/", "")
}

// FromBinanceSymbol converts Binance "BTCUSDT" to internal "BTC/USDT".
func FromBinanceSymbol(symbol string) string {
	s := strings.ToUpper(strings.TrimSpace(symbol))
	for _, quote := range []string{"USDT", "USDC", "BUSD", "BTC", "ETH", "BNB"} {
		if strings.HasSuffix(s, quote) && len(s) > len(quote) {
			base := s[:len(s)-len(quote)]
			return base + "/" + quote
		}
	}
	return s
}

// QuantityToBinance converts internal fixed-point quantity to Binance decimal string.
// Internal: 30 = 0.30 BTC (scale 100).
func QuantityToBinance(qty int64) string {
	return formatScaled(qty, decimal.QuantityScale)
}

// QuantityFromBinance parses Binance quantity string to internal fixed-point.
func QuantityFromBinance(s string) (int64, error) {
	return parseScaled(s, decimal.QuantityScale)
}

// PriceToBinance converts internal whole-number price to Binance decimal string.
func PriceToBinance(price int64) string {
	return strconv.FormatInt(price, 10)
}

// PriceFromBinance parses Binance price string to internal fixed-point.
func PriceFromBinance(s string) (int64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}

func formatScaled(value int64, scale int64) string {
	whole := value / scale
	frac := value % scale
	if frac == 0 {
		return strconv.FormatInt(whole, 10)
	}
	fracStr := fmt.Sprintf("%0*d", digits(scale), frac)
	fracStr = strings.TrimRight(fracStr, "0")
	return fmt.Sprintf("%d.%s", whole, fracStr)
}

func parseScaled(s string, scale int64) (int64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(f * float64(scale)), nil
}

func digits(scale int64) int {
	d := 0
	for s := scale; s > 1; s /= 10 {
		d++
	}
	return d
}
