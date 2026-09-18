"use client";

import { useCallback, useEffect, useState } from "react";
import { getCandles, getOrderBook, getTicker, getTrades } from "@/lib/api/market";
import { candleLimit, type CandleInterval } from "@/lib/chart-intervals";
import type { Candle, MarketTrade, OrderBook, Ticker } from "@/types";

const POLL_MS = 3000;

export function useMarketData(symbol: string, interval: CandleInterval = "1m") {
  const [ticker, setTicker] = useState<Ticker | null>(null);
  const [orderBook, setOrderBook] = useState<OrderBook | null>(null);
  const [trades, setTrades] = useState<MarketTrade[]>([]);
  const [candles, setCandles] = useState<Candle[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    try {
      const [t, ob, tr, c] = await Promise.all([
        getTicker(symbol).catch(() => null),
        getOrderBook(symbol).catch(() => null),
        getTrades(symbol, 30).catch(() => []),
        getCandles(symbol, interval, candleLimit(interval)).catch(() => []),
      ]);
      if (t) setTicker(t);
      if (ob) setOrderBook(ob);
      setTrades(tr);
      setCandles(c);
      setError(null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load market data");
    } finally {
      setLoading(false);
    }
  }, [symbol, interval]);

  useEffect(() => {
    setCandles([]);
    setLoading(true);
    refresh();
    const id = setInterval(refresh, POLL_MS);
    return () => clearInterval(id);
  }, [refresh]);

  return { ticker, orderBook, trades, candles, loading, error, refresh };
}
