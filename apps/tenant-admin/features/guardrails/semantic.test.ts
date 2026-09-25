import { describe, expect, it } from "vitest";

import { emptyForm, validateStep } from "./form";
import {
  defaultPackRules,
  newSemanticRule,
  nextRuleKey,
  packSelection,
  packTargets,
  semanticBehaviorError,
  semanticConfig,
  semanticConfigError,
  semanticCostPerThousand,
  semanticFieldsFromConfig,
  statementWarnings,
} from "./semantic";
import type { GuardrailFormInput } from "./types";

const t = (key: string) => key;

function customForm(patch: Partial<GuardrailFormInput> = {}): GuardrailFormInput {
  const rule = { ...newSemanticRule("financial_commitment"), statement: "The reply promises a refund.", examplesTrue: "We will refund you.\n\n", review: "0.3", block: "0.6" };

  return { ...emptyForm, type: "semantic_policy", semanticSource: "custom", semanticTarget: "output", semanticRules: [rule], ...patch };
}

describe("semantic policy config", () => {
  it("round-trips a custom policy through the gateway config", () => {
    const config = semanticConfig({ ...customForm(), semanticSeverity: true, semanticSeverityBlockAt: "2.5" });

    expect(config).toEqual({
      target: "output",
      rules: [
        {
          key: "financial_commitment",
          statement: "The reply promises a refund.",
          examples: { true: ["We will refund you."], false: [] },
          review_threshold: 0.3,
          block_threshold: 0.6,
        },
      ],
      finglish_thresholds: { review: 0.3, block: 0.5 },
      severity: { enabled: true, block_at: 2.5 },
    });

    const fields = semanticFieldsFromConfig(config);
    expect(fields.semanticSource).toBe("custom");
    expect(fields.semanticTarget).toBe("output");
    expect(fields.semanticRules[0]).toMatchObject({ key: "financial_commitment", examplesTrue: "We will refund you.", review: "0.3", block: "0.6" });
    expect(fields.semanticSeverity).toBe(true);
    expect(fields.semanticSeverityBlockAt).toBe("2.5");
  });

  it("reads pack threshold overrides and keeps defaults for the rest", () => {
    const fields = semanticFieldsFromConfig({ pack: "security-baseline@1", rules: [{ key: "jailbreak", block_threshold: 0.9 }] });

    expect(fields.semanticSource).toBe("pack");
    expect(fields.semanticFinglish).toBe(false);
    expect(fields.semanticPackRules.map((rule) => [rule.key, rule.block])).toEqual([
      ["jailbreak", "0.9"],
      ["system_prompt_extraction", "0.7"],
      ["evaluator_address", "0.7"],
    ]);
  });

  it("round-trips an industry pack on the output hook with its review-only defaults", () => {
    const fields = semanticFieldsFromConfig({ pack: "banking@1", target: "output", rules: [{ key: "financial_commitment", block_threshold: 0.8 }] });

    expect(fields.semanticPack).toBe("banking@1");
    expect(fields.semanticTarget).toBe("output");
    expect(fields.semanticPackRules.map((rule) => [rule.key, rule.block])).toEqual([
      ["financial_commitment", "0.8"],
      ["personal_financial_advice", "1"],
    ]);

    const config = semanticConfig({ ...customForm(), ...fields });
    expect(config).toMatchObject({ pack: "banking@1", target: "output" });
    expect(config.rules).toEqual([
      { key: "financial_commitment", review_threshold: 0.35, block_threshold: 0.8 },
      { key: "personal_financial_advice", review_threshold: 0.35, block_threshold: 1 },
    ]);
  });

  it("includes the security baseline in an industry pack's input rules", () => {
    expect(defaultPackRules("healthcare@1", "input").map((rule) => rule.key)).toEqual([
      "jailbreak",
      "system_prompt_extraction",
      "evaluator_address",
      "self_harm_risk",
    ]);
  });

  it("moves to the input hook when the pack has no output rules", () => {
    expect(packTargets("security-baseline@1")).toEqual(["input"]);
    expect(packSelection("security-baseline@1", "output").semanticTarget).toBe("input");

    const config = semanticConfig({ ...customForm(), semanticSource: "pack", semanticPack: "security-baseline@1", semanticTarget: "output" });
    expect(config.target).toBe("input");
  });

  it("drops threshold rows that do not belong to the selected pack and hook", () => {
    const form = { ...customForm(), ...packSelection("hr@1", "output"), semanticSource: "pack" as const };
    const config = semanticConfig({ ...form, semanticPackRules: [...form.semanticPackRules, { key: "jailbreak", review: "0.1", block: "0.2" }] });

    expect((config.rules as { key: string }[]).map((rule) => rule.key)).toEqual(["hiring_discrimination_reply", "protected_attribute_inference"]);
  });

  it("never sends custom rules for a pack policy", () => {
    const config = semanticConfig({ ...customForm(), semanticSource: "pack" });

    expect(config.pack).toBe("security-baseline@1");
    expect(JSON.stringify(config)).not.toContain("financial_commitment");
  });
});

