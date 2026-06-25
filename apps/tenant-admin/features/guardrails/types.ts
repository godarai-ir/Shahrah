export type GuardrailType = "keyword_blocklist" | "regex" | "secrets_detection" | "basic_pii";
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
  redactionText: string;
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
