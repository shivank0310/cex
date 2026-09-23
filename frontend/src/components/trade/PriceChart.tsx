"use client";

import { memo, useId, useMemo, useState } from "react";
import { CANDLE_INTERVALS, intervalLabel, type CandleInterval } from "@/lib/chart-intervals";
import { formatPrice } from "@/lib/format";
import type { Candle } from "@/types";

export type ChartType = "area" | "line" | "candlestick" | "ohlc";

interface PriceChartProps {
  candles: Candle[];
  candlesLoading?: boolean;
  candlesError?: string | null;
  lastPrice?: number;
  interval: CandleInterval;
  onIntervalChange: (interval: CandleInterval) => void;
}

const CHART_TYPES: { id: ChartType; label: string }[] = [
  { id: "area", label: "Area" },
  { id: "line", label: "Line" },
  { id: "candlestick", label: "Candles" },
  { id: "ohlc", label: "OHLC" },
];

const W = 600;
const H = 200;
const PAD_X = 8;

function bounds(candles: Candle[], type: ChartType) {
  if (type === "line" || type === "area") {
    const prices = candles.map((c) => c.close);
    return { min: Math.min(...prices), max: Math.max(...prices) };
  }
  return {
    min: Math.min(...candles.map((c) => c.low)),
    max: Math.max(...candles.map((c) => c.high)),
  };
}

function toY(price: number, min: number, max: number) {
  const range = max - min || 1;
  return H - ((price - min) / range) * H;
}

function candleColor(c: Candle) {
  return c.close >= c.open ? "#34d399" : "#fb7185";
}

function AreaChart({ candles, gradId }: { candles: Candle[]; gradId: string }) {
  const { min, max } = bounds(candles, "area");
  const innerW = W - PAD_X * 2;
  const step = innerW / Math.max(candles.length - 1, 1);
  const points = candles
    .map((c, i) => `${PAD_X + i * step},${toY(c.close, min, max)}`)
    .join(" ");
  const isUp = candles[candles.length - 1].close >= candles[0].open;
  const stroke = isUp ? "#34d399" : "#fb7185";

  return (
    <>
      <defs>
        <linearGradient id={gradId} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor={stroke} stopOpacity="0.35" />
          <stop offset="100%" stopColor={stroke} stopOpacity="0" />
        </linearGradient>
      </defs>
      <polygon points={`${PAD_X},${H} ${points} ${PAD_X + innerW},${H}`} fill={`url(#${gradId})`} />
      <polyline points={points} fill="none" stroke={stroke} strokeWidth="2" vectorEffect="non-scaling-stroke" />
    </>
  );
}

function LineChart({ candles }: { candles: Candle[] }) {
  const { min, max } = bounds(candles, "line");
  const innerW = W - PAD_X * 2;
  const step = innerW / Math.max(candles.length - 1, 1);
  const points = candles
    .map((c, i) => `${PAD_X + i * step},${toY(c.close, min, max)}`)
    .join(" ");
  const isUp = candles[candles.length - 1].close >= candles[0].open;

  return (
    <polyline
      points={points}
      fill="none"
      stroke={isUp ? "#34d399" : "#fb7185"}
      strokeWidth="2"
      vectorEffect="non-scaling-stroke"
    />
  );
}

function CandlestickChart({ candles }: { candles: Candle[] }) {
  const { min, max } = bounds(candles, "candlestick");
  const innerW = W - PAD_X * 2;
  const slot = innerW / candles.length;
  const bodyW = Math.max(2, slot * 0.55);

  return (
    <>
      {candles.map((c, i) => {
        const cx = PAD_X + i * slot + slot / 2;
        const color = candleColor(c);
        const openY = toY(c.open, min, max);
        const closeY = toY(c.close, min, max);
        const highY = toY(c.high, min, max);
        const lowY = toY(c.low, min, max);
        const bodyTop = Math.min(openY, closeY);
        const bodyH = Math.max(1, Math.abs(closeY - openY));

        return (
          <g key={`${c.open_time}-${i}`}>
            <line x1={cx} y1={highY} x2={cx} y2={lowY} stroke={color} strokeWidth="1" vectorEffect="non-scaling-stroke" />
            <rect
              x={cx - bodyW / 2}
              y={bodyTop}
              width={bodyW}
              height={bodyH}
              fill={color}
              stroke={color}
              strokeWidth="0.5"
            />
          </g>
        );
      })}
    </>
  );
}

function OhlcChart({ candles }: { candles: Candle[] }) {
  const { min, max } = bounds(candles, "ohlc");
  const innerW = W - PAD_X * 2;
  const slot = innerW / candles.length;
  const tickW = Math.max(2, slot * 0.22);

  return (
    <>
      {candles.map((c, i) => {
        const cx = PAD_X + i * slot + slot / 2;
        const color = candleColor(c);
        const openY = toY(c.open, min, max);
        const closeY = toY(c.close, min, max);
        const highY = toY(c.high, min, max);
        const lowY = toY(c.low, min, max);

        return (
          <g key={`${c.open_time}-${i}`}>
            <line x1={cx} y1={highY} x2={cx} y2={lowY} stroke={color} strokeWidth="1.5" vectorEffect="non-scaling-stroke" />
            <line x1={cx - tickW} y1={openY} x2={cx} y2={openY} stroke={color} strokeWidth="1.5" vectorEffect="non-scaling-stroke" />
            <line x1={cx} y1={closeY} x2={cx + tickW} y2={closeY} stroke={color} strokeWidth="1.5" vectorEffect="non-scaling-stroke" />
          </g>
        );
      })}
    </>
  );
}

