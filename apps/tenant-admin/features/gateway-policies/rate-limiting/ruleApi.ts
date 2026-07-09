import type { RateLimitRule } from "./ruleTypes";
import type { RateLimitRuleFormInput } from "./ruleValidation";

type ListResponse = {
  data: RateLimitRule[];
  total: number;
  limit: number;
  offset: number;
};

type EffectiveEntitlements = {
  features?: Record<string, unknown>;
} | null;

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

async function requestJSON<T>(path: string, options?: RequestInit): Promise<T> {
  const headers = new Headers(options?.headers ?? {});
  if (options?.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(path, { ...options, headers, cache: "no-store" });
  const text = await response.text();
  let payload: Record<string, unknown> = {};
  if (text.trim()) {
    try {
      payload = JSON.parse(text) as Record<string, unknown>;
    } catch {
      payload = { message: text };
    }
  }
  if (!response.ok) {
    const rawMessage = sanitizeServerMessage(
      (payload?.error as { message?: string } | undefined)?.message ??
        (payload?.message as string | undefined) ??
        `Request failed (${response.status})`,
      response.status,
    );
    throw new Error(mapAPIErrorToFa(rawMessage, response.status));
  }
  return payload as T;
}

export function listRateLimitRules(query = "") {
  return requestJSON<ListResponse>(`/api/gateway-policies/rate-limit-rules?query=${encodeURIComponent(query)}`);
}

export function getRateLimitEntitlements() {
  return requestJSON<EffectiveEntitlements>("/api/entitlements");
}

export function getRateLimitRule(id: string) {
  return requestJSON<RateLimitRule>(`/api/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}`);
}

export function createRateLimitRule(body: RateLimitRuleFormInput) {
  return requestJSON<RateLimitRule>("/api/gateway-policies/rate-limit-rules", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateRateLimitRule(id: string, body: Partial<RateLimitRuleFormInput>) {
  return requestJSON<RateLimitRule>(`/api/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteRateLimitRule(id: string) {
  return requestJSON<{ success: true }>(`/api/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
}

export function updateRateLimitRuleStatus(id: string, enabled: boolean) {
  return requestJSON<RateLimitRule>(`/api/gateway-policies/rate-limit-rules/${encodeURIComponent(id)}/status`, {
    method: "PATCH",
    body: JSON.stringify({ enabled }),
  });
}
