import { z } from "zod";

export const providerAuthModeSchema = z.string();

export const modelPricingRateSchema = z.object({
  key: z.string(),
  dimension: z.string(),
  unit: z.string(),
  unit_price_usd: z.string(),
  condition: z.record(z.string(), z.unknown()).optional(),
  region: z.string().optional(),
  source_url: z.string().optional(),
});

export const providerCatalogProviderSchema = z.object({
  provider: z.string(),
  display_name: z.string(),
  description: z.string().optional(),
  docs_url: z.string().optional(),
  docsUrl: z.string().optional(),
  logo_url: z.string().optional(),
  model_count: z.number().optional(),
  enabled: z.boolean(),
  auth_modes: z.array(providerAuthModeSchema),
  supports_bring_your_own_key: z.boolean(),
  supports_virtual_keys: z.boolean(),
});

export const providerCatalogModelSchema = z.object({
  provider: z.string(),
  model: z.string(),
  display_name: z.string(),
  mode: z.array(z.string()).optional(),
  modes: z.array(z.string()).optional(),
  types: z.array(z.string()),
  context_window: z.number().optional(),
  input_modalities: z.array(z.string()).optional(),
  output_modalities: z.array(z.string()).optional(),
  capabilities: z.array(z.string()).optional(),
  features: z.array(z.string()).optional(),
  public_pricing: z.object({
    input_usd_per_1m: z.number().optional(),
    output_usd_per_1m: z.number().optional(),
    rates: z.array(modelPricingRateSchema).optional(),
  }).optional(),
});

export const providersResponseSchema = z.object({
  data: z.array(providerCatalogProviderSchema).optional(),
});

export const providerModelsResponseSchema = z.object({
  data: z.array(providerCatalogModelSchema).optional(),
});

export type ProviderAuthMode = z.infer<typeof providerAuthModeSchema>;
export type ModelType = string;
export type ModelPricingRate = z.infer<typeof modelPricingRateSchema>;
export type ProviderCatalogProvider = z.infer<typeof providerCatalogProviderSchema>;
export type ProviderCatalogModel = z.infer<typeof providerCatalogModelSchema>;
export type ProvidersResponse = z.infer<typeof providersResponseSchema>;
export type ProviderModelsResponse = z.infer<typeof providerModelsResponseSchema>;
