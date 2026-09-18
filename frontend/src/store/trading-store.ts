import { create } from "zustand";
import { DEFAULT_SYMBOL, DEMO_TOKENS } from "@/lib/constants";

interface TradingState {
  symbol: string;
  userId: string;
  authToken: string;
  side: "BUY" | "SELL";
  orderType: "LIMIT" | "MARKET";
  price: string;
  quantity: string;
  setSymbol: (s: string) => void;
  setUserId: (id: string) => void;
  setSide: (s: "BUY" | "SELL") => void;
  setOrderType: (t: "LIMIT" | "MARKET") => void;
  setPrice: (p: string) => void;
  setQuantity: (q: string) => void;
}

export const useTradingStore = create<TradingState>((set) => ({
  symbol: DEFAULT_SYMBOL,
  userId: "user-a",
  authToken: DEMO_TOKENS["user-a"],
  side: "BUY",
  orderType: "LIMIT",
  price: "",
  quantity: "",
  setSymbol: (symbol) => set({ symbol }),
  setUserId: (userId) =>
    set({
      userId,
      authToken: DEMO_TOKENS[userId as keyof typeof DEMO_TOKENS] ?? DEMO_TOKENS["user-a"],
    }),
  setSide: (side) => set({ side }),
  setOrderType: (orderType) => set({ orderType }),
  setPrice: (price) => set({ price }),
  setQuantity: (quantity) => set({ quantity }),
}));
