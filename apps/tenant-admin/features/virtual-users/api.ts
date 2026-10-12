import { z } from "zod";

import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const virtualUserSchema = z.object({
  id: z.string(),
  tenant_id: z.string(),
  alias: z.string(),
  display_name: z.string().optional(),
  org_unit_id: z.string().optional(),
  team_id: z.string().optional(),
  active: z.boolean(),
  metadata: z.record(z.string(), z.unknown()).optional(),
  created_at: z.string(),
  updated_at: z.string(),
});

const vatTokenSchema = z.object({
  id: z.string(),
  key: z.string().optional(),
  key_alias: z.string(),
  key_name: z.string().optional(),
  key_type: z.string().optional(),
  allowed_models: z.array(z.string()).optional(),
  models: z.array(z.string()).optional(),
  team_id: z.string().optional(),
  org_unit_id: z.string().optional(),
  virtual_user_id: z.string().optional(),
  created_at: z.string().optional(),
  expires_at: z.string().optional(),
  last_used_at: z.string().optional(),
  revoked_at: z.string().optional(),
  blocked: z.boolean().optional(),
  on_behalf_of: z.object({ enabled: z.boolean(), product: z.string().optional() }).optional(),
});

const vatOnBehalfOfSchema = z.object({ enabled: z.boolean(), product: z.string().optional() });

const virtualUserListSchema = z.object({ data: z.array(virtualUserSchema).optional(), total: z.number().optional() });
const vatTokenListSchema = z.object({ data: z.array(vatTokenSchema).optional(), total: z.number().optional() });
const allowedModelListSchema = z.object({ data: z.array(z.object({ id: z.string() })).optional() });
const entitlementSchema = z.object({ features: z.record(z.string(), z.unknown()).optional() }).nullable();

export type VirtualUser = z.infer<typeof virtualUserSchema>;
export type VATToken = z.infer<typeof vatTokenSchema>;
export type EffectiveEntitlements = z.infer<typeof entitlementSchema>;
export type VirtualUserPayload = {
  alias: string;
  display_name?: string;
  team_id?: string;
  org_unit_id?: string;
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

/** Enables or disables on-behalf-of for a service token (tenant admins only). */
export function setVATOnBehalfOf(virtualUserID: string, tokenID: string, enabled: boolean) {
  return requestParsedJSON(
    vatOnBehalfOfSchema,
    `/api/virtual-users/${encodeURIComponent(virtualUserID)}/tokens/${encodeURIComponent(tokenID)}/on-behalf-of`,
    { method: "PUT", body: JSON.stringify({ enabled }) },
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
