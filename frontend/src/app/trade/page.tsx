"use client";

import { Suspense, useEffect, useRef, useState } from "react";
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
  const setPrice = useTradingStore((s) => s.setPrice);
  const [candleInterval, setCandleInterval] = useState<CandleInterval>("1m");
  const priceSyncedRef = useRef(false);
  const { ticker, orderBook, trades, candles, loading, candlesLoading, error, candlesError } =
    useMarketData(symbol, candleInterval);

  useEffect(() => {
    const param = searchParams.get("symbol");
    if (param) setSymbol(param);
  }, [searchParams, setSymbol]);

  useEffect(() => {
    priceSyncedRef.current = false;
  }, [symbol]);

  useEffect(() => {
    if (ticker?.last_price && !priceSyncedRef.current) {
      setPrice(String(ticker.last_price));
      priceSyncedRef.current = true;
    }
  }, [ticker?.last_price, setPrice, symbol]);

  return (
    <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6">
      <div className="mb-6 flex flex-wrap items-center gap-3">
        <h1 className="text-2xl font-bold text-white">Trading Terminal</h1>
        <VenueBadge symbol={symbol} />
      </div>

      {error && (
        <div className="mb-4 rounded-lg border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-sm text-amber-200">
          {error}
        </div>
      )}

      <TickerBar ticker={ticker} loading={loading} />

      <div className="mt-4 grid gap-4 lg:grid-cols-12">
        <div className="lg:col-span-8 space-y-4">
          <PriceChart
            candles={candles}
            candlesLoading={candlesLoading}
            candlesError={candlesError}
            lastPrice={ticker?.last_price}
            interval={candleInterval}
            onIntervalChange={setCandleInterval}
          />
          <TradeHistory trades={trades} />
        </div>
        <div className="lg:col-span-4 space-y-4">
          <AuthGuard title="Sign in to place orders">
            <OrderForm />
          </AuthGuard>
          <OrderBook bids={orderBook?.bids ?? []} asks={orderBook?.asks ?? []} />
        </div>
      </div>
    </div>
  );
}

export default function TradePage() {
  return (
    <Suspense fallback={<div className="p-6 text-slate-400">Loading terminal...</div>}>
      <TradeContent />
    </Suspense>
  );
}
