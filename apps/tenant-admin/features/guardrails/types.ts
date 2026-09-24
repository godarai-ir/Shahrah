import { z } from "zod";

export const guardrailTypeSchema = z.enum([
  "keyword_blocklist", "regex", "secrets_detection", "basic_pii", "jwt_validation",
  "regex_replace", "sentence_count", "word_count", "character_count", "json_schema",
  "json_keys", "contains_code", "not_null", "inline_image_urls", "semantic_policy",
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

/** Which message a semantic policy judges: the user's request or the model's reply. */
export type SemanticTarget = "input" | "output";

/** A semantic policy starts from the built-in pack or from the tenant's own rules. UI-only; only one is sent. */
export type SemanticSource = "pack" | "custom";

/** Threshold fields stay strings while editing, like every other numeric form field. */
export type SemanticThresholdInput = { review: string; block: string };

export type SemanticPackRuleInput = SemanticThresholdInput & { key: string };

export type SemanticRuleInput = SemanticThresholdInput & {
  /** A stable client-side identity for list rendering; never sent. */
  id: string;
  key: string;
  statement: string;
  trueWhen: string;
  falseWhen: string;
  examplesTrue: string;
  examplesFalse: string;
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
  semanticSource: SemanticSource;
  semanticTarget: SemanticTarget;
  semanticPackRules: SemanticPackRuleInput[];
  semanticRules: SemanticRuleInput[];
  semanticFinglish: boolean;
  semanticFinglishThresholds: SemanticThresholdInput;
  semanticSeverity: boolean;
  semanticSeverityBlockAt: string;
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

export const semanticDecisionSchema = z.enum(["pass", "review", "block"]);

export const semanticPolicyTestResultSchema = z.object({
  decision: semanticDecisionSchema,
  language: z.enum(["persian", "finglish", "english"]),
  rules: z.array(z.object({ key: z.string(), probability: z.number(), decision: semanticDecisionSchema })),
  severity: z.number().optional(),
  latency_ms: z.number(),
  truncated: z.boolean(),
});

export type SemanticDecision = z.infer<typeof semanticDecisionSchema>;
export type SemanticPolicyTestResult = z.infer<typeof semanticPolicyTestResultSchema>;
