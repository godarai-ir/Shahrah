package platform

import "time"

// ProviderAuthMode declares how a provider account can authenticate.
type ProviderAuthMode string

const (
	ProviderAuthModeAPIKey               ProviderAuthMode = "api_key"
	ProviderAuthModeAzureEntraCert       ProviderAuthMode = "azure_entra_cert"
	ProviderAuthModeAWSSigV4             ProviderAuthMode = "aws_sigv4"
	ProviderAuthModeGoogleServiceAccount ProviderAuthMode = "google_service_account"
	ProviderAuthModeNoCredential         ProviderAuthMode = "no_credential"
)

// ModelType is the logical usage type exposed to tenants.
type ModelType string

const (
	ModelTypeChat               ModelType = "chat"
	ModelTypeCompletion         ModelType = "completion"
	ModelTypeEmbedding          ModelType = "embedding"
	ModelTypeRealtime           ModelType = "realtime"
	ModelTypeTextToSpeech       ModelType = "text_to_speech"
	ModelTypeAudioTranscription ModelType = "audio_transcription"
	ModelTypeAudioTranslation   ModelType = "audio_translation"
	ModelTypeImageGeneration    ModelType = "image_generation"
	ModelTypeModeration         ModelType = "moderation"
	ModelTypeRerank             ModelType = "rerank"
	ModelTypeSearch             ModelType = "search"
)

// PricingMode defines whether price is inherited from public catalog or private override.
type PricingMode string

const (
	PricingModePublic  PricingMode = "public"
	PricingModePrivate PricingMode = "private"
)

// PrincipalType identifies who receives a role on a provider account.
type PrincipalType string

const (
	PrincipalTypeUser     PrincipalType = "user"
	PrincipalTypeTeam     PrincipalType = "team"
	PrincipalTypeOrgUnit  PrincipalType = "org_unit"
	PrincipalTypeEveryone PrincipalType = "everyone"
)

// AccessRole specifies which permission bucket a principal belongs to.
type AccessRole string

const (
	AccessRoleManager AccessRole = "manager"
	AccessRoleUser    AccessRole = "user"
)

// AccessScope determines whether a rule applies to account-wide models or a specific model.
type AccessScope string

const (
	AccessScopeAccount AccessScope = "account"
	AccessScopeModel   AccessScope = "model"
)

// ModelPrice is normalized as USD per 1M tokens.
type ModelPrice struct {
	InputUSDPer1M  *float64 `json:"input_usd_per_1m,omitempty"`
	OutputUSDPer1M *float64 `json:"output_usd_per_1m,omitempty"`
}

// ProviderCatalogProvider is returned by the provider picker endpoint.
type ProviderCatalogProvider struct {
	Provider                string             `json:"provider"`
	DisplayName             string             `json:"display_name"`
	Description             string             `json:"description,omitempty"`
	DocsURL                 string             `json:"docs_url,omitempty"`
	LogoURL                 string             `json:"logo_url,omitempty"`
	ModelCount              int                `json:"model_count"`
	Enabled                 bool               `json:"enabled"`
	AuthModes               []ProviderAuthMode `json:"auth_modes"`
	SupportsBringYourOwnKey bool               `json:"supports_bring_your_own_key"`
	SupportsVirtualKeys     bool               `json:"supports_virtual_keys"`
}

// ProviderCatalogModel is returned by provider model discovery endpoints.
type ProviderCatalogModel struct {
	Provider         string                 `json:"provider"`
	Model            string                 `json:"model"`
	DisplayName      string                 `json:"display_name"`
	Modes            []string               `json:"mode,omitempty"`
	Types            []ModelType            `json:"types"`
	ContextWindow    *int                   `json:"context_window,omitempty"`
	InputModalities  []string               `json:"input_modalities,omitempty"`
	OutputModalities []string               `json:"output_modalities,omitempty"`
	Capabilities     []string               `json:"capabilities,omitempty"`
	Features         []string               `json:"features,omitempty"`
	PublicPricing    *ModelPrice            `json:"public_pricing,omitempty"`
	SettingsSchema   map[string]interface{} `json:"settings_schema,omitempty"`
	IsDeprecated     bool                   `json:"is_deprecated,omitempty"`
	ReplacementModel string                 `json:"replacement_model,omitempty"`
}

