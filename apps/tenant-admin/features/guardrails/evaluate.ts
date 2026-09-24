import { operationForType } from "./catalog";
import { compilePattern, lines } from "./form";
import type { GuardrailFormInput } from "./types";

export type EvaluationStatus = "pass" | "mutate" | "flag" | "block" | "blockNoLog" | "na";

export type EvaluationDetailKey =
  | "noTerms" | "termsFound" | "termsNone"
  | "noPatterns" | "patternInvalid" | "patternMatches" | "patternNone"
  | "replaceCount" | "replaceNone"
  | "found" | "secretsNone" | "piiNone"
  | "count"
  | "empty" | "notEmpty"
  | "codeFound" | "codeNone"
  | "jsonInvalid" | "notObject" | "missingKeys" | "schemaOk"
  | "noKeys" | "keysPresent"
  | "na";

export type EvaluationDetail = {
  key: EvaluationDetailKey;
  count?: number;
  total?: number;
  items?: string[];
};

export type Evaluation = {
  status: EvaluationStatus;
  detail: EvaluationDetail;
  /** The text that reaches the model, when the guardrail rewrites it. */
  output?: string;
};

type Check = { violated: boolean; detail: EvaluationDetail; output?: string };

const SECRET_PATTERNS = [
  /sk-[A-Za-z0-9_-]{12,}/g,
  /AKIA[0-9A-Z]{16}/g,
  /gh[pousr]_[A-Za-z0-9]{20,}/g,
  /(?:password|passwd|secret)\s*[:=]\s*\S+/gi,
];

const PII_PATTERNS = [
  /[\w.+-]+@[\w-]+\.[\w.]+/g,
  /(?:\+98|0)9\d{9}/g,
  /\b(?:\d{4}[- ]?){3}\d{4}\b/g,
];

