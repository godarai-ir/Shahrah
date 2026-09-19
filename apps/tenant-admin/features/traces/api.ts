import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

import { parseTraceAttempts, parseTraceDetail, parseTraceSpans } from "./detail-contracts";

const requestTraceSchema = z.object({
  trace_id: z.string(), tenant_id: z.string(), request_id: z.string().optional(),
  model: z.string().optional(), provider: z.string().optional(), route: z.string(),
  status_code: z.number(), started_at: z.string(), duration_ms: z.number(),
  error_code: z.string().optional(), sampled: z.boolean(), attempts: z.number().optional(),
});
const facetBucketSchema = z.object({ value: z.string(), count: z.number() });
const statusFacetBucketSchema = z.object({ status_code: z.number(), count: z.number() });
const traceFacetsSchema = z.object({
  providers: z.array(facetBucketSchema).optional(), models: z.array(facetBucketSchema).optional(),
  routes: z.array(facetBucketSchema).optional(), error_codes: z.array(facetBucketSchema).optional(),
  status_codes: z.array(statusFacetBucketSchema).optional(),
});
const traceListSchema = z.object({
  data: z.array(requestTraceSchema).optional(), total: z.number().optional(),
  limit: z.number().optional(), offset: z.number().optional(), next_offset: z.number().optional(),
});
const facetsResponseSchema = z.object({ data: traceFacetsSchema.optional() });
const tracePrivacySettingsSchema = z.object({
  tenant_id: z.string().optional(),
  capture_policy: z.enum(["none", "metadata-only", "sampled-redacted", "full-secure"]),
  updated_at: z.string().optional(), updated_by_user_id: z.string().optional(),
});

export type TracePrivacySettings = z.infer<typeof tracePrivacySettingsSchema>;

export async function loadTraces(listQuery: string, filtersQuery: string) {
  const [list, facets] = await Promise.all([
    requestParsedJSON(traceListSchema, `/api/traces?${listQuery}`),
    requestParsedJSON(facetsResponseSchema, `/api/traces/facets${filtersQuery ? `?${filtersQuery}` : ""}`),
  ]);
  return { list, facets };
}

export function updateTracePrivacySettings(capturePolicy: TracePrivacySettings["capture_policy"]) {
  return requestParsedJSON(tracePrivacySettingsSchema, "/api/settings/trace-privacy", {
    method: "PUT",
    body: JSON.stringify({ capture_policy: capturePolicy }),
  });
}

export function parseTraceListPayload(payload: unknown) {
  const parsed = traceListSchema.safeParse(payload);
  return parsed.success ? parsed.data : { data: [], total: 0 };
}

export function parseTracePrivacySettings(payload: unknown): TracePrivacySettings | null {
  const parsed = tracePrivacySettingsSchema.safeParse(payload);
  return parsed.success ? parsed.data : null;
}

/** One trace's summary. Null when the payload does not match the contract. */
export async function getTraceDetail(traceID: string, signal?: AbortSignal) {
  return parseTraceDetail(await requestJSON(`/api/traces/${encodeURIComponent(traceID)}`, { signal }));
}

/** Span timeline of a sampled trace. */
export async function getTraceTimeline(traceID: string, signal?: AbortSignal) {
  return parseTraceSpans(await requestJSON(`/api/traces/${encodeURIComponent(traceID)}/spans`, { signal }));
}

/** Provider attempts, including retry and fallback reasons. */
export async function getTraceAttempts(traceID: string, signal?: AbortSignal) {
  return parseTraceAttempts(await requestJSON(`/api/traces/${encodeURIComponent(traceID)}/attempts`, { signal }));
}
