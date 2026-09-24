import type { GuardrailFormInput, GuardrailPayload, TenantGuardrail } from "./types";
import { operationForType } from "./catalog";

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
    if (form.sharedSecret.trim()) config.shared_secret = form.sharedSecret.trim();
    if (form.inlineJwks.trim()) config.inline_jwks = parseJSONOrEmpty(form.inlineJwks);
    if (form.jwksUri.trim()) config.jwks_uri = form.jwksUri.trim();
    if (form.introspectionUrl.trim()) config.introspection_url = form.introspectionUrl.trim();
    if (form.introspectionAuthHeader.trim()) config.introspection_auth_header = form.introspectionAuthHeader.trim();
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
