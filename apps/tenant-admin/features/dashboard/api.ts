import { requestParsedJSON } from "@/lib/http/api-client";
import { z } from "zod";
import { effectiveEntitlementsSchema, tenantSubscriptionSchema } from "./contracts";

const analyticsPointSchema = z.object({
  timestamp: z.string().optional(),
  group: z.string().optional(),
  request_count: z.number().optional(),
  error_count: z.number().optional(),
  cost_usd: z.number().optional(),
});
const analyticsResponseSchema = z.object({ data: z.array(analyticsPointSchema).optional() });

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
