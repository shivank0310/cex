import { QUANTITY_SCALE } from "./constants";

export function formatPrice(price: number): string {
  return price.toLocaleString("en-US", { maximumFractionDigits: 2 });
}

export function formatQuantity(qty: number, decimals = 2): string {
  return (qty / QUANTITY_SCALE).toFixed(decimals);
}

export function parseQuantity(input: string): number {
  const n = parseFloat(input);
  if (isNaN(n) || n <= 0) return 0;
  return Math.round(n * QUANTITY_SCALE);
}

export function formatPct(pct: number): string {
  const sign = pct >= 0 ? "+" : "";
  return `${sign}${(pct / 100).toFixed(2)}%`;
}

export function formatUSD(amount: number): string {
  return `$${formatPrice(amount)}`;
}

export function shortAddress(addr: string, chars = 6): string {
  if (addr.length <= chars * 2) return addr;
  return `${addr.slice(0, chars)}...${addr.slice(-chars)}`;
}
