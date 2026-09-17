import { requestParsedJSON } from "@/lib/http/api-client";
import { z } from "zod";
import { effectiveEntitlementsSchema, tenantSubscriptionSchema } from "./contracts";

const analyticsPointSchema = z.object({
  timestamp: z.string().optional(),
  group: z.string().optional(),
  request_count: z.number().optional(),
  error_count: z.number().optional(),
  cost_usd: z.number().optional(),
  // Not yet emitted by the backend; optional so older responses keep parsing.
  latency_p95_ms: z.number().optional(),
  cache_hit_count: z.number().optional(),
});
const analyticsResponseSchema = z.object({ data: z.array(analyticsPointSchema).optional() });

export type AnalyticsPoint = z.infer<typeof analyticsPointSchema>;

export function getDashboardSubscription() {
  return requestParsedJSON(tenantSubscriptionSchema, "/api/subscription");
}

export function getDashboardEntitlements() {
  return requestParsedJSON(effectiveEntitlementsSchema, "/api/entitlements");
}

export async function getDashboardAnalytics(query: URLSearchParams, signal?: AbortSignal) {
  const response = await requestParsedJSON(
    analyticsResponseSchema,
    `/api/logs/analytics?${query.toString()}`,
    { signal },
  );
  return response.data ?? [];
}

export function getDashboardTimeseries(
  window: { from: string; to: string; intervalSeconds: number },
  signal?: AbortSignal,
) {
  return getDashboardAnalytics(
    new URLSearchParams({
      from: window.from,
      to: window.to,
      type: "timeseries",
      group_by: "",
      interval_seconds: String(window.intervalSeconds),
    }),
    signal,
  );
}
