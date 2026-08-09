import { z } from "zod";

import { APIClientError, requestParsedJSON } from "@/lib/http/api-client";

import { type GuardrailPayload, tenantGuardrailSchema } from "./types";

const listResponseSchema = z.object({
  data: z.array(tenantGuardrailSchema),
  total: z.number(),
  limit: z.number(),
  offset: z.number(),
});
const successResponseSchema = z.union([
  z.object({ success: z.literal(true) }),
  z.object({}).strict().transform(() => ({ success: true as const })),
]);

export { APIClientError as GuardrailAPIError };

export function listGuardrails(query = "") {
  return requestParsedJSON(listResponseSchema, `/api/guardrails?query=${encodeURIComponent(query)}`);
}

export function createGuardrail(body: GuardrailPayload) {
  return requestParsedJSON(tenantGuardrailSchema, "/api/guardrails", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateGuardrail(slug: string, body: GuardrailPayload) {
  return requestParsedJSON(tenantGuardrailSchema, `/api/guardrails/${encodeURIComponent(slug)}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteGuardrail(slug: string) {
  return requestParsedJSON(successResponseSchema, `/api/guardrails/${encodeURIComponent(slug)}`, {
    method: "DELETE",
  });
}
