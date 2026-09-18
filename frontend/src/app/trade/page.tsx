"use client";

import { Suspense, useEffect, useState } from "react";
import type { CandleInterval } from "@/lib/chart-intervals";
import { useSearchParams } from "next/navigation";
import { AuthGuard } from "@/components/auth/AuthGuard";
import { OrderBook } from "@/components/trade/OrderBook";
import { OrderForm } from "@/components/trade/OrderForm";
import { PriceChart } from "@/components/trade/PriceChart";
import { TickerBar } from "@/components/trade/TickerBar";
import { TradeHistory } from "@/components/trade/TradeHistory";
import { VenueBadge } from "@/components/trade/VenueBadge";
import { useMarketData } from "@/hooks/useMarketData";
import { useTradingStore } from "@/store/trading-store";

function TradeContent() {
  const searchParams = useSearchParams();
  const symbol = useTradingStore((s) => s.symbol);
  const setSymbol = useTradingStore((s) => s.setSymbol);
  const [candleInterval, setCandleInterval] = useState<CandleInterval>("1m");
  const { ticker, orderBook, trades, candles, error } = useMarketData(symbol, candleInterval);

  useEffect(() => {
    const param = searchParams.get("symbol");
    if (param) setSymbol(param);
  }, [searchParams, setSymbol]);

  return (
    <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6">
      <div className="mb-6 flex flex-wrap items-center gap-3">
        <h1 className="text-2xl font-bold text-white">Trading Terminal</h1>
        <VenueBadge symbol={symbol} />
      </div>
      {error && (
        <div className="mb-4 rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-200">
          API: {error} — ensure backend stack is running (nginx :80 → api-gateway)
        </div>
      )}
      <TickerBar ticker={ticker} />
      <div className="mt-4 grid gap-4 lg:grid-cols-12">
        <div className="lg:col-span-8 space-y-4">
          <PriceChart
            candles={candles}
            lastPrice={ticker?.last_price}
            interval={candleInterval}
            onIntervalChange={setCandleInterval}
          />
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

export default function TradePage() {
  return (
    <AuthGuard title="Sign in to trade">
      <Suspense fallback={<div className="p-6 text-slate-400">Loading terminal...</div>}>
        <TradeContent />
      </Suspense>
    </AuthGuard>
  );
}
