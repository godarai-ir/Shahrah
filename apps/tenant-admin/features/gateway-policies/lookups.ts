import { z } from "zod";

import type { ModelSelection } from "@/components/app-ui/ModelEntitySelect";
import type { PrincipalSearchKind, PrincipalSearchOption } from "@/components/app-ui/PrincipalEntitySelect";
import { requestParsedJSON } from "@/lib/http/api-client";

const principalListSchema = z.object({
  data: z.array(z.object({
    principal_id: z.string(),
    display_name: z.string(),
    secondary: z.string().optional(),
    type: z.string().optional(),
  })).optional(),
});
const entityListSchema = z.object({
  data: z.array(z.object({ id: z.string(), alias: z.string(), display_name: z.string().optional() })).optional(),
});
const modelListSchema = z.object({
  data: z.array(z.object({
    id: z.string().optional(),
    model: z.string().optional(),
    name: z.string().optional(),
    provider_name: z.string().optional(),
    provider: z.string().optional(),
  })).optional(),
});
const virtualModelListSchema = z.object({
  data: z.array(z.object({ id: z.string().optional(), alias: z.string().optional() })).optional(),
});

export async function searchPolicyPrincipals(
  kind: PrincipalSearchKind,
  query: string,
): Promise<PrincipalSearchOption[]> {
  if (kind === "user" || kind === "team") {
    const payload = await requestParsedJSON(
      principalListSchema,
      `/api/access-principals?kind=${kind}&q=${encodeURIComponent(query)}&limit=20`,
    );
    return (payload.data ?? []).map((item) => ({ ...item, type: kind }));
  }

  if (kind !== "org_unit" && kind !== "virtual_account") return [];
  const path = kind === "org_unit"
    ? "/api/gateway/lookups/organization-units"
    : "/api/gateway/lookups/virtual-accounts";
  const payload = await requestParsedJSON(entityListSchema, path);
  const normalizedQuery = query.trim().toLowerCase();
  return (payload.data ?? [])
    .filter((item) => `${item.display_name ?? ""} ${item.alias}`.toLowerCase().includes(normalizedQuery))
    .slice(0, 20)
    .map((item) => ({
      principal_id: item.id,
      display_name: item.display_name || item.alias,
      secondary: item.alias,
      type: kind,
    }));
}

export async function loadPolicyModelOptions(): Promise<{
  providerAllowedModels: ModelSelection[];
  virtualModels: ModelSelection[];
}> {
  const [modelsResult, virtualResult] = await Promise.allSettled([
    requestParsedJSON(modelListSchema, "/api/gateway/lookups/models"),
    requestParsedJSON(virtualModelListSchema, "/api/virtual-models"),
  ]);
  const providerAllowedModels = modelsResult.status === "fulfilled"
    ? (modelsResult.value.data ?? []).map((item) => {
        const name = item.model ?? item.name ?? item.id ?? "";
        return {
          id: name,
          name,
          kind: "provider_allowed" as const,
          secondary: item.provider_name ?? item.provider,
        };
      }).filter((item) => item.name)
    : [];
  const virtualModels = virtualResult.status === "fulfilled"
    ? (virtualResult.value.data ?? []).map((item) => ({
        id: item.id ?? item.alias ?? "",
        name: item.alias ?? item.id ?? "",
        kind: "virtual_model" as const,
      })).filter((item) => item.id && item.name)
    : [];

  return { providerAllowedModels, virtualModels };
}
