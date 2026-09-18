import { apiFetch } from "./client";
import type { Candle, MarketTrade, OrderBook, Ticker } from "@/types";

export async function getTicker(symbol: string): Promise<Ticker> {
  const encoded = symbol.replace("/", "%2F");
  return apiFetch<Ticker>(`/api/v1/market/ticker/${encoded}`);
}

export async function getTickers(): Promise<Ticker[]> {
  return apiFetch<Ticker[]>("/api/v1/market/tickers");
}

export async function getOrderBook(symbol: string): Promise<OrderBook> {
  const encoded = symbol.replace("/", "%2F");
  return apiFetch<OrderBook>(`/api/v1/market/orderbook/${encoded}`);
}

export async function getTrades(symbol: string, limit = 50): Promise<MarketTrade[]> {
  const encoded = symbol.replace("/", "%2F");
  return apiFetch<MarketTrade[]>(`/api/v1/market/trades/${encoded}?limit=${limit}`);
}

export async function getCandles(
  symbol: string,
  interval = "1m",
  limit = 100
): Promise<Candle[]> {
  const encoded = symbol.replace("/", "%2F");
  return apiFetch<Candle[]>(
    `/api/v1/market/candles/${encoded}?interval=${interval}&limit=${limit}`
  );
}
