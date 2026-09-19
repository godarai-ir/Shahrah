import type { BudgetLimitRule } from "./budget-limiting/ruleTypes";
import type { RateLimitRule } from "./rate-limiting/ruleTypes";
import { readRuleTargets, type RuleTargets } from "./rule-targets";

export type PolicyRuleType = "rate" | "budget";

export type PolicyRule =
  | { type: "rate"; rule: RateLimitRule }
  | { type: "budget"; rule: BudgetLimitRule };

export function policyRuleKey(item: PolicyRule): string {
  return `${item.type}:${item.rule.rule_id}`;
}

export function policyRuleName(item: PolicyRule): string {
  return item.rule.name?.trim() || item.rule.rule_id;
}

export function policyRuleTargets(item: PolicyRule): RuleTargets {
  return readRuleTargets(item.rule.conditions?.conditions);
}

export type RuleSchedule = "active" | "scheduled" | "expired";

function parseTime(value: string | undefined): number | null {
  if (!value?.trim()) return null;
  const time = Date.parse(value);
  return Number.isNaN(time) ? null : time;
}

/** Budget rules have no schedule; rate rules honor effective_from and expires_at. */
export function ruleSchedule(item: PolicyRule, now: Date): RuleSchedule {
  if (item.type === "budget") return "active";
  const from = parseTime(item.rule.effective_from);
  const until = parseTime(item.rule.expires_at);
  if (from !== null && now.getTime() < from) return "scheduled";
  if (until !== null && now.getTime() >= until) return "expired";
  return "active";
}

/** Gateway evaluation order (runtimeprofile.sortProfilePolicies): priority ascending, then rule_id. */
export function compareByEvaluationOrder(a: PolicyRule, b: PolicyRule): number {
  if (a.rule.priority !== b.rule.priority) return a.rule.priority - b.rule.priority;
  return a.rule.rule_id < b.rule.rule_id ? -1 : a.rule.rule_id > b.rule.rule_id ? 1 : 0;
}

/**
 * Whether one request could satisfy both target sets. An empty group means
 * every value. Subjects of different kinds read different identities of the
 * same request (user, key team, …), so they can always coincide. Metadata is
 * not compared: it is sent per request, so any two metadata conditions may hold together.
 */
export function targetsOverlap(a: RuleTargets, b: RuleTargets): boolean {
  if (a.legacySubject || b.legacySubject) return false;

  if (a.subject && b.subject && a.subject.kind === b.subject.kind) {
    const ids = new Set(a.subject.ids);
    if (!b.subject.ids.some((id) => ids.has(id))) return false;
  }

  if (a.models && b.models) {
    const models = new Set(a.models);
    if (!b.models.some((model) => models.has(model))) return false;
  }

  return true;
}

/**
 * Rules that take part in enforcement: enabled, not expired for rate rules, and
 * blocking for budget rules (the gateway skips non-blocking budgets entirely).
 */
export function isRuleLive(item: PolicyRule, now: Date): boolean {
  if (item.type === "budget" && !item.rule.block_on_budget_exceed) return false;
  return item.rule.enabled && ruleSchedule(item, now) !== "expired";
}

/** Other live rules of the same type that can apply to the same request as `targets`. */
export function findOverlappingRules(
  type: PolicyRuleType,
  targets: RuleTargets,
  rules: PolicyRule[],
  now: Date,
  excludeKey?: string,
): PolicyRule[] {
  return rules
    .filter((other) => other.type === type && policyRuleKey(other) !== excludeKey && isRuleLive(other, now))
    .filter((other) => targetsOverlap(targets, policyRuleTargets(other)))
    .sort(compareByEvaluationOrder);
}
