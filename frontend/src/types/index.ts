export interface Ticker {
  symbol: string;
  last_price: number;
  best_bid: number;
  best_ask: number;
  high_24h: number;
  low_24h: number;
  volume_24h: number;
  quote_volume_24h: number;
  price_change_24h: number;
  price_change_pct_24h: number;
  trade_count_24h: number;
  updated_at: string;
}

export interface DepthLevel {
  price: number;
  quantity: number;
}

export interface OrderBook {
  symbol: string;
  bids: DepthLevel[];
  asks: DepthLevel[];
  updated_at: string;
}

export interface MarketTrade {
  id: string;
  symbol: string;
  price: number;
  quantity: number;
  notional: number;
  timestamp: string;
}

export interface Candle {
  symbol: string;
  interval: string;
  open_time: string;
  close_time: string;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
  quote_volume: number;
  trade_count: number;
}

export interface Order {
  id: string;
  user_id: string;
  symbol: string;
  side: "BUY" | "SELL";
  type: "LIMIT" | "MARKET";
  status: string;
  price: number;
  quantity: number;
  remaining: number;
  filled: number;
  sequence: number;
  created_at: string;
  updated_at: string;
}

export interface PlaceOrderRequest {
  symbol: string;
  side: "BUY" | "SELL";
  type: "LIMIT" | "MARKET";
  price: number;
  quantity: number;
}

export interface PlaceOrderResponse {
  order: Order;
  trades: Array<{
    id: string;
    symbol: string;
    price: number;
    quantity: number;
    timestamp: string;
  }>;
}

export interface Balance {
  user_id: string;
  asset: string;
  available: number;
  locked: number;
  total: number;
}

export interface WalletAddress {
  id: string;
  user_id: string;
  asset: string;
  chain: string;
  address: string;
  created_at: string;
}

export interface ApiError {
  code: string;
  message: string;
}
