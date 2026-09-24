import { createTranslator } from "next-intl";
import { describe, expect, it } from "vitest";

import en from "@/messages/modules/guardrails/en.json";

import { emptyForm, firstInvalidStep, type GuardrailDraft, type GuardrailTranslator, validateStep } from "./form";

const t = createTranslator({
  locale: "en",
  messages: { guardrails: en },
  namespace: "guardrails",
  onError: (error) => {
    throw error;
  },
}) as unknown as GuardrailTranslator;

const draft = (overrides: Partial<GuardrailDraft> = {}): GuardrailDraft => ({ ...emptyForm, type: null, ...overrides });

describe("validateStep", () => {
  it("requires a type on the type step only", () => {
    expect(validateStep(0, draft(), t)).toBe(en.validation.type);
    expect(validateStep(1, draft(), t)).toBe("");
    expect(validateStep(0, draft({ type: "not_null" }), t)).toBe("");
  });

  it("reports config errors on the config step only", () => {
    const form = draft({ type: "keyword_blocklist", terms: "" });

    expect(validateStep(0, form, t)).toBe("");
    expect(validateStep(1, form, t)).toBe(en.validation.terms);
    expect(validateStep(2, form, t)).toBe("");
    expect(validateStep(3, { ...form, name: "Terms", slug: "terms" }, t)).toBe("");
  });

  it("reports naming errors on the review step only", () => {
    const form = draft({ type: "not_null", name: "", slug: "x" });

    expect(validateStep(1, form, t)).toBe("");
    expect(validateStep(3, form, t)).toBe(en.validation.name);
    expect(validateStep(3, { ...form, name: "Not null" }, t)).toBe(en.validation.slug);
  });

  it("never fails the behavior step", () => {
    expect(validateStep(2, draft(), t)).toBe("");
  });

  it.each([
    ["regex", { patterns: "" }, en.validation.patterns],
    ["regex", { patterns: "ok\n([a-z" }, en.validation.patternInvalid],
    ["regex_replace", { pattern: " " }, en.validation.pattern],
    ["regex_replace", { pattern: "(" }, en.validation.patternInvalid],
    ["word_count", { min: "", max: "" }, en.validation.range],
    ["word_count", { min: "50", max: "10" }, en.validation.rangeOrder],
    ["json_schema", { schema: "{" }, en.validation.json],
    ["json_keys", { keys: "\n" }, en.validation.keys],
    ["jwt_validation", { jwtVerifier: "jwks", jwksUri: "", sharedSecret: "left over" }, en.validation.jwtVerifier],
    ["jwt_validation", { jwtVerifier: "inline", inlineJwks: "{" }, en.validation.json],
  ] as const)("rejects an invalid %s config", (type, overrides, message) => {
    expect(validateStep(1, draft({ type, ...overrides }), t)).toBe(message);
  });

  it("accepts a one-sided count range", () => {
    expect(validateStep(1, draft({ type: "character_count", max: "200" }), t)).toBe("");
  });

  it("rejects a slug already in the loaded list when creating only", () => {
    const form = draft({ type: "not_null", name: "Not null", slug: "taken" });

    expect(validateStep(3, form, t, { existingSlugs: ["taken"] })).toBe(en.messages.duplicate);
    expect(validateStep(3, form, t, { mode: "edit", existingSlugs: ["taken"] })).toBe("");
  });
});

describe("firstInvalidStep", () => {
  it("finds the earliest failing step", () => {
    expect(firstInvalidStep(draft(), t)).toBe(0);
    expect(firstInvalidStep(draft({ type: "json_keys" }), t)).toBe(1);
    expect(firstInvalidStep(draft({ type: "not_null" }), t)).toBe(3);
    expect(firstInvalidStep(draft({ type: "not_null", name: "Empty", slug: "empty" }), t)).toBeNull();
  });
});
