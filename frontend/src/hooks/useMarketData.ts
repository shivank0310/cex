"use client";

import { useEffect, useState } from "react";
import { getCandles, getOrderBook, getTicker, getTrades } from "@/lib/api/market";
import { ApiClientError } from "@/lib/api/client";
import { candleLimit, type CandleInterval } from "@/lib/chart-intervals";
import type { Candle, MarketTrade, OrderBook, Ticker } from "@/types";

const POLL_MS = 8000;

export function useMarketData(symbol: string, interval: CandleInterval = "1m") {
  const [ticker, setTicker] = useState<Ticker | null>(null);
  const [orderBook, setOrderBook] = useState<OrderBook | null>(null);
  const [trades, setTrades] = useState<MarketTrade[]>([]);
  const [candles, setCandles] = useState<Candle[]>([]);
  const [loading, setLoading] = useState(true);
  const [candlesLoading, setCandlesLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [candlesError, setCandlesError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    let inFlight = false;
    let pollTimer: ReturnType<typeof setTimeout> | null = null;
    const abort = new AbortController();

    const poll = async (isInitial: boolean) => {
      if (cancelled || inFlight) return;
      inFlight = true;

      if (isInitial) {
        setCandles([]);
        setCandlesLoading(true);
        setCandlesError(null);
        setLoading(true);
        setError(null);
      }

      const signal = abort.signal;

      try {
        const [tickerRes, orderBookRes, tradesRes, candlesRes] = await Promise.allSettled([
          getTicker(symbol, signal),
          getOrderBook(symbol, signal),
          getTrades(symbol, 30, signal),
          getCandles(symbol, interval, candleLimit(interval), signal),
        ]);

        if (cancelled) return;

        if (tickerRes.status === "fulfilled") {
          setTicker(tickerRes.value);
          setError(null);
        } else if (!isAbortError(tickerRes.reason)) {
          setError(
            "Market data unavailable — run: docker compose -f docker/docker-compose.yml up -d"
          );
        }

        if (orderBookRes.status === "fulfilled") {
          setOrderBook(orderBookRes.value);
        }

        if (tradesRes.status === "fulfilled") {
          setTrades(tradesRes.value);
        }

        if (candlesRes.status === "fulfilled") {
          setCandles(candlesRes.value);
          setCandlesError(null);
        } else if (!isAbortError(candlesRes.reason)) {
          setCandles([]);
          setCandlesError("Could not load chart data");
        }
      } finally {
        inFlight = false;
        if (!cancelled) {
          setLoading(false);
          setCandlesLoading(false);
        }
      }
    };

    const scheduleNext = () => {
      pollTimer = setTimeout(() => {
        void poll(false).finally(() => {
          if (!cancelled) scheduleNext();
        });
      }, POLL_MS);
    };

    void poll(true).finally(() => {
      if (!cancelled) scheduleNext();
    });

    return () => {
      cancelled = true;
      abort.abort();
      if (pollTimer) clearTimeout(pollTimer);
    };
  }, [symbol, interval]);

  return {
    ticker,
    orderBook,
    trades,
    candles,
    loading,
    candlesLoading,
    error,
    candlesError,
    refresh: () => {},
  };
}

function isAbortError(err: unknown): boolean {
  return err instanceof ApiClientError && err.code === "ABORTED";
}
