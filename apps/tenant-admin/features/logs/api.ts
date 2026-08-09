import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const bucketSchema = z.object({ value: z.string(), count: z.number() });
const statusBucketSchema = z.object({ status_code: z.number(), count: z.number() });
const rangeSchema = z.object({ min: z.number().optional(), max: z.number().optional() });

const requestLogSchema = z.object({
  log_id: z.string(),
  request_id: z.string(),
  trace_id: z.string().optional(),
  provider: z.string().optional(),
  provider_model: z.string().optional(),
  model: z.string().optional(),
  virtual_model: z.string().optional(),
  route: z.string().optional(),
  request_type: z.string().optional(),
  status: z.string().optional(),
  status_code: z.number(),
  error_type: z.string().optional(),
  started_at: z.string(),
  duration_ms: z.number(),
  input_tokens: z.number().optional(),
  output_tokens: z.number().optional(),
  total_tokens: z.number().optional(),
  cost_usd: z.number().optional(),
  cost_irr: z.number().optional(),
  cache_status: z.string().optional(),
  retry_count: z.number().optional(),
  fallback_count: z.number().optional(),
  fallback_status: z.string().optional(),
  load_balance_status: z.string().optional(),
  payload_mode: z.string().optional(),
});

const logsSchema = z.object({
  data: z.array(requestLogSchema).optional(),
  total: z.number().optional(),
  limit: z.number().optional(),
  offset: z.number().optional(),
  next_offset: z.number().optional(),
});

const facetsSchema = z.object({
  providers: z.array(bucketSchema).optional(),
  models: z.array(bucketSchema).optional(),
  virtual_models: z.array(bucketSchema).optional(),
  routes: z.array(bucketSchema).optional(),
  request_types: z.array(bucketSchema).optional(),
  error_types: z.array(bucketSchema).optional(),
  cache_statuses: z.array(bucketSchema).optional(),
  fallback_statuses: z.array(bucketSchema).optional(),
  load_balance_statuses: z.array(bucketSchema).optional(),
  payload_modes: z.array(bucketSchema).optional(),
  status_codes: z.array(statusBucketSchema).optional(),
});

const rangeStatsSchema = z.object({
  duration_ms: rangeSchema.optional(),
  total_tokens: rangeSchema.optional(),
  cost_usd: rangeSchema.optional(),
  cost_irr: rangeSchema.optional(),
});

const summarySchema = z.object({
  request_count: z.number().optional(),
  error_count: z.number().optional(),
  error_rate: z.number().optional(),
  latency_avg_ms: z.number().optional(),
  latency_p50_ms: z.number().optional(),
  latency_p95_ms: z.number().optional(),
  latency_p99_ms: z.number().optional(),
  total_tokens: z.number().optional(),
  cost_usd: z.number().optional(),
  cost_irr: z.number().optional(),
  cache_hit_count: z.number().optional(),
  cache_hit_rate: z.number().optional(),
  retry_count: z.number().optional(),
  fallback_count: z.number().optional(),
  load_balance_count: z.number().optional(),
  payload_modes: z.array(bucketSchema).optional(),
  status_codes: z.array(statusBucketSchema).optional(),
  providers: z.array(bucketSchema).optional(),
  models: z.array(bucketSchema).optional(),
  routes: z.array(bucketSchema).optional(),
});

const analyticsPointSchema = z.object({
  timestamp: z.string().optional(),
  group: z.string().optional(),
  request_count: z.number(),
  error_count: z.number(),
  error_rate: z.number(),
  latency_p95_ms: z.number(),
  total_tokens: z.number(),
  cost_usd: z.number(),
  cache_hit_count: z.number(),
  retry_count: z.number(),
  fallback_count: z.number(),
  load_balance_count: z.number(),
});

const savedViewSchema = z.object({
  saved_view_id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  filters: z.record(z.string(), z.unknown()),
  visible_columns: z.array(z.string()),
  shared_scope: z.string(),
});

const stringRecordSchema = z.record(z.string(), z.string());
const unknownRecordSchema = z.record(z.string(), z.unknown());
const logDetailSchema = z.object({
  log: requestLogSchema,
  content_visibility: z.object({
    can_view_content: z.boolean(),
    request_reason: z.string().optional(),
    response_reason: z.string().optional(),
    request_available: z.boolean(),
    response_available: z.boolean(),
  }),
  request_payload: unknownRecordSchema.optional(),
  response_payload: unknownRecordSchema.optional(),
  request_payload_meta: unknownRecordSchema.optional(),
  response_payload_meta: unknownRecordSchema.optional(),
  metadata: unknownRecordSchema.optional(),
  actor: stringRecordSchema.optional(),
  scope: stringRecordSchema.optional(),
});

const principalSchema = z.object({
  principal_id: z.string(),
  display_name: z.string(),
  secondary: z.string().optional(),
  type: z.string().optional(),
});
const lookupEntitySchema = z.object({
  id: z.string().optional(),
  org_unit_id: z.string().optional(),
  virtual_user_id: z.string().optional(),
  alias: z.string().optional(),
  display_name: z.string().optional(),
  name: z.string().optional(),
});

const requestLoggingSettingsSchema = z.object({
  tenant_id: z.string().optional(),
  logging_mode: z.enum(["none", "metadata-only", "redacted-content", "full-content"]),
  updated_at: z.string().optional(),
  updated_by_user_id: z.string().optional(),
});

