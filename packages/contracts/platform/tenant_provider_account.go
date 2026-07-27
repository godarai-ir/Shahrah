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
	ModelTypeVideoGeneration    ModelType = "video_generation"
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

// ProviderAccountReadiness is a derived setup state. It is never persisted.
type ProviderAccountReadiness string

const (
	ProviderAccountReadinessReady             ProviderAccountReadiness = "ready"
	ProviderAccountReadinessNeedsModels       ProviderAccountReadiness = "needs_models"
	ProviderAccountReadinessNeedsCredentials  ProviderAccountReadiness = "needs_credentials"
	ProviderAccountReadinessNeedsAccess       ProviderAccountReadiness = "needs_access"
	ProviderAccountReadinessPricingIncomplete ProviderAccountReadiness = "pricing_incomplete"
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

// PricingTier describes an optional usage tier for one rate.
type PricingTier struct {
	From *int   `json:"from,omitempty"`
	To   *int   `json:"to,omitempty"`
	Mode string `json:"mode,omitempty"`
}

// PricingRate is one exact, dimensioned catalog price. Amount remains a decimal
// string so JSON round-trips do not lose precision.
type PricingRate struct {
	Key         string                 `json:"key"`
	Usage       string                 `json:"usage"`
	Unit        string                 `json:"unit"`
	Amount      string                 `json:"amount"`
	Quantity    int                    `json:"quantity"`
	ServiceTier string                 `json:"service_tier"`
	Resolution  string                 `json:"resolution,omitempty"`
	Tier        *PricingTier           `json:"tier,omitempty"`
	Qualifiers  map[string]interface{} `json:"qualifiers,omitempty"`
}

// PricingProfile groups rates that share a region, currency and validity window.
type PricingProfile struct {
	Region        string        `json:"region"`
	Currency      string        `json:"currency"`
	EffectiveFrom string        `json:"effective_from,omitempty"`
	EffectiveTo   string        `json:"effective_to,omitempty"`
	SourceURL     string        `json:"source_url,omitempty"`
	Rates         []PricingRate `json:"rates"`
}

// ModelPrice is the canonical unit-aware pricing payload.
type ModelPrice struct {
	Status   string           `json:"status,omitempty"`
	Profiles []PricingProfile `json:"profiles"`
}

// UsageMeasure is one provider-independent billable quantity. Decimal values
// are strings so rate calculations and persisted snapshots remain exact.
type UsageMeasure struct {
	Usage       string                 `json:"usage"`
	Unit        string                 `json:"unit"`
	Amount      string                 `json:"amount"`
	Resolution  string                 `json:"resolution,omitempty"`
	ServiceTier string                 `json:"service_tier,omitempty"`
	Qualifiers  map[string]interface{} `json:"qualifiers,omitempty"`
}

// AppliedRate records the exact catalog rate selected for a usage measure.
type AppliedRate struct {
	ProfileKey       string                 `json:"profile_key"`
	RateKey          string                 `json:"rate_key"`
	Usage            string                 `json:"usage"`
	Unit             string                 `json:"unit"`
	RateAmount       string                 `json:"rate_amount"`
	RateQuantity     int                    `json:"rate_quantity"`
	UsageAmount      string                 `json:"usage_amount"`
	ComponentCostUSD string                 `json:"component_cost_usd"`
	Region           string                 `json:"region,omitempty"`
	Currency         string                 `json:"currency"`
	Resolution       string                 `json:"resolution,omitempty"`
	ServiceTier      string                 `json:"service_tier,omitempty"`
	PricingSource    string                 `json:"pricing_source,omitempty"`
	Qualifiers       map[string]interface{} `json:"qualifiers,omitempty"`
}

// RatedUsage is the canonical, auditable result of rating one request.
type RatedUsage struct {
	Status          string         `json:"status"`
	Model           string         `json:"model"`
	Provider        string         `json:"provider,omitempty"`
	Region          string         `json:"region,omitempty"`
	PricingSource   string         `json:"pricing_source,omitempty"`
	Measures        []UsageMeasure `json:"measures"`
	Components      []AppliedRate  `json:"components,omitempty"`
	TotalCostUSD    string         `json:"total_cost_usd"`
	ExchangeRate    string         `json:"exchange_rate"`
	TotalCostIRR    string         `json:"total_cost_irr"`
	PricingSnapshot *ModelPrice    `json:"pricing_snapshot,omitempty"`
	SnapshotHash    string         `json:"snapshot_hash,omitempty"`
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

// TestProviderConnectionRequest validates a provider connection without
// persisting a tenant provider account or its write-only credentials.
type TestProviderConnectionRequest struct {
	Provider                      string                      `json:"provider"`
	AuthMode                      ProviderAuthMode            `json:"auth_mode"`
	ProviderConfig                map[string]interface{}      `json:"provider_config,omitempty"`
	Credentials                   *ProviderAccountCredentials `json:"credentials,omitempty"`
	UsePlatformManagedCredentials bool                        `json:"use_platform_managed_credentials"`
	ModelHint                     string                      `json:"model_hint,omitempty"`
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
	Success     bool   `json:"success"`
	Unsupported bool   `json:"unsupported,omitempty"`
	Message     string `json:"message,omitempty"`
}

// TenantProviderAccountDetailsResponse returns full account data for wizard editing.
type TenantProviderAccountDetailsResponse struct {
	Account TenantProviderAccount              `json:"account"`
	Models  []TenantProviderAccountModel       `json:"models"`
	Access  TenantProviderAccountAccessControl `json:"access"`
}

// TenantProviderAccountSummary contains list-level, derived account state.
type TenantProviderAccountSummary struct {
	ModelCount             int                      `json:"model_count"`
	ManagerCount           int                      `json:"manager_count"`
	UserCount              int                      `json:"user_count"`
	PricingUnresolvedCount int                      `json:"pricing_unresolved_count"`
	Readiness              ProviderAccountReadiness `json:"readiness"`
}

// TenantProviderAccountCatalogRecord is the flat storage projection used to
// build account summaries and the tenant model catalog without N+1 queries.
type TenantProviderAccountCatalogRecord struct {
	Account          TenantProviderAccount
	ManagerCount     int
	UserCount        int
	Model            string
	Types            []ModelType
	InputModalities  []string
	OutputModalities []string
	PricingMode      PricingMode
	PublicPricing    *ModelPrice
	PrivatePricing   *ModelPrice
}

// TenantProviderAccountModelAvailability describes one callable account path
// for a catalog model.
type TenantProviderAccountModelAvailability struct {
	AccountID             string                   `json:"account_id"`
	AccountName           string                   `json:"account_name"`
	FQN                   string                   `json:"fqn"`
	CredentialSource      string                   `json:"credential_source"`
	EffectiveCapabilities []ModelType              `json:"effective_capabilities"`
	PricingMode           PricingMode              `json:"pricing_mode"`
	EffectivePricing      *ModelPrice              `json:"effective_pricing,omitempty"`
	PricingResolved       bool                     `json:"pricing_resolved"`
	Readiness             ProviderAccountReadiness `json:"readiness"`
}

// TenantProviderAccountModelCatalogItem is a tenant-visible model enriched
// with model-discovery metadata and all callable account paths.
type TenantProviderAccountModelCatalogItem struct {
	Provider         string                                   `json:"provider"`
	Model            string                                   `json:"model"`
	DisplayName      string                                   `json:"display_name"`
	Modes            []string                                 `json:"mode,omitempty"`
	Types            []ModelType                              `json:"types"`
	ContextWindow    *int                                     `json:"context_window,omitempty"`
	InputModalities  []string                                 `json:"input_modalities,omitempty"`
	OutputModalities []string                                 `json:"output_modalities,omitempty"`
	Capabilities     []string                                 `json:"capabilities,omitempty"`
	Features         []string                                 `json:"features,omitempty"`
	PublicPricing    *ModelPrice                              `json:"public_pricing,omitempty"`
	IsDeprecated     bool                                     `json:"is_deprecated,omitempty"`
	ReplacementModel string                                   `json:"replacement_model,omitempty"`
	Availability     []TenantProviderAccountModelAvailability `json:"availability"`
}

type TenantProviderAccountModelCatalogResponse struct {
	Data   []TenantProviderAccountModelCatalogItem `json:"data"`
	Total  int                                     `json:"total"`
	Limit  int                                     `json:"limit"`
	Offset int                                     `json:"offset"`
}

// TenantProviderAccountListResponse lists provider accounts for the tenant.
type TenantProviderAccountListResponse struct {
	Data      []TenantProviderAccount                 `json:"data"`
	Summaries map[string]TenantProviderAccountSummary `json:"summaries,omitempty"`
}

// AccessPrincipalListResponse contains principal picker search results.
type AccessPrincipalListResponse struct {
	Data []AccessPrincipalCandidate `json:"data"`
}
