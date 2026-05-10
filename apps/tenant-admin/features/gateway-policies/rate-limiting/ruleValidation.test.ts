import { describe, expect, it } from "vitest";

import { rateLimitRuleSchema } from "./ruleValidation";

const valid = {
  rule_id: "valid_rule",
  enabled: true,
  priority: 100,
  conditions: { combinator: "AND", conditions: [] },
  limit: { dimension: "requests", max: 1, window: { value: 1, unit: "hour" } },
  conflict_behavior: "most_restrictive",
  response_behavior: "return_429",
  soft_limit: false,
  retry_after_header: true,
};

describe("rateLimitRuleSchema", () => {
  it("requires rule id", () => {
    const result = rateLimitRuleSchema.safeParse({ ...valid, rule_id: "" });
    expect(result.success).toBe(false);
  });

  it("validates rule id format", () => {
    const result = rateLimitRuleSchema.safeParse({ ...valid, rule_id: "Invalid ID" });
    expect(result.success).toBe(false);
  });

  it("validates positive limit", () => {
    const result = rateLimitRuleSchema.safeParse({ ...valid, limit: { ...valid.limit, max: 0 } });
    expect(result.success).toBe(false);
  });

  it("requires metadata key", () => {
    const result = rateLimitRuleSchema.safeParse({
      ...valid,
      conditions: { combinator: "AND", conditions: [{ field: "metadata", operator: "equals", value: "x" }] },
    });
    expect(result.success).toBe(false);
  });
});
