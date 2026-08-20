import { z } from "zod";

import {
  chatbotAnalyticsSchema,
  chatbotListSchema,
  chatbotPublicationSchema,
  chatbotSessionTokenSchema,
  chatbotSchema,
  documentListSchema,
  knowledgeBaseDocumentSchema,
  knowledgeBaseListSchema,
  knowledgeBaseSchema,
} from "@/lib/contracts/managed-rag";
import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const base = "/api/managed-rag";

const gatewayModelListSchema = z.object({
  data: z.array(z.object({
    id: z.string().min(1),
    supported_modes: z.array(z.string()).optional(),
    types: z.array(z.string()).optional(),
  })).optional(),
});

export async function listAnswerModels() {
  const payload = await requestParsedJSON(gatewayModelListSchema, "/api/models");
  return (payload.data ?? [])
    .filter((model) => {
      const modes = [...(model.supported_modes ?? []), ...(model.types ?? [])];
      return modes.length === 0 || modes.some((mode) => ["chat", "chat_completions", "responses"].includes(mode));
    })
    .map((model) => model.id);
}

export function listKnowledgeBases() {
  return requestParsedJSON(knowledgeBaseListSchema, `${base}/knowledge-bases?limit=100`);
}

export function createKnowledgeBase(input: { name: string; description: string }) {
  return requestParsedJSON(knowledgeBaseSchema, `${base}/knowledge-bases`, {
    method: "POST", body: JSON.stringify(input),
  });
}

export function listDocuments(knowledgeBaseID: string) {
  return requestParsedJSON(documentListSchema, `${base}/knowledge-bases/${encodeURIComponent(knowledgeBaseID)}/documents?limit=100`);
}

export function createTextDocument(knowledgeBaseID: string, input: { title: string; text: string }) {
  return requestParsedJSON(knowledgeBaseDocumentSchema, `${base}/knowledge-bases/${encodeURIComponent(knowledgeBaseID)}/documents`, {
    method: "POST", body: JSON.stringify(input),
  });
}

export function uploadDocument(knowledgeBaseID: string, file: File) {
  const form = new FormData();
  form.set("file", file);
  return requestParsedJSON(knowledgeBaseDocumentSchema, `${base}/knowledge-bases/${encodeURIComponent(knowledgeBaseID)}/documents`, {
    method: "POST", body: form,
  });
}

export function listChatbots() {
  return requestParsedJSON(chatbotListSchema, `${base}/chatbots?limit=100`);
}

export function getChatbot(id: string) {
  return requestParsedJSON(chatbotSchema, `${base}/chatbots/${encodeURIComponent(id)}`);
}

export type ChatbotInput = {
  name: string;
  description: string;
  system_instruction: string;
  answer_model: string;
  tone: string;
  default_locale: "fa-IR" | "en-US";
  answer_language: "query" | "fa" | "en";
  access_mode: "private" | "public";
  retention_days: 0 | 7 | 30;
  knowledge_base_ids: string[];
  allowed_origins: string[];
  branding: Record<string, unknown>;
};

export function createChatbot(input: ChatbotInput) {
  return requestParsedJSON(chatbotSchema, `${base}/chatbots`, {
    method: "POST", body: JSON.stringify(input),
  });
}

export function publishChatbot(id: string) {
  return requestParsedJSON(chatbotPublicationSchema, `${base}/chatbots/${encodeURIComponent(id)}/publish`, {
    method: "POST", body: JSON.stringify({}),
  });
}

export async function unpublishChatbot(id: string) {
  await requestJSON(`${base}/chatbots/${encodeURIComponent(id)}/unpublish`, {
    method: "POST", body: JSON.stringify({}),
  });
}

export function getChatbotAnalytics(id: string) {
  return requestParsedJSON(chatbotAnalyticsSchema, `${base}/chatbots/${encodeURIComponent(id)}/analytics`);
}

export function createChatbotSessionToken(id: string) {
  return requestParsedJSON(chatbotSessionTokenSchema, `${base}/chatbots/${encodeURIComponent(id)}/session-tokens`, {
    method: "POST", body: JSON.stringify({ ttl_seconds: 3600 }),
  });
}
