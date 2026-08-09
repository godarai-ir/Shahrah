import type {
  ProviderAccountReadiness,
  TenantProviderAccount,
  TenantProviderAccountSummary,
} from "@/features/provider-accounts/types";

export type ModelPrice = {
  input_usd_per_1m?: number;
  output_usd_per_1m?: number;
  rates?: ModelPricingRate[];
};

export type ModelPricingRate = {
  key: string;
  dimension: string;
  unit: string;
  unit_price_usd: string;
  condition?: Record<string, unknown>;
  region?: string;
  source_url?: string;
};

export type CapabilityReadiness = {
  endpoint: string;
  method?: string;
  input_modality?: string;
  output_modality?: string;
  transport?: string;
  state:
    | "contract_ready"
    | "blocked_lifecycle"
    | "pricing_incomplete"
    | "unsupported_transport";
  reason?: string;
  audit_revision?: string;
  live_validated: boolean;
};

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
  capability_readiness?: CapabilityReadiness[];
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
  readiness?: CapabilityReadiness[];
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
