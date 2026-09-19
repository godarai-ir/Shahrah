import { doConditionsMatch } from "./rate-limiting/ruleMatchers";
import type { GatewayRequestContext } from "./rate-limiting/ruleTypes";
import type { BudgetLimitRule } from "./budget-limiting/ruleTypes";
import type { RateLimitRule } from "./rate-limiting/ruleTypes";
import { compareByEvaluationOrder, ruleSchedule, type PolicyRule, type RuleSchedule } from "./policy-rules";

/**
 * - enforced: the rule's counter counts this request, and exceeding it rejects with 429.
 * - countOnly: a soft limit; the counter counts but never rejects.
 * - notReached: an earlier matching rule set to stop evaluation ended the pass.
 */
export type RateOutcome = "enforced" | "countOnly" | "notReached";

/** - enforced: the request reserves against this budget. - noEffect: blocking is off, so the gateway skips the rule. */
export type BudgetOutcome = "enforced" | "noEffect";

export type DormantReason = "disabled" | Exclude<RuleSchedule, "active">;

export type SimulationResult = {
  rate: Array<{ rule: RateLimitRule; outcome: RateOutcome }>;
  /** The first-match rule that ended rate evaluation, if any. */
  stoppedBy: RateLimitRule | null;
  budget: Array<{ rule: BudgetLimitRule; outcome: BudgetOutcome }>;
  /** Rules whose conditions match but that are off or outside their schedule. */
  dormant: Array<{ item: PolicyRule; reason: DormantReason }>;
};

/**
 * Replays the gateway's rule pass for one request
 * (billing.checkRateLimitPolicies and budgetPolicyTargets). Every matching
 * rate rule is checked against its own counter; the pass stops only after a
 * matching rule whose conflict_behavior is first_match. most_restrictive and
 * apply_all enforce the same way and differ only in which rule's numbers the
 * response headers report.
 */
export function simulateRequest(rules: PolicyRule[], request: GatewayRequestContext, now: Date): SimulationResult {
  const result: SimulationResult = { rate: [], stoppedBy: null, budget: [], dormant: [] };

  for (const item of [...rules].sort(compareByEvaluationOrder)) {
    if (!doConditionsMatch(request, item.rule.conditions?.conditions)) continue;

    if (!item.rule.enabled) {
      result.dormant.push({ item, reason: "disabled" });
      continue;
    }

    const schedule = ruleSchedule(item, now);
    if (schedule !== "active") {
      result.dormant.push({ item, reason: schedule });
      continue;
    }

    if (item.type === "budget") {
      result.budget.push({ rule: item.rule, outcome: item.rule.block_on_budget_exceed ? "enforced" : "noEffect" });
      continue;
    }

    if (result.stoppedBy) {
      result.rate.push({ rule: item.rule, outcome: "notReached" });
      continue;
    }

    result.rate.push({ rule: item.rule, outcome: item.rule.soft_limit ? "countOnly" : "enforced" });
    if (item.rule.conflict_behavior === "first_match") result.stoppedBy = item.rule;
  }

  return result;
}
