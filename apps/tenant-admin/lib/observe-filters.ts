/**
 * The filter state shared by the three observe views (/logs, /traces, /metrics).
 *
 * Only the intersection lives here: each view keeps its own specific filters and reads or writes
 * these keys through the URL, so moving between tabs keeps the search, range and filters.
 */

export const OBSERVE_TABS = ["logs", "traces", "metrics"] as const;
export type ObserveTab = (typeof OBSERVE_TABS)[number];

export const OBSERVE_RANGE_PRESETS = [
  "last15m", "last1h", "last6h", "last24h", "yesterday", "last3d", "last7d", "last14d", "last30d", "custom",
] as const;
export type ObserveRangePreset = (typeof OBSERVE_RANGE_PRESETS)[number];
export const DEFAULT_OBSERVE_RANGE: ObserveRangePreset = "last24h";

export type ObserveFilters = {
  /** Request ID lookup. Every observe backend filters on the request ID; none offers free-text search. */
  q: string;
  from: string;
  to: string;
  range: ObserveRangePreset;
  model: string;
  provider: string;
  /** HTTP status code, three digits. */
  status: string;
  errorCode: string;
};

export type ObserveFilterKey = Exclude<keyof ObserveFilters, "from" | "to" | "range">;

export const EMPTY_OBSERVE_FILTERS: ObserveFilters = {
  q: "",
  from: "",
  to: "",
  range: DEFAULT_OBSERVE_RANGE,
  model: "",
  provider: "",
  status: "",
  errorCode: "",
};

/** URL parameter for each filter key. */
const PARAM_BY_KEY: Record<ObserveFilterKey, string> = {
  q: "q",
  model: "model",
  provider: "provider",
  status: "status",
  errorCode: "error_code",
};

export const OBSERVE_FILTER_KEYS: ObserveFilterKey[] = ["q", "model", "provider", "status", "errorCode"];

/** URL parameters owned by the shared filter; a view's own parameters are everything else. */
export const OBSERVE_PARAM_NAMES = ["q", "range", "from", "to", ...OBSERVE_FILTER_KEYS.map((key) => PARAM_BY_KEY[key])];

/**
 * Which shared filters each view can apply. Logs and metrics only know an error *type*, not the
 * trace error code, so `errorCode` stays visible there but is marked as not applied.
 */
const SUPPORT: Record<ObserveTab, Record<ObserveFilterKey, boolean>> = {
  logs: { q: true, model: true, provider: true, status: true, errorCode: false },
  traces: { q: true, model: true, provider: true, status: true, errorCode: true },
  metrics: { q: true, model: true, provider: true, status: true, errorCode: false },
};

/** Range values the logs view stored before the ranges were unified (saved views, old links). */
const LEGACY_RANGE_ALIASES: Record<string, ObserveRangePreset> = {
  "1h": "last1h",
  "6h": "last6h",
  "24h": "last24h",
  "7d": "last7d",
};

const DAY_MS = 24 * 60 * 60 * 1000;
const RANGE_DURATION_MS: Record<Exclude<ObserveRangePreset, "custom" | "yesterday">, number> = {
  last15m: 15 * 60 * 1000,
  last1h: 60 * 60 * 1000,
  last6h: 6 * 60 * 60 * 1000,
  last24h: DAY_MS,
  last3d: 3 * DAY_MS,
  last7d: 7 * DAY_MS,
  last14d: 14 * DAY_MS,
  last30d: 30 * DAY_MS,
};

export function parseObserveRangePreset(value: string | null | undefined): ObserveRangePreset | null {
  const trimmed = value?.trim() ?? "";
  if ((OBSERVE_RANGE_PRESETS as readonly string[]).includes(trimmed)) return trimmed as ObserveRangePreset;
  return LEGACY_RANGE_ALIASES[trimmed] ?? null;
}