const summaryEnvelopeSchema = z.object({ data: summarySchema });
const facetsEnvelopeSchema = z.object({ data: facetsSchema });
const rangeStatsEnvelopeSchema = z.object({ data: rangeStatsSchema.optional() });
const analyticsEnvelopeSchema = z.object({ data: z.array(analyticsPointSchema) });
const savedViewsEnvelopeSchema = z.object({ data: z.array(savedViewSchema) });
const principalsEnvelopeSchema = z.object({ data: z.array(principalSchema).optional() });
const lookupEntitiesEnvelopeSchema = z.object({ data: z.array(lookupEntitySchema).optional() });
const exportJobSchema = z.object({ export_id: z.string() });

export type RequestLoggingSettings = z.infer<typeof requestLoggingSettingsSchema>;

export function getLogs(query: string, signal?: AbortSignal) {
  return requestParsedJSON(logsSchema, `/api/logs?${query}`, { signal });
}

export function getLogSummary(query: string, signal?: AbortSignal) {
  return requestParsedJSON(summaryEnvelopeSchema, `/api/logs/summary?${query}`, { signal });
}

export function getLogFacets(query: string, signal?: AbortSignal) {
  return requestParsedJSON(facetsEnvelopeSchema, `/api/logs/facets?${query}`, { signal });
}

export function getLogAnalytics(query: string, signal?: AbortSignal) {
  return requestParsedJSON(analyticsEnvelopeSchema, `/api/logs/analytics?${query}`, { signal });
}

export function getLogRangeStats(query: string, signal?: AbortSignal) {
  const suffix = query ? `?${query}` : "";
  return requestParsedJSON(rangeStatsEnvelopeSchema, `/api/logs/range-stats${suffix}`, { signal });
}

export function getLogDetail(logID: string, signal?: AbortSignal) {
  return requestParsedJSON(logDetailSchema, `/api/logs/${encodeURIComponent(logID)}`, { signal });
}

export async function traceExists(traceID: string, signal?: AbortSignal) {
  await requestJSON(`/api/traces/${encodeURIComponent(traceID)}`, { signal });
}

export function searchAccessPrincipals(kind: "user" | "team", query: string) {
  return requestParsedJSON(
    principalsEnvelopeSchema,
    `/api/access-principals?kind=${kind}&q=${encodeURIComponent(query)}&limit=20`,
  );
}

export function getOrganizationUnitLookups() {
  return requestParsedJSON(lookupEntitiesEnvelopeSchema, "/api/gateway/lookups/organization-units");
}

export function getVirtualAccountLookups() {
  return requestParsedJSON(lookupEntitiesEnvelopeSchema, "/api/gateway/lookups/virtual-accounts");
}

export function createSavedView(input: {
  name: string;
  filters: Record<string, string>;
  visible_columns: string[];
  shared_scope: string;
}) {
  return requestParsedJSON(savedViewSchema, "/api/logs/saved-views", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function deleteSavedView(viewID: string) {
  await requestJSON(`/api/logs/saved-views/${encodeURIComponent(viewID)}`, { method: "DELETE" });
}

export function createLogExport(input: {
  filters: Record<string, string>;
  format: "csv" | "jsonl";
  include_content: boolean;
}) {
  return requestParsedJSON(exportJobSchema, "/api/logs/exports", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getRequestLoggingSettings() {
  return requestParsedJSON(requestLoggingSettingsSchema, "/api/settings/request-logging");
}

export function updateRequestLoggingSettings(loggingMode: RequestLoggingSettings["logging_mode"]) {
  return requestParsedJSON(requestLoggingSettingsSchema, "/api/settings/request-logging", {
    method: "PUT",
    body: JSON.stringify({ logging_mode: loggingMode }),
  });
}

export function parseInitialLogs(value: unknown) {
  return parseOr(logsSchema, value, { data: [], total: 0, limit: 25, offset: 0 });
}

export function parseInitialSummary(value: unknown) {
  return parseEnvelopeOr(summaryEnvelopeSchema, value, {});
}

export function parseInitialFacets(value: unknown) {
  return parseEnvelopeOr(facetsEnvelopeSchema, value, {});
}

export function parseInitialRangeStats(value: unknown) {
  return parseEnvelopeOr(rangeStatsEnvelopeSchema, value, {});
}

export function parseInitialSavedViews(value: unknown) {
  return parseEnvelopeOr(savedViewsEnvelopeSchema, value, []);
}

export function parseInitialAnalytics(value: unknown) {
  return parseEnvelopeOr(analyticsEnvelopeSchema, value, []);
}

export function parseRequestLoggingSettings(payload: unknown): RequestLoggingSettings | null {
  const parsed = requestLoggingSettingsSchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

function parseOr<T>(schema: z.ZodType<T>, value: unknown, fallback: T): T {
  const parsed = schema.safeParse(value);
  return parsed.success ? parsed.data : fallback;
}

function parseEnvelopeOr<T>(schema: z.ZodType<{ data?: T }>, value: unknown, fallback: T): T {
  const parsed = schema.safeParse(value);
  return parsed.success ? parsed.data.data ?? fallback : fallback;
}
