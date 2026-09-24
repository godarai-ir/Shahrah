import { z } from "zod";

export const guardrailTypeSchema = z.enum([
  "keyword_blocklist", "regex", "secrets_detection", "basic_pii", "jwt_validation",
  "regex_replace", "sentence_count", "word_count", "character_count", "json_schema",
  "json_keys", "contains_code", "not_null", "inline_image_urls",
]);
export const guardrailOperationSchema = z.enum(["validate", "mutate"]);
export const guardrailEnforcementSchema = z.enum(["audit", "enforce", "enforce_ignore_error"]);

export const tenantGuardrailSchema = z.object({
  id: z.string(),
  slug: z.string(),
  selector: z.string(),
  name: z.string(),
  description: z.string().optional(),
  type: guardrailTypeSchema,
  operation: guardrailOperationSchema,
  enforcement: guardrailEnforcementSchema,
  enabled: z.boolean(),
  config: z.record(z.string(), z.unknown()),
  created_at: z.string(),
  updated_at: z.string(),
});

export type GuardrailType = z.infer<typeof guardrailTypeSchema>;
export type GuardrailOperation = z.infer<typeof guardrailOperationSchema>;
export type GuardrailEnforcement = z.infer<typeof guardrailEnforcementSchema>;
export type TenantGuardrail = z.infer<typeof tenantGuardrailSchema>;

/** How a JWT guardrail verifies tokens. A UI-only choice; only the chosen method's field is sent. */
export type JwtVerifier = "jwks" | "inline" | "shared" | "introspection";

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
  jwtVerifier: JwtVerifier;
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

export const guardrailPayloadSchema = z.object({
  slug: z.string().optional(),
  name: z.string().optional(),
  description: z.string().optional(),
  type: guardrailTypeSchema.optional(),
  operation: guardrailOperationSchema.optional(),
  enforcement: guardrailEnforcementSchema.optional(),
  enabled: z.boolean().optional(),
  config: z.record(z.string(), z.unknown()).optional(),
});

export type GuardrailPayload = z.infer<typeof guardrailPayloadSchema>;
