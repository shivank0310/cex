"use client";

import { useState } from "react";
import clsx from "clsx";
import { Button } from "@/components/ui/Button";
import { placeOrder } from "@/lib/api/orders";
import { parseQuantity } from "@/lib/format";
import { TRADING_PAIRS } from "@/lib/constants";
import { useTradingStore } from "@/store/trading-store";

export function OrderForm() {
  const {
    symbol, setSymbol, side, setSide, orderType, setOrderType,
    price, setPrice, quantity, setQuantity, authToken,
  } = useTradingStore();
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: "ok" | "err"; text: string } | null>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setMessage(null);
    try {
      const qty = parseQuantity(quantity);
      if (qty <= 0) throw new Error("Invalid quantity");
      const priceNum = orderType === "LIMIT" ? parseInt(price, 10) : 0;
      if (orderType === "LIMIT" && priceNum <= 0) throw new Error("Invalid price");

      const res = await placeOrder(
        { symbol, side, type: orderType, price: priceNum, quantity: qty },
        authToken
      );
      setMessage({
        type: "ok",
        text: `Order ${res.order.id} — ${res.order.status}${res.trades.length ? ` (${res.trades.length} trades)` : ""}`,
      });
      setQuantity("");
    } catch (err) {
      setMessage({
        type: "err",
        text: err instanceof Error ? err.message : "Order failed",
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="rounded-xl border border-white/10 bg-[#12121a] p-4">
      <div className="mb-4 flex gap-2">
        <button
          type="button"
          onClick={() => setSide("BUY")}
          className={clsx(
            "flex-1 rounded-lg py-2 text-sm font-semibold transition-colors",
            side === "BUY" ? "bg-emerald-500 text-white" : "bg-white/5 text-slate-400"
          )}
        >
          Buy
        </button>
        <button
          type="button"
          onClick={() => setSide("SELL")}
          className={clsx(
            "flex-1 rounded-lg py-2 text-sm font-semibold transition-colors",
            side === "SELL" ? "bg-rose-500 text-white" : "bg-white/5 text-slate-400"
          )}
        >
          Sell
        </button>
      </div>

      <select
        value={symbol}
        onChange={(e) => setSymbol(e.target.value)}
        className="mb-3 w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white"
      >
        {TRADING_PAIRS.map((p) => (
          <option key={p} value={p}>{p}</option>
        ))}
      </select>

      <div className="mb-3 flex gap-2">
        {(["LIMIT", "MARKET"] as const).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setOrderType(t)}
            className={clsx(
              "rounded-lg px-3 py-1 text-xs font-medium",
              orderType === t ? "bg-violet-600 text-white" : "bg-white/5 text-slate-400"
            )}
          >
            {t}
          </button>
        ))}
      </div>

      {orderType === "LIMIT" && (
        <label className="mb-3 block">
          <span className="text-xs text-slate-500">Price (USDT)</span>
          <input
            type="number"
            value={price}
            onChange={(e) => setPrice(e.target.value)}
            placeholder="101100"
            className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white"
          />
        </label>
      )}

      <label className="mb-4 block">
        <span className="text-xs text-slate-500">Quantity (BTC)</span>
        <input
          type="number"
          step="0.01"
          value={quantity}
          onChange={(e) => setQuantity(e.target.value)}
          placeholder="0.10"
          className="mt-1 w-full rounded-lg border border-white/10 bg-white/5 px-3 py-2 text-sm text-white"
        />
      </label>

      <Button
        type="submit"
        variant={side === "BUY" ? "buy" : "sell"}
        className="w-full"
        disabled={loading}
      >
        {loading ? "Submitting..." : `${side} ${symbol}`}
      </Button>

      {message && (
        <p className={`mt-3 text-xs ${message.type === "ok" ? "text-emerald-400" : "text-rose-400"}`}>
          {message.text}
        </p>
      )}
    </form>
  );
}
