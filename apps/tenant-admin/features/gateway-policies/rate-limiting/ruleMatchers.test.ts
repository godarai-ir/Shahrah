import { describe, expect, it } from "vitest";

import { doConditionsMatch, doesRequestMatchRateLimitRule } from "./ruleMatchers";
import type { RateLimitRule } from "./ruleTypes";

const baseRule: RateLimitRule = {
  id: "1",
  rule_id: "image-model-hourly-limit",
  enabled: true,
  priority: 100,
  conditions: {
    combinator: "AND",
    conditions: [
      { field: "subject", operator: "in", subjectKind: "team", values: ["team-a", "team-b"] },
      { field: "model", operator: "in", values: ["chatgpt-image-latest", "gpt-4o-mini-2024-07-18"] },
      { field: "metadata", operator: "equals", metadataKey: "my-key", value: "7090" },
    ],
  },
  limit: { dimension: "requests", max: 90, window: { value: 1, unit: "hour" } },
  conflict_behavior: "most_restrictive",
  response_behavior: "return_429",
  soft_limit: false,
  retry_after_header: true,
  created_at: "2026-01-01T00:00:00.000Z",
  updated_at: "2026-01-01T00:00:00.000Z",
};

describe("doesRequestMatchRateLimitRule", () => {
  it("matches the subject through its kind, plus model and metadata", () => {
    expect(
      doesRequestMatchRateLimitRule({ teamId: "team-a", model: "chatgpt-image-latest", metadata: { "my-key": "7090" } }, baseRule),
    ).toBe(true);
  });

  it("does not read a team subject from the user identity", () => {
    expect(
      doesRequestMatchRateLimitRule({ userId: "team-a", model: "chatgpt-image-latest", metadata: { "my-key": "7090" } }, baseRule),
    ).toBe(false);
  });

  it("reads the user identity when subjectKind is missing, as the gateway does", () => {
    const conditions = [{ field: "subject" as const, operator: "in" as const, values: ["user-1"] }];
    expect(doConditionsMatch({ userId: "user-1" }, conditions)).toBe(true);
    // Legacy `kind:id` tokens are compared verbatim and never equal a real ID.
    expect(doConditionsMatch({ userId: "user-1" }, [{ ...conditions[0], values: ["user:user-1"] }])).toBe(false);
  });

  it("ignores disabled rules", () => {
    expect(doesRequestMatchRateLimitRule({ teamId: "team-a", model: "chatgpt-image-latest", metadata: { "my-key": "7090" } }, { ...baseRule, enabled: false })).toBe(false);
  });

  it("supports metadata exists and model not_in", () => {
    const conditions = [
      { field: "model" as const, operator: "not_in" as const, values: ["blocked-model"] },
      { field: "metadata" as const, operator: "exists" as const, metadataKey: "k1" },
    ];
    expect(doConditionsMatch({ model: "gpt-4o", metadata: { k1: "1" } }, conditions)).toBe(true);
    expect(doConditionsMatch({ model: "gpt-4o", metadata: { k1: " " } }, conditions)).toBe(false);
  });

  it("supports contains and numeric comparisons", () => {
    expect(doConditionsMatch({ metadata: { k: "xyzabcuvw" } }, [{ field: "metadata", operator: "contains", metadataKey: "k", value: "abc" }])).toBe(true);
    expect(doConditionsMatch({ metadata: { tier: 3 } }, [{ field: "metadata", operator: "gte", metadataKey: "tier", value: "2" }])).toBe(true);
    expect(doConditionsMatch({}, [{ field: "metadata", operator: "lt", metadataKey: "tier", value: "2" }])).toBe(false);
  });
});
