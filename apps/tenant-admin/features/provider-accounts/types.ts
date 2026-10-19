import type { PrincipalSearchKind } from "@/components/app-ui/PrincipalEntitySelect";
import type {
  ModelPricingRate,
  ModelType,
  ProviderAuthMode,
} from "@/lib/contracts/provider-catalog";

export type {
  ModelPricingRate,
  ModelType,
  ProviderAuthMode,
  ProviderCatalogModel,
  ProviderCatalogProvider,
  ProviderModelsResponse,
  ProvidersResponse,
} from "@/lib/contracts/provider-catalog";

export type ProviderConnectionSchemaField = {
  key: string;
  label: string;
  type: "text" | "url" | "password" | "textarea" | "multi_select" | string;
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

export type PrincipalType = "user" | "team" | "everyone";
export type SetupStep = 1 | 2 | 3 | 4 | 5;
export type WorkspaceTab = "overview" | "models" | "access" | "credentials";
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

export type TenantProviderAccount = {
  id: string;
  provider: string;
  name: string;
  auth_mode: ProviderAuthMode;
  provider_config?: Record<string, unknown>;
  has_credential: boolean;
  credential_hint?: string;
  credential_rotation_required?: boolean;
  use_platform_managed_credentials: boolean;
  created_at?: string;
  updated_at?: string;
};

export type CredentialTestResult = {
  success: boolean;
  authenticated?: boolean;
  ready?: boolean;
  category?: string;
  unsupported?: boolean;
  message?: string;
  details?: {
    authenticated: boolean;
    ready: boolean;
    permissions?: string[];
    credential_disabled?: boolean;
    credential_blocked?: boolean;
    account_blocked?: boolean;
  };
  wif_effective_scopes?: string[];
  wif_expires_at?: string;
  wif_last_refresh_at?: string;
};

export type ProviderAccountReadiness =
  | "ready"
  | "needs_models"
  | "needs_credentials"
  | "needs_access"
  | "pricing_incomplete";

export type TenantProviderAccountSummary = {
  model_count: number;
  manager_count: number;
  user_count: number;
  pricing_unresolved_count: number;
  readiness: ProviderAccountReadiness;
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
    rates?: ModelPricingRate[];
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

export type ProviderConnectionSchemasResponse = { data?: ProviderConnectionSchema[] };
export type AccountsResponse = {
  data?: TenantProviderAccount[];
  summaries?: Record<string, TenantProviderAccountSummary>;
};
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
  kind: "success" | "error" | "warning";
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
  rateOverrides?: ModelPricingRate[];
};

export type Readiness = "ready" | "connection_failed" | "needs_credentials" | "needs_access" | "pricing_incomplete";

export type ProviderAccountWizardMode = "manage" | "create";

export const CURL_GUIDE_LANGUAGE: Record<CurlGuideShell, "powershell" | "bash"> = {
  powershell: "powershell",
  bash: "bash",
};

export type SupportedCurlGuideMode =
  | "chat"
  | "completion"
  | "embedding"
  | "responses"
  | "moderation"
  | "image_generation"
  | "audio_transcription"
  | "text_to_speech";

export const SUPPORTED_CURL_GUIDE_MODES: SupportedCurlGuideMode[] = [
  "chat",
  "completion",
  "responses",
  "embedding",
  "moderation",
  "image_generation",
  "audio_transcription",
  "text_to_speech",
];


