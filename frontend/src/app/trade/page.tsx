"use client";

import { OrderBook } from "@/components/trade/OrderBook";
import { OrderForm } from "@/components/trade/OrderForm";
import { PriceChart } from "@/components/trade/PriceChart";
import { TickerBar } from "@/components/trade/TickerBar";
import { TradeHistory } from "@/components/trade/TradeHistory";
import { useMarketData } from "@/hooks/useMarketData";
import { useTradingStore } from "@/store/trading-store";

export default function TradePage() {
  const symbol = useTradingStore((s) => s.symbol);
  const { ticker, orderBook, trades, candles, error } = useMarketData(symbol);

  return (
    <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6">
      <h1 className="mb-6 text-2xl font-bold text-white">Trading Terminal</h1>
      {error && (
        <div className="mb-4 rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-200">
          API: {error} — ensure backend services are running (nginx :80 or order-service :8081)
        </div>
      )}
      <TickerBar ticker={ticker} />
      <div className="mt-4 grid gap-4 lg:grid-cols-12">
        <div className="lg:col-span-8 space-y-4">
          <PriceChart candles={candles} lastPrice={ticker?.last_price} />
          <TradeHistory trades={trades} />
        </div>
        <div className="lg:col-span-4 space-y-4">
          <OrderForm />
          <OrderBook bids={orderBook?.bids ?? []} asks={orderBook?.asks ?? []} />
        </div>
      </div>
    </div>
  );
}
