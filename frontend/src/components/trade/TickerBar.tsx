"use client";

import type { Ticker } from "@/types";
import { formatPct, formatPrice } from "@/lib/format";

export function TickerBar({ ticker }: { ticker: Ticker | null }) {
  if (!ticker) {
    return (
      <div className="rounded-xl border border-white/10 bg-[#12121a] px-4 py-3 text-slate-500 text-sm">
        Loading ticker...
      </div>
    );
  }

  const up = ticker.price_change_pct_24h >= 0;

  return (
    <div className="flex flex-wrap items-center gap-6 rounded-xl border border-white/10 bg-[#12121a] px-4 py-3">
      <div>
        <span className="text-xs text-slate-500">{ticker.symbol}</span>
        <p className="text-xl font-bold font-mono text-white">{formatPrice(ticker.last_price)}</p>
      </div>
      <div>
        <span className="text-xs text-slate-500">24h Change</span>
        <p className={`font-semibold ${up ? "text-emerald-400" : "text-rose-400"}`}>
          {formatPct(ticker.price_change_pct_24h)}
        </p>
      </div>
      <div>
        <span className="text-xs text-slate-500">24h High</span>
        <p className="font-mono text-slate-300">{formatPrice(ticker.high_24h)}</p>
      </div>
      <div>
        <span className="text-xs text-slate-500">24h Low</span>
        <p className="font-mono text-slate-300">{formatPrice(ticker.low_24h)}</p>
      </div>
      <div>
        <span className="text-xs text-slate-500">24h Volume</span>
        <p className="font-mono text-slate-300">{ticker.volume_24h}</p>
      </div>
    </div>
  );
}
