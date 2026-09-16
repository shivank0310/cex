package tests

import (
	"testing"

	"github.com/shivank0310/cex.git/matching-engine/internal/engine"
	"github.com/shivank0310/cex.git/matching-engine/internal/order"
	"github.com/shivank0310/cex.git/matching-engine/internal/settlement"
	"github.com/shivank0310/cex.git/matching-engine/pkg/decimal"
)

const (
	symbol    = "BTC/USDT"
	baseAsset = "BTC"
	quoteAsset = "USDT"
)

func newTestEngine() *engine.Engine {
	ledger := settlement.NewLedger()
	eng := engine.New(ledger, engine.FeeConfig{MakerBasisPoints: 0, TakerBasisPoints: 0})
	eng.RegisterSymbol(symbol, engine.SymbolConfig{BaseAsset: baseAsset, QuoteAsset: quoteAsset})
	return eng
}

func seedBook(t *testing.T, eng *engine.Engine) {
	ledger := eng.Ledger()

	// Bids
	seedOrder(t, eng, ledger, "bid1", "user-bid-1", order.Buy, 101000, 50)
	seedOrder(t, eng, ledger, "bid2", "user-bid-2", order.Buy, 100900, 120)
	seedOrder(t, eng, ledger, "bid3", "user-bid-3", order.Buy, 100800, 200)

	// Asks
	seedOrder(t, eng, ledger, "ask1", "user-ask-1", order.Sell, 101100, 30)
	seedOrder(t, eng, ledger, "ask2", "user-ask-2", order.Sell, 101200, 100)
	seedOrder(t, eng, ledger, "ask3", "user-ask-3", order.Sell, 101300, 250)
}

func seedOrder(t *testing.T, eng *engine.Engine, ledger *settlement.Ledger, id, userID string, side order.Side, price, qty int64) {
	if side == order.Buy {
		ledger.Deposit(userID, quoteAsset, decimal.Notional(price, qty))
	} else {
		ledger.Deposit(userID, baseAsset, qty)
	}
	o := order.NewLimit(id, userID, symbol, side, price, qty, 0)
	res := eng.SubmitOrder(o)
	if res.Error != nil {
		t.Fatalf("seed %s: %v", id, res.Error)
	}
}

func TestUserAScenario_LimitBuyMatchesBestAsk(t *testing.T) {
	eng := newTestEngine()
	seedBook(t, eng)

	ledger := eng.Ledger()
	ledger.Deposit("user-a", quoteAsset, decimal.Notional(101100, 30))

	buy := order.NewLimit("buy-a", "user-a", symbol, order.Buy, 101100, 30, 0)
	res := eng.SubmitOrder(buy)
	if res.Error != nil {
		t.Fatalf("submit: %v", res.Error)
	}
	if len(res.Trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(res.Trades))
	}

	tr := res.Trades[0]
	if tr.Price != 101100 || tr.Quantity != 30 {
		t.Fatalf("unexpected trade: price=%d qty=%d", tr.Price, tr.Quantity)
	}
	if tr.MakerOrderID != "ask1" || tr.TakerOrderID != "buy-a" {
		t.Fatalf("maker/taker: maker=%s taker=%s", tr.MakerOrderID, tr.TakerOrderID)
	}
	if buy.Status != order.StatusFilled {
		t.Fatalf("buyer order status: %s", buy.Status)
	}

	// Buyer: USDT -30330, BTC +0.30
	buyerUSDT := ledger.Get("user-a", quoteAsset)
	buyerBTC := ledger.Get("user-a", baseAsset)
	if buyerUSDT.Available != 0 || buyerUSDT.Locked != 0 {
		t.Fatalf("buyer USDT: avail=%d locked=%d", buyerUSDT.Available, buyerUSDT.Locked)
	}
	if buyerBTC.Available != 30 {
		t.Fatalf("buyer BTC: %d", buyerBTC.Available)
	}

	// Seller: BTC -0.30, USDT +30330
	sellerBTC := ledger.Get("user-ask-1", baseAsset)
	sellerUSDT := ledger.Get("user-ask-1", quoteAsset)
	if sellerBTC.Available != 0 || sellerBTC.Locked != 0 {
		t.Fatalf("seller BTC: avail=%d locked=%d", sellerBTC.Available, sellerBTC.Locked)
	}
	if sellerUSDT.Available != decimal.Notional(101100, 30) {
		t.Fatalf("seller USDT: %d", sellerUSDT.Available)
	}

	book := eng.GetBook(symbol)
	asks := book.Asks()
	if len(asks) != 2 || asks[0].Price != 101200 {
		t.Fatalf("ask book after match: %+v", asks)
	}
}

