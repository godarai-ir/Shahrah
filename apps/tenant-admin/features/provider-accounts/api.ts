import { z } from "zod";

import {
  providerCatalogModelSchema,
  providerCatalogProviderSchema,
} from "@/lib/contracts/provider-catalog";
import { APIContractError, requestJSON, requestParsedJSON } from "@/lib/http/api-client";

const principalKindSchema = z.enum(["user", "team", "org_unit", "virtual_account", "external_identity"]);
const accessPrincipalSchema = z.object({
  type: z.enum(["user", "team", "org_unit", "everyone"]),
  principal_id: z.string().optional(),
  display_name: z.string().optional(),
});
const accessPrincipalCandidateSchema = z.object({
  type: principalKindSchema,
  principal_id: z.string(),
  display_name: z.string(),
  secondary: z.string().optional(),
});

const providerAccountSchema = z.object({
  id: z.string(),
  provider: z.string(),
  name: z.string(),
  auth_mode: z.string(),
  provider_config: z.record(z.string(), z.unknown()).optional(),
  has_credential: z.boolean(),
  credential_hint: z.string().optional(),
  credential_rotation_required: z.boolean().optional(),
  use_platform_managed_credentials: z.boolean(),
  created_at: z.string().optional(),
  updated_at: z.string().optional(),
});

const providerAccountSummarySchema = z.object({
  model_count: z.number(),
  manager_count: z.number(),
  user_count: z.number(),
  pricing_unresolved_count: z.number(),
  readiness: z.enum(["ready", "needs_models", "needs_credentials", "needs_access", "pricing_incomplete"]),
});

const tenantProviderAccountModelSchema = z.object({
  model: z.string(),
  enabled: z.boolean(),
  types: z.array(z.string()),
  input_modalities: z.array(z.string()).optional(),
  output_modalities: z.array(z.string()).optional(),
  pricing_mode: z.enum(["public", "private"]),
  public_pricing: z.object({
    input_usd_per_1m: z.number().optional(),
    output_usd_per_1m: z.number().optional(),
  }).optional(),
  private_pricing: z.object({
    input_usd_per_1m: z.number().optional(),
    output_usd_per_1m: z.number().optional(),
    rates: z.array(z.object({
      key: z.string(),
      dimension: z.string(),
      unit: z.string(),
      unit_price_usd: z.string(),
      condition: z.record(z.string(), z.unknown()).optional(),
      region: z.string().optional(),
      source_url: z.string().optional(),
    })).optional(),
  }).optional(),
  settings: z.record(z.string(), z.unknown()).optional(),
});

const accountDetailsSchema = z.object({
  account: providerAccountSchema,
  models: z.array(tenantProviderAccountModelSchema),
  access: z.object({
    managers: z.array(accessPrincipalSchema),
    users: z.array(accessPrincipalSchema),
  }),
});

const connectionSchemaFieldSchema = z.object({
  key: z.string(),
  label: z.string(),
  type: z.string(),
  secret: z.boolean(),
  required: z.boolean().optional(),
  required_when: z.record(z.string(), z.unknown()).optional(),
  validation: z.record(z.string(), z.unknown()).optional(),
  scope: z.enum(["connection", "credential"]),
  runtime_mapping: z.string().optional(),
  help_text: z.string().optional(),
  documentation_url: z.string().optional(),
  redaction_policy: z.string().optional(),
});
const connectionSchema = z.object({
  provider_slug: z.string(),
  auth_mode: z.string(),
  version: z.number(),
  runtime_ready: z.boolean(),
  fields: z.array(connectionSchemaFieldSchema),
});

