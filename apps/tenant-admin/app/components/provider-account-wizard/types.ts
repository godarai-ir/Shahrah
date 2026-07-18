import type { PrincipalSearchKind } from "@/app/components/principal-search-input";
export type ProviderAuthMode =
  | "api_key"
  | "azure_entra_cert"
  | "aws_sigv4"
  | "google_service_account"
  | "no_credential"
  | (string & {});

export type ProviderConnectionSchemaField = {
  key: string;
  label: string;
  type: "text" | "url" | "password" | "textarea" | string;
  secret: boolean;
  required?: boolean;
  required_when?: Record<string, unknown>;
  validation?: { default?: unknown } & Record<string, unknown>;
  scope: "connection" | "credential";
  runtime_mapping?: string;
  help_text?: string;
  documentation_url?: string;
  redaction_policy?: string;
};

export type ProviderConnectionSchema = {
  provider_slug: string;
  auth_mode: ProviderAuthMode;
  version: number;
  runtime_ready: boolean;
  fields: ProviderConnectionSchemaField[];
};

export type ModelType = string;

export type PrincipalType = "user" | "team" | "org_unit" | "everyone";
export type SetupStep = 1 | 2 | 3 | 4 | 5;
export type WorkspaceTab = "models" | "access" | "credentials";
export type WorkspaceMode = "overview" | "setup" | "workspace";

export type AccessPrincipal = {
  type: PrincipalType;
  principal_id?: string;
  display_name?: string;
};

export type AccessPrincipalCandidate = {
  type: PrincipalSearchKind;
  principal_id: string;
  display_name: string;
  secondary?: string;
};

export type ProviderCatalogProvider = {
  provider: string;
  display_name: string;
  description?: string;
  docs_url?: string;
  docsUrl?: string;
  logo_url?: string;
  model_count?: number;
  enabled: boolean;
  auth_modes: ProviderAuthMode[];
  supports_bring_your_own_key: boolean;
  supports_virtual_keys: boolean;
};

export type ProviderCatalogModel = {
  provider: string;
  model: string;
  display_name: string;
  mode?: string[];
  modes?: string[];
  types: ModelType[];
  context_window?: number;
  input_modalities?: string[];
  output_modalities?: string[];
  capabilities?: string[];
  features?: string[];
  public_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
};

export type TenantProviderAccount = {
  id: string;
  provider: string;
  name: string;
  auth_mode: ProviderAuthMode;
  provider_config?: Record<string, unknown>;
  has_credential: boolean;
  credential_hint?: string;
  use_platform_managed_credentials: boolean;
};

export type TenantProviderAccountModel = {
  model: string;
  enabled: boolean;
  types: ModelType[];
  input_modalities?: string[];
  output_modalities?: string[];
  pricing_mode: "public" | "private";
  public_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
  private_pricing?: {
    input_usd_per_1m?: number;
    output_usd_per_1m?: number;
  };
  settings?: Record<string, unknown>;
};

export type AccountDetailsResponse = {
  account: TenantProviderAccount;
  models: TenantProviderAccountModel[];
  access: {
    managers: AccessPrincipal[];
    users: AccessPrincipal[];
  };
};

export type EffectiveModelView = {
  model: string;
  types?: string[];
  effective_capabilities?: string[];
  pricing_mode: "public" | "private";
  pricing_resolved: boolean;
  allowed: boolean;
  denied_reason?: string;
};

export type EffectiveModelsResponse = {
  data?: EffectiveModelView[];
};

export type ProvidersResponse = { data?: ProviderCatalogProvider[] };
export type ProviderModelsResponse = { data?: ProviderCatalogModel[] };
export type ProviderConnectionSchemasResponse = { data?: ProviderConnectionSchema[] };
export type AccountsResponse = { data?: TenantProviderAccount[] };
export type AccessPrincipalResponse = { data?: AccessPrincipalCandidate[] };
export type DisplaySettingsResponse = {
  data?: {
    currencyUnit?: "USD" | "IRR" | "TOMAN";
    priceScale?: "PER_MILLION_TOKENS" | "PER_TOKEN";
    exchangeRate?: {
      usdToIrr?: number;
    };
  };
};

export type Banner = {
  kind: "success" | "error";
  message: string;
};

export type CurlGuideShell = "powershell" | "bash";

export type CurlGuide = {
  modelName: string;
  modelFqn: string;
  mode: SupportedCurlGuideMode;
  commands: Record<CurlGuideShell, string>;
  sampleResponse?: string;
};

export type ModelDraft = {
  enabled: boolean;
  selectedCapabilities: string[];
  pricingMode: "public" | "private";
  inputPrice: string;
  outputPrice: string;
};

export type Readiness = "ready" | "needs_credentials" | "needs_access" | "pricing_incomplete";

export type ProviderAccountWizardMode = "manage" | "create";

export const CURL_GUIDE_LANGUAGE: Record<CurlGuideShell, "powershell" | "bash"> = {
  powershell: "powershell",
  bash: "bash",
};

export type SupportedCurlGuideMode = "chat" | "completion" | "embedding" | "responses" | "moderation" | "image_generation";

export const SUPPORTED_CURL_GUIDE_MODES: SupportedCurlGuideMode[] = [
  "chat",
  "completion",
  "responses",
  "embedding",
  "moderation",
  "image_generation",
];