// ProviderCatalogListResponse contains providers available to a tenant.
type ProviderCatalogListResponse struct {
	Data []ProviderCatalogProvider `json:"data"`
}

// ProviderModelCatalogListResponse contains models of one provider.
type ProviderModelCatalogListResponse struct {
	Data []ProviderCatalogModel `json:"data"`
}

type ProviderCatalogExchangeRate struct {
	USDToIRR  float64 `json:"usdToIrr"`
	FetchedAt string  `json:"fetchedAt,omitempty"`
	Source    string  `json:"source,omitempty"`
}

type ProviderCatalogDisplaySettings struct {
	CurrencyUnit string                       `json:"currencyUnit,omitempty"`
	DateCalendar string                       `json:"dateCalendar,omitempty"`
	PriceScale   string                       `json:"priceScale,omitempty"`
	ExchangeRate *ProviderCatalogExchangeRate `json:"exchangeRate,omitempty"`
}

type ProviderCatalogDisplaySettingsResponse struct {
	Data ProviderCatalogDisplaySettings `json:"data"`
}

// ProviderAccountCredentials captures credential input for account setup.
type ProviderAccountCredentials struct {
	APIKey       string                 `json:"api_key,omitempty"`
	SecretRef    string                 `json:"secret_ref,omitempty"`
	Raw          map[string]interface{} `json:"raw,omitempty"`
	SkipRequired bool                   `json:"skip_required,omitempty"`
}

// TenantProviderAccount stores tenant-level provider account configuration.
type TenantProviderAccount struct {
	ID                            string                 `json:"id"`
	TenantID                      string                 `json:"tenant_id"`
	Provider                      string                 `json:"provider"`
	Name                          string                 `json:"name"`
	AuthMode                      ProviderAuthMode       `json:"auth_mode"`
	ProviderConfig                map[string]interface{} `json:"provider_config,omitempty"`
	HasCredential                 bool                   `json:"has_credential"`
	CredentialHint                string                 `json:"credential_hint,omitempty"`
	UsePlatformManagedCredentials bool                   `json:"use_platform_managed_credentials"`
	CreatedAt                     time.Time              `json:"created_at"`
	UpdatedAt                     time.Time              `json:"updated_at"`
}

// TenantProviderAccountModel is tenant-selected model config inside one account.
type TenantProviderAccountModel struct {
	Model            string                 `json:"model"`
	Enabled          bool                   `json:"enabled"`
	Types            []ModelType            `json:"types"`
	SupportedModes   []ModelType            `json:"supported_modes,omitempty"`
	InputModalities  []string               `json:"input_modalities,omitempty"`
	OutputModalities []string               `json:"output_modalities,omitempty"`
	PricingMode      PricingMode            `json:"pricing_mode"`
	PublicPricing    *ModelPrice            `json:"public_pricing,omitempty"`
	PrivatePricing   *ModelPrice            `json:"private_pricing,omitempty"`
	Settings         map[string]interface{} `json:"settings,omitempty"`
}

// AccessPrincipal is one user/team/everyone principal entry.
type AccessPrincipal struct {
	Type        PrincipalType `json:"type"`
	PrincipalID string        `json:"principal_id,omitempty"`
	DisplayName string        `json:"display_name,omitempty"`
}

// AccessPrincipalCandidate is returned by principal picker search APIs.
type AccessPrincipalCandidate struct {
	Type        PrincipalType `json:"type"`
	PrincipalID string        `json:"principal_id"`
	DisplayName string        `json:"display_name"`
	Secondary   string        `json:"secondary,omitempty"`
}

// TenantProviderAccountAccessRule binds a principal to manager/user role.
type TenantProviderAccountAccessRule struct {
	Role        AccessRole      `json:"role"`
	Principal   AccessPrincipal `json:"principal"`
	AccessScope AccessScope     `json:"access_scope,omitempty"`
	Model       string          `json:"model,omitempty"`
}

