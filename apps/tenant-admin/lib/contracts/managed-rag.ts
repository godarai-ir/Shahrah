import { z } from "zod";

const dateSchema = z.string();

export const knowledgeBaseSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  description: z.string().default(""),
  rag_profile_revision: z.string(),
  document_count: z.number().int(),
  ready_document_count: z.number().int(),
  created_at: dateSchema,
  updated_at: dateSchema,
});

export const knowledgeBaseListSchema = z.object({
  items: z.array(knowledgeBaseSchema),
  total: z.number().int(),
  limit: z.number().int(),
  offset: z.number().int(),
});

export const documentStatusSchema = z.enum([
  "queued", "scanning", "parsing", "chunking", "embedding", "indexing", "ready", "failed", "deleting",
]);

export const knowledgeBaseDocumentSchema = z.object({
  id: z.string().uuid(),
  knowledge_base_id: z.string().uuid(),
  title: z.string(),
  source_type: z.enum(["file", "text"]),
  media_type: z.string(),
  status: documentStatusSchema,
  active_version_id: z.string().uuid().nullable().optional(),
  processing_version_id: z.string().uuid().nullable().optional(),
  size_bytes: z.number(),
  page_count: z.number().int().nullable().optional(),
  chunk_count: z.number().int(),
  detected_language: z.string().default(""),
  failure_code: z.string().optional(),
  failure_message: z.string().optional(),
  created_at: dateSchema,
  updated_at: dateSchema,
});

export const documentListSchema = z.object({
  items: z.array(knowledgeBaseDocumentSchema),
  total: z.number().int(),
  limit: z.number().int(),
  offset: z.number().int(),
});

export const chatbotSchema = z.object({
  id: z.string().uuid(),
  public_id: z.string(),
  name: z.string(),
  description: z.string().default(""),
  system_instruction: z.string().default(""),
  answer_model: z.string(),
  tone: z.string(),
  default_locale: z.enum(["fa-IR", "en-US"]),
  answer_language: z.enum(["query", "fa", "en"]),
  access_mode: z.enum(["private", "public"]),
  retention_days: z.union([z.literal(0), z.literal(7), z.literal(30)]),
  knowledge_base_ids: z.array(z.string().uuid()).max(5),
  allowed_origins: z.array(z.string()),
  branding: z.record(z.string(), z.unknown()),
  published_version_id: z.string().uuid().nullable().optional(),
  created_at: dateSchema,
  updated_at: dateSchema,
});

export const chatbotListSchema = z.object({
  items: z.array(chatbotSchema),
  total: z.number().int(),
  limit: z.number().int(),
  offset: z.number().int(),
});

export const chatbotPublicationSchema = z.object({
  version_id: z.string().uuid(),
  chatbot_id: z.string().uuid(),
  revision: z.number().int(),
  rag_profile_revision: z.string(),
  published_at: dateSchema,
});

export const chatbotAnalyticsSchema = z.object({
  query_count: z.number().int(),
  refusal_count: z.number().int(),
  citation_failure_count: z.number().int(),
  average_latency_ms: z.number(),
  provider_cost_usd: z.number(),
  platform_cost_irr: z.number(),
});

export const citationSchema = z.object({
  id: z.string(),
  document_id: z.string().uuid(),
  title: z.string(),
  page: z.number().int().nullable().optional(),
  quote: z.string(),
});

export const chatbotResponseSchema = z.object({
  request_id: z.string(),
  trace_id: z.string(),
  session_id: z.string().uuid(),
  answer: z.string(),
  citations: z.array(citationSchema),
  refusal: z.boolean(),
  actual_model: z.string(),
});

export const chatbotSessionTokenSchema = z.object({
  token: z.string().min(1),
  expires_at: dateSchema,
});

export type KnowledgeBase = z.infer<typeof knowledgeBaseSchema>;
export type KnowledgeBaseDocument = z.infer<typeof knowledgeBaseDocumentSchema>;
export type Chatbot = z.infer<typeof chatbotSchema>;
export type ChatbotAnalytics = z.infer<typeof chatbotAnalyticsSchema>;
export type Citation = z.infer<typeof citationSchema>;
export type ChatbotResponse = z.infer<typeof chatbotResponseSchema>;
export type ChatbotSessionToken = z.infer<typeof chatbotSessionTokenSchema>;
