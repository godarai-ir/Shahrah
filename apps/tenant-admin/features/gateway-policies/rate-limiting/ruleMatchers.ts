import type { GatewayRequestContext, RateLimitCondition, RateLimitRule } from "./ruleTypes";

// Mirrors services/control-plane/internal/billing/budget_policies.go
// (budgetStructuredConditionMatches), which evaluates both rate and budget rules.

function toStringValue(value: unknown): string {
  if (value === undefined || value === null) return "";
  return String(value).trim();
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
  if (Number.isNaN(actual) || Number.isNaN(expected)) return false;
  if (operator === "gt") return actual > expected;
  if (operator === "gte") return actual >= expected;
  if (operator === "lt") return actual < expected;
  if (operator === "lte") return actual <= expected;
  return false;
}

/** The request identity a subject condition reads, chosen by its subjectKind. */
function subjectValue(request: GatewayRequestContext, kind: RateLimitCondition["subjectKind"]): string {
  if (kind === "team") return toStringValue(request.teamId);
  if (kind === "virtual_account") return toStringValue(request.apiKeyId);
  return toStringValue(request.userId);
}

function doesConditionMatch(request: GatewayRequestContext, condition: RateLimitCondition): boolean {
  if (condition.field === "metadata") {
    const raw = request.metadata?.[condition.metadataKey?.trim() ?? ""];
    const actual = toStringValue(raw);
    if (condition.operator === "exists") return raw !== undefined && actual !== "";
    if (condition.operator === "not_exists") return raw === undefined || actual === "";
    if (["gt", "gte", "lt", "lte"].includes(condition.operator)) {
      return raw !== undefined && compareNumeric(condition.operator, Number(actual), Number(toStringValue(condition.value)));
    }
    return compareText(condition.operator, actual, toStringValue(condition.value));
  }

  const actual = condition.field === "model" ? toStringValue(request.model) : subjectValue(request, condition.subjectKind);
  const values = (condition.values ?? []).map(toStringValue).filter(Boolean);
  if (condition.operator === "in") return values.includes(actual);
  if (condition.operator === "not_in") return !values.includes(actual);
  return false;
}

/** Whether the conditions alone match, ignoring enabled state and schedule. */
export function doConditionsMatch(request: GatewayRequestContext, conditions: RateLimitCondition[] | undefined): boolean {
  return (conditions ?? []).every((condition) => doesConditionMatch(request, condition));
}

export function doesRequestMatchRateLimitRule(request: GatewayRequestContext, rule: RateLimitRule): boolean {
  if (!rule.enabled) return false;
  return doConditionsMatch(request, rule.conditions?.conditions);
}
