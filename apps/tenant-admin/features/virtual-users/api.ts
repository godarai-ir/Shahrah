import { z } from "zod";

import { virtualUserSchema } from "@/lib/contracts/virtual-users";
import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const vatTokenSchema = z.object({
  id: z.string(),
  key: z.string().optional(),
  key_alias: z.string(),
  key_name: z.string().optional(),
  key_type: z.string().optional(),
  allowed_models: z.array(z.string()).optional(),
  models: z.array(z.string()).optional(),
  team_id: z.string().optional(),
  virtual_user_id: z.string().optional(),
  created_at: z.string().optional(),
  expires_at: z.string().optional(),
  last_used_at: z.string().optional(),
  revoked_at: z.string().optional(),
  blocked: z.boolean().optional(),
});

const virtualUserListSchema = z.object({ data: z.array(virtualUserSchema).optional(), total: z.number().optional() });
const vatTokenListSchema = z.object({ data: z.array(vatTokenSchema).optional(), total: z.number().optional() });
const allowedModelListSchema = z.object({ data: z.array(z.object({ id: z.string() })).optional() });
const entitlementSchema = z.object({ features: z.record(z.string(), z.unknown()).optional() }).nullable();

export type VATToken = z.infer<typeof vatTokenSchema>;
export type EffectiveEntitlements = z.infer<typeof entitlementSchema>;
export type VirtualUserPayload = {
  alias: string;
  display_name?: string;
  team_id?: string;
  metadata: Record<string, unknown>;
  active?: boolean;
};

export function getVirtualUserEntitlements() {
  return requestParsedJSON(entitlementSchema, "/api/entitlements");
}

export function listVirtualUsers(params: URLSearchParams) {
  return requestParsedJSON(virtualUserListSchema, `/api/virtual-users?${params.toString()}`);
}

export function listVATTokens(virtualUserID: string) {
  return requestParsedJSON(
    vatTokenListSchema,
    `/api/virtual-users/${encodeURIComponent(virtualUserID)}/tokens?limit=50&offset=0`,
  );
}

export async function listAllowedVATModels(teamID: string): Promise<string[]> {
  const payload = await requestParsedJSON(
    allowedModelListSchema,
    `/api/allowed-models?key_type=vat&team_id=${encodeURIComponent(teamID)}`,
  );
  return (payload.data ?? []).map((item) => item.id.trim()).filter(Boolean);
}

export function createVirtualUser(body: VirtualUserPayload) {
  return requestParsedJSON(virtualUserSchema, "/api/virtual-users", { method: "POST", body: JSON.stringify(body) });
}

export function updateVirtualUser(id: string, body: VirtualUserPayload) {
  return requestParsedJSON(virtualUserSchema, `/api/virtual-users/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(body),
  });
}

export async function deleteVirtualUser(id: string): Promise<void> {
  await requestJSON(`/api/virtual-users/${encodeURIComponent(id)}`, { method: "DELETE" });
}

export function createVATToken(
  virtualUserID: string,
  body: { key_alias: string; expires_in?: string; allowed_models: string[] },
) {
  return requestParsedJSON(vatTokenSchema, `/api/virtual-users/${encodeURIComponent(virtualUserID)}/tokens`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export async function revokeVATToken(virtualUserID: string, tokenID: string): Promise<void> {
  await requestJSON(
    `/api/virtual-users/${encodeURIComponent(virtualUserID)}/tokens/${encodeURIComponent(tokenID)}/revoke`,
    { method: "POST", body: JSON.stringify({ reason: "tenant_admin_ui" }) },
  );
}

export function rotateVATToken(virtualUserID: string, tokenID: string) {
  return requestParsedJSON(
    vatTokenSchema,
    `/api/virtual-users/${encodeURIComponent(virtualUserID)}/tokens/${encodeURIComponent(tokenID)}/rotate`,
    { method: "POST", body: JSON.stringify({}) },
  );
}

export function parseVirtualUserListPayload(payload: unknown) {
  const parsed = virtualUserListSchema.safeParse(payload);
  return parsed.success ? parsed.data : { data: [], total: 0 };
}
