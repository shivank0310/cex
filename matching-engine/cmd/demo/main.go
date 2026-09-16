package main

import (
	"fmt"

	"github.com/shivank0310/cex.git/matching-engine/internal/engine"
	"github.com/shivank0310/cex.git/matching-engine/internal/order"
	"github.com/shivank0310/cex.git/matching-engine/internal/settlement"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
)

const (
	symbol     = "BTC/USDT"
	baseAsset  = "BTC"
	quoteAsset = "USDT"
)

func main() {
	ledger := settlement.NewLedger()
	eng := engine.New(ledger, engine.FeeConfig{MakerBasisPoints: 0, TakerBasisPoints: 0})
	eng.RegisterSymbol(symbol, engine.SymbolConfig{BaseAsset: baseAsset, QuoteAsset: quoteAsset})

	fmt.Println("=== BTC/USDT Matching Engine Demo ===")
	fmt.Println()

	// Seed order book
	place(eng, ledger, "bid1", "bidder-1", order.Buy, 101000, 50)
	place(eng, ledger, "bid2", "bidder-2", order.Buy, 100900, 120)
	place(eng, ledger, "bid3", "bidder-3", order.Buy, 100800, 200)

	place(eng, ledger, "ask1", "seller-1", order.Sell, 101100, 30)
	place(eng, ledger, "ask2", "seller-2", order.Sell, 101200, 100)
	place(eng, ledger, "ask3", "seller-3", order.Sell, 101300, 250)

	printBook(eng)

	fmt.Println()
	fmt.Println("--- User A: BUY 0.30 BTC @ 101100 ---")
	fmt.Println()
	ledger.Deposit("user-a", quoteAsset, decimal.Notional(101100, 30))

	buy := order.NewLimit("buy-a", "user-a", symbol, order.Buy, 101100, 30, 0)
	res := eng.SubmitOrder(buy)

	for _, tr := range res.Trades {
		fmt.Printf("TRADE  seq=%d  %.2f BTC @ %d USDT  maker=%s taker=%s\n",
			tr.Sequence, qty(tr.Quantity), tr.Price, tr.MakerOrderID, tr.TakerOrderID)
	}

	fmt.Println()
	printBalances(ledger, "user-a", "User A (buyer)")
	printBalances(ledger, "seller-1", "Seller-1")
	printBook(eng)
}

func place(eng *engine.Engine, ledger *settlement.Ledger, id, user string, side order.Side, price, qty int64) {
	if side == order.Buy {
		ledger.Deposit(user, quoteAsset, decimal.Notional(price, qty))
	} else {
		ledger.Deposit(user, baseAsset, qty)
	}
	eng.SubmitOrder(order.NewLimit(id, user, symbol, side, price, qty, 0))
}

func printBook(eng *engine.Engine) {
	book := eng.GetBook(symbol)
	fmt.Println("BUY ORDERS")
	fmt.Println("Price       Amount")
	fmt.Println("───────────")
	for _, b := range book.Bids() {
		fmt.Printf("%-11d %.2f\n", b.Price, qty(b.Quantity))
	}
	fmt.Println()
	fmt.Println("SELL ORDERS")
	fmt.Println("Price       Amount")
	fmt.Println("───────────")
	for _, a := range book.Asks() {
		fmt.Printf("%-11d %.2f\n", a.Price, qty(a.Quantity))
	}
}

func printBalances(ledger *settlement.Ledger, userID, label string) {
	usdt := ledger.Get(userID, quoteAsset)
	btc := ledger.Get(userID, baseAsset)
	fmt.Printf("%s:\n", label)
	fmt.Printf("  USDT  available=%d  locked=%d\n", usdt.Available, usdt.Locked)
	fmt.Printf("  BTC   available=%.2f  locked=%.2f\n", qty(btc.Available), qty(btc.Locked))
	fmt.Println()
}

func qty(units int64) float64 {
	return float64(units) / float64(decimal.QuantityScale)
}
