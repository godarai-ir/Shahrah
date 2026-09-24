import { operationForType } from "./catalog";
import type { GuardrailFormInput, GuardrailPayload, GuardrailType, JwtVerifier, TenantGuardrail } from "./types";

/** A form whose type has not been picked yet, as the create wizard starts. */
export type GuardrailDraft = Omit<GuardrailFormInput, "type"> & { type: GuardrailType | null };

/** A translator bound to the `guardrails` namespace. */
export type GuardrailTranslator = (key: string, values?: Record<string, string | number>) => string;

export type WizardStep = 0 | 1 | 2 | 3;

export const WIZARD_STEPS = ["type", "config", "behavior", "review"] as const;

export const JWT_VERIFIERS: JwtVerifier[] = ["jwks", "inline", "shared", "introspection"];

const jwtVerifierField = {
  jwks: "jwksUri",
  inline: "inlineJwks",
  shared: "sharedSecret",
  introspection: "introspectionUrl",
} as const satisfies Record<JwtVerifier, keyof GuardrailFormInput>;

const SLUG_PATTERN = /^[a-zA-Z0-9][a-zA-Z0-9_-]{1,95}$/;

export const emptyForm: GuardrailFormInput = {
  slug: "",
  name: "",
  description: "",
  type: "keyword_blocklist",
  operation: "validate",
  enforcement: "enforce",
  enabled: true,
  terms: "",
  patterns: "",
  pattern: "",
  replacement: "[***]",
  redactionText: "[***]",
  min: "",
  max: "",
  schema: '{\n  "type": "object"\n}',
  keys: "",
  keyMode: "all",
  languages: "",
  codeMode: "must_contain",
  jwtVerifier: "jwks",
  tokenRegex: "",
  issuer: "",
  audience: "",
  sharedSecret: "",
  inlineJwks: "",
  jwksUri: "",
  introspectionUrl: "",
  introspectionAuthHeader: "",
  maxBytes: "1048576",
  allowedHosts: "",
  timeoutMs: "3000",
  allowPrivateNetworks: false,
};

export function lines(value: string) {
  return value
    .split(/\r?\n/)
    .map((line) => line.trim())
    .filter(Boolean);
}

function numberOrUndefined(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);
  return Number.isFinite(parsed) ? parsed : undefined;
}

export function parseJSONOrEmpty(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return undefined;
  return JSON.parse(trimmed) as unknown;
}

function stringFromConfig(config: Record<string, unknown>, key: string, fallback = "") {
  const value = config[key];
  return typeof value === "string" ? value : fallback;
}

function numberStringFromConfig(config: Record<string, unknown>, key: string) {
  const value = config[key];
  return typeof value === "number" ? String(value) : "";
}

function listStringFromConfig(config: Record<string, unknown>, key: string) {
  const value = config[key];
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string").join("\n") : "";
}

function jsonStringFromConfig(config: Record<string, unknown>, key: string, fallback: string) {
  const value = config[key];
  if (!value || typeof value !== "object") return fallback;
  return JSON.stringify(value, null, 2);
}

/** The verifier of a saved JWT guardrail: the first method whose field is filled. */
function jwtVerifierFromConfig(config: Record<string, unknown>): JwtVerifier {
  const configKey: Record<JwtVerifier, string> = { jwks: "jwks_uri", inline: "inline_jwks", shared: "shared_secret", introspection: "introspection_url" };

  return JWT_VERIFIERS.find((verifier) => {
    const value = config[configKey[verifier]];
    return typeof value === "string" ? value.trim() !== "" : Boolean(value && typeof value === "object");
  }) ?? "jwks";
}

