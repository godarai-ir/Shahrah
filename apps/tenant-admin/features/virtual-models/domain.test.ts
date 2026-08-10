import { describe, expect, it } from "vitest";

import {
  buildModeSummary,
  canonicalModelFQN,
  createCopyAlias,
  makeDefaultForm,
  MAX_TARGETS,
  normalizeTargets,
  rebalanceTargetsForWeighted,
  validateForm,
} from "./domain";

describe("virtual model domain", () => {
  it("canonicalizes provider-prefixed model identifiers", () => {
    expect(canonicalModelFQN("openai:primary:openai/gpt-4.1")).toBe("openai:primary:gpt-4.1");
  });

  it("creates a unique copy alias", () => {
    expect(createCopyAlias("router", new Set(["copy_of_router", "copy_of_2_router"])))
      .toBe("copy_of_3_router");
  });

  it("normalizes missing target policy defaults", () => {
    const [target] = normalizeTargets([{
      order: 0,
      provider: "openai",
      model: "openai:primary:gpt-4.1",
      policy: { retry: { attempts: 0, delay_ms: 0 }, fallback: { candidate: false } },
    }]);

    expect(target.order).toBe(1);
    expect(target.policy.timeout_sec).toBe(60);
  });

  it("rebalances weighted targets to exactly one hundred", () => {
    const targets = normalizeTargets(undefined);
    const weighted = rebalanceTargetsForWeighted([...targets, { ...targets[0], order: 2 }, { ...targets[0], order: 3 }]);

    expect(weighted.reduce((sum, target) => sum + (target.weight ?? 0), 0)).toBe(100);
  });

  it("rejects forms with more than three targets", () => {
    const form = makeDefaultForm("Everyone");
    form.alias = "router";
    form.targets = Array.from({ length: MAX_TARGETS + 1 }, (_, index) => ({
      ...form.targets[0],
      order: index + 1,
    }));

    expect(validateForm(form, (key) => key)).toBe("messages.targetsLimit");
  });

  it("separates common and partial endpoint modes", () => {
    const targets = normalizeTargets(undefined);
    targets[0].model = "openai:primary:gpt-4.1";
    const second = { ...targets[0], order: 2, model: "anthropic:primary:claude" };

    expect(buildModeSummary([targets[0], second], {
      "openai:primary:gpt-4.1": ["chat", "responses"],
      "anthropic:primary:claude": ["chat"],
    })).toEqual({ intersection: ["chat"], partialModes: ["responses"], ready: true });
  });
});
