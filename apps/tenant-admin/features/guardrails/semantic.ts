import type {
  GuardrailFormInput,
  SemanticPackRuleInput,
  SemanticRuleInput,
  SemanticThresholdInput,
} from "./types";

/** The built-in pack. Its wording lives on the gateway; the form only carries threshold overrides. */
export const SECURITY_BASELINE_PACK = "security-baseline@1";
export const SEMANTIC_PACK_RULE_KEYS = ["jailbreak", "system_prompt_extraction", "evaluator_address"] as const;

export const MAX_SEMANTIC_RULES = 20;
export const MAX_SEMANTIC_STATEMENT = 500;
export const MAX_SEMANTIC_CRITERIA = 500;
export const MAX_SEMANTIC_EXAMPLES = 5;
export const MAX_SEMANTIC_EXAMPLE = 300;
export const MAX_SEMANTIC_SEVERITY = 3;

const RULE_KEY_PATTERN = /^[a-z][a-z0-9_]{1,47}$/;

export const DEFAULT_SEMANTIC_THRESHOLDS: SemanticThresholdInput = { review: "0.35", block: "0.7" };
/** Phase 0 showed Finglish scores run lower than Persian for the same meaning, so it gets lower cut points. */
export const DEFAULT_FINGLISH_THRESHOLDS: SemanticThresholdInput = { review: "0.3", block: "0.5" };
export const DEFAULT_SEVERITY_BLOCK_AT = "2";

/** TypeSafe list price per million input tokens; output tokens are free. */
const USD_PER_MILLION_TOKENS = 0.042;
const BASE_TOKENS_PER_CHECK = 400;
const TOKENS_PER_RULE = 120;
const AVERAGE_MESSAGE_TOKENS = 300;

let ruleSequence = 0;

export function newSemanticRule(key: string): SemanticRuleInput {
  ruleSequence += 1;

  return {
    id: `semantic-rule-${ruleSequence}`,
    key,
    statement: "",
    trueWhen: "",
    falseWhen: "",
    examplesTrue: "",
    examplesFalse: "",
    ...DEFAULT_SEMANTIC_THRESHOLDS,
  };
}

export function defaultPackRules(): SemanticPackRuleInput[] {
  return SEMANTIC_PACK_RULE_KEYS.map((key) => ({ key, ...DEFAULT_SEMANTIC_THRESHOLDS }));
}

/** The next unused `rule_N` key, so a new rule is valid before the user renames it. */
export function nextRuleKey(rules: readonly SemanticRuleInput[]) {
  const used = new Set(rules.map((rule) => rule.key));
  let index = rules.length + 1;
  while (used.has(`rule_${index}`)) index += 1;

  return `rule_${index}`;
}

function splitLines(value: string) {
  return value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);
}

function toNumber(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);

  return Number.isFinite(parsed) ? parsed : undefined;
}

function thresholdString(value: unknown, fallback: string) {
  return typeof value === "number" && Number.isFinite(value) ? String(value) : fallback;
}

function stringValue(value: unknown) {
  return typeof value === "string" ? value : "";
}

function stringList(value: unknown) {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string").join("\n") : "";
}

function record(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? (value as Record<string, unknown>) : {};
}

type SemanticFormFields = Pick<
  GuardrailFormInput,
  | "semanticSource"
  | "semanticTarget"
  | "semanticPackRules"
  | "semanticRules"
  | "semanticFinglish"
  | "semanticFinglishThresholds"
  | "semanticSeverity"
  | "semanticSeverityBlockAt"
>;

export function emptySemanticFields(): SemanticFormFields {
  return {
    semanticSource: "pack",
    semanticTarget: "input",
    semanticPackRules: defaultPackRules(),
    semanticRules: [newSemanticRule("rule_1")],
    semanticFinglish: true,
    semanticFinglishThresholds: { ...DEFAULT_FINGLISH_THRESHOLDS },
    semanticSeverity: false,
    semanticSeverityBlockAt: DEFAULT_SEVERITY_BLOCK_AT,
  };
}