export function formFromGuardrail(item: TenantGuardrail): GuardrailFormInput {
  return {
    slug: item.slug,
    name: item.name,
    description: item.description ?? "",
    type: item.type,
    operation: item.operation,
    enforcement: item.enforcement,
    enabled: item.enabled,
    terms: listStringFromConfig(item.config, "terms"),
    patterns: listStringFromConfig(item.config, "patterns"),
    pattern: stringFromConfig(item.config, "pattern"),
    replacement: stringFromConfig(item.config, "replacement", "[***]"),
    redactionText: typeof item.config.redaction_text === "string" ? item.config.redaction_text : "[***]",
    min: numberStringFromConfig(item.config, "min"),
    max: numberStringFromConfig(item.config, "max"),
    schema: jsonStringFromConfig(item.config, "schema", '{\n  "type": "object"\n}'),
    keys: listStringFromConfig(item.config, "keys"),
    keyMode: item.config.mode === "any" || item.config.mode === "none" ? item.config.mode : "all",
    languages: listStringFromConfig(item.config, "languages"),
    codeMode: item.config.mode === "must_not_contain" ? "must_not_contain" : "must_contain",
    jwtVerifier: jwtVerifierFromConfig(item.config),
    tokenRegex: stringFromConfig(item.config, "token_regex"),
    issuer: stringFromConfig(item.config, "issuer"),
    audience: stringFromConfig(item.config, "audience"),
    sharedSecret: stringFromConfig(item.config, "shared_secret"),
    inlineJwks: jsonStringFromConfig(item.config, "inline_jwks", ""),
    jwksUri: stringFromConfig(item.config, "jwks_uri"),
    introspectionUrl: stringFromConfig(item.config, "introspection_url"),
    introspectionAuthHeader: stringFromConfig(item.config, "introspection_auth_header"),
    maxBytes: numberStringFromConfig(item.config, "max_bytes") || "1048576",
    allowedHosts: listStringFromConfig(item.config, "allowed_hosts"),
    timeoutMs: numberStringFromConfig(item.config, "timeout_ms") || "3000",
    allowPrivateNetworks: item.config.allow_private_networks === true,
  };
}

export function buildPayload(form: GuardrailFormInput, mode: "create" | "edit"): GuardrailPayload {
  const type = form.type;
  const config: Record<string, unknown> = {};
  if (type === "keyword_blocklist") {
    config.terms = lines(form.terms);
  } else if (type === "regex") {
    config.patterns = lines(form.patterns);
    config.redaction_text = form.redactionText.trim() || "[***]";
  } else if (type === "regex_replace") {
    config.pattern = form.pattern.trim();
    config.replacement = form.replacement;
    config.case_sensitive = true;
  } else if (type === "sentence_count" || type === "word_count" || type === "character_count") {
    const min = numberOrUndefined(form.min);
    const max = numberOrUndefined(form.max);
    if (min !== undefined) config.min = min;
    if (max !== undefined) config.max = max;
  } else if (type === "json_schema") {
    config.schema = parseJSONOrEmpty(form.schema);
  } else if (type === "json_keys") {
    config.keys = lines(form.keys);
    config.mode = form.keyMode;
  } else if (type === "contains_code") {
    config.languages = lines(form.languages);
    config.mode = form.codeMode;
    config.must_contain = form.codeMode === "must_contain";
  } else if (type === "jwt_validation") {
    if (form.tokenRegex.trim()) config.token_regex = form.tokenRegex.trim();
    if (form.issuer.trim()) config.issuer = form.issuer.trim();
    if (form.audience.trim()) config.audience = form.audience.trim();
    // Only the chosen verifier is sent, so a value left over from another method never reaches the server.
    if (form.jwtVerifier === "shared" && form.sharedSecret.trim()) config.shared_secret = form.sharedSecret.trim();
    if (form.jwtVerifier === "inline" && form.inlineJwks.trim()) config.inline_jwks = parseJSONOrEmpty(form.inlineJwks);
    if (form.jwtVerifier === "jwks" && form.jwksUri.trim()) config.jwks_uri = form.jwksUri.trim();
    if (form.jwtVerifier === "introspection" && form.introspectionUrl.trim()) config.introspection_url = form.introspectionUrl.trim();
    if (form.jwtVerifier === "introspection" && form.introspectionAuthHeader.trim()) config.introspection_auth_header = form.introspectionAuthHeader.trim();
  } else if (type === "inline_image_urls") {
    const maxBytes = numberOrUndefined(form.maxBytes);
    const timeoutMs = numberOrUndefined(form.timeoutMs);
    if (maxBytes !== undefined) config.max_bytes = maxBytes;
    if (timeoutMs !== undefined) config.timeout_ms = timeoutMs;
    config.allowed_hosts = lines(form.allowedHosts);
    config.allow_private_networks = form.allowPrivateNetworks;
  } else if (type !== "not_null") {
    config.redaction_text = form.redactionText.trim() || "[***]";
  }

  return {
    ...(mode === "create" ? { slug: form.slug.trim() } : {}),
    name: form.name.trim(),
    description: form.description.trim(),
    type,
    operation: operationForType(type, form.operation),
    enforcement: form.enforcement,
    enabled: form.enabled,
    config,
  };
}