describe("semantic policy validation", () => {
  it("reports rule errors on the settings step and threshold errors on the behavior step", () => {
    const duplicate = customForm({ semanticRules: [...customForm().semanticRules, { ...customForm().semanticRules[0], id: "other" }] });
    const badKey = customForm({ semanticRules: [{ ...customForm().semanticRules[0], key: "Bad Key" }] });
    const badThresholds = customForm({ semanticRules: [{ ...customForm().semanticRules[0], review: "0.8", block: "0.2" }] });

    expect(semanticConfigError(duplicate)).toBe("semanticRuleKeyDuplicate");
    expect(semanticConfigError(badKey)).toBe("semanticRuleKey");
    expect(semanticConfigError(customForm({ semanticRules: [] }))).toBe("semanticRules");
    expect(semanticConfigError(badThresholds)).toBeNull();
    expect(semanticBehaviorError(badThresholds)).toBe("semanticThresholds");
    expect(validateStep(2, badThresholds, t)).toBe("validation.semanticThresholds");
    expect(validateStep(1, badKey, t)).toBe("validation.semanticRuleKey");
    expect(semanticBehaviorError(customForm({ semanticSeverity: true, semanticSeverityBlockAt: "4" }))).toBe("semanticSeverity");
  });

  it("does not validate a pack's custom rules", () => {
    expect(semanticConfigError({ ...customForm({ semanticRules: [] }), semanticSource: "pack" })).toBeNull();
  });

  it("suggests the next free rule key", () => {
    expect(nextRuleKey([newSemanticRule("rule_1"), newSemanticRule("rule_3")])).toBe("rule_4");
    expect(nextRuleKey([newSemanticRule("rule_2")])).toBe("rule_3");
  });
});

describe("statement warnings", () => {
  it("flags compound and negated statements in English and Persian", () => {
    expect(statementWarnings("The user is angry and asks for a refund.")).toEqual(["compound"]);
    expect(statementWarnings("کاربر عصبانی است و بازپرداخت می‌خواهد.")).toEqual(["compound"]);
    expect(statementWarnings("Not a question about the product.")).toEqual(["negation"]);
    expect(statementWarnings("نباید درباره‌ی رقبا حرف بزند.")).toEqual(["negation"]);
    expect(statementWarnings("The reply promises a refund.")).toEqual([]);
    expect(statementWarnings("   ")).toEqual([]);
  });
});

describe("semantic cost estimate", () => {
  it("grows with the number of rules and severity", () => {
    const one = semanticCostPerThousand(customForm());
    const withSeverity = semanticCostPerThousand(customForm({ semanticSeverity: true }));

    expect(one).toBeCloseTo(0.0344, 4);
    expect(withSeverity).toBeGreaterThan(one);
  });
});
