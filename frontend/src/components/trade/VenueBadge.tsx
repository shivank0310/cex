"use client";

import { getTradingVenue } from "@/lib/constants";

export function VenueBadge({ symbol }: { symbol: string }) {
  const venue = getTradingVenue(symbol);
  if (venue === "binance") {
    return (
      <span className="inline-flex items-center rounded-full bg-amber-500/15 px-2.5 py-0.5 text-xs font-medium text-amber-300 ring-1 ring-amber-500/30">
        Binance via adapter
      </span>
    );
  }
  return (
    <span className="inline-flex items-center rounded-full bg-violet-500/15 px-2.5 py-0.5 text-xs font-medium text-violet-300 ring-1 ring-violet-500/30">
      Internal engine
    </span>
  );
}
