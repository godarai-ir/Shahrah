import {
  promptEnvelopeSchema,
  promptListResponseSchema,
  promptVersionListResponseSchema,
  tenantAdminUserLookupResponseSchema,
} from "@/lib/contracts/prompts";
import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

export function listPrompts(query: string, offset: number, limit: number) {
  const params = new URLSearchParams({ query, offset: String(offset), limit: String(limit) });
  return requestParsedJSON(promptListResponseSchema, `/api/prompts?${params.toString()}`);
}

export async function getPrompt(id: string) {
  const payload = await requestParsedJSON(promptEnvelopeSchema, `/api/prompts/${encodeURIComponent(id)}`);
  return "data" in payload ? payload.data : payload;
}

export function listPromptVersions(id: string, limit = 100) {
  return requestParsedJSON(
    promptVersionListResponseSchema,
    `/api/prompts/${encodeURIComponent(id)}/versions?limit=${limit}`,
  );
}

export function listPromptUsers() {
  return requestParsedJSON(tenantAdminUserLookupResponseSchema, "/api/gateway/lookups/users");
}

export async function deletePrompt(id: string): Promise<void> {
  await requestJSON(`/api/prompts/${encodeURIComponent(id)}`, { method: "DELETE" });
}