const CODE_PATTERN = /```|\b(?:SELECT|INSERT|UPDATE|DELETE)\b|\b(?:def|function|const|import|class|return)\s|[{};]\s*$/im;

/** Converts Persian and Arabic-Indic digits to ASCII so numeric patterns match them. */
export function normalizeDigits(text: string) {
  return text
    .replace(/[۰-۹]/g, (digit) => String(digit.charCodeAt(0) - 0x06f0))
    .replace(/[٠-٩]/g, (digit) => String(digit.charCodeAt(0) - 0x0660));
}

function collectMatches(patterns: RegExp[], text: string) {
  return patterns.flatMap((pattern) => text.match(pattern) ?? []);
}

function redact(patterns: RegExp[], text: string, replacement: string) {
  return patterns.reduce((result, pattern) => result.replace(pattern, () => replacement), text);
}

function redactionText(form: GuardrailFormInput) {
  return form.redactionText.trim() || "[***]";
}

function detectPatterns(patterns: RegExp[], text: string, form: GuardrailFormInput, noneKey: EvaluationDetailKey): Check {
  const matches = collectMatches(patterns, text);

  return {
    violated: matches.length > 0,
    detail: matches.length > 0 ? { key: "found", items: matches } : { key: noneKey },
    output: redact(patterns, text, redactionText(form)),
  };
}

function checkRegex(form: GuardrailFormInput, text: string): Check {
  const sources = lines(form.patterns);
  if (sources.length === 0) return { violated: false, detail: { key: "noPatterns" } };
  const patterns = sources.map((source) => compilePattern(source));
  if (patterns.some((pattern) => !pattern)) return { violated: false, detail: { key: "patternInvalid" } };
  const compiled = patterns.filter((pattern): pattern is RegExp => pattern !== null);
  const matches = collectMatches(compiled, text);

  return {
    violated: matches.length > 0,
    detail: matches.length > 0 ? { key: "patternMatches", count: matches.length, items: matches } : { key: "patternNone" },
    output: redact(compiled, text, redactionText(form)),
  };
}

function checkRegexReplace(form: GuardrailFormInput, text: string): Check {
  if (!form.pattern.trim()) return { violated: false, detail: { key: "noPatterns" } };
  const pattern = compilePattern(form.pattern.trim());
  if (!pattern) return { violated: false, detail: { key: "patternInvalid" } };
  const matches = collectMatches([pattern], text);

  return {
    violated: matches.length > 0,
    detail: matches.length > 0 ? { key: "replaceCount", count: matches.length } : { key: "replaceNone" },
    output: text.replace(pattern, () => form.replacement),
  };
}

function checkKeywords(form: GuardrailFormInput, text: string): Check {
  const terms = lines(form.terms);
  if (terms.length === 0) return { violated: false, detail: { key: "noTerms" } };
  const lowered = text.toLowerCase();
  const hits = terms.filter((term) => lowered.includes(term.toLowerCase()));

  return { violated: hits.length > 0, detail: hits.length > 0 ? { key: "termsFound", items: hits } : { key: "termsNone" } };
}

function countUnits(type: GuardrailFormInput["type"], text: string) {
  if (type === "sentence_count") return text.split(/[.!?؟\n]+/).filter((part) => part.trim()).length;
  if (type === "word_count") return text.split(/\s+/).filter(Boolean).length;

  return [...text].length;
}

function checkCount(form: GuardrailFormInput, text: string): Check {
  const count = countUnits(form.type, text);
  const min = form.min.trim() === "" ? undefined : Number(form.min);
  const max = form.max.trim() === "" ? undefined : Number(form.max);
  const violated = (min !== undefined && count < min) || (max !== undefined && count > max);

  return { violated, detail: { key: "count", count } };
}

function parseJSONObject(text: string): { ok: true; value: unknown } | { ok: false } {
  try {
    return { ok: true, value: JSON.parse(text) as unknown };
  } catch {
    return { ok: false };
  }
}

function hasKey(value: unknown, key: string) {
  return typeof value === "object" && value !== null && !Array.isArray(value) && key in value;
}

function checkJSONSchema(form: GuardrailFormInput, text: string): Check {
  const parsed = parseJSONObject(text);
  if (!parsed.ok) return { violated: true, detail: { key: "jsonInvalid" } };
  const schema = parseJSONObject(form.schema);
  const schemaValue = schema.ok && typeof schema.value === "object" && schema.value !== null ? schema.value as { type?: unknown; required?: unknown } : {};
  const isObject = typeof parsed.value === "object" && parsed.value !== null && !Array.isArray(parsed.value);
  if (schemaValue.type === "object" && !isObject) return { violated: true, detail: { key: "notObject" } };
  const required = Array.isArray(schemaValue.required) ? schemaValue.required.filter((key): key is string => typeof key === "string") : [];
  const missing = required.filter((key) => !hasKey(parsed.value, key));
  if (missing.length > 0) return { violated: true, detail: { key: "missingKeys", items: missing } };

  return { violated: false, detail: { key: "schemaOk" } };
}

function checkJSONKeys(form: GuardrailFormInput, text: string): Check {
  const parsed = parseJSONObject(text);
  if (!parsed.ok) return { violated: true, detail: { key: "jsonInvalid" } };
  const keys = lines(form.keys);
  if (keys.length === 0) return { violated: false, detail: { key: "noKeys" } };
  const present = keys.filter((key) => hasKey(parsed.value, key)).length;
  const violated = form.keyMode === "all" ? present < keys.length : form.keyMode === "any" ? present === 0 : present > 0;

  return { violated, detail: { key: "keysPresent", count: present, total: keys.length } };
}

function checkSample(form: GuardrailFormInput, text: string): Check | null {
  switch (form.type) {
    case "keyword_blocklist":
      return checkKeywords(form, text);
    case "regex":
      return checkRegex(form, text);
    case "regex_replace":
      return checkRegexReplace(form, text);
    case "secrets_detection":
      return detectPatterns(SECRET_PATTERNS, text, form, "secretsNone");
    case "basic_pii":
      return detectPatterns(PII_PATTERNS, normalizeDigits(text), form, "piiNone");
    case "sentence_count":
    case "word_count":
    case "character_count":
      return checkCount(form, text);
    case "not_null":
      return { violated: !text.trim(), detail: { key: text.trim() ? "notEmpty" : "empty" } };
    case "contains_code": {
      const hasCode = CODE_PATTERN.test(text);
      return { violated: form.codeMode === "must_contain" ? !hasCode : hasCode, detail: { key: hasCode ? "codeFound" : "codeNone" } };
    }
    case "json_schema":
      return checkJSONSchema(form, text);
    case "json_keys":
      return checkJSONKeys(form, text);
    default:
      // JWT validation and image inlining depend on services outside the text.
      return null;
  }
}

/**
 * Runs an approximate, browser-only version of the guardrail on a sample text. It never
 * throws: an invalid pattern is reported as a detail instead.
 */
export function evaluateSample(form: GuardrailFormInput, text: string): Evaluation {
  const check = checkSample(form, text);
  if (!check) return { status: "na", detail: { key: "na" } };
  if (!check.violated) return { status: "pass", detail: check.detail };
  if (operationForType(form.type, form.operation) === "mutate") {
    return { status: "mutate", detail: check.detail, output: check.output ?? text };
  }
  if (form.enforcement === "audit") return { status: "flag", detail: check.detail };

  return { status: form.enforcement === "enforce" ? "block" : "blockNoLog", detail: check.detail };
}
