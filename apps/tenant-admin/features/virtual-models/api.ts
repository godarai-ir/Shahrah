import { z } from "zod";

import { providersResponseSchema } from "@/lib/contracts/provider-catalog";
import { requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const principalSchema = z.object({
  type: z.enum(["user", "team", "org_unit", "virtual_account", "everyone"]),
  principal_id: z.string().optional(),
  display_name: z.string().optional(),
});

const virtualModelPolicySchema = z.object({
  timeout_sec: z.number().optional(),
  retry: z.object({
    attempts: z.number(),
    delay_ms: z.number(),
    status_codes: z.array(z.number()).optional(),
  }),
  fallback: z.object({
    status_codes: z.array(z.number()).optional(),
    candidate: z.boolean(),
  }),
});

const virtualModelTargetSchema = z.object({
  id: z.string().optional(),
  order: z.number(),
  weight: z.number().optional(),
  provider: z.string(),
  model: z.string(),
  policy: virtualModelPolicySchema,
});

const virtualModelSchema = z.object({
  id: z.string(),
  alias: z.string(),
  routing_strategy: z.string(),
  enabled: z.boolean(),
  exposed_modes: z.array(z.string()).optional(),
  version: z.number(),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
  created_by: z.string().optional(),
  updated_by: z.string().optional(),
  targets: z.array(virtualModelTargetSchema),
  access: z.object({
    managers: z.array(principalSchema).optional(),
    users: z.array(principalSchema).optional(),
  }).optional(),
});

const virtualModelListSchema = z.object({ data: z.array(virtualModelSchema).optional() });
const saveVirtualModelResponseSchema = z.object({ id: z.string().optional() }).passthrough();

const providerAccountSchema = z.object({
  id: z.string(),
  provider: z.string(),
  name: z.string(),
});
const providerAccountModelSchema = z.object({
  model: z.string(),
  enabled: z.boolean(),
  types: z.array(z.string()).optional(),
  effective_capabilities: z.array(z.string()).optional(),
  supported_modes: z.array(z.string()).optional(),
});
const providerAccountDetailsSchema = z.object({
  account: providerAccountSchema.optional(),
  models: z.array(providerAccountModelSchema).optional(),
});
const providerAccountEnvelopeSchema = providerAccountDetailsSchema.extend({
  data: providerAccountDetailsSchema.optional(),
});
const providerAccountListSchema = z.object({ data: z.array(providerAccountSchema).optional() });

const accessPrincipalSchema = z.object({
  principal_id: z.string(),
  display_name: z.string(),
  secondary: z.string().optional(),
  type: z.enum(["user", "team", "org_unit", "virtual_account", "external_identity"]).optional(),
});
const accessPrincipalListSchema = z.object({ data: z.array(accessPrincipalSchema).optional() });

const userLookupSchema = z.object({
  id: z.string(),
  email: z.string().optional(),
  mobile: z.string().optional(),
  display_name: z.string().optional(),
});
const userLookupListSchema = z.object({
  data: z.array(userLookupSchema).optional(),
  total: z.number().optional(),
});

export type VirtualModelPolicy = z.infer<typeof virtualModelPolicySchema>;
export type VirtualModelPrincipal = z.infer<typeof principalSchema>;
export type VirtualModelTarget = z.infer<typeof virtualModelTargetSchema>;
export type VirtualModel = z.infer<typeof virtualModelSchema>;
export type TenantProviderAccount = z.infer<typeof providerAccountSchema>;
export type TenantProviderAccountModel = z.infer<typeof providerAccountModelSchema>;
export type AccessPrincipalCandidate = z.infer<typeof accessPrincipalSchema>;
export type TenantAdminUserLookup = z.infer<typeof userLookupSchema>;

export type SaveVirtualModelPayload = {
  alias: string;
  routing_strategy: "priority" | "weighted";
  enabled: boolean;
  exposed_modes: string[];
  targets: VirtualModelTarget[];
  access: {
    managers: z.infer<typeof principalSchema>[];
    users: z.infer<typeof principalSchema>[];
  };
};

export async function listVirtualModels(): Promise<VirtualModel[]> {
  const payload = await requestParsedJSON(virtualModelListSchema, "/api/virtual-models");
  return payload.data ?? [];
}

export function getVirtualModel(id: string): Promise<VirtualModel> {
  return requestParsedJSON(virtualModelSchema, `/api/virtual-models/${encodeURIComponent(id)}`);
}

export function saveVirtualModel(id: string | undefined, body: SaveVirtualModelPayload) {
  const path = id ? `/api/virtual-models/${encodeURIComponent(id)}` : "/api/virtual-models";
  return requestParsedJSON(saveVirtualModelResponseSchema, path, {
    method: id ? "PATCH" : "POST",
    body: JSON.stringify(body),
  });
}

export async function deleteVirtualModel(id: string): Promise<void> {
  await requestJSON(`/api/virtual-models/${encodeURIComponent(id)}`, { method: "DELETE" });
}

export async function listProviderAccounts(): Promise<TenantProviderAccount[]> {
  const payload = await requestParsedJSON(providerAccountListSchema, "/api/tenant-provider-accounts");
  return payload.data ?? [];
}

export async function getProviderAccountDetails(id: string) {
  const payload = await requestParsedJSON(
    providerAccountEnvelopeSchema,
    `/api/tenant-provider-accounts/${encodeURIComponent(id)}`,
  );
  return payload.data ?? payload;
}

export async function searchAccessPrincipals(kind: string, query: string): Promise<AccessPrincipalCandidate[]> {
  const payload = await requestParsedJSON(
    accessPrincipalListSchema,
    `/api/access-principals?kind=${encodeURIComponent(kind)}&q=${encodeURIComponent(query)}&limit=20`,
  );
  return payload.data ?? [];
}

export async function listTenantUsers(): Promise<TenantAdminUserLookup[]> {
  const payload = await requestParsedJSON(userLookupListSchema, "/api/gateway/lookups/users");
  return payload.data ?? [];
}

export function getProviderCatalog() {
  return requestParsedJSON(providersResponseSchema, "/api/provider-catalog/providers").then((response) => response.data ?? []);
}