export function semanticFieldsFromConfig(config: Record<string, unknown>): SemanticFormFields {
  const fields = emptySemanticFields();
  const configuredRules = Array.isArray(config.rules) ? config.rules.map(record) : [];
  const severity = record(config.severity);
  const finglish = config.finglish_thresholds ? record(config.finglish_thresholds) : null;

  fields.semanticTarget = config.target === "output" ? "output" : "input";
  fields.semanticSeverity = severity.enabled === true;
  fields.semanticSeverityBlockAt = thresholdString(severity.block_at, DEFAULT_SEVERITY_BLOCK_AT);
  fields.semanticFinglish = finglish !== null;
  if (finglish) {
    fields.semanticFinglishThresholds = {
      review: thresholdString(finglish.review, DEFAULT_FINGLISH_THRESHOLDS.review),
      block: thresholdString(finglish.block, DEFAULT_FINGLISH_THRESHOLDS.block),
    };
  }

  if (typeof config.pack === "string" && config.pack) {
    const overrides = new Map(configuredRules.map((rule) => [stringValue(rule.key), rule]));
    fields.semanticSource = "pack";
    fields.semanticPackRules = SEMANTIC_PACK_RULE_KEYS.map((key) => {
      const override = overrides.get(key) ?? {};

      return {
        key,
        review: thresholdString(override.review_threshold, DEFAULT_SEMANTIC_THRESHOLDS.review),
        block: thresholdString(override.block_threshold, DEFAULT_SEMANTIC_THRESHOLDS.block),
      };
    });

    return fields;
  }

  fields.semanticSource = "custom";
  if (configuredRules.length > 0) {
    fields.semanticRules = configuredRules.map((rule) => {
      const examples = record(rule.examples);

      return {
        ...newSemanticRule(stringValue(rule.key)),
        statement: stringValue(rule.statement),
        trueWhen: stringValue(rule.true_when),
        falseWhen: stringValue(rule.false_when),
        examplesTrue: stringList(examples.true),
        examplesFalse: stringList(examples.false),
        review: thresholdString(rule.review_threshold, DEFAULT_SEMANTIC_THRESHOLDS.review),
        block: thresholdString(rule.block_threshold, DEFAULT_SEMANTIC_THRESHOLDS.block),
      };
    });
  }

  return fields;
}

function thresholdConfig({ review, block }: SemanticThresholdInput) {
  return { review_threshold: toNumber(review), block_threshold: toNumber(block) };
}

/** The config the gateway validates with ParseSemanticPolicyConfig. The unused source's fields are never sent. */
export function semanticConfig(form: SemanticFormFields): Record<string, unknown> {
  const config: Record<string, unknown> = {};

  if (form.semanticSource === "pack") {
    config.pack = SECURITY_BASELINE_PACK;
    config.target = "input";
    config.rules = form.semanticPackRules.map((rule) => ({ key: rule.key, ...thresholdConfig(rule) }));
  } else {
    config.target = form.semanticTarget;
    config.rules = form.semanticRules.map((rule) => {
      const examplesTrue = splitLines(rule.examplesTrue);
      const examplesFalse = splitLines(rule.examplesFalse);

      return {
        key: rule.key.trim(),
        statement: rule.statement.trim(),
        ...(rule.trueWhen.trim() ? { true_when: rule.trueWhen.trim() } : {}),
        ...(rule.falseWhen.trim() ? { false_when: rule.falseWhen.trim() } : {}),
        ...(examplesTrue.length > 0 || examplesFalse.length > 0 ? { examples: { true: examplesTrue, false: examplesFalse } } : {}),
        ...thresholdConfig(rule),
      };
    });
  }

  if (form.semanticFinglish) {
    config.finglish_thresholds = {
      review: toNumber(form.semanticFinglishThresholds.review),
      block: toNumber(form.semanticFinglishThresholds.block),
    };
  }

  if (form.semanticSeverity) {
    config.severity = { enabled: true, block_at: toNumber(form.semanticSeverityBlockAt) };
  }

  return config;
}

export type SemanticValidationKey =
  | "semanticRules"
  | "semanticRuleKey"
  | "semanticRuleKeyDuplicate"
  | "semanticStatement"
  | "semanticCriteria"
  | "semanticExamples"
  | "semanticThresholds"
  | "semanticSeverity";

