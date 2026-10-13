import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const personalAccessTokenSchema = z.object({
  id: z.string(),
  key_alias: z.string(),
  key_name: z.string().optional(),
  key_type: z.string().optional(),
  created_at: z.string().optional(),
  expires_at: z.string().optional(),
  last_used_at: z.string().optional(),
  revoked_at: z.string().optional(),
});

const createdPersonalAccessTokenSchema = personalAccessTokenSchema.extend({ key: z.string() });
const personalAccessTokenListSchema = z.object({
  data: z.array(personalAccessTokenSchema).optional(),
  total: z.number().optional(),
});

export type PersonalAccessToken = z.infer<typeof personalAccessTokenSchema>;

export function listPersonalAccessTokens(params: URLSearchParams) {
  return requestParsedJSON(personalAccessTokenListSchema, `/api/personal-access-tokens?${params.toString()}`);
}

export function createPersonalAccessToken(body: { name: string; expiresIn?: string }) {
  return requestParsedJSON(createdPersonalAccessTokenSchema, "/api/personal-access-tokens", {
    method: "POST",
    body: JSON.stringify({ name: body.name, expires_in: body.expiresIn }),
  });
}

export async function revokePersonalAccessToken(id: string): Promise<void> {
  await requestJSON(`/api/personal-access-tokens/${encodeURIComponent(id)}`, { method: "DELETE" });
}

export async function revokeAllPersonalAccessTokens(): Promise<void> {
  await requestJSON("/api/personal-access-tokens/revoke-all", { method: "DELETE" });
}

export function parsePersonalAccessTokenListPayload(payload: unknown) {
  const parsed = personalAccessTokenListSchema.safeParse(payload);
  return parsed.success ? parsed.data : { data: [], total: 0 };
}
