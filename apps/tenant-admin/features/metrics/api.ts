import { z } from "zod";

import { requestParsedJSON } from "@/lib/http/api-client";

const metricsRowSchema = z.record(z.string(), z.unknown());
const metricsQueryResponseSchema = z.object({ data: z.array(metricsRowSchema).optional() });
const metricsFacetsSchema = z.object({
  models: z.unknown().optional(),
  providers: z.unknown().optional(),
});

export function queryMetrics(body: Record<string, unknown>) {
  return requestParsedJSON(metricsQueryResponseSchema, "/api/metrics/query", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function getMetricsFacets() {
  return requestParsedJSON(metricsFacetsSchema, "/api/metrics/facets");
}
