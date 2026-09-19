import { formatCompactCurrency, formatCompactIrr, formatCompactNumber, formatExactIrr } from "@/lib/compact-number";
import {
  DEFAULT_OBSERVE_RANGE,
  observeFiltersFromSearchParams,
  parseObserveRangePreset,
  resolveObserveRange,
  type ObserveFilters,
} from "@/lib/observe-filters";

import {
  COLUMN_CHOICES,
  DEFAULT_COLUMNS,
  DEFAULT_FILTERS,
  MAX_CHART_POINTS,
  OPTIONAL_COLUMNS,
  type AnalyticsPoint,
  type Filters,
  type LogRange,
  type LogsFormatter,
  type RequestLog,
  type SavedView,
} from "./logs-types";
import type { TimingSpan } from "./api";

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
    const range = resolveObserveRange(filters);
    output.from = range.from;
    output.to = range.to;
  }
  return output;
}

export function filtersFromSearchParams(searchParams: Pick<URLSearchParams, "get">): Filters {
  return logsFiltersFromObserve(DEFAULT_FILTERS, observeFiltersFromSearchParams(searchParams));
}

/** Writes the shared observe filters onto a logs filter set. The request ID lookup is the logs `request_id` filter. */
export function logsFiltersFromObserve(base: Filters, observe: ObserveFilters): Filters {
  return {
    ...base,
    request_id: observe.q,
    range: observe.range,
    from: observe.range === "custom" ? observe.from : "",
    to: observe.range === "custom" ? observe.to : "",
    model: observe.model,
    provider: observe.provider,
    status_code: observe.status,
  };
}

/**
 * Reads the shared part of a logs filter set. Logs have no trace error code, so the shared one is
 * passed through untouched to survive a round trip through this view.
 */
export function observeFromLogsFilters(filters: Filters, errorCode: string): ObserveFilters {
  const custom = filters.range === "custom";
  return {
    q: filters.request_id.trim(),
    range: filters.range,
    from: custom ? filters.from : "",
    to: custom ? filters.to : "",
    model: filters.model.trim(),
    provider: filters.provider.trim(),
    status: filters.status_code.trim(),
    errorCode,
  };
}

type FilterSetInput = Partial<Record<keyof Filters, unknown>>;

/**
 * Reduces a filter set to the values that change the query: defaults fill blanks, unknown keys
 * are dropped, and `from`/`to` only count for a custom range because relative ranges derive them.
 */
export function normalizeFilterSet(filters: FilterSetInput): Record<string, string> {
  const keys = Object.keys(DEFAULT_FILTERS) as Array<keyof Filters>;
  const text = (key: keyof Filters) => {
    const value = filters[key];
    const trimmed = value === undefined || value === null ? "" : String(value).trim();
    return trimmed || String(DEFAULT_FILTERS[key]);
  };
  // Older saved views stored "24h"-style ranges; compare them by the preset they stand for.
  const range = parseObserveRangePreset(text("range")) ?? DEFAULT_OBSERVE_RANGE;
  const output: Record<string, string> = {};
  for (const key of keys) {
    if ((key === "from" || key === "to") && range !== "custom") continue;
    const value = key === "range" ? range : text(key);
    if (value) output[key] = value;
  }
  return output;
}

export function isSameFilterSet(a: FilterSetInput, b: FilterSetInput): boolean {
  const left = normalizeFilterSet(a);
  const right = normalizeFilterSet(b);
  const leftKeys = Object.keys(left);
  return leftKeys.length === Object.keys(right).length && leftKeys.every((key) => left[key] === right[key]);
}

/** The filter set a saved view stores: the relative range itself rather than timestamps derived from it. */
export function savedViewFilters(filters: Filters): Record<string, string> {
  return normalizeFilterSet(filters);
}

/** Restores a saved view's filters. Older views stored derived from/to without a range; those fall back to the default range. */
export function filtersFromSavedView(view: Pick<SavedView, "filters">): Filters {
  const normalized = normalizeFilterSet(view.filters);
  const range = parseObserveRangePreset(normalized.range) ?? DEFAULT_FILTERS.range;
  const sortOrder = normalized.sort_order === "asc" ? "asc" : "desc";
  return { ...DEFAULT_FILTERS, ...normalized, range, sort_order: sortOrder };
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

export type TimingSegment = { spanID: string; name: string; durationMs: number; share: number };
export type TimingBreakdown = {
  totalMs: number;
  segments: TimingSegment[];
  /** False when child spans overlap, so their shares cannot be stacked into one bar. */
  stackable: boolean;
};

/**
 * Splits a request's response time by the real spans of its trace: the children of the root span,
 * each as a share of the root's duration. Returns null when there is nothing to show; it never
 * estimates a split from the log's own duration.
 */
export function buildTimingBreakdown(spans: Array<Pick<TimingSpan, "span_id" | "parent_span_id" | "name" | "started_at" | "duration_ms">>): TimingBreakdown | null {
  const valid = spans.filter((span) => Number.isFinite(Date.parse(span.started_at)) && Number.isFinite(span.duration_ms) && span.duration_ms >= 0);
  if (valid.length === 0) return null;
  const root = valid.find((span) => !span.parent_span_id) ?? null;
  const startOf = (span: (typeof valid)[number]) => Date.parse(span.started_at);
  const totalMs = root
    ? root.duration_ms
    : Math.max(...valid.map((span) => startOf(span) + span.duration_ms)) - Math.min(...valid.map(startOf));
  if (!(totalMs > 0)) return null;

  // Direct children only: nested spans are already inside their parent's time.
  const children = valid.filter((span) => (root ? span.parent_span_id === root.span_id : true)).sort((a, b) => startOf(a) - startOf(b));
  const parts = children.length > 0 ? children : root ? [root] : [];
  const segments = parts.map((span) => ({
    spanID: span.span_id,
    name: span.name,
    durationMs: span.duration_ms,
    share: Math.min(1, span.duration_ms / totalMs),
  }));
  const stackable = parts.every((span, index) => index === 0 || startOf(span) >= startOf(parts[index - 1]) + parts[index - 1].duration_ms)
    && segments.reduce((sum, segment) => sum + segment.share, 0) <= 1;
  return { totalMs, segments, stackable };
}

/** Request signals shown in the table and the detail sheet. */
export function requestSignals(log: Pick<RequestLog, "retry_count" | "cache_status" | "fallback_count" | "fallback_status">) {
  return {
    retries: log.retry_count ?? 0,
    cacheHit: (log.cache_status ?? "").toLowerCase().includes("hit"),
    fallback: (log.fallback_count ?? 0) > 0 || log.fallback_status === "active",
  };
}
