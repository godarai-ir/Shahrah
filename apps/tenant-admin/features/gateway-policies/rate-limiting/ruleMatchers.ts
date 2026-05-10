import type { GatewayRequestContext, RateLimitCondition, RateLimitRule } from "./ruleTypes";

function toStringValue(value: unknown): string {
  if (value === undefined || value === null) return "";
  return String(value);
}

function compareText(operator: RateLimitCondition["operator"], actual: string, expected: string): boolean {
  if (operator === "equals") return actual === expected;
  if (operator === "not_equals") return actual !== expected;
  if (operator === "contains") return actual.includes(expected);
  if (operator === "starts_with") return actual.startsWith(expected);
  if (operator === "ends_with") return actual.endsWith(expected);
  return false;
}

function compareNumeric(operator: RateLimitCondition["operator"], actual: number, expected: number): boolean {
  if (operator === "gt") return actual > expected;
  if (operator === "gte") return actual >= expected;
  if (operator === "lt") return actual < expected;
  if (operator === "lte") return actual <= expected;
  return false;
}

function getRequestField(request: GatewayRequestContext, field: RateLimitCondition["field"]): unknown {
  switch (field) {
    case "subject":
      return request.subject;
    case "model":
      return request.model;
    default:
      return undefined;
  }
}

function doesConditionMatch(request: GatewayRequestContext, condition: RateLimitCondition): boolean {
  if (condition.field === "metadata") {
    const key = condition.metadataKey ?? "";
    const metadataValue = request.metadata?.[key];
    if (condition.operator === "exists") return metadataValue !== undefined;
    if (condition.operator === "not_exists") return metadataValue === undefined;
    if (metadataValue === undefined) return false;
    if (["gt", "gte", "lt", "lte"].includes(condition.operator)) {
      return compareNumeric(condition.operator, Number(metadataValue), Number(condition.value));
    }
    return compareText(condition.operator, toStringValue(metadataValue), toStringValue(condition.value));
  }

  const requestValue = toStringValue(getRequestField(request, condition.field));
  const values = (condition.values ?? []).map((v) => toStringValue(v));
  if (condition.operator === "in") return values.includes(requestValue);
  if (condition.operator === "not_in") return !values.includes(requestValue);
  return compareText(condition.operator, requestValue, toStringValue(condition.value));
}

export function doesRequestMatchRateLimitRule(request: GatewayRequestContext, rule: RateLimitRule): boolean {
  if (!rule.enabled) return false;
  if (!rule.conditions?.conditions?.length) return true;
  return rule.conditions.conditions.every((condition) => doesConditionMatch(request, condition));
}

export function buildRateLimitKey(request: GatewayRequestContext, rule: RateLimitRule): string {
  const parts = [`rate_limit`, rule.rule_id];
  for (const condition of rule.conditions.conditions) {
    if (condition.field === "metadata") {
      const key = condition.metadataKey ?? "unknown";
      parts.push(`meta:${key}=${toStringValue(request.metadata?.[key])}`);
      continue;
    }
    const value = toStringValue(getRequestField(request, condition.field));
    parts.push(`${condition.field}:${value}`);
  }
  return parts.join(":");
}
