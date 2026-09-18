"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getTickers } from "@/lib/api/market";
import { formatPct, formatPrice } from "@/lib/format";
import type { Ticker } from "@/types";

export default function MarketsPage() {
  const [tickers, setTickers] = useState<Ticker[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getTickers()
      .then(setTickers)
      .catch(() => setTickers([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
      <h1 className="text-2xl font-bold text-white">Markets</h1>
      <p className="mt-2 text-slate-400">Real-time tickers from market-data service</p>

      <div className="mt-8 overflow-hidden rounded-2xl border border-white/10">
        <table className="w-full text-sm">
          <thead className="bg-white/5 text-slate-400">
            <tr>
              <th className="px-4 py-3 text-left">Pair</th>
              <th className="px-4 py-3 text-right">Last Price</th>
              <th className="px-4 py-3 text-right">24h Change</th>
              <th className="px-4 py-3 text-right">24h High</th>
              <th className="px-4 py-3 text-right">24h Low</th>
              <th className="px-4 py-3 text-right">Volume</th>
              <th className="px-4 py-3 text-right">Action</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-slate-500">Loading...</td></tr>
            ) : tickers.length === 0 ? (
              <tr><td colSpan={7} className="px-4 py-8 text-center text-slate-500">No market data — start order-service & market-data</td></tr>
            ) : (
              tickers.map((t) => {
                const up = t.price_change_pct_24h >= 0;
                return (
                  <tr key={t.symbol} className="border-t border-white/5 hover:bg-white/5">
                    <td className="px-4 py-3 font-semibold text-white">{t.symbol}</td>
                    <td className="px-4 py-3 text-right font-mono">{formatPrice(t.last_price)}</td>
                    <td className={`px-4 py-3 text-right font-mono ${up ? "text-emerald-400" : "text-rose-400"}`}>
                      {formatPct(t.price_change_pct_24h)}
                    </td>
                    <td className="px-4 py-3 text-right font-mono text-slate-300">{formatPrice(t.high_24h)}</td>
                    <td className="px-4 py-3 text-right font-mono text-slate-300">{formatPrice(t.low_24h)}</td>
                    <td className="px-4 py-3 text-right font-mono text-slate-300">{t.volume_24h}</td>
                    <td className="px-4 py-3 text-right">
                      <Link href="/trade" className="text-cyan-400 hover:underline">Trade</Link>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
