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

/** Filters only the traces view has; the rest come from the shared observe filters. */
export type TraceLocalFilters = Pick<TraceFilters, "traceID" | "route" | "sortBy" | "sortOrder">;

export const PAGE_SIZE = 25;
export const MAX_BREAKDOWN_ITEMS = 5;

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
