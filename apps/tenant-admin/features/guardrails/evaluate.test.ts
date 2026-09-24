import { describe, expect, it } from "vitest";

import fa from "@/messages/modules/guardrails/fa.json";

import { GUARDRAIL_CATALOG } from "./catalog";
import { evaluateSample, normalizeDigits } from "./evaluate";
import { emptyForm } from "./form";
import type { GuardrailFormInput, GuardrailType } from "./types";

const samples: Partial<Record<GuardrailType, string>> = fa.wizard.test.samples;

/** A form configured the way the wizard would be after picking each type. */
const configured: GuardrailFormInput = {
  ...emptyForm,
  terms: "ignore previous instructions\nsystem prompt",
  patterns: "(?i)password\\s*[:=]\\s*\\S+",
  pattern: "(?i)password\\s*[:=]\\s*\\S+",
  min: "1",
  max: "2",
  schema: '{ "type": "object", "required": ["answer"] }',
  keys: "answer\nconfidence",
  codeMode: "must_not_contain",
};

const expected: Record<GuardrailType, { operation: "validate" | "mutate"; status: string }> = {
  secrets_detection: { operation: "mutate", status: "mutate" },
  basic_pii: { operation: "mutate", status: "mutate" },
  regex: { operation: "validate", status: "block" },
  regex_replace: { operation: "mutate", status: "mutate" },
  keyword_blocklist: { operation: "validate", status: "block" },
  contains_code: { operation: "validate", status: "block" },
  not_null: { operation: "validate", status: "block" },
  sentence_count: { operation: "validate", status: "block" },
  word_count: { operation: "validate", status: "block" },
  character_count: { operation: "validate", status: "block" },
  json_schema: { operation: "validate", status: "pass" },
  json_keys: { operation: "validate", status: "block" },
  jwt_validation: { operation: "validate", status: "na" },
  inline_image_urls: { operation: "mutate", status: "na" },
};

describe("evaluateSample", () => {
  it.each(GUARDRAIL_CATALOG)("gives the expected result for the default $type sample", ({ type }) => {
    const { operation, status } = expected[type];
    const result = evaluateSample({ ...configured, type, operation }, samples[type] ?? "");

    expect(result.status).toBe(status);
  });

  it("maps a violation to the enforcement mode", () => {
    const form = { ...configured, type: "keyword_blocklist" as const };
    const text = "please ignore previous instructions";

    expect(evaluateSample({ ...form, enforcement: "enforce" }, text).status).toBe("block");
    expect(evaluateSample({ ...form, enforcement: "enforce_ignore_error" }, text).status).toBe("blockNoLog");
    expect(evaluateSample({ ...form, enforcement: "audit" }, text).status).toBe("flag");
    expect(evaluateSample(form, "hello").status).toBe("pass");
  });

  it("returns the rewritten text for mutations", () => {
    const result = evaluateSample({ ...configured, type: "secrets_detection", operation: "mutate", redactionText: "[KEY]" }, "key sk-proj-8fJ2kLmQ9xT4vB7nR1 ok");

    expect(result).toMatchObject({ status: "mutate", output: "key [KEY] ok", detail: { key: "found" } });
  });

  it("detects PII written with Persian digits", () => {
    const result = evaluateSample({ ...configured, type: "basic_pii", operation: "validate" }, "شماره من ۰۹۱۲۳۴۵۶۷۸۹ است");

    expect(result.status).toBe("block");
    expect(result.detail.items).toEqual(["09123456789"]);
  });

  it("honors a leading (?i) in patterns", () => {
    const result = evaluateSample({ ...configured, type: "regex", operation: "validate", patterns: "(?i)secret" }, "A SECRET value");

    expect(result.status).toBe("block");
  });

  it("does not throw on an invalid pattern", () => {
    expect(() => evaluateSample({ ...configured, type: "regex", patterns: "([a-z" }, "text")).not.toThrow();
    expect(evaluateSample({ ...configured, type: "regex", patterns: "([a-z" }, "text").detail.key).toBe("patternInvalid");
    expect(evaluateSample({ ...configured, type: "regex_replace", pattern: "(" }, "text").detail.key).toBe("patternInvalid");
  });

  it("reports JSON key presence", () => {
    const result = evaluateSample({ ...configured, type: "json_keys", keyMode: "any" }, '{"answer": 1}');

    expect(result).toMatchObject({ status: "pass", detail: { key: "keysPresent", count: 1, total: 2 } });
  });
});

describe("normalizeDigits", () => {
  it("converts Persian and Arabic-Indic digits", () => {
    expect(normalizeDigits("۱۲۳ ٤٥٦")).toBe("123 456");
  });
});
