import { formatCompactCurrency, formatCompactIrr, formatCompactNumber, formatExactIrr } from "@/lib/compact-number";

import {
  COLUMN_CHOICES,
  DEFAULT_COLUMNS,
  DEFAULT_FILTERS,
  MAX_CHART_POINTS,
  OPTIONAL_COLUMNS,
  rangeMs,
  type AnalyticsPoint,
  type Filters,
  type LogRange,
  type LogsFormatter,
} from "./logs-types";

export function isAbortError(error: unknown) {
  return error instanceof Error && error.name === "AbortError";
}

export function buildQuery(filters: Filters, limit?: number, offset?: number) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(serializeFilters(filters))) {
    if (value !== "") params.set(key, value);
  }
  if (limit !== undefined) params.set("limit", String(limit));
  if (offset !== undefined) params.set("offset", String(offset));
  return params.toString();
}

export function buildRangeStatsQuery(filters: Filters) {
  const params = new URLSearchParams();
  const rangeKeys = new Set([
    "duration_min", "duration_max", "tokens_min", "tokens_max",
    "cost_usd_min", "cost_usd_max", "cost_irr_min", "cost_irr_max",
  ]);
  for (const [key, value] of Object.entries(serializeFilters(filters))) {
    if (!rangeKeys.has(key) && value !== "") params.set(key, value);
  }
  return params.toString();
}

export function serializeFilters(filters: Filters): Record<string, string> {
  const output: Record<string, string> = {};
  for (const [key, value] of Object.entries(filters)) {
    if (key !== "range" && value) output[key] = String(value);
  }
  if (filters.range !== "custom") {
    output.from = new Date(Date.now() - rangeMs[filters.range]).toISOString();
    output.to = new Date().toISOString();
  }
  return output;
}

export function filtersFromSearchParams(searchParams: Pick<URLSearchParams, "get">): Filters {
  return {
    ...DEFAULT_FILTERS,
    provider: searchParams.get("provider")?.trim() ?? "",
    model: searchParams.get("model")?.trim() ?? "",
  };
}

export function normalizeVisibleColumns(columns?: string[]) {
  if (!columns?.length) return DEFAULT_COLUMNS;
  const expanded = columns.flatMap((column) => (column === "routing" ? OPTIONAL_COLUMNS : [column]));
  return Array.from(new Set(expanded)).filter((column) => COLUMN_CHOICES.includes(column));
}

export function normalizeRange(range?: LogRange): { min: number; max: number } {
  const min = Number(range?.min ?? 0);
  const max = Number(range?.max ?? 0);
  if (!Number.isFinite(min) || !Number.isFinite(max)) return { min: 0, max: 0 };
  return { min: Math.min(min, max), max: Math.max(min, max) };
}

export function normalizeRangeValue(minValue: string, maxValue: string, bounds: { min: number; max: number }): [number, number] {
  const min = parseFiniteNumber(minValue) ?? bounds.min;
  const max = parseFiniteNumber(maxValue) ?? bounds.max;
  return [
    clampNumber(Math.min(min, max), bounds.min, bounds.max),
    clampNumber(Math.max(min, max), bounds.min, bounds.max),
  ];
}

export function rangeStep(bounds: { min: number; max: number }) {
  const span = bounds.max - bounds.min;
  if (span <= 1) return Math.max(span / 100, 0.0001);
  return Math.max(1, Math.round(span / 100));
}

export function formatFilterNumber(value: number) {
  if (Number.isInteger(value)) return String(value);
  return value.toFixed(6).replace(/0+$/, "").replace(/\.$/, "");
}

export function createLogFormatters(locale: string): LogsFormatter {
  const number = new Intl.NumberFormat(locale);
  const percent = new Intl.NumberFormat(locale, { style: "percent", maximumFractionDigits: 1 });
  const shortTime = new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit" });
  const weekday = new Intl.DateTimeFormat(locale, { weekday: "short" });
  const fullTime = new Intl.DateTimeFormat(locale, { hour: "2-digit", minute: "2-digit", second: "2-digit" });

  return {
    number: (value) => number.format(value),
    numberDisplay: (value) => formatCompactNumber(value, { locale }),
    compactNumber: (value) => formatCompactNumber(value, { locale, compactThreshold: 1_000 }).value,
    percent: (value) => percent.format(value),
    usd: (value) => formatExactCurrencyValue(value, locale),
    usdDisplay: (value) => formatCompactCurrency(value, { locale, currency: "USD", maximumFractionDigits: 4 }),
    irr: (value) => formatExactIrr(value, { locale }),
    irrDisplay: (value) => formatCompactIrr(value, { locale }),
    ms: (value) => `${number.format(Math.round(value))} ms`,
    msValue: (value) => number.format(Math.round(value)),
    timeLabel: (value) => (value ? shortTime.format(new Date(value)) : undefined),
    dateTime: (value) => {
      if (!value) return "-";
      const date = new Date(value);
      return `${weekday.format(date)}${locale.startsWith("fa") ? "،" : ","} ${fullTime.format(date)}`;
    },
  };
}

export function formatTimeseries(points: AnalyticsPoint[], formatters: LogsFormatter) {
  return points.map((point) => ({
    label: formatters.timeLabel(point.timestamp) ?? point.group,
    requests: point.request_count,
    errors: point.error_count,
    latency: Math.round(point.latency_p95_ms ?? 0),
    cost: Number((point.cost_usd ?? 0).toFixed(4)),
  }));
}

export function limitChartPoints<T>(points: T[]): T[] {
  if (points.length <= MAX_CHART_POINTS) return points;
  const step = (points.length - 1) / (MAX_CHART_POINTS - 1);
  return Array.from({ length: MAX_CHART_POINTS }, (_, index) => points[Math.round(index * step)]);
}

export function formatCacheTimeseries(points: AnalyticsPoint[], formatters: LogsFormatter) {
  return points.map((point) => {
    const cacheHit = point.cache_hit_count ?? 0;
    return {
      label: formatters.timeLabel(point.timestamp) ?? point.group,
      cacheHit,
      cacheMiss: Math.max(0, (point.request_count ?? 0) - cacheHit),
    };
  });
}

export function payloadStatus(mode?: string) {
  if (!mode || mode === "none") return "inactive";
  if (mode.includes("redacted") || mode.includes("metadata")) return "warning";
  return "success";
}

function parseFiniteNumber(value: string): number | undefined {
  if (!value.trim()) return undefined;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : undefined;
}

function clampNumber(value: number, min: number, max: number) {
  return Math.min(Math.max(value, min), max);
}

function formatExactCurrencyValue(value: number, locale: string) {
  return formatCompactCurrency(value, {
    locale,
    currency: "USD",
    compactThreshold: Number.POSITIVE_INFINITY,
    maximumFractionDigits: 4,
  }).fullValue;
}
