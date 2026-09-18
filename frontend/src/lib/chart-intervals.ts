export const CANDLE_INTERVALS = [
  { value: "1m", label: "1m" },
  { value: "5m", label: "5m" },
  { value: "10m", label: "10m" },
  { value: "15m", label: "15m" },
  { value: "30m", label: "30m" },
  { value: "1h", label: "1H" },
  { value: "24h", label: "24H" },
] as const;

export type CandleInterval = (typeof CANDLE_INTERVALS)[number]["value"];

export function candleLimit(interval: CandleInterval): number {
  switch (interval) {
    case "1m":
      return 120;
    case "5m":
      return 96;
    case "10m":
      return 72;
    case "15m":
      return 96;
    case "30m":
      return 48;
    case "1h":
      return 72;
    case "24h":
      return 60;
  }
}

export function intervalLabel(interval: CandleInterval): string {
  return CANDLE_INTERVALS.find((i) => i.value === interval)?.label ?? interval;
}
