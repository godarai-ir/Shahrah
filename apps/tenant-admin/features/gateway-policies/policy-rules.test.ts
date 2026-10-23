import { createTranslator } from "next-intl";
import { describe, expect, it } from "vitest";

import fa from "@/messages/modules/gatewayPolicies/fa.json";

import type { BudgetLimitRule } from "./budget-limiting/ruleTypes";
import { compareByEvaluationOrder, findOverlappingRules, policyRuleTargets, ruleSchedule, type PolicyRule } from "./policy-rules";
import type { RateLimitRule } from "./rate-limiting/ruleTypes";
import { ruleSentence, type SentenceTranslator } from "./rule-sentence";
import { buildRuleConditions, readRuleTargets, type PolicyCondition } from "./rule-targets";
import { simulateRequest } from "./simulate-request";

const NOW = new Date("2026-09-19T12:00:00Z");

function rate(ruleID: string, overrides: Partial<RateLimitRule> = {}, conditions: PolicyCondition[] = []): PolicyRule {
  return {
    type: "rate",
    rule: {
      id: ruleID,
      rule_id: ruleID,
      enabled: true,
      priority: 100,
      conditions: { combinator: "AND", conditions },
      limit: { dimension: "requests", max: 600, window: { value: 1, unit: "minute" } },
      conflict_behavior: "most_restrictive",
      response_behavior: "return_429",
      soft_limit: false,
      retry_after_header: true,
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
      ...overrides,
    },
  };
}

function budget(ruleID: string, overrides: Partial<BudgetLimitRule> = {}, conditions: PolicyCondition[] = []): PolicyRule {
  return {
    type: "budget",
    rule: {
      id: ruleID,
      rule_id: ruleID,
      enabled: true,
      priority: 100,
      conditions: { combinator: "AND", conditions },
      limit: "5000",
      unit: "cost_per_month",
      on_exceed: "block",
      alert_thresholds: [80, 100],
      created_at: "2026-01-01T00:00:00Z",
      updated_at: "2026-01-01T00:00:00Z",
      ...overrides,
    },
  };
}

const team = (...ids: string[]): PolicyCondition => ({ field: "subject", operator: "in", subjectKind: "team", values: ids, valueLabels: { checkout: "تیم چک‌اوت" } });
const user = (...ids: string[]): PolicyCondition => ({ field: "subject", operator: "in", subjectKind: "user", values: ids });
const model = (...ids: string[]): PolicyCondition => ({ field: "model", operator: "in", values: ids });

describe("readRuleTargets / buildRuleConditions", () => {
  it("treats missing groups as every requester and every model", () => {
    expect(readRuleTargets([])).toMatchObject({ subject: null, models: null, metadata: [], legacySubject: false });
  });

  it("upgrades legacy kind:id subject tokens and flags dropped kinds", () => {
    const targets = readRuleTargets([{ field: "subject", operator: "in", values: ["team:t1", "team:t2", "user:u1"] }]);
    expect(targets.legacySubject).toBe(true);
    expect(targets.droppedLegacyKinds).toBe(true);
    expect(targets.subject).toEqual({ kind: "team", ids: ["t1", "t2"], labels: {} });
    expect(buildRuleConditions(targets)).toEqual([{ field: "subject", operator: "in", subjectKind: "team", values: ["t1", "t2"] }]);
  });

  it("writes raw IDs with subjectKind, keeps labels that differ from the ID, and drops empty groups", () => {
    const conditions = buildRuleConditions({
      subject: { kind: "virtual_account", ids: ["ou-1", "ou-2"], labels: { "ou-1": "Finance", "ou-2": "ou-2" } },
      models: [],
      metadata: [{ field: "metadata", operator: "exists", metadataKey: "project" }],
    });
    expect(conditions).toEqual([
      { field: "subject", operator: "in", subjectKind: "virtual_account", values: ["ou-1", "ou-2"], valueLabels: { "ou-1": "Finance" } },
      { field: "metadata", operator: "exists", metadataKey: "project" },
    ]);
  });
});

describe("evaluation order and schedule", () => {
  it("orders by priority, then rule_id, like the runtime profile", () => {
    const rules = [rate("b", { priority: 10 }), rate("a", { priority: 10 }), rate("c", { priority: 5 })];
    expect(rules.sort(compareByEvaluationOrder).map((item) => item.rule.rule_id)).toEqual(["c", "a", "b"]);
  });

  it("reads effective_from and expires_at", () => {
    expect(ruleSchedule(rate("x", { effective_from: "2026-10-01T00:00:00Z" }), NOW)).toBe("scheduled");
    expect(ruleSchedule(rate("x", { expires_at: "2026-09-01T00:00:00Z" }), NOW)).toBe("expired");
    expect(ruleSchedule(rate("x", { effective_from: "", expires_at: "" }), NOW)).toBe("active");
  });
});