const credentialTestSchema = z.object({
  success: z.boolean(),
  authenticated: z.boolean().optional(),
  ready: z.boolean().optional(),
  category: z.string().optional(),
  unsupported: z.boolean().optional(),
  message: z.string().optional(),
  details: z.object({
    authenticated: z.boolean(),
    ready: z.boolean(),
    permissions: z.array(z.string()).optional(),
    credential_disabled: z.boolean().optional(),
    credential_blocked: z.boolean().optional(),
    account_blocked: z.boolean().optional(),
  }).optional(),
  wif_effective_scopes: z.array(z.string()).optional(),
  wif_expires_at: z.string().optional(),
  wif_last_refresh_at: z.string().optional(),
});

const effectiveModelSchema = z.object({
  model: z.string(),
  types: z.array(z.string()).optional(),
  effective_capabilities: z.array(z.string()).optional(),
  pricing_mode: z.enum(["public", "private"]),
  pricing_resolved: z.boolean(),
  allowed: z.boolean(),
  denied_reason: z.string().optional(),
});

const displaySettingsSchema = z.object({
  currencyUnit: z.enum(["USD", "IRR", "TOMAN"]).optional(),
  priceScale: z.enum(["PER_MILLION_TOKENS", "PER_TOKEN"]).optional(),
  exchangeRate: z.object({ usdToIrr: z.number().optional() }).optional(),
});

const capabilityReadinessSchema = z.object({
  endpoint: z.string(),
  method: z.string().optional(),
  input_modality: z.string().optional(),
  output_modality: z.string().optional(),
  transport: z.string().optional(),
  state: z.enum(["contract_ready", "blocked_lifecycle", "pricing_incomplete", "unsupported_transport"]),
  reason: z.string().optional(),
  audit_revision: z.string().optional(),
  live_validated: z.boolean(),
});
const modelPriceSchema = z.object({
  input_usd_per_1m: z.number().optional(),
  output_usd_per_1m: z.number().optional(),
  rates: z.array(z.object({
    key: z.string(),
    dimension: z.string(),
    unit: z.string(),
    unit_price_usd: z.string(),
    condition: z.record(z.string(), z.unknown()).optional(),
    region: z.string().optional(),
    source_url: z.string().optional(),
  })).optional(),
});
const modelAvailabilitySchema = z.object({
  account_id: z.string(),
  account_name: z.string(),
  fqn: z.string(),
  credential_source: z.string(),
  effective_capabilities: z.array(z.string()),
  pricing_mode: z.enum(["public", "private"]),
  effective_pricing: modelPriceSchema.optional(),
  pricing_resolved: z.boolean(),
  readiness: z.enum(["ready", "needs_models", "needs_credentials", "needs_access", "pricing_incomplete"]),
  capability_readiness: z.array(capabilityReadinessSchema).optional(),
});
const catalogModelSchema = z.object({
  provider: z.string(),
  model: z.string(),
  display_name: z.string(),
  mode: z.array(z.string()).optional(),
  types: z.array(z.string()),
  context_window: z.number().optional(),
  input_modalities: z.array(z.string()).optional(),
  output_modalities: z.array(z.string()).optional(),
  capabilities: z.array(z.string()).optional(),
  features: z.array(z.string()).optional(),
  public_pricing: modelPriceSchema.optional(),
  readiness: z.array(capabilityReadinessSchema).optional(),
  is_deprecated: z.boolean().optional(),
  replacement_model: z.string().optional(),
  availability: z.array(modelAvailabilitySchema),
});

const accountsResponseSchema = z.object({
  data: z.array(providerAccountSchema).optional(),
  summaries: z.record(z.string(), providerAccountSummarySchema).optional(),
});
const providersResponseSchema = z.object({ data: z.array(providerCatalogProviderSchema).optional() });
const modelsResponseSchema = z.object({ data: z.array(providerCatalogModelSchema).optional() });
const connectionSchemasResponseSchema = z.object({ data: z.array(connectionSchema).optional() });
const principalsResponseSchema = z.object({ data: z.array(accessPrincipalCandidateSchema).optional() });
const effectiveModelsResponseSchema = z.object({ data: z.array(effectiveModelSchema).optional() });
const catalogModelsResponseSchema = z.object({
  data: z.array(catalogModelSchema).optional(),
  total: z.number().optional(),
  limit: z.number().optional(),
  offset: z.number().optional(),
});

