import type { ObserveFilters } from "@/lib/observe-filters";

export type MetricsRow = Record<string, unknown>;

export const METRICS_VIEW_BY = ["modelName", "virtualModelName", "userEmail", "virtualAccount", "team"] as const;
export const METRICS_PERCENTILES = ["p50", "p75", "p90", "p99"] as const;
/** What the focus chart and the breakdown table measure. */
export const METRIC_KINDS = ["requests", "latency", "tokens", "cost"] as const;

export type MetricsViewBy = (typeof METRICS_VIEW_BY)[number];
export type MetricsPercentile = (typeof METRICS_PERCENTILES)[number];
export type MetricKind = (typeof METRIC_KINDS)[number];

export type BreakdownRow = {
  label: string;
  value: number;
  /** Fraction of the total. Only for additive metrics. */
  share?: number;
  /** Multiple of the overall value. Only for latency, whose percentiles do not add up. */
  ratio?: number;
};

type CacheBreakdownItem = {
  mode: string;
  exactHits: number;
  semanticHits: number;
  misses: number;
  avoidedCost: number;
};

type SeriesPoint = {
  timestamp: string;
  label: string;
  tooltipLabel: string;
} & Record<string, number | string>;

type AxisLabelMode = "time" | "dateTime";

const MAX_SERIES = 5;

export function buildChartData(
  rows: MetricsRow[],
  viewBy: MetricsViewBy,
  pct: MetricsPercentile,
  unknownLabel: string,
  locale?: string,
  rangeFrom?: string,
  rangeTo?: string,
) {
  const timeSet = new Set<string>();
  const grouped = new Map<string, MetricsRow[]>();

  for (const row of rows) {
    const key = metricString(row, viewBy) || unknownLabel;
    const ts = metricString(row, "timestamp");
    if (!grouped.has(key)) grouped.set(key, []);
    grouped.get(key)!.push(row);
    if (ts) timeSet.add(ts);
  }

  const x = Array.from(timeSet).sort((a, b) => new Date(a).getTime() - new Date(b).getTime());
  const labelMode = getAxisLabelMode(x, rangeFrom, rangeTo);
  const series = Array.from(grouped.entries())
    .slice(0, MAX_SERIES)
    .map(([label], index) => ({ label, key: `series${index}` }));

  const points: SeriesPoint[] = x.map((timestamp) => {
    const point: SeriesPoint = {
      timestamp,
      label: formatAxisLabel(timestamp, labelMode, locale),
      tooltipLabel: formatLocalDateTime(timestamp, locale),
    };

    series.forEach((seriesItem) => {
      const item = grouped.get(seriesItem.label)?.find((row) => metricString(row, "timestamp") === timestamp);
      point[seriesItem.key] = metricNumber(item, "countModelName", "countAll");
    });

    return point;
  });

  const latencyPoints: SeriesPoint[] = x.map((timestamp) => {
    const point: SeriesPoint = {
      timestamp,
      label: formatAxisLabel(timestamp, labelMode, locale),
      tooltipLabel: formatLocalDateTime(timestamp, locale),
    };
    const latencyKey = `${pct}LatencyMs`;
    const latencyKeyAlt = `${pct}LatencyMS`;

    series.forEach((seriesItem) => {
      const item = grouped.get(seriesItem.label)?.find((row) => metricString(row, "timestamp") === timestamp);
      point[seriesItem.key] = metricNumber(item, latencyKey, latencyKeyAlt);
    });

    return point;
  });

  const seriesPoints = (valueOf: (item: MetricsRow | undefined) => number): SeriesPoint[] => x.map((timestamp) => {
    const point: SeriesPoint = {
      timestamp,
      label: formatAxisLabel(timestamp, labelMode, locale),
      tooltipLabel: formatLocalDateTime(timestamp, locale),
    };
    series.forEach((seriesItem) => {
      const item = grouped.get(seriesItem.label)?.find((row) => metricString(row, "timestamp") === timestamp);
      point[seriesItem.key] = valueOf(item);
    });
    return point;
  });
  const tokenPoints = seriesPoints((item) => metricNumber(item, "sumInputTokens") + metricNumber(item, "sumOutputTokens"));
  const costPoints = seriesPoints((item) => Number(metricNumber(item, "sumCostInUSD", "sumCostUsd").toFixed(6)));

  const tokenCostPoints = x.map((timestamp) => {
    let inputTokens = 0;
    let outputTokens = 0;
    let costUsd = 0;

    for (const [, items] of grouped) {
      const item = items.find((row) => metricString(row, "timestamp") === timestamp);
      inputTokens += metricNumber(item, "sumInputTokens");
      outputTokens += metricNumber(item, "sumOutputTokens");
      costUsd += metricNumber(item, "sumCostInUSD", "sumCostUsd");
    }

    return {
      timestamp,
      label: formatAxisLabel(timestamp, labelMode, locale),
      tooltipLabel: formatLocalDateTime(timestamp, locale),
      inputTokens,
      outputTokens,
      costUsd: Number(costUsd.toFixed(6)),
    };
  });

  return { points, latencyPoints, tokenPoints, costPoints, tokenCostPoints, series };
}

