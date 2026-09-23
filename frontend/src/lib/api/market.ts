import { apiFetch } from "./client";
import type { Candle, MarketTrade, OrderBook, Ticker } from "@/types";

export async function getTicker(symbol: string, signal?: AbortSignal): Promise<Ticker> {
  const encoded = symbol.replace("/", "%2F");
  return apiFetch<Ticker>(`/api/v1/market/ticker/${encoded}`, { signal });
}

export async function getTickers(signal?: AbortSignal): Promise<Ticker[]> {
  return apiFetch<Ticker[]>("/api/v1/market/tickers", { signal });
}

export async function getOrderBook(symbol: string, signal?: AbortSignal): Promise<OrderBook> {
  const encoded = symbol.replace("/", "%2F");
  return apiFetch<OrderBook>(`/api/v1/market/orderbook/${encoded}`, { signal });
}

export async function getTrades(
  symbol: string,
  limit = 50,
  signal?: AbortSignal
): Promise<MarketTrade[]> {
  const encoded = symbol.replace("/", "%2F");
  return apiFetch<MarketTrade[]>(`/api/v1/market/trades/${encoded}?limit=${limit}`, { signal });
}

export async function getCandles(
  symbol: string,
  interval = "1m",
  limit = 100,
  signal?: AbortSignal
): Promise<Candle[]> {
  const encoded = symbol.replace("/", "%2F");
  return apiFetch<Candle[]>(
    `/api/v1/market/candles/${encoded}?interval=${interval}&limit=${limit}`,
    { signal }
  );
}