export function getProviderCatalog() {
  return requestParsedJSON(providersResponseSchema, "/api/provider-catalog/providers").then((response) => response.data ?? []);
}

export function getProviderConnectionSchemas(provider: string) {
  return requestParsedJSON(
    connectionSchemasResponseSchema,
    `/api/provider-catalog/providers/${encodeURIComponent(provider)}/connection-schemas`,
  ).then((response) => response.data ?? []);
}

export function getProviderModels(provider: string) {
  return requestParsedJSON(
    modelsResponseSchema,
    `/api/provider-catalog/providers/${encodeURIComponent(provider)}/models`,
  ).then((response) => response.data ?? []);
}

export function getProviderAccounts() {
  return requestParsedJSON(accountsResponseSchema, "/api/tenant-provider-accounts");
}

export function getProviderAccount(accountID: string) {
  return requestDataOrDirect(
    accountDetailsSchema,
    `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}`,
  );
}

export function createProviderAccount(input: Record<string, unknown>) {
  return requestDataOrDirect(providerAccountSchema, "/api/tenant-provider-accounts", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function deleteProviderAccount(accountID: string) {
  await requestJSON(`/api/tenant-provider-accounts/${encodeURIComponent(accountID)}`, { method: "DELETE" });
}

export async function updateProviderAccountCredentials(accountID: string, input: Record<string, unknown>) {
  await requestJSON(`/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/credentials`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function updateProviderAccountModels(accountID: string, models: unknown[]) {
  await requestJSON(`/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/models`, {
    method: "PUT",
    body: JSON.stringify({ models }),
  });
}

export async function updateProviderAccountAccess(accountID: string, input: {
  managers: Array<Record<string, unknown>>;
  users: Array<Record<string, unknown>>;
}) {
  await requestJSON(`/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/access`, {
    method: "PUT",
    body: JSON.stringify({ access: input }),
  });
}

export function testProviderConnection(input: Record<string, unknown>) {
  return requestDataOrDirect(credentialTestSchema, "/api/provider-connections/test", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function testProviderAccountCredentials(accountID: string, modelHint?: string) {
  return requestDataOrDirect(
    credentialTestSchema,
    `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/test-credentials`,
    { method: "POST", body: JSON.stringify({ model_hint: modelHint || undefined }) },
  );
}

export function searchProviderAccessPrincipals(kind: string, query: string) {
  return requestParsedJSON(
    principalsResponseSchema,
    `/api/access-principals?kind=${encodeURIComponent(kind)}&q=${encodeURIComponent(query)}&limit=20`,
  ).then((response) => response.data ?? []);
}

export function getEffectiveProviderModels(accountID: string, query: URLSearchParams) {
  return requestParsedJSON(
    effectiveModelsResponseSchema,
    `/api/tenant-provider-accounts/${encodeURIComponent(accountID)}/effective-models?${query.toString()}`,
  ).then((response) => response.data ?? []);
}

export function getProviderDisplaySettings() {
  return requestDataOrDirect(displaySettingsSchema, "/api/provider-catalog/display-settings");
}

export function getCatalogModels(query: string, signal?: AbortSignal) {
  return requestParsedJSON(catalogModelsResponseSchema, `/api/tenant-provider-account-models?${query}`, { signal });
}

async function requestDataOrDirect<T>(schema: z.ZodType<T>, path: string, options?: RequestInit): Promise<T> {
  const payload = await requestJSON(path, options);
  const direct = schema.safeParse(payload);
  if (direct.success) return direct.data;

  const enveloped = z.object({ data: schema }).safeParse(payload);
  if (enveloped.success) return enveloped.data.data;

  throw new APIContractError(path, [...direct.error.issues, ...enveloped.error.issues]);
}
