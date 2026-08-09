import type { TracePrivacySettings } from "@/features/traces/api";

export type RequestTrace = {
  trace_id: string;
  tenant_id: string;
  request_id?: string;
  model?: string;
  provider?: string;
  route: string;
  status_code: number;
  started_at: string;
  duration_ms: number;
  error_code?: string;
  sampled: boolean;
  attempts?: number;
};

export type FacetBucket = { value: string; count: number };
export type StatusFacetBucket = { status_code: number; count: number };

export type TraceFacets = {
  providers?: FacetBucket[];
  models?: FacetBucket[];
  routes?: FacetBucket[];
  error_codes?: FacetBucket[];
  status_codes?: StatusFacetBucket[];
};

export type TraceFilters = {
  traceID: string;
  requestID: string;
  model: string;
  provider: string;
  route: string;
  statusCode: string;
  errorCode: string;
  from: string;
  to: string;
  sortBy: string;
  sortOrder: "asc" | "desc";
};

export const PAGE_SIZE = 25;
export const MAX_BREAKDOWN_ITEMS = 5;
export const RANGE_PRESETS = ["last15m", "last1h", "last24h", "yesterday", "last3d", "last7d", "last14d", "last30d", "custom"] as const;
export const DEFAULT_RANGE_PRESET = "last24h";
export const DEFAULT_RANGE_MS = 24 * 60 * 60 * 1000;
export const ALL_MODELS_VALUE = "__all_models__";
export const ALL_PROVIDERS_VALUE = "__all_providers__";
export const TRACE_COLUMN_KEYS = ["startedAt", "status", "duration", "trace", "context", "endpoint", "error"] as const;

export type RangePreset = (typeof RANGE_PRESETS)[number];
export type TraceColumnKey = (typeof TRACE_COLUMN_KEYS)[number];

export type TraceManagementProps = {
  initialData: RequestTrace[];
  initialTotal: number;
  initialError: string | null;
  initialTracePrivacySettings: TracePrivacySettings | null;
  initialTracePrivacyError: string | null;
  initialTracePrivacyOpen?: boolean;
  currentUserID: string;
  currentUserLabel: string;
};

export type BreakdownBucket = { value: string; label: string; count: number };

export const BASE_FILTERS: Omit<TraceFilters, "from" | "to"> = {
  traceID: "",
  requestID: "",
  model: "",
  provider: "",
  route: "",
  statusCode: "",
  errorCode: "",
  sortBy: "started_at",
  sortOrder: "desc",
};
