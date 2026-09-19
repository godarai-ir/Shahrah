import type { ObserveFilters } from "@/lib/observe-filters";

import {
  BASE_FILTERS,
  type BreakdownBucket,
  type FacetBucket,
  type RequestTrace,
  type StatusFacetBucket,
  type TraceFilters,
  type TraceLocalFilters,
} from "./trace-types";

export function createDefaultLocalFilters(): TraceLocalFilters {
  return { traceID: "", route: "", sortBy: BASE_FILTERS.sortBy, sortOrder: BASE_FILTERS.sortOrder };
}

/** The trace list query: shared filters plus trace-only ones. `range` is the resolved window. */
export function traceFiltersFromObserve(local: TraceLocalFilters, observe: ObserveFilters, range: { from: string; to: string }): TraceFilters {
  return normalizeFilters({
    ...BASE_FILTERS,
    ...local,
    requestID: observe.q,
    model: observe.model,
    provider: observe.provider,
    statusCode: observe.status,
    errorCode: observe.errorCode,
    from: range.from,
    to: range.to,
  });
}

export function buildTraceFiltersQuery(filters: TraceFilters): string {
  const params = new URLSearchParams();
  const entries: Array<[string, string]> = [
    ["trace_id", filters.traceID],
    ["request_id", filters.requestID],
    ["model", filters.model],
    ["provider", filters.provider],
    ["route", filters.route],
    ["status_code", filters.statusCode],
    ["error_code", filters.errorCode],
    ["from", filters.from],
    ["to", filters.to],
    ["sort_by", filters.sortBy],
    ["sort_order", filters.sortOrder],
  ];
  for (const [key, value] of entries) {
    if (value) params.set(key, value);
  }
  return params.toString();
}

export function normalizeError(error: unknown, fallback: string) {
  if (error instanceof Error) return error.message;
  return fallback;
}

export function normalizeFilters(filters: TraceFilters): TraceFilters {
  return {
    ...filters,
    traceID: filters.traceID.trim(),
    requestID: filters.requestID.trim(),
    model: filters.model.trim(),
    provider: filters.provider.trim(),
    route: filters.route.trim(),
    statusCode: filters.statusCode.trim(),
    errorCode: filters.errorCode.trim(),
    from: filters.from.trim(),
    to: filters.to.trim(),
    sortBy: (filters.sortBy || "started_at").trim(),
    sortOrder: filters.sortOrder === "asc" ? "asc" : "desc",
  };
}

export function normalizeFacetBuckets(buckets: FacetBucket[] | undefined, emptyLabel: string): BreakdownBucket[] {
  return (buckets ?? [])
    .map((item) => ({ value: item.value || "", label: item.value || emptyLabel, count: item.count }))
    .sort(sortBuckets);
}

export function normalizeComboboxOptions(buckets: FacetBucket[] | undefined) {
  return Array.from(new Set((buckets ?? []).map((item) => item.value.trim()).filter(Boolean)))
    .sort((a, b) => a.localeCompare(b))
    .map((value) => ({ value, label: value }));
}

export function aggregateStatusBuckets(items: RequestTrace[]): StatusFacetBucket[] {
  const counts = new Map<number, number>();
  for (const item of items) counts.set(item.status_code, (counts.get(item.status_code) ?? 0) + 1);
  return Array.from(counts.entries()).map(([status_code, count]) => ({ status_code, count }));
}

export function statusBadgeVariant(statusCode: number): "success" | "info" | "warning" | "destructive" | "outline" {
  if (statusCode >= 500) return "destructive";
  if (statusCode >= 400) return "warning";
  if (statusCode >= 300) return "info";
  if (statusCode >= 200) return "success";
  return "outline";
}

export function compactID(value: string) {
  if (value.length <= 22) return value;
  return `${value.slice(0, 12)}...${value.slice(-6)}`;
}

export function formatPercent(value: number, locale: string) {
  return new Intl.NumberFormat(locale, { style: "percent", maximumFractionDigits: 1 }).format(value);
}

export function formatDuration(value: number, msLabel: string, sLabel: string, locale: string) {
  if (!Number.isFinite(value)) return `0 ${msLabel}`;
  if (value >= 1000) {
    return `${new Intl.NumberFormat(locale, { maximumFractionDigits: 1 }).format(value / 1000)} ${sLabel}`;
  }
  return `${new Intl.NumberFormat(locale, { maximumFractionDigits: 0 }).format(value)} ${msLabel}`;
}

export function durationColor(value: number) {
  if (!Number.isFinite(value)) return "hsl(212 90% 45%)";
  const clamped = Math.max(0, Math.min(value, 5000));
  const progress = clamped / 5000;
  const hue = 212 - progress * 212;
  const saturation = 90 - progress * 8;
  const lightness = 45 + progress * 8;
  return `hsl(${hue.toFixed(0)} ${saturation.toFixed(0)}% ${lightness.toFixed(0)}%)`;
}

export function formatLocalDateTime(value: string, locale: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "-";
  return new Intl.DateTimeFormat(locale, {
    month: "short", day: "numeric", hour: "2-digit", minute: "2-digit", second: "2-digit",
  }).format(date);
}

export function sortBuckets(a: { count: number; value: string }, b: { count: number; value: string }) {
  if (b.count !== a.count) return b.count - a.count;
  return a.value.localeCompare(b.value);
}