/**
 * Breakdown table rows from the grouped distribution query. Requests, tokens and cost are additive
 * and get a share of the total; latency percentiles are not, so each group is compared with the
 * overall percentile instead.
 */
export function buildBreakdownRows(
  rows: MetricsRow[],
  viewBy: MetricsViewBy,
  metric: MetricKind,
  pct: MetricsPercentile,
  overallLatencyMs: number,
  unknownLabel: string,
): BreakdownRow[] {
  const values = rows.map((row) => ({ label: metricString(row, viewBy) || unknownLabel, value: breakdownValue(row, metric, pct) }));
  values.sort((a, b) => b.value - a.value);
  if (metric === "latency") {
    return values.map((item) => ({ ...item, ratio: overallLatencyMs > 0 ? item.value / overallLatencyMs : undefined }));
  }
  const total = values.reduce((sum, item) => sum + item.value, 0);
  return values.map((item) => ({ ...item, share: total > 0 ? item.value / total : 0 }));
}

export function latencyValue(row: MetricsRow | undefined, pct: MetricsPercentile) {
  return metricNumber(row, `${pct}LatencyMs`, `${pct}LatencyMS`);
}

function breakdownValue(row: MetricsRow, metric: MetricKind, pct: MetricsPercentile) {
  if (metric === "latency") return latencyValue(row, pct);
  if (metric === "tokens") return metricNumber(row, "sumInputTokens") + metricNumber(row, "sumOutputTokens");
  if (metric === "cost") return metricNumber(row, "sumCostInUSD", "sumCostUsd");
  return metricNumber(row, "countModelName", "countAll");
}

/** Maps the shared observe filters onto the metrics query. The trace error code has no metrics field. */
export function buildMetricsFilters(observe: ObserveFilters): Array<Record<string, unknown>> {
  const filters: Array<Record<string, unknown>> = [];
  if (observe.model) filters.push({ fieldName: "modelName", operator: "STRING_CONTAINS", value: observe.model });
  if (observe.provider) filters.push({ fieldName: "providerName", operator: "EQUAL", value: observe.provider });
  if (observe.q) filters.push({ fieldName: "requestID", operator: "EQUAL", value: observe.q });
  if (observe.status) filters.push({ fieldName: "statusCode", operator: "EQUAL", value: Number(observe.status) });
  return filters;
}

export function buildCacheBreakdown(rows: MetricsRow[]): CacheBreakdownItem[] {
  const items = new Map<string, CacheBreakdownItem>([
    ["simple", { mode: "simple", exactHits: 0, semanticHits: 0, misses: 0, avoidedCost: 0 }],
    ["smart", { mode: "smart", exactHits: 0, semanticHits: 0, misses: 0, avoidedCost: 0 }],
  ]);
  for (const row of rows) {
    const mode = metricString(row, "cacheMode");
    const item = items.get(mode);
    if (!item) continue;
    const status = metricString(row, "cacheStatus");
    const match = metricString(row, "cacheMatch");
    const count = metricNumber(row, "countModelName", "countAll");
    if (status === "hit" && match === "semantic") item.semanticHits += count;
    if (status === "hit" && match === "exact") item.exactHits += count;
    if (status === "miss") item.misses += count;
    item.avoidedCost += metricNumber(row, "sumAvoidedCostUSD");
  }
  return Array.from(items.values());
}

export function buildSavingsChartData(
  rows: MetricsRow[],
  locale?: string,
  rangeFrom?: string,
  rangeTo?: string,
) {
  const buckets = new Map<string, { actualCost: number; simpleSavings: number; smartSavings: number }>();
  for (const row of rows) {
    const timestamp = metricString(row, "timestamp");
    if (!timestamp) continue;
    const bucket = buckets.get(timestamp) ?? { actualCost: 0, simpleSavings: 0, smartSavings: 0 };
    bucket.actualCost += metricNumber(row, "sumCostInUSD", "sumCostUsd");
    const avoided = metricNumber(row, "sumAvoidedCostUSD");
    const mode = metricString(row, "cacheMode");
    if (mode === "simple") bucket.simpleSavings += avoided;
    if (mode === "smart") bucket.smartSavings += avoided;
    buckets.set(timestamp, bucket);
  }
  const timestamps = Array.from(buckets.keys()).sort((a, b) => new Date(a).getTime() - new Date(b).getTime());
  const labelMode = getAxisLabelMode(timestamps, rangeFrom, rangeTo);
  return timestamps.map((timestamp) => {
    const bucket = buckets.get(timestamp)!;
    return {
      timestamp,
      label: formatAxisLabel(timestamp, labelMode, locale),
      tooltipLabel: formatLocalDateTime(timestamp, locale),
      actualCost: Number(bucket.actualCost.toFixed(8)),
      simpleSavings: Number(bucket.simpleSavings.toFixed(8)),
      smartSavings: Number(bucket.smartSavings.toFixed(8)),
    };
  });
}

