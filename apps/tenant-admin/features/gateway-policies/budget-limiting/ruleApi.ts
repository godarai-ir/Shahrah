import { z } from "zod";

import { requestParsedJSON } from "@/lib/http/api-client";

import type { BudgetLimitRuleFormInput } from "./ruleValidation";
import { budgetLimitRuleResponseSchema } from "./ruleValidation";

const listResponseSchema = z.object({
  data: z.array(budgetLimitRuleResponseSchema),
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

export function listBudgetLimitRules(query = "") {
  return requestParsedJSON(listResponseSchema, `/api/gateway-policies/budget-limit-rules?query=${encodeURIComponent(query)}`);
}

export function getBudgetLimitEntitlements() {
  return requestParsedJSON(effectiveEntitlementsSchema, "/api/entitlements");
}

export function getBudgetLimitRule(id: string) {
  return requestParsedJSON(budgetLimitRuleResponseSchema, `/api/gateway-policies/budget-limit-rules/${encodeURIComponent(id)}`);
}

export function createBudgetLimitRule(body: BudgetLimitRuleFormInput) {
  return requestParsedJSON(budgetLimitRuleResponseSchema, "/api/gateway-policies/budget-limit-rules", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateBudgetLimitRule(id: string, body: Partial<BudgetLimitRuleFormInput>) {
  return requestParsedJSON(budgetLimitRuleResponseSchema, `/api/gateway-policies/budget-limit-rules/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteBudgetLimitRule(id: string) {
  return requestParsedJSON(successResponseSchema, `/api/gateway-policies/budget-limit-rules/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
}

export function updateBudgetLimitRuleStatus(id: string, enabled: boolean) {
  return requestParsedJSON(budgetLimitRuleResponseSchema, `/api/gateway-policies/budget-limit-rules/${encodeURIComponent(id)}/status`, {
    method: "PATCH",
    body: JSON.stringify({ enabled }),
  });
}
