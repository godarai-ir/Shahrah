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

function mapAPIErrorToFa(message: string, status?: number): string {
  const m = message.toLowerCase();
  if (m.includes("no rows in result set")) {
    return "ذخیره قانون انجام نشد. لطفا فیلترها و مقادیر واردشده را بررسی کرده و دوباره تلاش کنید.";
  }
  if (m.includes("invalid request body")) {
    return "داده‌های فرم معتبر نیست. لطفا فیلدها را بررسی کنید.";
  }
  if (m.includes("already exists") || status === 409) {
    return "شناسه قانون تکراری است. یک شناسه یکتا وارد کنید.";
  }
  if (m.includes("internal server error") || status === 500) {
    return "خطای داخلی در ذخیره قانون رخ داد. لطفا دوباره تلاش کنید یا با پشتیبانی تماس بگیرید.";
  }
  return message;
}

function looksLikeHTMLDocument(value: string): boolean {
  const normalized = value.trim().toLowerCase();
  return normalized.startsWith("<!doctype") || normalized.startsWith("<html") || normalized.startsWith("<body");
}

function sanitizeServerMessage(message: string | undefined, status?: number): string {
  const normalized = message?.trim();
  if (!normalized) {
    return `Request failed (${status ?? "unknown"})`;
  }
  if (looksLikeHTMLDocument(normalized)) {
    if (status === 401) return "برای ادامه باید وارد حساب کاربری شوید.";
    if (status === 403) return "دسترسی شما به این بخش محدود است.";
    if (status === 404) return "منبع درخواستی پیدا نشد.";
    if (status === 500) return "خطای داخلی در پردازش درخواست رخ داد. لطفا دوباره تلاش کنید.";
    return "در پردازش درخواست خطایی رخ داد. لطفا دوباره تلاش کنید.";
  }
  return normalized;
}

async function requestRateLimitJSON<T>(schema: z.ZodType<T>, path: string, options?: RequestInit): Promise<T> {
  try {
    return await requestParsedJSON(schema, path, options);
  } catch (error) {
    if (!(error instanceof APIClientError)) throw error;
    const message = sanitizeServerMessage(error.message, error.status);
    throw new Error(mapAPIErrorToFa(message, error.status), { cause: error });
  }
}

export function listRateLimitRules(query = "") {
  return requestRateLimitJSON(listResponseSchema, `/api/gateway-policies/rate-limit-rules?query=${encodeURIComponent(query)}`);
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
