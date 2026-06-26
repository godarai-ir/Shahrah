export type GuardrailType =
  | "keyword_blocklist"
  | "regex"
  | "secrets_detection"
  | "basic_pii"
  | "jwt_validation"
  | "regex_replace"
  | "sentence_count"
  | "word_count"
  | "character_count"
  | "json_schema"
  | "json_keys"
  | "contains_code"
  | "not_null"
  | "inline_image_urls";
export type GuardrailOperation = "validate" | "mutate";
export type GuardrailEnforcement = "audit" | "enforce" | "enforce_ignore_error";

export type TenantGuardrail = {
  id: string;
  slug: string;
  selector: string;
  name: string;
  description?: string;
  type: GuardrailType;
  operation: GuardrailOperation;
  enforcement: GuardrailEnforcement;
  enabled: boolean;
  config: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type GuardrailFormInput = {
  slug: string;
  name: string;
  description: string;
  type: GuardrailType;
  operation: GuardrailOperation;
  enforcement: GuardrailEnforcement;
  enabled: boolean;
  terms: string;
  patterns: string;
  pattern: string;
  replacement: string;
  redactionText: string;
  min: string;
  max: string;
  schema: string;
  keys: string;
  keyMode: "any" | "all" | "none";
  languages: string;
  codeMode: "must_contain" | "must_not_contain";
  tokenRegex: string;
  issuer: string;
  audience: string;
  sharedSecret: string;
  inlineJwks: string;
  jwksUri: string;
  introspectionUrl: string;
  introspectionAuthHeader: string;
  maxBytes: string;
  allowedHosts: string;
  timeoutMs: string;
  allowPrivateNetworks: boolean;
};

export type GuardrailPayload = {
  slug?: string;
  name?: string;
  description?: string;
  type?: GuardrailType;
  operation?: GuardrailOperation;
  enforcement?: GuardrailEnforcement;
  enabled?: boolean;
  config?: Record<string, unknown>;
};
