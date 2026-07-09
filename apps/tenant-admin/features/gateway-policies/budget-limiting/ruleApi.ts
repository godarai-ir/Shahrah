import type { BudgetLimitRule } from "./ruleTypes";
import type { BudgetLimitRuleFormInput } from "./ruleValidation";

type ListResponse = {
  data: BudgetLimitRule[];
  total: number;
  limit: number;
  offset: number;
};

type EffectiveEntitlements = {
  features?: Record<string, unknown>;
} | null;

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
    const message =
      (payload?.error as { message?: string } | undefined)?.message ??
      (payload?.message as string | undefined) ??
      `Request failed (${response.status})`;
    throw new Error(message);
  }
  return payload as T;
}

export function listBudgetLimitRules(query = "") {
  return requestJSON<ListResponse>(`/api/gateway-policies/budget-limit-rules?query=${encodeURIComponent(query)}`);
}

export function getBudgetLimitEntitlements() {
  return requestJSON<EffectiveEntitlements>("/api/entitlements");
}

export function getBudgetLimitRule(id: string) {
  return requestJSON<BudgetLimitRule>(`/api/gateway-policies/budget-limit-rules/${encodeURIComponent(id)}`);
}

export function createBudgetLimitRule(body: BudgetLimitRuleFormInput) {
  return requestJSON<BudgetLimitRule>("/api/gateway-policies/budget-limit-rules", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateBudgetLimitRule(id: string, body: Partial<BudgetLimitRuleFormInput>) {
  return requestJSON<BudgetLimitRule>(`/api/gateway-policies/budget-limit-rules/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteBudgetLimitRule(id: string) {
  return requestJSON<{ success: true }>(`/api/gateway-policies/budget-limit-rules/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
}

export function updateBudgetLimitRuleStatus(id: string, enabled: boolean) {
  return requestJSON<BudgetLimitRule>(`/api/gateway-policies/budget-limit-rules/${encodeURIComponent(id)}/status`, {
    method: "PATCH",
    body: JSON.stringify({ enabled }),
  });
}
