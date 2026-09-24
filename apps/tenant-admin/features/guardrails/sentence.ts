import { catalogEntry, operationForType } from "./catalog";
import { type GuardrailTranslator, lines } from "./form";
import type { GuardrailFormInput } from "./types";

export type SentenceFormat = {
  t: GuardrailTranslator;
  formatNumber: (value: number) => string;
};

function countValue(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);

  return Number.isFinite(parsed) ? parsed : undefined;
}

function conditionPhrase(form: GuardrailFormInput, { t, formatNumber }: SentenceFormat) {
  const type = form.type;
  if (type === "keyword_blocklist" || type === "regex") {
    const count = lines(type === "regex" ? form.patterns : form.terms).length;
    return t(`sentence.cond.${type}`, { n: count, count: formatNumber(count) });
  }
  if (type === "sentence_count" || type === "word_count" || type === "character_count") {
    const unit = t(`sentence.units.${catalogEntry(type).unit}`);
    const min = countValue(form.min);
    const max = countValue(form.max);
    if (min !== undefined && max !== undefined) return t("sentence.cond.countRange", { unit, min: formatNumber(min), max: formatNumber(max) });
    if (max !== undefined) return t("sentence.cond.countMax", { unit, max: formatNumber(max) });
    if (min !== undefined) return t("sentence.cond.countMin", { unit, min: formatNumber(min) });
    return t("sentence.cond.countAny", { unit });
  }
  if (type === "json_keys") {
    const keys = lines(form.keys);
    const keyList = keys.length > 0
      ? t("sentence.keyList", { keys: keys.join(t("sentence.listSeparator")) })
      : t("sentence.keysFallback");
    return t(`sentence.cond.json_keys.${form.keyMode}`, { keys: keyList });
  }
  if (type === "contains_code") return t(`sentence.cond.contains_code.${form.codeMode}`);

  return t(`sentence.cond.${type}`);
}

function actionPhrase(form: GuardrailFormInput, t: GuardrailTranslator) {
  if (form.type === "regex_replace") return t("sentence.action.regex_replace", { text: form.replacement });
  if (form.type === "inline_image_urls") return t("sentence.action.inline_image_urls");
  if (operationForType(form.type, form.operation) === "mutate") {
    return t("sentence.action.mutate", { text: form.redactionText.trim() || "[***]" });
  }

  return t(`sentence.action.${form.enforcement}`);
}

/**
 * The one-sentence description of a guardrail — condition plus action — shared by the wizard
 * summary, the success page, and the delete confirmation, e.g. «اگر متن شامل یکی از ۳ عبارت
 * مسدود باشد، درخواست متوقف و در لاگ ثبت می‌شود.»
 */
export function guardrailSentence(form: GuardrailFormInput, format: SentenceFormat): string {
  const { t } = format;
  const condition = conditionPhrase(form, format);
  const action = actionPhrase(form, t);
  // Replacement types describe the matched part itself, so the action continues the condition directly.
  const direct = form.type === "regex_replace" || form.type === "inline_image_urls";
  const sentence = t(direct ? "sentence.direct" : "sentence.full", { condition, action });

  return form.enabled ? sentence : t("sentence.disabledSuffix", { sentence });
}
