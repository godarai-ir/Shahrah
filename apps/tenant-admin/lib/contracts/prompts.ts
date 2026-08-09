import { z } from "zod";

const promptMessageSchema = z.object({
  role: z.string(),
  content: z.unknown().optional(),
  tool_call_id: z.string().optional(),
});

export const promptManifestSchema = z.object({
  type: z.string(),
  name: z.string(),
  description: z.string().optional(),
  messages: z.array(promptMessageSchema),
  variables: z.record(z.string(), z.string()).optional(),
  model: z.string().optional(),
  parameters: z.record(z.string(), z.unknown()).optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  version_alias: z.string().optional(),
  model_settings: z.object({
    stream: z.boolean().optional(),
    temperature: z.number().optional(),
    top_p: z.number().optional(),
    max_tokens: z.number().optional(),
  }).optional(),
  cache: z.object({
    enabled: z.boolean(),
    type: z.enum(["simple", "smart"]).optional(),
    ttl: z.number().optional(),
    namespace: z.string().optional(),
    similarity_threshold: z.number().optional(),
  }).optional(),
  guardrails: z.object({
    input: z.array(z.string()).optional(),
    output: z.array(z.string()).optional(),
    scope: z.enum(["all", "last"]).optional(),
  }).optional(),
  structured_output: z.object({
    enabled: z.boolean(),
    format: z.enum(["json_schema", "json_object"]).optional(),
    strict: z.boolean().optional(),
    schema: z.record(z.string(), z.unknown()).optional(),
  }).optional(),
  logging: z.object({ enabled: z.boolean(), store_logs: z.boolean() }).optional(),
});

export const promptVersionSchema = z.object({
  id: z.string(),
  prompt_id: z.string(),
  version: z.number(),
  fqn: z.string(),
  version_alias: z.string().optional(),
  name: z.string(),
  description: z.string().optional(),
  messages: z.array(promptMessageSchema),
  variables: z.record(z.string(), z.string()).optional(),
  model: z.string().optional(),
  parameters: z.record(z.string(), z.unknown()).optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  manifest: promptManifestSchema.optional(),
  commit_message: z.string().optional(),
  created_by: z.string().optional(),
  created_at: z.string().optional(),
});

export const promptSchema = z.object({
  id: z.string(),
  name: z.string(),
  slug: z.string(),
  description: z.string().optional(),
  latest_version: z.number(),
  latest_version_detail: promptVersionSchema.optional(),
  created_by: z.string().optional(),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

export const promptListResponseSchema = z.object({
  data: z.array(promptSchema).optional(),
  total: z.number().optional(),
  limit: z.number().optional(),
  offset: z.number().optional(),
  next_offset: z.number().nullable().optional(),
});

export const promptVersionListResponseSchema = z.object({ data: z.array(promptVersionSchema).optional() });
export const promptEnvelopeSchema = z.union([promptSchema, z.object({ data: promptSchema })]);
export const promptVersionEnvelopeSchema = z.union([promptVersionSchema, z.object({ data: promptVersionSchema })]);

export const tenantAdminUserLookupSchema = z.object({
  id: z.string(),
  email: z.string().optional(),
  mobile: z.string().optional(),
  display_name: z.string().optional(),
});
export const tenantAdminUserLookupResponseSchema = z.object({
  data: z.array(tenantAdminUserLookupSchema).optional(),
  total: z.number().optional(),
});

export const promptRunResponseSchema = z.object({
  choices: z.array(z.object({
    message: z.object({ role: z.string().optional(), content: z.unknown().optional() }).optional(),
  })).optional(),
  usage: z.object({
    prompt_tokens: z.number().optional(),
    completion_tokens: z.number().optional(),
    total_tokens: z.number().optional(),
  }).optional(),
  model: z.string().optional(),
  id: z.string().optional(),
});

export type PromptManifest = z.infer<typeof promptManifestSchema>;
export type PromptVersion = z.infer<typeof promptVersionSchema>;
export type Prompt = z.infer<typeof promptSchema>;
export type PromptListResponse = z.infer<typeof promptListResponseSchema>;
export type PromptVersionListResponse = z.infer<typeof promptVersionListResponseSchema>;
export type TenantAdminUserLookup = z.infer<typeof tenantAdminUserLookupSchema>;
export type TenantAdminUserLookupResponse = z.infer<typeof tenantAdminUserLookupResponseSchema>;
export type PromptRunResponse = z.infer<typeof promptRunResponseSchema>;