export function metricNumber(row: MetricsRow | undefined, ...keys: string[]) {
  const value = metricValue(row, keys);
  if (typeof value === "number") return Number.isFinite(value) ? value : 0;
  if (typeof value === "bigint") return Number(value);
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

export function metricString(row: MetricsRow | undefined, ...keys: string[]) {
  const value = metricValue(row, keys);
  if (value == null) return "";
  if (value instanceof Date) return value.toISOString();
  return String(value);
}

export function normalizeStringList(value: unknown) {
  if (!Array.isArray(value)) return [];
  return Array.from(
    new Set(
      value
        .map((item) => (typeof item === "string" ? item.trim() : ""))
        .filter(Boolean),
    ),
  ).sort((a, b) => a.localeCompare(b));
}

function metricValue(row: MetricsRow | undefined, keys: string[]) {
  if (!row) return undefined;
  const candidates = new Set<string>();
  for (const key of keys) {
    candidates.add(key);
    candidates.add(key.toLowerCase());
    candidates.add(toSnakeCase(key));
    for (const alias of metricAliases[key] ?? []) {
      candidates.add(alias);
      candidates.add(alias.toLowerCase());
    }
  }
  for (const key of candidates) {
    if (Object.prototype.hasOwnProperty.call(row, key)) {
      return row[key];
    }
  }
  return undefined;
}

function toSnakeCase(value: string) {
  return value.replace(/([a-z0-9])([A-Z])/g, "$1_$2").toLowerCase();
}

const metricAliases: Record<string, string[]> = {
  modelName: ["model_name", "view_model_name"],
  virtualModelName: ["virtual_model_name", "view_virtual_model_name"],
  userEmail: ["user_email", "view_user_email"],
  virtualAccount: ["virtual_account", "virtual_account_name", "view_virtual_account_name"],
  team: ["team_name", "view_team_name"],
  errorCode: ["error_code", "status_code"],
  timestamp: ["bucket_ts"],
  countModelName: ["count_model_name", "countmodelname", "request_count"],
  countAll: ["count_all", "countall"],
  sumInputTokens: ["sum_input_tokens", "suminputtokens", "input_tokens_sum"],
  sumOutputTokens: ["sum_output_tokens", "sumoutputtokens", "output_tokens_sum"],
  sumCostInUSD: ["sum_cost_in_usd", "sumcostinusd", "sum_cost_usd", "cost_usd_sum"],
  sumCostUsd: ["sum_cost_usd", "sumcostusd", "cost_usd_sum"],
  cacheMode: ["cache_mode"],
  cacheStatus: ["cache_status"],
  cacheMatch: ["cache_match"],
  sumAvoidedInputTokens: ["sum_avoided_input_tokens", "sumavoidedinputtokens"],
  sumAvoidedOutputTokens: ["sum_avoided_output_tokens", "sumavoidedoutputtokens"],
  sumAvoidedCostUSD: ["sum_avoided_cost_usd", "sumavoidedcostusd"],
  p50LatencyMs: ["p50_latency_ms", "p50latencyms", "latency_p50"],
  p75LatencyMs: ["p75_latency_ms", "p75latencyms", "latency_p75"],
  p90LatencyMs: ["p90_latency_ms", "p90latencyms", "latency_p90"],
  p99LatencyMs: ["p99_latency_ms", "p99latencyms", "latency_p99"],
};

function getAxisLabelMode(timestamps: string[], rangeFrom?: string, rangeTo?: string): AxisLabelMode {
  const rangeStart = parseTime(rangeFrom);
  const rangeEnd = parseTime(rangeTo);
  if (rangeStart != null && rangeEnd != null && rangeEnd - rangeStart >= 24 * 60 * 60 * 1000) {
    return "dateTime";
  }

  const first = parseTime(timestamps[0]);
  const last = parseTime(timestamps[timestamps.length - 1]);
  if (first != null && last != null && last - first >= 24 * 60 * 60 * 1000) {
    return "dateTime";
  }

  return "time";
}

function parseTime(value: string | undefined) {
  if (!value) return undefined;
  const time = new Date(value).getTime();
  return Number.isNaN(time) ? undefined : time;
}

function formatAxisLabel(value: string, mode: AxisLabelMode, locale?: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  if (mode === "dateTime") {
    return new Intl.DateTimeFormat(locale, {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  }
  return new Intl.DateTimeFormat(locale, {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

function formatLocalDateTime(value: string, locale?: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(locale, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}
