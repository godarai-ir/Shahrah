export type CompactNumberOptions = {
  locale?: string;
  maximumFractionDigits?: number;
};

const COMPACT_NUMBER_SCALES = [
  { threshold: 1_000_000, divisor: 1_000_000, suffix: "M" },
  { threshold: 1_000, divisor: 1_000, suffix: "K" },
] as const;

export function formatCompactNumber(
  value: number,
  {
    locale = "fa-IR",
    maximumFractionDigits = 1,
  }: CompactNumberOptions = {},
): string {
  if (!Number.isFinite(value)) return String(value);

  const scale = COMPACT_NUMBER_SCALES.find(
    ({ threshold }) => Math.abs(value) >= threshold,
  );
  const scaledValue = scale ? value / scale.divisor : value;
  const formatted = new Intl.NumberFormat(locale, {
    maximumFractionDigits,
    minimumFractionDigits: 0,
  }).format(scaledValue);

  return scale ? `${formatted}${scale.suffix}` : formatted;
}
