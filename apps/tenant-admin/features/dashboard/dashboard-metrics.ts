import type { AnalyticsPoint } from "./api";

export const HOUR_MS = 60 * 60 * 1_000;
export const DAY_MS = 24 * HOUR_MS;

export const ERROR_RATE_ALERT_THRESHOLD = 0.05;
export const ERROR_RATE_SPIKE_FACTOR = 2;
export const SPARKLINE_POINTS = 9;

export type AnalyticsRangeKey = "24h" | "7d" | "30d";

export const ANALYTICS_RANGE_KEYS: AnalyticsRangeKey[] = ["24h", "7d", "30d"];

export const RANGE_CONFIG: Record<AnalyticsRangeKey, { spanMs: number; intervalSeconds: number }> = {
  "24h": { spanMs: DAY_MS, intervalSeconds: 3_600 },
  "7d": { spanMs: 7 * DAY_MS, intervalSeconds: 6 * 3_600 },
  "30d": { spanMs: 30 * DAY_MS, intervalSeconds: 86_400 },
};

export const DEFAULT_ANALYTICS_RANGE: AnalyticsRangeKey = "24h";
export const ANALYTICS_RANGE_STORAGE_KEY = "dashboard.range";

export function isAnalyticsRangeKey(value: unknown): value is AnalyticsRangeKey {
  return typeof value === "string" && (ANALYTICS_RANGE_KEYS as string[]).includes(value);
}

export function readStoredAnalyticsRange(): AnalyticsRangeKey {
  try {
    const stored = window.localStorage.getItem(ANALYTICS_RANGE_STORAGE_KEY);
    return isAnalyticsRangeKey(stored) ? stored : DEFAULT_ANALYTICS_RANGE;
  } catch {
    return DEFAULT_ANALYTICS_RANGE;
  }
}

export function storeAnalyticsRange(range: AnalyticsRangeKey) {
  try {
    window.localStorage.setItem(ANALYTICS_RANGE_STORAGE_KEY, range);
  } catch {
    // Storage can be blocked; the selection simply won't persist.
  }
}

/** Query window for a range, with `from` snapped to a bucket boundary. */
export function resolveRangeWindow(range: AnalyticsRangeKey, now = new Date()) {
  const { spanMs, intervalSeconds } = RANGE_CONFIG[range];
  const intervalMs = intervalSeconds * 1_000;
  const from = new Date(Math.floor((now.getTime() - spanMs) / intervalMs) * intervalMs);
  return { from: from.toISOString(), to: now.toISOString(), intervalSeconds };
}

type DatedPoint = { point: AnalyticsPoint; date: Date };