describe("findOverlappingRules", () => {
  const target = readRuleTargets([team("checkout"), model("gpt-4o-mini")]);

  it("finds same-type live rules whose subject and model sets intersect, with empty meaning all", () => {
    const rules = [
      rate("all-traffic"),
      rate("same-team", {}, [team("checkout", "payments")]),
      rate("other-team", {}, [team("search")]),
      rate("other-model", {}, [model("gpt-4o")]),
      rate("user-rule", {}, [user("u1")]),
      rate("disabled", { enabled: false }),
      rate("expired", { expires_at: "2026-01-01T00:00:00Z" }),
      budget("budget-all"),
    ];
    expect(findOverlappingRules("rate", target, rules, NOW).map((item) => item.rule.rule_id)).toEqual(["all-traffic", "same-team", "user-rule"]);
  });

  it("never overlaps a legacy rule, which matches nothing at runtime", () => {
    const legacy = rate("legacy", {}, [{ field: "subject", operator: "in", values: ["team:checkout"] }]);
    expect(findOverlappingRules("rate", target, [legacy], NOW)).toEqual([]);
  });

  it("excludes the rule being edited", () => {
    expect(findOverlappingRules("rate", target, [rate("self")], NOW, "rate:self")).toEqual([]);
  });
});

describe("simulateRequest", () => {
  const request = { userId: "u1", teamId: "checkout", model: "gpt-4o-mini", metadata: { project: "p1" } };

  it("enforces every matching rate rule until one stops evaluation", () => {
    const rules = [
      rate("tenant-wide", { priority: 50 }),
      rate("soft", { priority: 60, soft_limit: true }, [team("checkout")]),
      rate("stopper", { priority: 70, conflict_behavior: "first_match" }, [model("gpt-4o-mini")]),
      rate("after-stop", { priority: 80 }),
      rate("no-match", { priority: 10 }, [team("search")]),
    ];
    const result = simulateRequest(rules, request, NOW);
    expect(result.rate.map(({ rule, outcome }) => [rule.rule_id, outcome])).toEqual([
      ["tenant-wide", "enforced"],
      ["soft", "countOnly"],
      ["stopper", "enforced"],
      ["after-stop", "notReached"],
    ]);
    expect(result.stoppedBy?.rule_id).toBe("stopper");
  });

  it("enforces every matching blocking budget and reports the rest", () => {
    const rules = [
      budget("blocking"),
      budget("alert-only", { on_exceed: "alert" }),
      budget("off", { enabled: false }, [team("checkout")]),
      rate("later", { effective_from: "2026-12-01T00:00:00Z" }),
    ];
    const result = simulateRequest(rules, request, NOW);
    expect(result.budget.map(({ rule, outcome }) => [rule.rule_id, outcome])).toEqual([["alert-only", "alertOnly"], ["blocking", "enforced"]]);
    expect(result.dormant.map(({ item, reason }) => [item.rule.rule_id, reason])).toEqual([["later", "scheduled"], ["off", "disabled"]]);
  });
});

describe("ruleSentence", () => {
  const t = createTranslator({ locale: "fa", messages: { gatewayPolicies: fa }, namespace: "gatewayPolicies" }) as unknown as SentenceTranslator;
  const format = { t, formatNumber: (value: number) => new Intl.NumberFormat("fa-IR").format(value) };

  it("reads a rate rule as one sentence", () => {
    expect(ruleSentence(rate("r", {}, [team("checkout"), model("gpt-4o-mini")]), format)).toBe(
      "برای تیم «تیم چک‌اوت» روی gpt-4o-mini حداکثر ۶۰۰ درخواست در هر دقیقه",
    );
  });

  it("names every target and the window size when there are several", () => {
    const item = rate("r", { limit: { dimension: "requests", max: 1000, window: { value: 5, unit: "hour" } } }, [
      user("u1", "u2"),
      { field: "metadata", operator: "exists", metadataKey: "project" },
    ]);
    expect(ruleSentence(item, format)).toBe("برای هر یک از کاربران «u1»، «u2» روی همه مدل‌ها حداکثر ۱٬۰۰۰ درخواست در هر ۵ ساعت، به شرط ۱ شرط متادیتا");
  });

  it("reads a budget rule with its applies-per scope", () => {
    expect(ruleSentence(budget("b", { budget_applies_per: "team" }), format)).toBe(
      "برای همه درخواست‌کننده‌ها روی همه مدل‌ها حداکثر ۵٬۰۰۰ اعتبار در هر ماه، جدا برای هر تیم",
    );
  });

  it("uses the targets reader for list search too", () => {
    expect(policyRuleTargets(rate("r", {}, [model("a", "b")])).models).toEqual(["a", "b"]);
  });
});