export function observeFiltersFromSearchParams(searchParams: Pick<URLSearchParams, "get">): ObserveFilters {
  const text = (name: string) => searchParams.get(name)?.trim() ?? "";
  const from = normalizeIso(text("from"));
  const to = normalizeIso(text("to"));
  const explicitRange = parseObserveRangePreset(searchParams.get("range"));
  // A bare from/to pair (older metrics links) means a custom range.
  let range = explicitRange ?? (from && to ? "custom" : DEFAULT_OBSERVE_RANGE);
  if (range === "custom" && (!from || !to)) range = DEFAULT_OBSERVE_RANGE;
  const status = text("status");

  return {
    q: text("q"),
    from: range === "custom" ? from : "",
    to: range === "custom" ? to : "",
    range,
    model: text("model"),
    provider: text("provider"),
    status: /^[1-5]\d\d$/u.test(status) ? status : "",
    errorCode: text("error_code"),
  };
}

/** Serializes the shared filters in a stable order, leaving out blanks and the default range. */
export function observeQueryString(filters: ObserveFilters): string {
  const params = new URLSearchParams();
  writeObserveParams(params, filters);
  return params.toString();
}

/**
 * Replaces the shared parameters on `current` and keeps the view's own ones, so a view can sync
 * its filters into the URL without dropping state such as the selected trace.
 */
export function mergeObserveSearchParams(current: Pick<URLSearchParams, "toString">, filters: ObserveFilters): URLSearchParams {
  const params = new URLSearchParams(current.toString());
  for (const name of OBSERVE_PARAM_NAMES) params.delete(name);
  writeObserveParams(params, filters);
  return params;
}

export function observeHref(tab: ObserveTab, filters: ObserveFilters, extra?: Record<string, string>): string {
  const params = new URLSearchParams(observeQueryString(filters));
  for (const [key, value] of Object.entries(extra ?? {})) {
    if (value) params.set(key, value);
  }
  const query = params.toString();
  return query ? `/${tab}?${query}` : `/${tab}`;
}

export function isObserveFilterSupported(tab: ObserveTab, key: ObserveFilterKey): boolean {
  return SUPPORT[tab][key];
}

/** Active shared filters the given view cannot apply. They are shown, not dropped. */
export function unsupportedObserveFilters(tab: ObserveTab, filters: ObserveFilters): ObserveFilterKey[] {
  return OBSERVE_FILTER_KEYS.filter((key) => filters[key] !== "" && !SUPPORT[tab][key]);
}

export function hasActiveObserveFilters(filters: ObserveFilters): boolean {
  return filters.range !== DEFAULT_OBSERVE_RANGE || OBSERVE_FILTER_KEYS.some((key) => filters[key] !== "");
}

/** Resolves a range to absolute timestamps. Relative presets are anchored at `now`. */
export function resolveObserveRange(filters: Pick<ObserveFilters, "range" | "from" | "to">, now: Date = new Date()): { from: string; to: string } {
  if (filters.range === "custom") return { from: filters.from, to: filters.to };
  if (filters.range === "yesterday") {
    const start = new Date(now);
    start.setHours(0, 0, 0, 0);
    start.setDate(start.getDate() - 1);
    const end = new Date(start);
    end.setDate(end.getDate() + 1);
    return { from: start.toISOString(), to: end.toISOString() };
  }
  return { from: new Date(now.getTime() - RANGE_DURATION_MS[filters.range]).toISOString(), to: now.toISOString() };
}

export function isSameObserveFilters(a: ObserveFilters, b: ObserveFilters): boolean {
  return observeQueryString(a) === observeQueryString(b);
}

function writeObserveParams(params: URLSearchParams, filters: ObserveFilters) {
  if (filters.q.trim()) params.set("q", filters.q.trim());
  if (filters.range === "custom" && filters.from && filters.to) {
    params.set("range", "custom");
    params.set("from", filters.from);
    params.set("to", filters.to);
  } else if (filters.range !== "custom" && filters.range !== DEFAULT_OBSERVE_RANGE) {
    params.set("range", filters.range);
  }
  for (const key of OBSERVE_FILTER_KEYS) {
    if (key === "q") continue;
    const value = filters[key].trim();
    if (value) params.set(PARAM_BY_KEY[key], value);
  }
}

function normalizeIso(value: string) {
  if (!value) return "";
  const time = Date.parse(value);
  return Number.isNaN(time) ? "" : new Date(time).toISOString();
}
