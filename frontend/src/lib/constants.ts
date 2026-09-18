export const QUANTITY_SCALE = 100;

export const DEFAULT_SYMBOL = "BTC/USDT";

export const TRADING_PAIRS = [
  "BTC/USDT",
  "ETH/USDT",
  "SOL/USDT",
] as const;

/** Symbols routed to binance-adapter-service when ROUTE_BINANCE=1 on order-service. */
export const BINANCE_VENUE_SYMBOLS = new Set<string>(["BTC/USDT"]);

export function getTradingVenue(symbol: string): "binance" | "internal" {
  return BINANCE_VENUE_SYMBOLS.has(symbol) ? "binance" : "internal";
}

export const FEATURES = [
  {
    title: "Spot Trading Engine",
    description: "In-memory matching with price-time priority. Sub-millisecond order execution off-chain.",
    icon: "Zap",
  },
  {
    title: "Crypto Wallet",
    description: "Hot/cold custody architecture. Deposits on-chain, balances in internal ledger.",
    icon: "Wallet",
  },
  {
    title: "Ledger & Settlement",
    description: "Double-entry accounting with balanced journals. Trade settlement after every match.",
    icon: "Scale",
  },
  {
    title: "Market Data",
    description: "Real-time tickers, order book depth, trades, and OHLCV candles from Kafka streams.",
    icon: "BarChart3",
  },
  {
    title: "Binance Adapter",
    description: "BTC/USDT orders route to binance-adapter-service for external liquidity execution.",
    icon: "Link",
  },
  {
    title: "Enterprise Security",
    description: "JWT auth, API gateway, rate limiting, session management, and risk checks.",
    icon: "Shield",
  },
] as const;

export const STATS = [
  { label: "Microservices", value: "12+" },
  { label: "Order Types", value: "Limit/Market" },
  { label: "Matching", value: "Off-chain" },
  { label: "Venues", value: "Internal + Binance" },
] as const;