function GridLines({ candles, type }: { candles: Candle[]; type: ChartType }) {
  const { min, max } = bounds(candles, type);
  const mid = (min + max) / 2;
  const lines = [max, mid, min];

  return (
    <>
      {lines.map((price) => (
        <g key={price}>
          <line
            x1={PAD_X}
            y1={toY(price, min, max)}
            x2={W - PAD_X}
            y2={toY(price, min, max)}
            stroke="rgba(255,255,255,0.06)"
            strokeWidth="1"
            vectorEffect="non-scaling-stroke"
          />
        </g>
      ))}
    </>
  );
}

function formatAxisTime(iso: string, interval: CandleInterval): string {
  const d = new Date(iso);
  if (interval === "24h") {
    return d.toLocaleDateString([], { month: "short", day: "numeric" });
  }
  if (interval === "1h" || interval === "30m") {
    return d.toLocaleString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" });
  }
  return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

const MemoCandlestickChart = memo(CandlestickChart);
const MemoOhlcChart = memo(OhlcChart);
const MemoAreaChart = memo(AreaChart);
const MemoLineChart = memo(LineChart);

export function PriceChart({
  candles,
  candlesLoading = false,
  candlesError = null,
  lastPrice,
  interval,
  onIntervalChange,
}: PriceChartProps) {
  const [chartType, setChartType] = useState<ChartType>("candlestick");
  const gradId = useId().replace(/:/g, "");

  const isUp = candles.length > 0 && candles[candles.length - 1].close >= candles[0].open;
  const activeLabel = CHART_TYPES.find((t) => t.id === chartType)?.label ?? "Chart";
  const showLoading = candlesLoading && candles.length === 0;
  const showEmpty = !candlesLoading && candles.length === 0;
  const showOverlay = candlesLoading && candles.length > 0;

  const chartBody = useMemo(() => {
    if (!candles.length) return null;
    return (
      <>
        <GridLines candles={candles} type={chartType} />
        {chartType === "area" && <MemoAreaChart candles={candles} gradId={gradId} />}
        {chartType === "line" && <MemoLineChart candles={candles} />}
        {chartType === "candlestick" && <MemoCandlestickChart candles={candles} />}
        {chartType === "ohlc" && <MemoOhlcChart candles={candles} />}
      </>
    );
  }, [candles, chartType, gradId]);

  function handleIntervalChange(next: CandleInterval) {
    if (next === interval) return;
    onIntervalChange(next);
  }

  return (
    <div className="rounded-xl border border-white/10 bg-[#12121a] p-4">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-wrap items-center gap-2">
          <h3 className="text-sm font-semibold text-white">Price Chart ({intervalLabel(interval)})</h3>
          <span className="text-xs text-slate-500">· {activeLabel}</span>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          {lastPrice !== undefined && (
            <span className={`text-lg font-bold font-mono ${isUp ? "text-emerald-400" : "text-rose-400"}`}>
              {formatPrice(lastPrice)}
            </span>
          )}
        </div>
      </div>

      <div className="mb-3 flex flex-wrap items-center gap-2">
        <div className="flex flex-wrap rounded-lg border border-white/10 bg-white/5 p-0.5">
          {CANDLE_INTERVALS.map((item) => (
            <button
              key={item.value}
              type="button"
              onClick={() => handleIntervalChange(item.value)}
              className={`rounded-md px-2 py-1 text-xs font-medium transition ${
                interval === item.value
                  ? "bg-cyan-600 text-white shadow-sm"
                  : "text-slate-400 hover:bg-white/10 hover:text-white"
              }`}
            >
              {item.label}
            </button>
          ))}
        </div>
        <div className="flex rounded-lg border border-white/10 bg-white/5 p-0.5">
          {CHART_TYPES.map((type) => (
            <button
              key={type.id}
              type="button"
              onClick={() => setChartType(type.id)}
              className={`rounded-md px-2.5 py-1 text-xs font-medium transition ${
                chartType === type.id
                  ? "bg-violet-600 text-white shadow-sm"
                  : "text-slate-400 hover:bg-white/10 hover:text-white"
              }`}
            >
              {type.label}
            </button>
          ))}
        </div>
      </div>

      <div className="relative">
        {showLoading ? (
          <div className="flex h-48 items-center justify-center rounded-lg border border-white/5 bg-black/20 text-sm text-slate-500">
            Loading {intervalLabel(interval)} candles...
          </div>
        ) : showEmpty ? (
          <div className="flex h-48 items-center justify-center rounded-lg border border-white/5 bg-black/20 text-sm text-slate-500">
            {candlesError ?? `No ${intervalLabel(interval)} candle data`}
          </div>
        ) : (
          <>
            <svg
              key={interval}
              viewBox={`0 0 ${W} ${H}`}
              className={`h-48 w-full transition-opacity duration-200 ${showOverlay ? "opacity-60" : "opacity-100"}`}
              preserveAspectRatio="none"
              role="img"
              aria-label={`${activeLabel} price chart`}
            >
              {chartBody}
            </svg>
            {candles.length > 0 && (
              <div className="mt-2 flex justify-between text-[10px] text-slate-500">
                <span>{formatAxisTime(candles[0].open_time, interval)}</span>
                <span className="text-slate-600">Green = up · Red = down</span>
                <span>{formatAxisTime(candles[candles.length - 1].close_time, interval)}</span>
              </div>
            )}
          </>
        )}
        {showOverlay && (
          <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
            <span className="rounded-full bg-black/50 px-3 py-1 text-xs text-slate-300">
              Updating {intervalLabel(interval)}…
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