/**
 * Compiles a gateway (RE2-style) pattern for browser use. A leading inline flag group such as
 * `(?i)` becomes a JavaScript flag and `(?P<name>` becomes `(?<name>`. Returns null when the
 * pattern cannot be compiled.
 */
export function compilePattern(pattern: string, flags = "g"): RegExp | null {
  let source = pattern;
  let compiledFlags = flags;
  const inlineFlags = /^\(\?([ims]+)\)/.exec(source);
  if (inlineFlags) {
    source = source.slice(inlineFlags[0].length);
    for (const flag of inlineFlags[1]) {
      if (!compiledFlags.includes(flag)) compiledFlags += flag;
    }
  }
  source = source.replaceAll("(?P<", "(?<");

  try {
    return new RegExp(source, compiledFlags);
  } catch {
    return null;
  }
}

function isValidJSON(value: string) {
  try {
    parseJSONOrEmpty(value);
    return true;
  } catch {
    return false;
  }
}

export function isValidSlug(slug: string) {
  return SLUG_PATTERN.test(slug.trim());
}

/** Derives a slug from a Latin name, or returns null when the name has non-ASCII characters. */
export function slugFromName(name: string): string | null {
  const trimmed = name.trim();
  if (!trimmed || !/^[\x20-\x7e]+$/.test(trimmed)) return null;

  return trimmed.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 96);
}

export function slugFromType(type: GuardrailType) {
  return type.replaceAll("_", "-");
}

export function jwtVerifierFieldName(verifier: JwtVerifier) {
  return jwtVerifierField[verifier];
}

export function jwtVerifierValue(form: GuardrailFormInput | GuardrailDraft) {
  return form[jwtVerifierField[form.jwtVerifier]];
}

function validateConfig(form: GuardrailDraft, t: GuardrailTranslator) {
  const type = form.type;
  if (type === "keyword_blocklist" && lines(form.terms).length === 0) return t("validation.terms");
  if (type === "regex") {
    const patterns = lines(form.patterns);
    if (patterns.length === 0) return t("validation.patterns");
    if (patterns.some((pattern) => !compilePattern(pattern))) return t("validation.patternInvalid");
  }
  if (type === "regex_replace") {
    if (!form.pattern.trim()) return t("validation.pattern");
    if (!compilePattern(form.pattern.trim())) return t("validation.patternInvalid");
  }
  if (type === "sentence_count" || type === "word_count" || type === "character_count") {
    const min = numberOrUndefined(form.min);
    const max = numberOrUndefined(form.max);
    if (min === undefined && max === undefined) return t("validation.range");
    if (min !== undefined && max !== undefined && min > max) return t("validation.rangeOrder");
  }
  if (type === "json_schema" && !isValidJSON(form.schema)) return t("validation.json");
  if (type === "json_keys" && lines(form.keys).length === 0) return t("validation.keys");
  if (type === "jwt_validation") {
    if (!jwtVerifierValue(form).trim()) return t("validation.jwtVerifier");
    if (form.jwtVerifier === "inline" && !isValidJSON(form.inlineJwks)) return t("validation.json");
  }

  return "";
}

export type StepValidationContext = {
  mode?: "create" | "edit";
  existingSlugs?: readonly string[];
};

/** The first error of one wizard step, or an empty string. Each step reports only its own fields. */
export function validateStep(step: WizardStep, form: GuardrailDraft, t: GuardrailTranslator, context: StepValidationContext = {}) {
  if (step === 0) return form.type ? "" : t("validation.type");
  if (step === 1) return form.type ? validateConfig(form, t) : "";
  if (step === 2) return "";

  if (!form.name.trim()) return t("validation.name");
  if (!isValidSlug(form.slug)) return t("validation.slug");
  if (context.mode !== "edit" && context.existingSlugs?.includes(form.slug.trim())) return t("messages.duplicate");

  return "";
}

/** The first step with an error, or null when every step is valid. */
export function firstInvalidStep(form: GuardrailDraft, t: GuardrailTranslator, context: StepValidationContext = {}): WizardStep | null {
  const steps: WizardStep[] = [0, 1, 2, 3];

  return steps.find((step) => validateStep(step, form, t, context)) ?? null;
}
