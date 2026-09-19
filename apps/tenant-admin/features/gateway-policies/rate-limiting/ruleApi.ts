import { z } from "zod";

import { APIClientError, requestParsedJSON } from "@/lib/http/api-client";

import type { RateLimitRuleFormInput } from "./ruleValidation";
import { rateLimitRuleResponseSchema } from "./ruleValidation";

const listResponseSchema = z.object({
  data: z.array(rateLimitRuleResponseSchema),
  total: z.number(),
  limit: z.number(),
  offset: z.number(),
});
const effectiveEntitlementsSchema = z.object({
  features: z.record(z.string(), z.unknown()).optional(),
}).nullable();
const successResponseSchema = z.union([
  z.object({ success: z.literal(true) }),
  z.object({}).strict().transform(() => ({ success: true as const })),
]);

/**
 * Stable identifiers for the failures this feature can explain to the user.
 * The display copy lives in the `gatewayPolicyErrors` message namespace.
 */
export type RuleErrorCode =
  | "saveRejected"
  | "invalidBody"
  | "duplicateRuleId"
  | "internal"
  | "unauthenticated"
  | "forbidden"
  | "notFound"
  | "unknown";

/** Error whose display copy is resolved at the UI boundary. */
export class RuleAPIError extends Error {
  readonly code: RuleErrorCode;

  constructor(code: RuleErrorCode, message: string, options?: ErrorOptions) {
    super(message, options);
    this.name = "RuleAPIError";
    this.code = code;
  }
}

function classifyAPIError(message: string, status?: number): RuleErrorCode | null {
  const normalized = message.toLowerCase();
  if (normalized.includes("no rows in result set")) return "saveRejected";
  if (normalized.includes("invalid request body")) return "invalidBody";
  if (normalized.includes("already exists") || status === 409) return "duplicateRuleId";
  if (normalized.includes("internal server error") || status === 500) return "internal";
  return null;
}

function looksLikeHTMLDocument(value: string): boolean {
  const normalized = value.trim().toLowerCase();
  return normalized.startsWith("<!doctype") || normalized.startsWith("<html") || normalized.startsWith("<body");
}

function statusErrorCode(status?: number): RuleErrorCode {
  if (status === 401) return "unauthenticated";
  if (status === 403) return "forbidden";
  if (status === 404) return "notFound";
  if (status === 500) return "internal";
  return "unknown";
}

/**
 * Upstream messages are shown verbatim, except HTML error pages and empty
 * bodies, which carry no useful copy and are replaced by a coded error.
 */
function toRuleAPIError(message: string | undefined, status: number | undefined, cause: unknown): RuleAPIError {
  const normalized = message?.trim();
  if (!normalized || looksLikeHTMLDocument(normalized)) {
    const code = statusErrorCode(status);
    return new RuleAPIError(code, `Request failed (${status ?? "unknown"})`, { cause });
  }

  const code = classifyAPIError(normalized, status);
  return new RuleAPIError(code ?? "unknown", normalized, { cause });
}

async function requestRateLimitJSON<T>(schema: z.ZodType<T>, path: string, options?: RequestInit): Promise<T> {
  try {
    return await requestParsedJSON(schema, path, options);
  } catch (error) {
    if (!(error instanceof APIClientError)) throw error;
    throw toRuleAPIError(error.message, error.status, error);
  }
}

/** The gateway compiles at most this many rules per type, so one page holds the whole set. */
export const RULE_LIST_LIMIT = 1000;

export function listRateLimitRules(query = "") {
  return requestRateLimitJSON(listResponseSchema, `/api/gateway-policies/rate-limit-rules?query=${encodeURIComponent(query)}&limit=${RULE_LIST_LIMIT}`);
}

export function getRateLimitEntitlements() {
  return requestRateLimitJSON(effectiveEntitlementsSchema, "/api/entitlements");
}

export function getRateLimitRule(id: string) {
  return requestRateLimitJSON(rateLimitRuleResponseSchema, `/api/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}`);
}

export function createRateLimitRule(body: RateLimitRuleFormInput) {
  return requestRateLimitJSON(rateLimitRuleResponseSchema, "/api/gateway-policies/rate-limit-rules", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateRateLimitRule(id: string, body: Partial<RateLimitRuleFormInput>) {
  return requestRateLimitJSON(rateLimitRuleResponseSchema, `/api/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteRateLimitRule(id: string) {
  return requestRateLimitJSON(successResponseSchema, `/api/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
}

export function updateRateLimitRuleStatus(id: string, enabled: boolean) {
  return requestRateLimitJSON(rateLimitRuleResponseSchema, `/api/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}/status`, {
    method: "PATCH",
    body: JSON.stringify({ enabled }),
  });
}
