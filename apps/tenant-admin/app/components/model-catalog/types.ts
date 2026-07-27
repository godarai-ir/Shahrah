import type {
  ProviderAccountReadiness,
  TenantProviderAccount,
  TenantProviderAccountSummary,
} from "@/app/components/provider-account-wizard/types";
import type { ModelPrice } from "@/app/components/pricing";

export type ModelAvailability = {
  account_id: string;
  account_name: string;
  fqn: string;
  credential_source: "managed" | "byok" | string;
  effective_capabilities: string[];
  pricing_mode: "public" | "private";
  effective_pricing?: ModelPrice;
  pricing_resolved: boolean;
  readiness: ProviderAccountReadiness;
};

export type CatalogModel = {
  provider: string;
  model: string;
  display_name: string;
  mode?: string[];
  types: string[];
  context_window?: number;
  input_modalities?: string[];
  output_modalities?: string[];
  capabilities?: string[];
  features?: string[];
  public_pricing?: ModelPrice;
  is_deprecated?: boolean;
  replacement_model?: string;
  availability: ModelAvailability[];
};

export type CatalogModelResponse = {
  data?: CatalogModel[];
  total?: number;
  limit?: number;
  offset?: number;
};

export type ProviderAccountsResponse = {
  data?: TenantProviderAccount[];
  summaries?: Record<string, TenantProviderAccountSummary>;
};