// TenantProviderAccountAccessControl groups manager/user access entries.
type TenantProviderAccountAccessControl struct {
	Managers []AccessPrincipal                 `json:"managers"`
	Users    []AccessPrincipal                 `json:"users"`
	Rules    []TenantProviderAccountAccessRule `json:"rules,omitempty"`
}

// CreateTenantProviderAccountRequest creates step-1 provider account.
type CreateTenantProviderAccountRequest struct {
	Provider                      string                      `json:"provider"`
	Name                          string                      `json:"name"`
	AuthMode                      ProviderAuthMode            `json:"auth_mode"`
	ProviderConfig                map[string]interface{}      `json:"provider_config,omitempty"`
	Credentials                   *ProviderAccountCredentials `json:"credentials,omitempty"`
	UsePlatformManagedCredentials bool                        `json:"use_platform_managed_credentials"`
}

// UpsertTenantProviderAccountModelSelection contains client-editable selection fields.
type UpsertTenantProviderAccountModelSelection struct {
	Model                 string      `json:"model"`
	EffectiveCapabilities []ModelType `json:"effective_capabilities,omitempty"`
	PricingMode           PricingMode `json:"pricing_mode,omitempty"`
	PrivatePricing        *ModelPrice `json:"private_pricing,omitempty"`
}

// UpsertTenantProviderAccountModelsRequest saves selected models from step-2.
type UpsertTenantProviderAccountModelsRequest struct {
	Models []UpsertTenantProviderAccountModelSelection `json:"models"`
}

type TenantProviderAccountModelEffective struct {
	Model                 string      `json:"model"`
	Enabled               bool        `json:"enabled"`
	EffectiveCapabilities []ModelType `json:"effective_capabilities"`
	PricingMode           PricingMode `json:"pricing_mode"`
	PricingResolved       bool        `json:"pricing_resolved"`
}

type TenantProviderAccountModelsUpsertResponse struct {
	Data []TenantProviderAccountModelEffective `json:"data"`
}

type EffectiveModelView struct {
	Model                 string      `json:"model"`
	Enabled               bool        `json:"enabled"`
	EffectiveCapabilities []ModelType `json:"effective_capabilities"`
	PricingMode           PricingMode `json:"pricing_mode"`
	PricingResolved       bool        `json:"pricing_resolved"`
	Allowed               bool        `json:"allowed"`
	DeniedReason          string      `json:"denied_reason,omitempty"`
}

type EffectiveModelsResponse struct {
	Data []EffectiveModelView `json:"data"`
}

// UpdateTenantProviderAccountAccessRequest saves step-3 access control.
type UpdateTenantProviderAccountAccessRequest struct {
	Access TenantProviderAccountAccessControl `json:"access"`
}

// TestTenantProviderAccountCredentialsRequest triggers credential validation.
type TestTenantProviderAccountCredentialsRequest struct {
	ModelHint string `json:"model_hint,omitempty"`
}

// UpdateTenantProviderAccountCredentialsRequest updates auth/credential settings for an existing account.
type UpdateTenantProviderAccountCredentialsRequest struct {
	AuthMode                      ProviderAuthMode            `json:"auth_mode,omitempty"`
	ProviderConfig                map[string]interface{}      `json:"provider_config,omitempty"`
	Credentials                   *ProviderAccountCredentials `json:"credentials,omitempty"`
	UsePlatformManagedCredentials *bool                       `json:"use_platform_managed_credentials,omitempty"`
}

// TestTenantProviderAccountCredentialsResponse returns test execution result.
type TestTenantProviderAccountCredentialsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// TenantProviderAccountDetailsResponse returns full account data for wizard editing.
type TenantProviderAccountDetailsResponse struct {
	Account TenantProviderAccount              `json:"account"`
	Models  []TenantProviderAccountModel       `json:"models"`
	Access  TenantProviderAccountAccessControl `json:"access"`
}

// TenantProviderAccountListResponse lists provider accounts for the tenant.
type TenantProviderAccountListResponse struct {
	Data []TenantProviderAccount `json:"data"`
}

// AccessPrincipalListResponse contains principal picker search results.
type AccessPrincipalListResponse struct {
	Data []AccessPrincipalCandidate `json:"data"`
}
