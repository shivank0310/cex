export const QUANTITY_SCALE = 100;

export const DEFAULT_SYMBOL = "BTC/USDT";

export const TRADING_PAIRS = [
  "BTC/USDT",
  "ETH/USDT",
  "SOL/USDT",
] as const;

export const DEMO_TOKENS = {
  "user-a": "token-user-a",
  "seller-1": "token-seller",
} as const;

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
    title: "Blockchain Service",
    description: "EVM isolation layer for deposits, withdrawals, and treasury — not order matching.",
    icon: "Link",
  },
  {
    title: "Enterprise Security",
    description: "Rate limiting, session management, risk checks, and withdrawal reservations.",
    icon: "Shield",
  },
] as const;

export const STATS = [
  { label: "Microservices", value: "12+" },
  { label: "Order Types", value: "Limit/Market" },
  { label: "Matching", value: "Off-chain" },
  { label: "Blockchains", value: "50+" },
] as const;