export function parseValidDate(value?: string | null) {
  if (!value) return null;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

export function orderPoints(points: AnalyticsPoint[]): DatedPoint[] {
  return points
    .map((point) => ({ point, date: parseValidDate(point.timestamp) }))
    .filter((item): item is DatedPoint => item.date !== null)
    .sort((a, b) => a.date.getTime() - b.date.getTime());
}

const requests = (point: AnalyticsPoint) => Math.max(0, Number(point.request_count ?? 0));
const errors = (point: AnalyticsPoint) => Math.max(0, Math.min(requests(point), Number(point.error_count ?? 0)));

/** Split a series into `count` contiguous chunks and fold each chunk. */
function chunk<T>(items: T[], count: number): T[][] {
  if (items.length === 0) return [];
  const size = Math.min(count, items.length);
  const chunks: T[][] = Array.from({ length: size }, () => []);
  items.forEach((item, index) => chunks[Math.floor((index * size) / items.length)].push(item));
  return chunks;
}

/** Fixed-length hourly buckets ending at the last complete hour; missing hours are empty. */
function hourlyBuckets(points: AnalyticsPoint[], now: Date, hours: number) {
  const lastCompleteStart = Math.floor(now.getTime() / HOUR_MS) * HOUR_MS - HOUR_MS;
  const firstStart = lastCompleteStart - (hours - 1) * HOUR_MS;
  const buckets = Array.from({ length: hours }, () => ({ requests: 0, errors: 0, latencyWeighted: 0, latencyRequests: 0 }));
  for (const { point, date } of orderPoints(points)) {
    const index = Math.floor((date.getTime() - firstStart) / HOUR_MS);
    if (index < 0 || index >= hours) continue;
    const bucket = buckets[index];
    bucket.requests += requests(point);
    bucket.errors += errors(point);
    if (typeof point.latency_p95_ms === "number" && Number.isFinite(point.latency_p95_ms)) {
      const weight = Math.max(1, requests(point));
      bucket.latencyWeighted += point.latency_p95_ms * weight;
      bucket.latencyRequests += weight;
    }
  }
  return buckets;
}

type HourlyBucket = ReturnType<typeof hourlyBuckets>[number];

const ratio = (numerator: number, denominator: number) => (denominator > 0 ? numerator / denominator : 0);
const sumBy = (buckets: HourlyBucket[], key: keyof HourlyBucket) => buckets.reduce((total, bucket) => total + bucket[key], 0);

export type ErrorRateSummary = {
  rate: number;
  previousRate: number | null;
  sparkline: number[];
  spike: { lastRate: number; baseline: number } | null;
};

/**
 * Error rate over the last 24 complete hours versus the 24 before, plus a spike
 * check: the last complete hour must exceed the absolute threshold and be more
 * than ERROR_RATE_SPIKE_FACTOR × the mean hourly rate of the preceding 23 hours.
 */
export function summarizeErrorRate(hourlyPoints: AnalyticsPoint[], now = new Date()): ErrorRateSummary | null {
  const buckets = hourlyBuckets(hourlyPoints, now, 48);
  const current = buckets.slice(24);
  const previous = buckets.slice(0, 24);
  const currentRequests = sumBy(current, "requests");
  if (currentRequests === 0 && sumBy(previous, "requests") === 0) return null;

  const previousRequests = sumBy(previous, "requests");
  const last = current[current.length - 1];
  const lastRate = ratio(last.errors, last.requests);
  const baselineRates = current.slice(0, -1).map((bucket) => ratio(bucket.errors, bucket.requests));
  const baseline = baselineRates.reduce((total, value) => total + value, 0) / baselineRates.length;
  const isSpike = lastRate > ERROR_RATE_ALERT_THRESHOLD && lastRate > ERROR_RATE_SPIKE_FACTOR * baseline;

  return {
    rate: ratio(sumBy(current, "errors"), currentRequests),
    previousRate: previousRequests > 0 ? ratio(sumBy(previous, "errors"), previousRequests) : null,
    sparkline: chunk(current, SPARKLINE_POINTS).map((group) => ratio(sumBy(group, "errors"), sumBy(group, "requests"))),
    spike: isSpike ? { lastRate, baseline } : null,
  };
}

export type LatencySummary = { valueMs: number; previousMs: number | null; sparkline: number[] };

/**
 * Request-weighted mean of hourly P95s. Only an approximation of a true 24h P95,
 * which the backend would have to compute; renders only when the field exists.
 */
export function summarizeLatencyP95(hourlyPoints: AnalyticsPoint[], now = new Date()): LatencySummary | null {
  const buckets = hourlyBuckets(hourlyPoints, now, 48);
  const weighted = (group: HourlyBucket[]) => {
    const weight = sumBy(group, "latencyRequests");
    return weight > 0 ? sumBy(group, "latencyWeighted") / weight : null;
  };
  const current = buckets.slice(24);
  const valueMs = weighted(current);
  if (valueMs === null) return null;
  return {
    valueMs,
    previousMs: weighted(buckets.slice(0, 24)),
    sparkline: chunk(current, SPARKLINE_POINTS)
      .map(weighted)
      .filter((value): value is number => value !== null),
  };
}

export type DailyCostSummary = { today: number; average7d: number; sparkline: number[] };

/** Today's cost (UTC day) against the mean of the previous seven full days. */
export function summarizeDailyCost(dailyPoints: AnalyticsPoint[], now = new Date()): DailyCostSummary | null {
  const todayStart = Math.floor(now.getTime() / DAY_MS) * DAY_MS;
  const days = Array.from({ length: SPARKLINE_POINTS }, () => 0);
  let seen = false;
  for (const { point, date } of orderPoints(dailyPoints)) {
    const index = SPARKLINE_POINTS - 1 - Math.floor((todayStart - Math.floor(date.getTime() / DAY_MS) * DAY_MS) / DAY_MS);
    if (index < 0 || index >= SPARKLINE_POINTS) continue;
    days[index] += Math.max(0, Number(point.cost_usd ?? 0));
    seen = true;
  }
  if (!seen) return null;
  const previousWeek = days.slice(-8, -1);
  return {
    today: days[SPARKLINE_POINTS - 1],
    average7d: previousWeek.reduce((total, value) => total + value, 0) / previousWeek.length,
    sparkline: days,
  };
}
