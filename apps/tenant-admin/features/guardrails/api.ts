import { z } from "zod";

import { APIClientError, requestParsedJSON } from "@/lib/http/api-client";

import { type GuardrailPayload, semanticPolicyTestResultSchema, tenantGuardrailSchema } from "./types";

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

export function getGuardrail(slug: string) {
  return requestParsedJSON(tenantGuardrailSchema, `/api/guardrails/${encodeURIComponent(slug)}`);
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

/** Runs one sample through an unsaved semantic policy. Each call is a paid judge request and is rate limited. */
export function testSemanticGuardrail(body: { config: Record<string, unknown>; text: string }, signal?: AbortSignal) {
  return requestParsedJSON(semanticPolicyTestResultSchema, "/api/guardrails/semantic/test", {
    method: "POST",
    body: JSON.stringify(body),
    signal,
  });
}
