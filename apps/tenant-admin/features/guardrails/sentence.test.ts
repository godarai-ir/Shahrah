import { createTranslator } from "next-intl";
import { describe, expect, it } from "vitest";

import en from "@/messages/modules/guardrails/en.json";
import fa from "@/messages/modules/guardrails/fa.json";

import { GUARDRAIL_CATALOG } from "./catalog";
import { emptyForm, type GuardrailTranslator } from "./form";
import { guardrailSentence, type SentenceFormat } from "./sentence";
import type { GuardrailEnforcement, GuardrailFormInput, GuardrailOperation } from "./types";

function formatFor(locale: "fa" | "en"): SentenceFormat {
  const t = createTranslator({
    locale,
    messages: { guardrails: locale === "fa" ? fa : en },
    namespace: "guardrails",
    onError: (error) => {
      throw error;
    },
  }) as unknown as GuardrailTranslator;
  const formatter = new Intl.NumberFormat(locale === "fa" ? "fa-IR" : "en-US");

  return { t, formatNumber: (value) => formatter.format(value) };
}

const fa_ = formatFor("fa");
const en_ = formatFor("en");

const configured: GuardrailFormInput = {
  ...emptyForm,
  terms: "jailbreak\nsystem prompt\ndrop table",
  patterns: "(?i)password",
  pattern: "secret-\\d+",
  replacement: "<hidden>",
  redactionText: "[***]",
  min: "5",
  max: "200",
  keys: "answer\nconfidence",
};

const operations: GuardrailOperation[] = ["validate", "mutate"];
const enforcements: GuardrailEnforcement[] = ["enforce", "audit", "enforce_ignore_error"];
const cases = GUARDRAIL_CATALOG.flatMap(({ type }) =>
  operations.flatMap((operation) => enforcements.map((enforcement) => ({ type, operation, enforcement }))),
);

describe("guardrailSentence", () => {
  it.each(cases)("describes $type · $operation · $enforcement in both locales", ({ type, operation, enforcement }) => {
    const form = { ...configured, type, operation, enforcement };
    const faSentence = guardrailSentence(form, fa_);
    const enSentence = guardrailSentence(form, en_);

    expect({ fa: faSentence, en: enSentence }).toMatchSnapshot();
    expect(faSentence).not.toMatch(/[{}]|sentence\./);
    expect(enSentence).not.toMatch(/[{}]|sentence\./);
  });

  it("builds the keyword condition from the term count", () => {
    const form = { ...configured, type: "keyword_blocklist" as const, enforcement: "enforce" as const };

    expect(guardrailSentence(form, fa_)).toBe("اگر متن شامل یکی از ۳ عبارت مسدود باشد، درخواست متوقف و در لاگ ثبت می‌شود.");
    expect(guardrailSentence({ ...form, terms: "" }, en_)).toBe("If the text contains a blocked term, the request is stopped and logged.");
  });

  it("uses the redaction text when a flexible type rewrites content", () => {
    const form = { ...configured, type: "secrets_detection" as const, operation: "mutate" as const, redactionText: "[KEY]" };

    expect(guardrailSentence(form, en_)).toBe("If a key or secret is found in the text, the part that is found is masked with “[KEY]” and the request continues.");
  });

  it("describes one-sided count ranges", () => {
    const form = { ...configured, type: "word_count" as const, min: "", max: "40" };

    expect(guardrailSentence(form, en_)).toBe("If the word count is above 40, the request is stopped and logged.");
    expect(guardrailSentence({ ...form, min: "3", max: "" }, fa_)).toBe("اگر تعداد کلمه‌ها کمتر از ۳ باشد، درخواست متوقف و در لاگ ثبت می‌شود.");
  });

  it("marks a disabled guardrail", () => {
    const form = { ...configured, type: "not_null" as const, enabled: false };

    expect(guardrailSentence(form, fa_)).toBe("اگر متن خالی باشد، درخواست متوقف و در لاگ ثبت می‌شود. (غیرفعال ذخیره می‌شود)");
  });
});
