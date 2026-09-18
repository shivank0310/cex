"use client";

import type { MarketTrade } from "@/types";
import { formatPrice, formatQuantity } from "@/lib/format";

export function TradeHistory({ trades }: { trades: MarketTrade[] }) {
  return (
    <div className="rounded-xl border border-white/10 bg-[#12121a] p-4">
      <h3 className="mb-3 text-sm font-semibold text-white">Recent Trades</h3>
      <div className="max-h-48 overflow-y-auto">
        <table className="w-full text-xs font-mono">
          <thead>
            <tr className="text-slate-500">
              <th className="pb-2 text-left">Price</th>
              <th className="pb-2 text-right">Qty</th>
              <th className="pb-2 text-right">Time</th>
            </tr>
          </thead>
          <tbody>
            {trades.length === 0 ? (
              <tr>
                <td colSpan={3} className="py-4 text-center text-slate-500">No trades yet</td>
              </tr>
            ) : (
              trades.map((t) => (
                <tr key={t.id} className="border-t border-white/5">
                  <td className="py-1.5 text-emerald-400">{formatPrice(t.price)}</td>
                  <td className="py-1.5 text-right text-slate-300">{formatQuantity(t.quantity)}</td>
                  <td className="py-1.5 text-right text-slate-500">
                    {new Date(t.timestamp).toLocaleTimeString()}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
