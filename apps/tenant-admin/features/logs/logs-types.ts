import type { CompactDisplayValue } from "@/lib/compact-number";
import type { ObserveRangePreset } from "@/lib/observe-filters";

export type RequestLog = {
  log_id: string;
  request_id: string;
  trace_id?: string;
  provider?: string;
  provider_model?: string;
  model?: string;
  virtual_model?: string;
  route?: string;
  request_type?: string;
  status?: string;
  status_code: number;
  error_type?: string;
  started_at: string;
  duration_ms: number;
  input_tokens?: number;
  output_tokens?: number;
  total_tokens?: number;
  cost_usd?: number;
  cost_irr?: number;
  cache_status?: string;
  retry_count?: number;
  fallback_count?: number;
  fallback_status?: string;
  load_balance_status?: string;
  payload_mode?: string;
};

export type LogsResponse = {
  data?: RequestLog[];
  total?: number;
  limit?: number;
  offset?: number;
  next_offset?: number;
};

export type Bucket = { value: string; count: number };
export type StatusBucket = { status_code: number; count: number };

export type LogFacets = {
  providers?: Bucket[];
  models?: Bucket[];
  virtual_models?: Bucket[];
  routes?: Bucket[];
  request_types?: Bucket[];
  error_types?: Bucket[];
  cache_statuses?: Bucket[];
  fallback_statuses?: Bucket[];
  load_balance_statuses?: Bucket[];
  payload_modes?: Bucket[];
  status_codes?: StatusBucket[];
};

export type LogRange = { min?: number; max?: number };
export type LogRangeStats = {
  duration_ms?: LogRange;
  total_tokens?: LogRange;
  cost_usd?: LogRange;
  cost_irr?: LogRange;
};

export type Summary = {
  request_count?: number;
  error_count?: number;
  error_rate?: number;
  latency_avg_ms?: number;
  latency_p50_ms?: number;
  latency_p95_ms?: number;
  latency_p99_ms?: number;
  total_tokens?: number;
  cost_usd?: number;
  cost_irr?: number;
  cache_hit_count?: number;
  cache_hit_rate?: number;
  retry_count?: number;
  fallback_count?: number;
  load_balance_count?: number;
  payload_modes?: Bucket[];
  status_codes?: StatusBucket[];
  providers?: Bucket[];
  models?: Bucket[];
  routes?: Bucket[];
};

export type AnalyticsPoint = {
  timestamp?: string;
  group?: string;
  request_count: number;
  error_count: number;
  error_rate: number;
  latency_p95_ms: number;
  total_tokens: number;
  cost_usd: number;
  cache_hit_count: number;
  retry_count: number;
  fallback_count: number;
  load_balance_count: number;
};

export type ChartDataPoint = {
  label?: string;
  requests?: number;
  errors?: number;
  latency?: number;
  cost?: number;
  cacheHit?: number;
  cacheMiss?: number;
};

export type ExpandedChartKey = "traffic" | "latencyCost" | "providerBreakdown" | "cacheHitMiss";

export type SavedView = {
  saved_view_id: string;
  name: string;
  description?: string;
  filters: Record<string, unknown>;
  visible_columns: string[];
  shared_scope: string;
};

export type LogDetail = {
  log: RequestLog;
  content_visibility: {
    can_view_content: boolean;
    request_reason?: string;
    response_reason?: string;
    request_available: boolean;
    response_available: boolean;
  };
  request_payload?: Record<string, unknown>;
  response_payload?: Record<string, unknown>;
  request_payload_meta?: Record<string, unknown>;
  response_payload_meta?: Record<string, unknown>;
  metadata?: Record<string, unknown>;
  actor?: Record<string, string>;
  scope?: Record<string, string>;
};

export type Filters = {
  /** Shared with traces and metrics; see lib/observe-filters. */
  range: ObserveRangePreset;
  from: string;
  to: string;
  log_id: string;
  request_id: string;
  trace_id: string;
  provider: string;
  model: string;
  virtual_model: string;
  route: string;
  request_type: string;
  status_code: string;
  error_type: string;
  cache_status: string;
  fallback_status: string;
  load_balance_status: string;
  payload_mode: string;
  actor_type: string;
  actor_id: string;
  team_id: string;
  duration_min: string;
  duration_max: string;
  tokens_min: string;
  tokens_max: string;
  cost_usd_min: string;
  cost_usd_max: string;
  cost_irr_min: string;
  cost_irr_max: string;
  sort_by: string;
  sort_order: "asc" | "desc";
};

export type LogsCommandCenterProps = {
  initialLogs?: unknown;
  initialSummary?: unknown;
  initialFacets?: unknown;
  initialSavedViews?: unknown;
  initialTimeseries?: unknown;
  initialBreakdown?: unknown;
  initialRangeStats?: unknown;
  initialError?: string;
};

export type LogsFormatter = {
  number: (value: number) => string;
  numberDisplay: (value: number) => CompactDisplayValue;
  compactNumber: (value: number) => string;
  percent: (value: number) => string;
  usd: (value: number) => string;
  usdDisplay: (value: number) => CompactDisplayValue;
  irr: (value: number) => string;
  irrDisplay: (value: number) => CompactDisplayValue;
  ms: (value: number) => string;
  msValue: (value: number) => string;
  timeLabel: (value?: string) => string | undefined;
  dateTime: (value: string) => string;
};

export const DEFAULT_FILTERS: Filters = {
  range: "last24h",
  from: "",
  to: "",
  log_id: "",
  request_id: "",
  trace_id: "",
  provider: "",
  model: "",
  virtual_model: "",
  route: "",
  request_type: "",
  status_code: "",
  error_type: "",
  cache_status: "",
  fallback_status: "",
  load_balance_status: "",
  payload_mode: "",
  actor_type: "",
  actor_id: "",
  team_id: "",
  duration_min: "",
  duration_max: "",
  tokens_min: "",
  tokens_max: "",
  cost_usd_min: "",
  cost_usd_max: "",
  cost_irr_min: "",
  cost_irr_max: "",
  sort_by: "started_at",
  sort_order: "desc",
};

export const DEFAULT_COLUMNS = [
  "status", "started_at", "provider", "model", "route", "latency", "tokens", "cost", "signals", "payload", "actions",
];

export const OPTIONAL_COLUMNS = ["cache_status", "retry_count", "fallback_status", "load_balance_status"];

export const COLUMN_CHOICES = [
  "status", "started_at", "provider", "model", "route", "latency", "tokens", "cost", "signals", "payload",
  "cache_status", "retry_count", "fallback_status", "load_balance_status", "actions",
];

export const MAX_CHART_POINTS = 360;
