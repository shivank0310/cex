"use client";

import type { DepthLevel } from "@/types";
import { formatPrice, formatQuantity } from "@/lib/format";

interface OrderBookProps {
  bids: DepthLevel[];
  asks: DepthLevel[];
}

function DepthRow({
  level,
  side,
  maxQty,
}: {
  level: DepthLevel;
  side: "bid" | "ask";
  maxQty: number;
}) {
  const pct = maxQty > 0 ? (level.quantity / maxQty) * 100 : 0;
  return (
    <div className="relative grid grid-cols-2 gap-2 py-0.5 text-xs font-mono">
      <div
        className={`absolute inset-y-0 ${side === "bid" ? "right-0 bg-emerald-500/10" : "left-0 bg-rose-500/10"}`}
        style={{ width: `${pct}%` }}
      />
      <span className={`relative z-10 ${side === "bid" ? "text-emerald-400" : "text-rose-400"}`}>
        {formatPrice(level.price)}
      </span>
      <span className="relative z-10 text-right text-slate-300">{formatQuantity(level.quantity)}</span>
    </div>
  );
}

export function OrderBook({ bids, asks }: OrderBookProps) {
  const topAsks = [...asks].reverse().slice(0, 12);
  const topBids = bids.slice(0, 12);
  const maxQty = Math.max(
    ...topBids.map((b) => b.quantity),
    ...topAsks.map((a) => a.quantity),
    1
  );

  return (
    <div className="rounded-xl border border-white/10 bg-[#12121a] p-4">
      <h3 className="mb-3 text-sm font-semibold text-white">Order Book</h3>
      <div className="mb-1 grid grid-cols-2 gap-2 text-xs text-slate-500">
        <span>Price (USDT)</span>
        <span className="text-right">Amount (BTC)</span>
      </div>
      <div className="space-y-0.5">
        {topAsks.map((a, i) => (
          <DepthRow key={`a-${i}`} level={a} side="ask" maxQty={maxQty} />
        ))}
      </div>
      <div className="my-2 border-y border-white/10 py-2 text-center text-sm font-semibold text-cyan-400">
        Spread
      </div>
      <div className="space-y-0.5">
        {topBids.map((b, i) => (
          <DepthRow key={`b-${i}`} level={b} side="bid" maxQty={maxQty} />
        ))}
      </div>
    </div>
  );
}