export function validThresholds({ review, block }: SemanticThresholdInput) {
  const reviewValue = toNumber(review);
  const blockValue = toNumber(block);

  return reviewValue !== undefined && blockValue !== undefined && reviewValue >= 0 && blockValue <= 1 && reviewValue <= blockValue;
}

function examplesInvalid(value: string) {
  const examples = splitLines(value);

  return examples.length > MAX_SEMANTIC_EXAMPLES || examples.some((example) => [...example].length > MAX_SEMANTIC_EXAMPLE);
}

/** Rule-content errors belong to the settings step. */
export function semanticConfigError(form: SemanticFormFields): SemanticValidationKey | null {
  if (form.semanticSource === "pack") return null;
  if (form.semanticRules.length === 0 || form.semanticRules.length > MAX_SEMANTIC_RULES) return "semanticRules";

  const keys = new Set<string>();
  for (const rule of form.semanticRules) {
    const key = rule.key.trim();
    if (!RULE_KEY_PATTERN.test(key)) return "semanticRuleKey";
    if (keys.has(key)) return "semanticRuleKeyDuplicate";
    keys.add(key);

    const statement = [...rule.statement.trim()].length;
    if (statement === 0 || statement > MAX_SEMANTIC_STATEMENT) return "semanticStatement";
    if ([...rule.trueWhen.trim()].length > MAX_SEMANTIC_CRITERIA || [...rule.falseWhen.trim()].length > MAX_SEMANTIC_CRITERIA) return "semanticCriteria";
    if (examplesInvalid(rule.examplesTrue) || examplesInvalid(rule.examplesFalse)) return "semanticExamples";
  }

  return null;
}

/** Threshold and severity errors belong to the behavior step, where those fields live. */
export function semanticBehaviorError(form: SemanticFormFields): SemanticValidationKey | null {
  const rules = form.semanticSource === "pack" ? form.semanticPackRules : form.semanticRules;
  if (!rules.every(validThresholds)) return "semanticThresholds";
  if (form.semanticFinglish && !validThresholds(form.semanticFinglishThresholds)) return "semanticThresholds";

  if (form.semanticSeverity) {
    const blockAt = toNumber(form.semanticSeverityBlockAt);
    if (blockAt === undefined || blockAt <= 0 || blockAt > MAX_SEMANTIC_SEVERITY) return "semanticSeverity";
  }

  return null;
}

export type StatementWarning = "compound" | "negation";

const COMPOUND_PATTERN = /(?:^|\s)(?:and|or|و|یا)(?:\s|$)/i;
// JavaScript's \b only knows ASCII letters, so the Persian alternatives end on whitespace instead.
const NEGATION_PATTERN = /^(?:not\b|no\b|never\b|نه\s|هیچ|بدون|نباید|نمی)/i;

/**
 * Soft hints from TypeSafe's Noul guidance: one proposition per rule, phrased so that "yes" means a violation.
 * They never block saving because a conjunction can be part of a single idea.
 */
export function statementWarnings(statement: string): StatementWarning[] {
  const trimmed = statement.trim();
  if (!trimmed) return [];
  const warnings: StatementWarning[] = [];
  if (COMPOUND_PATTERN.test(trimmed)) warnings.push("compound");
  if (NEGATION_PATTERN.test(trimmed)) warnings.push("negation");

  return warnings;
}

/** The number of rules one check asks the judge. */
export function semanticRuleCount(form: SemanticFormFields) {
  return form.semanticSource === "pack" ? form.semanticPackRules.length : form.semanticRules.length;
}

/** An approximate cost in US dollars of one thousand checks with an average-length message. */
export function semanticCostPerThousand(form: SemanticFormFields) {
  const tokens = BASE_TOKENS_PER_CHECK + TOKENS_PER_RULE * semanticRuleCount(form) + AVERAGE_MESSAGE_TOKENS + (form.semanticSeverity ? TOKENS_PER_RULE : 0);

  return (tokens * 1000 * USD_PER_MILLION_TOKENS) / 1_000_000;
}
