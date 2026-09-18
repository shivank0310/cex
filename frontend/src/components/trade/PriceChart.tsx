"use client";

import type { Candle } from "@/types";
import { formatPrice } from "@/lib/format";

interface PriceChartProps {
  candles: Candle[];
  lastPrice?: number;
}

export function PriceChart({ candles, lastPrice }: PriceChartProps) {
  if (!candles.length) {
    return (
      <div className="flex h-64 items-center justify-center rounded-xl border border-white/10 bg-[#12121a] text-slate-500 text-sm">
        Waiting for candle data...
      </div>
    );
  }

  const prices = candles.map((c) => c.close);
  const min = Math.min(...prices);
  const max = Math.max(...prices);
  const range = max - min || 1;
  const h = 200;
  const w = 600;
  const step = w / Math.max(candles.length - 1, 1);

  const points = candles
    .map((c, i) => {
      const x = i * step;
      const y = h - ((c.close - min) / range) * h;
      return `${x},${y}`;
    })
    .join(" ");

  const isUp = candles[candles.length - 1].close >= candles[0].open;

  return (
    <div className="rounded-xl border border-white/10 bg-[#12121a] p-4">
      <div className="mb-2 flex items-baseline justify-between">
        <h3 className="text-sm font-semibold text-white">Price Chart (1m)</h3>
        {lastPrice && (
          <span className={`text-lg font-bold font-mono ${isUp ? "text-emerald-400" : "text-rose-400"}`}>
            {formatPrice(lastPrice)}
          </span>
        )}
      </div>
      <svg viewBox={`0 0 ${w} ${h}`} className="w-full h-48" preserveAspectRatio="none">
        <defs>
          <linearGradient id="chartGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={isUp ? "#10b981" : "#f43f5e"} stopOpacity="0.3" />
            <stop offset="100%" stopColor={isUp ? "#10b981" : "#f43f5e"} stopOpacity="0" />
          </linearGradient>
        </defs>
        <polygon
          points={`0,${h} ${points} ${w},${h}`}
          fill="url(#chartGrad)"
        />
        <polyline
          points={points}
          fill="none"
          stroke={isUp ? "#34d399" : "#fb7185"}
          strokeWidth="2"
        />
      </svg>
    </div>
  );
}
