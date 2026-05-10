import { describe, expect, it } from "vitest";

import { buildRateLimitKey, doesRequestMatchRateLimitRule } from "./ruleMatchers";
import type { RateLimitRule } from "./ruleTypes";

const baseRule: RateLimitRule = {
  id: "1",
  rule_id: "image-model-hourly-limit",
  enabled: true,
  priority: 100,
  conditions: {
    combinator: "AND",
    conditions: [
      { field: "subject", operator: "in", values: ["Iman Zareh", "Team A"] },
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
  it("matches subject/model/metadata equals", () => {
    expect(
      doesRequestMatchRateLimitRule(
        { subject: "Team A", model: "chatgpt-image-latest", metadata: { "my-key": "7090" } },
        baseRule,
      ),
    ).toBe(true);
  });

  it("supports metadata exists and not_in", () => {
    const rule = {
      ...baseRule,
      conditions: {
        combinator: "AND" as const,
        conditions: [
          { field: "model" as const, operator: "not_in" as const, values: ["blocked-model"] },
          { field: "metadata" as const, operator: "exists" as const, metadataKey: "k1" },
        ],
      },
    };
    expect(doesRequestMatchRateLimitRule({ model: "gpt-4o", metadata: { k1: "1" } }, rule)).toBe(true);
  });

  it("supports contains behavior", () => {
    const rule = {
      ...baseRule,
      conditions: { combinator: "AND" as const, conditions: [{ field: "metadata" as const, operator: "contains" as const, metadataKey: "k", value: "abc" }] },
    };
    expect(doesRequestMatchRateLimitRule({ metadata: { k: "xyzabcuvw" } }, rule)).toBe(true);
  });
});

describe("buildRateLimitKey", () => {
  it("builds deterministic key", () => {
    const key = buildRateLimitKey(
      { subject: "Team A", model: "chatgpt-image-latest", metadata: { "my-key": "7090" } },
      baseRule,
    );
    expect(key).toBe("rate_limit:image-model-hourly-limit:subject:Team A:model:chatgpt-image-latest:meta:my-key=7090");
  });
});