func TestPartialFill(t *testing.T) {
	eng := newTestEngine()
	ledger := eng.Ledger()

	ledger.Deposit("seller", baseAsset, 100)
	ledger.Deposit("buyer", quoteAsset, decimal.Notional(101000, 100))

	sell := order.NewLimit("sell-1", "seller", symbol, order.Sell, 101000, 100, 0)
	eng.SubmitOrder(sell)

	buy := order.NewLimit("buy-1", "buyer", symbol, order.Buy, 101000, 40, 0)
	res := eng.SubmitOrder(buy)

	if len(res.Trades) != 1 || res.Trades[0].Quantity != 40 {
		t.Fatalf("partial fill trade: %+v", res.Trades)
	}
	if sell.Remaining != 60 || sell.Status != order.StatusPartiallyFilled {
		t.Fatalf("seller remaining=%d status=%s", sell.Remaining, sell.Status)
	}
}

func TestPriceTimePriority(t *testing.T) {
	eng := newTestEngine()
	ledger := eng.Ledger()

	ledger.Deposit("s1", baseAsset, 100)
	ledger.Deposit("s2", baseAsset, 100)
	ledger.Deposit("buyer", quoteAsset, decimal.Notional(101000, 100))

	o1 := order.NewLimit("s1", "s1", symbol, order.Sell, 101000, 50, 0)
	o2 := order.NewLimit("s2", "s2", symbol, order.Sell, 101000, 50, 0)
	eng.SubmitOrder(o1)
	eng.SubmitOrder(o2)

	buy := order.NewLimit("buy", "buyer", symbol, order.Buy, 101000, 50, 0)
	res := eng.SubmitOrder(buy)

	if res.Trades[0].MakerOrderID != "s1" {
		t.Fatalf("expected FIFO at same price, got maker %s", res.Trades[0].MakerOrderID)
	}
}

func TestMarketSell(t *testing.T) {
	eng := newTestEngine()
	ledger := eng.Ledger()

	ledger.Deposit("bidder", quoteAsset, decimal.Notional(100000, 100))
	ledger.Deposit("seller", baseAsset, 50)

	bid := order.NewLimit("bid", "bidder", symbol, order.Buy, 100000, 100, 0)
	eng.SubmitOrder(bid)

	mkt := order.NewMarket("mkt-sell", "seller", symbol, order.Sell, 50, 0)
	res := eng.SubmitOrder(mkt)

	if len(res.Trades) != 1 || res.Trades[0].Price != 100000 {
		t.Fatalf("market sell: %+v", res.Trades)
	}
	if mkt.Status != order.StatusFilled {
		t.Fatalf("market order status: %s", mkt.Status)
	}
}

func TestCancelOrder(t *testing.T) {
	eng := newTestEngine()
	ledger := eng.Ledger()

	ledger.Deposit("user", quoteAsset, decimal.Notional(101000, 50))
	o := order.NewLimit("cancel-me", "user", symbol, order.Buy, 101000, 50, 0)
	eng.SubmitOrder(o)

	cancelled, err := eng.CancelOrder(symbol, "cancel-me")
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancelled.Status != order.StatusCancelled {
		t.Fatalf("status: %s", cancelled.Status)
	}

	bal := ledger.Get("user", quoteAsset)
	if bal.Available != decimal.Notional(101000, 50) || bal.Locked != 0 {
		t.Fatalf("funds not unlocked: avail=%d locked=%d", bal.Available, bal.Locked)
	}
}

func TestPriceImprovement(t *testing.T) {
	eng := newTestEngine()
	ledger := eng.Ledger()

	ledger.Deposit("seller", baseAsset, 30)
	ledger.Deposit("buyer", quoteAsset, decimal.Notional(101200, 30))

	sell := order.NewLimit("ask", "seller", symbol, order.Sell, 101100, 30, 0)
	eng.SubmitOrder(sell)

	buy := order.NewLimit("buy", "buyer", symbol, order.Buy, 101200, 30, 0)
	res := eng.SubmitOrder(buy)

	if res.Trades[0].Price != 101100 {
		t.Fatalf("trade at maker price")
	}

	// Buyer pays 30330 not 30360; 30 USDT refunded
	buyerUSDT := ledger.Get("buyer", quoteAsset)
	if buyerUSDT.Available != 30 {
		t.Fatalf("price improvement refund: %d", buyerUSDT.Available)
	}
}

func TestMakerTakerFees(t *testing.T) {
	ledger := settlement.NewLedger()
	eng := engine.New(ledger, engine.FeeConfig{MakerBasisPoints: 10, TakerBasisPoints: 20})
	eng.RegisterSymbol(symbol, engine.SymbolConfig{BaseAsset: baseAsset, QuoteAsset: quoteAsset})

	ledger.Deposit("seller", baseAsset, 30)
	ledger.Deposit("buyer", quoteAsset, decimal.Notional(101100, 30))

	eng.SubmitOrder(order.NewLimit("ask", "seller", symbol, order.Sell, 101100, 30, 0))
	res := eng.SubmitOrder(order.NewLimit("buy", "buyer", symbol, order.Buy, 101100, 30, 0))

	tr := res.Trades[0]
	if tr.MakerFee == 0 || tr.TakerFee == 0 {
		t.Fatalf("expected non-zero fees: maker=%d taker=%d", tr.MakerFee, tr.TakerFee)
	}
	if tr.TakerFee <= tr.MakerFee {
		t.Fatalf("taker fee should exceed maker fee")
	}
}
