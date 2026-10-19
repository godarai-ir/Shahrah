package platform

import (
	"context"
	"io"
	"strings"
	"time"
)

const (
	CredentialSourceGodarManagedKey = "godar_managed_key"
	CredentialSourceUserBYOK        = "user_byok"

	BillingModeWalletBilled   = "wallet_billed"
	BillingModeExternalBilled = "external_billed"

	AnthropicCodeExecutionPricingCondition        = "anthropic_code_execution_billing"
	AnthropicCodeExecutionFreeWithCurrentWebTools = "free_with_web_search_or_fetch_20260209_plus"
)

// Tenant represents the root tenant boundary for Shahrah.
type Tenant struct {
	ID     string `json:"id"`
	Slug   string `json:"slug"`
	Domain string `json:"domain"`
	Status string `json:"status"`
	// BillingUnit is irr or credit; empty means irr.
	BillingUnit string    `json:"billing_unit,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuthContext is the effective request auth context.
type AuthContext struct {
	TenantID                    string                          `json:"tenant_id"`
	TeamID                      string                          `json:"team_id,omitempty"`
	UserID                      string                          `json:"user_id,omitempty"`
	VirtualUserID               string                          `json:"virtual_user_id,omitempty"`
	KeyID                       string                          `json:"key_id"`
	KeyAlias                    string                          `json:"key_alias,omitempty"`
	Role                        string                          `json:"role"`
	Permissions                 []string                        `json:"permissions,omitempty"`
	KeyType                     string                          `json:"key_type,omitempty"`
	CreditProduct               string                          `json:"credit_product,omitempty"`
	TeamMemberPermissions       []string                        `json:"team_member_permissions,omitempty"`
	AllowedModels               []string                        `json:"allowed_models,omitempty"`
	AllowedModelTypes           map[string][]string             `json:"allowed_model_types,omitempty"`
	AllowedModelPricingMode     map[string]string               `json:"allowed_model_pricing_mode,omitempty"`
	AllowedModelPricingResolved map[string]bool                 `json:"allowed_model_pricing_resolved,omitempty"`
	VirtualModelSnapshots       map[string]VirtualModelSnapshot `json:"virtual_model_snapshots,omitempty"`
	EnforceModelAllowlist       bool                            `json:"enforce_model_allowlist,omitempty"`
	Spend                       float64                         `json:"spend"`
	ExpiresAt                   *time.Time                      `json:"expires_at,omitempty"`
	Blocked                     bool                            `json:"blocked"`
	TokenHash                   string                          `json:"token_hash,omitempty"`
}

// Wallet tracks tenant balance in IRR.
type Wallet struct {
	ID                    string     `json:"id"`
	TenantID              string     `json:"tenant_id"`
	BalanceIRR            int64      `json:"balance_irr"`
	PendingIRR            int64      `json:"pending_irr"`
	LedgerBalanceIRR      int64      `json:"ledger_balance_irr,omitempty"`
	AvailableBalanceIRR   int64      `json:"available_balance_irr,omitempty"`
	ReservedIRR           int64      `json:"reserved_irr,omitempty"`
	CommittedUnsettledIRR int64      `json:"committed_unsettled_irr,omitempty"`
	AsOf                  *time.Time `json:"as_of,omitempty"`
	SyncLagSeconds        int64      `json:"sync_lag_seconds,omitempty"`
	SyncStatus            string     `json:"sync_status,omitempty"`
	WindowID              string     `json:"window_id,omitempty"`
}

// Transaction records wallet changes.
type Transaction struct {
	ID           string                 `json:"id"`
	WalletID     string                 `json:"wallet_id"`
	TenantID     string                 `json:"tenant_id"`
	Type         string                 `json:"type"`
	AmountIRR    int64                  `json:"amount_irr"`
	BalanceAfter int64                  `json:"balance_after_irr,omitempty"`
	AmountUSD    *float64               `json:"amount_usd,omitempty"`
	ExchangeRate *float64               `json:"exchange_rate,omitempty"`
	Status       string                 `json:"status"`
	Description  string                 `json:"description,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	SourceType   string                 `json:"source_type,omitempty"`
	SourceID     string                 `json:"source_id,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
}

// WalletTopUpStatus defines payment lifecycle states.
type WalletTopUpStatus string

const (
	WalletTopUpStatusPending   WalletTopUpStatus = "PENDING"
	WalletTopUpStatusCompleted WalletTopUpStatus = "COMPLETED"
	WalletTopUpStatusFailed    WalletTopUpStatus = "FAILED"
	WalletTopUpStatusCanceled  WalletTopUpStatus = "CANCELED"
)

// WalletTopUpIntent tracks a pending external payment before wallet credit.
type WalletTopUpIntent struct {
	ID            string                 `json:"id"`
	PublicID      string                 `json:"public_id"`
	TenantID      string                 `json:"tenant_id"`
	Provider      string                 `json:"provider"`
	Status        WalletTopUpStatus      `json:"status"`
	AmountUSD     float64                `json:"amount_usd"`
	AmountIRR     float64                `json:"amount_irr"`
	ExchangeRate  float64                `json:"exchange_rate"`
	TaxPercent    float64                `json:"tax_percent"`
	FeePercent    float64                `json:"fee_percent"`
	TaxAmountIRR  float64                `json:"tax_amount_irr"`
	FeeAmountIRR  float64                `json:"fee_amount_irr"`
	PayableIRR    float64                `json:"payable_irr"`
	Authority     string                 `json:"authority,omitempty"`
	PaymentURL    string                 `json:"payment_url,omitempty"`
	RefID         string                 `json:"ref_id,omitempty"`
	GatewayTxnID  string                 `json:"gateway_txn_id,omitempty"`
	CardMask      string                 `json:"card_mask,omitempty"`
	PayerPhone    string                 `json:"payer_phone,omitempty"`
	CallbackURL   string                 `json:"callback_url,omitempty"`
	ReturnURL     string                 `json:"return_url,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	RawPayload    map[string]interface{} `json:"raw_payload,omitempty"`
	WalletTxID    string                 `json:"wallet_tx_id,omitempty"`
	FailureReason string                 `json:"failure_reason,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

// WalletTopUpSettings controls supported top-up ranges and pricing parameters.
type WalletTopUpSettings struct {
	MinAmountUSD float64   `json:"min_amount_usd"`
	MaxAmountUSD float64   `json:"max_amount_usd"`
	TaxPercent   float64   `json:"tax_percent"`
	FeePercent   float64   `json:"fee_percent"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TenantTracePrivacySettings controls tenant-level trace payload capture behavior.
type TenantTracePrivacySettings struct {
	TenantID        string    `json:"tenant_id"`
	CapturePolicy   string    `json:"capture_policy"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedByUserID string    `json:"updated_by_user_id,omitempty"`
}

// TenantRequestLoggingSettings controls tenant-level request log content capture behavior.
type TenantRequestLoggingSettings struct {
	TenantID        string    `json:"tenant_id"`
	LoggingMode     string    `json:"logging_mode"`
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedByUserID string    `json:"updated_by_user_id,omitempty"`
}

// ExchangeRate records a USD to IRR rate snapshot.
type ExchangeRate struct {
	ID        string    `json:"id"`
	USDtoIRR  float64   `json:"usd_to_irr"`
	Source    string    `json:"source"`
	FetchedAt time.Time `json:"fetched_at"`
}

// SpendLog records request billing and usage.
type SpendLog struct {
	ID              int64                  `json:"id"`
	RequestID       string                 `json:"request_id"`
	TokenHash       string                 `json:"token_hash"`
	TenantID        string                 `json:"tenant_id"`
	TeamID          string                 `json:"team_id,omitempty"`
	UserID          string                 `json:"user_id,omitempty"`
	Model           string                 `json:"model"`
	InputTokens     int                    `json:"input_tokens"`
	OutputTokens    int                    `json:"output_tokens"`
	TotalTokens     int                    `json:"total_tokens"`
	CostUSD         float64                `json:"cost_usd"`
	CostIRR         float64                `json:"cost_irr"`
	ResponseTimeMS  int                    `json:"response_time_ms"`
	Status          string                 `json:"status"`
	RequestMetadata map[string]interface{} `json:"request_metadata,omitempty"`
	UsageBreakdown  *UsageCostBreakdown    `json:"usage_breakdown,omitempty"`
	PricingSnapshot *UsagePricingSnapshot  `json:"pricing_snapshot,omitempty"`
	Credits         *SpendLogCredits       `json:"credits,omitempty"`
	CompletionStart time.Time              `json:"completion_start_time"`
	CompletionEnd   time.Time              `json:"completion_end_time"`
}

// SpendLogCredits is the credit computation attached to one spend log.
type SpendLogCredits struct {
	Millicredits        int64  `json:"credits_mc"`
	AvoidedMillicredits int64  `json:"credits_avoided_mc,omitempty"`
	RateCardVersion     int    `json:"rate_card_version"`
	Product             string `json:"product"`
	BillingMode         string `json:"billing_mode,omitempty"`
	Source              string `json:"credit_source"`
}

// Credit product labels identify which GodarAI surface produced usage.
const (
	CreditProductGateway       = "gateway"
	CreditProductLibreChat     = "librechat"
	CreditProductGuardrail     = "guardrail"
	CreditProductSemanticCache = "semantic_cache"
	CreditProductEmbeddings    = "embeddings"
	CreditProductOCR           = "ocr"
	CreditProductPlayground    = "playground"
	CreditProductOIDCApp       = "oidc_app"
)

// UsageCostBreakdown captures estimated or final cost.
type UsageCostComponent struct {
	Dimension    string                 `json:"dimension"`
	Quantity     string                 `json:"quantity"`
	Unit         string                 `json:"unit"`
	UnitPriceUSD string                 `json:"unit_price_usd"`
	CostUSD      string                 `json:"cost_usd"`
	Condition    map[string]interface{} `json:"condition,omitempty"`
}

type UsagePricingSnapshot struct {
	Provider    string                        `json:"provider"`
	Revision    string                        `json:"revision"`
	GeneratedAt time.Time                     `json:"generated_at"`
	Rates       map[string]UsageCostComponent `json:"rates"`
	RateList    []UsageCostComponent          `json:"rate_list,omitempty"`
}

type UsageCostBreakdown struct {
	Model                      string                      `json:"model"`
	RequestType                string                      `json:"request_type,omitempty"`
	Endpoint                   string                      `json:"endpoint,omitempty"`
	Metadata                   map[string]interface{}      `json:"metadata,omitempty"`
	InputTokens                int                         `json:"input_tokens"`
	OutputTokens               int                         `json:"output_tokens"`
	TotalTokens                int                         `json:"total_tokens"`
	InputCostUSD               float64                     `json:"input_cost_usd"`
	OutputCostUSD              float64                     `json:"output_cost_usd"`
	TotalCostUSD               float64                     `json:"total_cost_usd"`
	ExchangeRate               float64                     `json:"exchange_rate"`
	TotalCostIRR               float64                     `json:"total_cost_irr"`
	CacheCreation5mInputTokens int                         `json:"cache_creation_5m_input_tokens,omitempty"`
	CacheCreation1hInputTokens int                         `json:"cache_creation_1h_input_tokens,omitempty"`
	CacheReadInputTokens       int                         `json:"cache_read_input_tokens,omitempty"`
	ServerToolUse              map[string]int              `json:"server_tool_use,omitempty"`
	ServiceTier                string                      `json:"service_tier,omitempty"`
	Speed                      string                      `json:"speed,omitempty"`
	Batch                      bool                        `json:"batch,omitempty"`
	PricingRevision            string                      `json:"pricing_revision,omitempty"`
	PricingConditions          map[string]string           `json:"pricing_conditions,omitempty"`
	ModelRuntimeProfile        *ModelRuntimeProfileBinding `json:"model_runtime_profile,omitempty"`
	Components                 []UsageCostComponent        `json:"components,omitempty"`
	Quantities                 map[string]string           `json:"quantities,omitempty"`
	PricingSnapshot            *UsagePricingSnapshot       `json:"pricing_snapshot,omitempty"`
}

// BudgetReservationTarget is one window reserved for a request.
type BudgetReservationTarget struct {
	Kind           string  `json:"kind"`
	RuleID         string  `json:"rule_id,omitempty"`
	Scope          string  `json:"scope,omitempty"`
	ScopeValue     string  `json:"scope_value,omitempty"`
	WindowID       string  `json:"window_id"`
	WindowKey      string  `json:"window_key"`
	ReservationKey string  `json:"reservation_key"`
	EstimatedIRR   float64 `json:"estimated_irr"`
	// EstimatedMC is the reserved amount of a credit-unit target, in millicredits.
	EstimatedMC int64 `json:"estimated_mc,omitempty"`
	// OnExceed is the rule's exceed behavior for credit-unit policy targets.
	OnExceed  string    `json:"on_exceed,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// BudgetReservation tracks request-time spend reservations in active budget windows.
type BudgetReservation struct {
	ID             string  `json:"id"`
	TenantID       string  `json:"tenant_id"`
	WindowID       string  `json:"window_id"`
	WindowKey      string  `json:"window_key"`
	ReservationKey string  `json:"reservation_key"`
	EstimatedIRR   float64 `json:"estimated_irr"`
	// Unit is the window unit of every target: irr (empty) or credit.
	Unit string `json:"unit,omitempty"`
	// BillingMode is the deployment billing mode the reservation was made for.
	BillingMode string `json:"billing_mode,omitempty"`
	// BYOKFeePercent is the platform fee charged to the tenant pool for
	// externally billed credit requests, as a decimal string.
	BYOKFeePercent string                    `json:"byok_fee_percent,omitempty"`
	CreatedAt      time.Time                 `json:"created_at"`
	Targets        []BudgetReservationTarget `json:"targets,omitempty"`
}

// Deployment represents a configured provider target.
type Deployment struct {
	ID                  string                 `json:"id"`
	TenantID            string                 `json:"tenant_id,omitempty"`
	ModelName           string                 `json:"model_name"`
	Provider            string                 `json:"provider"`
	Model               string                 `json:"model"`
	APIBase             string                 `json:"api_base,omitempty"`
	APIKey              string                 `json:"api_key,omitempty"`
	APIVersion          string                 `json:"api_version,omitempty"`
	ProjectID           string                 `json:"project_id,omitempty"`
	EndpointID          string                 `json:"endpoint_id,omitempty"`
	RecognizerID        string                 `json:"recognizer_id,omitempty"`
	ProviderHeaders     map[string]string      `json:"provider_headers,omitempty"`
	RegionName          string                 `json:"region_name,omitempty"`
	RPM                 int64                  `json:"rpm,omitempty"`
	TPM                 int64                  `json:"tpm,omitempty"`
	Weight              float64                `json:"weight,omitempty"`
	TimeoutSec          float64                `json:"timeout,omitempty"`
	CredentialSource    string                 `json:"credential_source,omitempty"`
	BillingMode         string                 `json:"billing_mode,omitempty"`
	PricingMode         PricingMode            `json:"pricing_mode,omitempty"`
	ProviderAccountID   string                 `json:"provider_account_id,omitempty"`
	ProviderAccountName string                 `json:"provider_account_name,omitempty"`
	ConnectionID        string                 `json:"connection_id,omitempty"`
	CredentialVersionID string                 `json:"credential_version_id,omitempty"`
	ProviderAuth        *ResolvedProviderAuth  `json:"provider_auth,omitempty"`
	ModelInfo           map[string]interface{} `json:"model_info,omitempty"`
	EffectivePricing    *ModelPrice            `json:"effective_pricing,omitempty"`
}

// TenantStore resolves tenants for requests.
type TenantStore interface {
	GetTenantByHost(ctx context.Context, host string) (*Tenant, error)
}

// KeyStore resolves and persists virtual keys.
type KeyStore interface {
	GetKeyByHash(ctx context.Context, hash string) (*AuthContext, error)
	TouchKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error
	CreateKey(ctx context.Context, params CreateKeyParams) (*CreatedKey, error)
	ListKeys(ctx context.Context, tenantID string) ([]CreatedKey, error)
	ListPATs(ctx context.Context, tenantID string, userID string, query string, limit int, offset int) ([]CreatedKey, int, error)
	PATExists(ctx context.Context, tenantID string, userID string, name string) (bool, error)
	RevokePAT(ctx context.Context, tenantID string, userID string, keyID string, revokedByUserID string) (bool, error)
	RevokeAllPATs(ctx context.Context, tenantID string, userID string, revokedByUserID string) (int64, error)
	ListVATTokens(ctx context.Context, tenantID string, virtualUserID string, query string, limit int, offset int) ([]CreatedKey, int, error)
	UpdateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, req UpdateVATTokenRequest) (*CreatedKey, error)
	RevokeVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, revokedByUserID string, reason string) (bool, error)
	RotateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, rotatedByUserID string, gracePeriod string) (*CreatedKey, error)
}

// BudgetEnforcer validates budgets and rate limits.
type BudgetEnforcer interface {
	CheckRateLimit(ctx context.Context, auth *AuthContext, model string) error
	CheckBudget(ctx context.Context, auth *AuthContext, estimated UsageCostBreakdown) error
}

// WalletStore manages wallet and billing persistence.
type WalletStore interface {
	GetWallet(ctx context.Context, tenantID string) (*Wallet, error)
	TopUpWallet(ctx context.Context, tenantID string, amountIRR float64, metadata map[string]interface{}) (*Transaction, error)
	DebitWallet(ctx context.Context, tenantID string, breakdown UsageCostBreakdown, requestID string) (*Transaction, error)
	GetLatestExchangeRate(ctx context.Context) (*ExchangeRate, error)
	UpsertExchangeRate(ctx context.Context, rate ExchangeRate) (*ExchangeRate, error)
}

// UsageLogger records request and spend outcomes.
type UsageLogger interface {
	LogRequest(ctx context.Context, entry RequestLogEntry) error
	LogSpend(ctx context.Context, log SpendLog) error
}

// TunnelTransport forwards proxied requests into the Europe path.
type TunnelTransport interface {
	Proxy(ctx context.Context, req TunnelRequest) (*TunnelResponse, error)
	Stream(ctx context.Context, req TunnelRequest) (TunnelStream, error)
	Healthy(ctx context.Context) error
}

// TunnelRawTransport streams large provider payloads without embedding them in
// the JSON tunnel envelope.
type TunnelRawTransport interface {
	ProxyRaw(ctx context.Context, req TunnelRequest, body io.Reader, contentLength int64) (*TunnelRawResponse, error)
}

// ProviderForwarder forwards provider requests on the Europe side.
type ProviderForwarder interface {
	Forward(ctx context.Context, req TunnelRequest) (*TunnelResponse, error)
	ForwardStream(ctx context.Context, req TunnelRequest) (TunnelStream, error)
}

// CreateKeyParams are the inputs for creating a virtual key.
type CreateKeyParams struct {
	TenantID      string                 `json:"tenant_id"`
	KeyAlias      string                 `json:"key_alias"`
	Role          string                 `json:"role,omitempty"`
	KeyType       string                 `json:"key_type,omitempty"`
	Models        []string               `json:"models,omitempty"`
	AllowedModels []string               `json:"allowed_models,omitempty"`
	TeamID        string                 `json:"team_id,omitempty"`
	UserID        string                 `json:"user_id,omitempty"`
	VirtualUserID string                 `json:"virtual_user_id,omitempty"`
	ExpiresIn     string                 `json:"expires_in,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CreatedKey is the persisted key metadata returned by admin APIs.
type CreatedKey struct {
	ID              string     `json:"id"`
	Key             string     `json:"key,omitempty"`
	KeyAlias        string     `json:"key_alias"`
	MaskedKey       string     `json:"key_name,omitempty"`
	TenantID        string     `json:"tenant_id"`
	Role            string     `json:"role"`
	KeyType         string     `json:"key_type,omitempty"`
	Models          []string   `json:"models"`
	AllowedModels   []string   `json:"allowed_models,omitempty"`
	TeamID          string     `json:"team_id,omitempty"`
	UserID          string     `json:"user_id,omitempty"`
	VirtualUserID   string     `json:"virtual_user_id,omitempty"`
	Spend           float64    `json:"spend"`
	Blocked         bool       `json:"blocked"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
	RevokedAt       *time.Time `json:"revoked_at,omitempty"`
	CreatedByUserID string     `json:"created_by_user_id,omitempty"`
	TokenHash       string     `json:"token_hash,omitempty"`
}

type ListPATResponse struct {
	Data       []CreatedKey `json:"data"`
	Total      int          `json:"total"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
	NextOffset *int         `json:"next_offset,omitempty"`
}

type CreatePATRequest struct {
	Name      string `json:"name"`
	ExpiresIn string `json:"expires_in,omitempty"`
}

type PATExistsResponse struct {
	Exists bool `json:"exists"`
}

type RevokeAllPATResponse struct {
	Revoked int64 `json:"revoked"`
}

type ListVATTokensResponse struct {
	Data       []CreatedKey `json:"data"`
	Total      int          `json:"total"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
	NextOffset *int         `json:"next_offset,omitempty"`
}

type UpdateVATTokenRequest struct {
	KeyAlias      string                 `json:"key_alias,omitempty"`
	AllowedModels []string               `json:"allowed_models,omitempty"`
	Models        []string               `json:"models,omitempty"`
	ExpiresIn     string                 `json:"expires_in,omitempty"`
	Blocked       *bool                  `json:"blocked,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type RevokeVATTokenRequest struct {
	Reason string `json:"reason,omitempty"`
}

type RotateVATTokenRequest struct {
	GracePeriod string `json:"grace_period,omitempty"`
}

const (
	LegacyKeyRoleVirtualKey     = "virtual_key"
	LegacyKeyRoleServiceAccount = "service_account"
	KeyTypePAT                  = "pat"
	KeyTypeVAT                  = "vat"
)

func (p *CreateKeyParams) Normalize() {
	p.KeyAlias = strings.TrimSpace(p.KeyAlias)
	p.Role = strings.TrimSpace(strings.ToLower(p.Role))
	p.KeyType = strings.TrimSpace(strings.ToLower(p.KeyType))
	p.TeamID = strings.TrimSpace(p.TeamID)
	p.UserID = strings.TrimSpace(p.UserID)
	p.VirtualUserID = strings.TrimSpace(p.VirtualUserID)
	p.ExpiresIn = strings.TrimSpace(p.ExpiresIn)

	if len(p.AllowedModels) == 0 && len(p.Models) > 0 {
		p.AllowedModels = append([]string(nil), p.Models...)
	}

	if len(p.Models) == 0 && len(p.AllowedModels) > 0 {
		p.Models = append([]string(nil), p.AllowedModels...)
	}

	if p.KeyType == "" {
		switch p.Role {
		case LegacyKeyRoleServiceAccount:
			p.KeyType = KeyTypeVAT
		case "", LegacyKeyRoleVirtualKey:
			p.KeyType = KeyTypePAT
		}
	}

	if p.Role == "" {
		switch p.KeyType {
		case KeyTypeVAT:
			p.Role = LegacyKeyRoleServiceAccount
		default:
			p.Role = LegacyKeyRoleVirtualKey
		}
	}
}

// TunnelRequest is the internal request sent over the Shahrah tunnel path.
type ProviderOperation string

const (
	ProviderOperationCredentialProbe                    ProviderOperation = "credential_probe"
	ProviderOperationChatCompletions                    ProviderOperation = "chat_completions"
	ProviderOperationZAILayoutParsing                   ProviderOperation = "zai_layout_parsing"
	ProviderOperationXAIDeferredChatGet                 ProviderOperation = "xai_deferred_chat_get"
	ProviderOperationXAIFilesInitialize                 ProviderOperation = "xai_files_initialize"
	ProviderOperationXAIFilesUploadChunks               ProviderOperation = "xai_files_upload_chunks"
	ProviderOperationXAILanguageModelsList              ProviderOperation = "xai_language_models_list"
	ProviderOperationXAILanguageModelsGet               ProviderOperation = "xai_language_models_get"
	ProviderOperationXAIImageGenerationModelsList       ProviderOperation = "xai_image_generation_models_list"
	ProviderOperationXAIImageGenerationModelsGet        ProviderOperation = "xai_image_generation_models_get"
	ProviderOperationOpenAICompletions                  ProviderOperation = "openai_completions"
	ProviderOperationOpenAIResponsesCreate              ProviderOperation = "openai_responses_create"
	ProviderOperationOpenAIResponsesGet                 ProviderOperation = "openai_responses_get"
	ProviderOperationOpenAIResponsesDelete              ProviderOperation = "openai_responses_delete"
	ProviderOperationOpenAIResponsesCancel              ProviderOperation = "openai_responses_cancel"
	ProviderOperationOpenAIResponsesCompact             ProviderOperation = "openai_responses_compact"
	ProviderOperationOpenAIResponsesInputItems          ProviderOperation = "openai_responses_input_items"
	ProviderOperationOpenAIResponsesInputTokens         ProviderOperation = "openai_responses_input_tokens"
	ProviderOperationOpenAIConversationsCreate          ProviderOperation = "openai_conversations_create"
	ProviderOperationOpenAIConversationsList            ProviderOperation = "openai_conversations_list"
	ProviderOperationOpenAIConversationsGet             ProviderOperation = "openai_conversations_get"
	ProviderOperationOpenAIConversationsUpdate          ProviderOperation = "openai_conversations_update"
	ProviderOperationOpenAIConversationsDelete          ProviderOperation = "openai_conversations_delete"
	ProviderOperationOpenAIConversationItemsCreate      ProviderOperation = "openai_conversation_items_create"
	ProviderOperationOpenAIConversationItemsList        ProviderOperation = "openai_conversation_items_list"
	ProviderOperationOpenAIConversationItemsGet         ProviderOperation = "openai_conversation_items_get"
	ProviderOperationOpenAIConversationItemsDelete      ProviderOperation = "openai_conversation_items_delete"
	ProviderOperationOpenAIEmbeddings                   ProviderOperation = "openai_embeddings"
	ProviderOperationOpenAIModerations                  ProviderOperation = "openai_moderations"
	ProviderOperationOpenAIImagesGenerations            ProviderOperation = "openai_images_generations"
	ProviderOperationOpenAIImagesEdits                  ProviderOperation = "openai_images_edits"
	ProviderOperationOpenAIImagesVariations             ProviderOperation = "openai_images_variations"
	ProviderOperationOpenAIAudioSpeech                  ProviderOperation = "openai_audio_speech"
	ProviderOperationOpenAIAudioTranscriptions          ProviderOperation = "openai_audio_transcriptions"
	ProviderOperationOpenAIAudioTranslations            ProviderOperation = "openai_audio_translations"
	ProviderOperationOpenAIBatchesCreate                ProviderOperation = "openai_batches_create"
	ProviderOperationOpenAIBatchesList                  ProviderOperation = "openai_batches_list"
	ProviderOperationOpenAIBatchesGet                   ProviderOperation = "openai_batches_get"
	ProviderOperationOpenAIBatchesCancel                ProviderOperation = "openai_batches_cancel"
	ProviderOperationXAIBatchesCreate                   ProviderOperation = "xai_batches_create"
	ProviderOperationXAIBatchesList                     ProviderOperation = "xai_batches_list"
	ProviderOperationXAIBatchesGet                      ProviderOperation = "xai_batches_get"
	ProviderOperationXAIBatchRequestsList               ProviderOperation = "xai_batch_requests_list"
	ProviderOperationXAIBatchRequestsAdd                ProviderOperation = "xai_batch_requests_add"
	ProviderOperationXAIBatchResults                    ProviderOperation = "xai_batch_results"
	ProviderOperationXAIBatchesCancel                   ProviderOperation = "xai_batches_cancel"
	ProviderOperationOpenAIFilesCreate                  ProviderOperation = "openai_files_create"
	ProviderOperationOpenAIFilesList                    ProviderOperation = "openai_files_list"
	ProviderOperationOpenAIFilesGet                     ProviderOperation = "openai_files_get"
	ProviderOperationOpenAIFilesDelete                  ProviderOperation = "openai_files_delete"
	ProviderOperationOpenAIFilesContent                 ProviderOperation = "openai_files_content"
	ProviderOperationXAIFilesUpdate                     ProviderOperation = "xai_files_update"
	ProviderOperationXAIFilePublicURLCreate             ProviderOperation = "xai_file_public_url_create"
	ProviderOperationXAIFilePublicURLRevoke             ProviderOperation = "xai_file_public_url_revoke"
	ProviderOperationOpenAIUploadsCreate                ProviderOperation = "openai_uploads_create"
	ProviderOperationOpenAIUploadsAddPart               ProviderOperation = "openai_uploads_add_part"
	ProviderOperationOpenAIUploadsComplete              ProviderOperation = "openai_uploads_complete"
	ProviderOperationOpenAIUploadsCancel                ProviderOperation = "openai_uploads_cancel"
	ProviderOperationOpenAIVectorStoresCreate           ProviderOperation = "openai_vector_stores_create"
	ProviderOperationOpenAIVectorStoresList             ProviderOperation = "openai_vector_stores_list"
	ProviderOperationOpenAIVectorStoresGet              ProviderOperation = "openai_vector_stores_get"
	ProviderOperationOpenAIVectorStoresUpdate           ProviderOperation = "openai_vector_stores_update"
	ProviderOperationOpenAIVectorStoresDelete           ProviderOperation = "openai_vector_stores_delete"
	ProviderOperationOpenAIVectorStoresSearch           ProviderOperation = "openai_vector_stores_search"
	ProviderOperationOpenAIVectorStoreFilesCreate       ProviderOperation = "openai_vector_store_files_create"
	ProviderOperationOpenAIVectorStoreFilesList         ProviderOperation = "openai_vector_store_files_list"
	ProviderOperationOpenAIVectorStoreFilesGet          ProviderOperation = "openai_vector_store_files_get"
	ProviderOperationOpenAIVectorStoreFilesUpdate       ProviderOperation = "openai_vector_store_files_update"
	ProviderOperationOpenAIVectorStoreFilesDelete       ProviderOperation = "openai_vector_store_files_delete"
	ProviderOperationOpenAIVectorStoreFilesContent      ProviderOperation = "openai_vector_store_files_content"
	ProviderOperationOpenAIVectorStoreFileBatchesCreate ProviderOperation = "openai_vector_store_file_batches_create"
	ProviderOperationOpenAIVectorStoreFileBatchesList   ProviderOperation = "openai_vector_store_file_batches_list"
	ProviderOperationOpenAIVectorStoreFileBatchesGet    ProviderOperation = "openai_vector_store_file_batches_get"
	ProviderOperationOpenAIVectorStoreFileBatchesCancel ProviderOperation = "openai_vector_store_file_batches_cancel"
	ProviderOperationOpenAIContainersCreate             ProviderOperation = "openai_containers_create"
	ProviderOperationOpenAIContainersList               ProviderOperation = "openai_containers_list"
	ProviderOperationOpenAIContainersGet                ProviderOperation = "openai_containers_get"
	ProviderOperationOpenAIContainersDelete             ProviderOperation = "openai_containers_delete"
	ProviderOperationOpenAIContainerFilesCreate         ProviderOperation = "openai_container_files_create"
	ProviderOperationOpenAIContainerFilesList           ProviderOperation = "openai_container_files_list"
	ProviderOperationOpenAIContainerFilesGet            ProviderOperation = "openai_container_files_get"
	ProviderOperationOpenAIContainerFilesDelete         ProviderOperation = "openai_container_files_delete"
	ProviderOperationOpenAIContainerFilesContent        ProviderOperation = "openai_container_files_content"
	ProviderOperationOpenAIVideosCreate                 ProviderOperation = "openai_videos_create"
	ProviderOperationOpenAIVideosList                   ProviderOperation = "openai_videos_list"
	ProviderOperationOpenAIVideosGet                    ProviderOperation = "openai_videos_get"
	ProviderOperationOpenAIVideosDelete                 ProviderOperation = "openai_videos_delete"
	ProviderOperationOpenAIVideosContent                ProviderOperation = "openai_videos_content"
	ProviderOperationOpenAIVideosEdit                   ProviderOperation = "openai_videos_edit"
	ProviderOperationOpenAIVideosExtend                 ProviderOperation = "openai_videos_extend"
	ProviderOperationOpenAIVideosRemix                  ProviderOperation = "openai_videos_remix"
	ProviderOperationOpenAIModelsList                   ProviderOperation = "openai_models_list"
	ProviderOperationOpenAIModelsGet                    ProviderOperation = "openai_models_get"
	ProviderOperationOpenAIRealtime                     ProviderOperation = "openai_realtime"
	ProviderOperationOpenAIRealtimeTranscription        ProviderOperation = "openai_realtime_transcription"
	ProviderOperationOpenAIRealtimeTranslation          ProviderOperation = "openai_realtime_translation"
	ProviderOperationAnthropicMessages                  ProviderOperation = "anthropic_messages"
	ProviderOperationAnthropicCountTokens               ProviderOperation = "anthropic_count_tokens"
	ProviderOperationAnthropicModelsList                ProviderOperation = "anthropic_models_list"
	ProviderOperationAnthropicModelsGet                 ProviderOperation = "anthropic_models_get"
	ProviderOperationAnthropicFilesCreate               ProviderOperation = "anthropic_files_create"
	ProviderOperationAnthropicFilesList                 ProviderOperation = "anthropic_files_list"
	ProviderOperationAnthropicFilesGet                  ProviderOperation = "anthropic_files_get"
	ProviderOperationAnthropicFilesDelete               ProviderOperation = "anthropic_files_delete"
	ProviderOperationAnthropicFilesContent              ProviderOperation = "anthropic_files_content"
	ProviderOperationAnthropicBatchesCreate             ProviderOperation = "anthropic_batches_create"
	ProviderOperationAnthropicBatchesList               ProviderOperation = "anthropic_batches_list"
	ProviderOperationAnthropicBatchesGet                ProviderOperation = "anthropic_batches_get"
	ProviderOperationAnthropicBatchesCancel             ProviderOperation = "anthropic_batches_cancel"
	ProviderOperationAnthropicBatchesResults            ProviderOperation = "anthropic_batches_results"
	ProviderOperationGeminiGenerateContent              ProviderOperation = "gemini_generate_content"
	ProviderOperationGeminiStreamGenerateContent        ProviderOperation = "gemini_stream_generate_content"
	ProviderOperationGeminiCountTokens                  ProviderOperation = "gemini_count_tokens"
	ProviderOperationGeminiEmbedContent                 ProviderOperation = "gemini_embed_content"
	ProviderOperationGeminiBatchEmbedContents           ProviderOperation = "gemini_batch_embed_contents"
	ProviderOperationGeminiInteractionsCreate           ProviderOperation = "gemini_interactions_create"
	ProviderOperationGeminiInteractionsGet              ProviderOperation = "gemini_interactions_get"
	ProviderOperationGeminiInteractionsDelete           ProviderOperation = "gemini_interactions_delete"
	ProviderOperationGeminiInteractionsCancel           ProviderOperation = "gemini_interactions_cancel"
	ProviderOperationGeminiModelsList                   ProviderOperation = "gemini_models_list"
	ProviderOperationGeminiModelsGet                    ProviderOperation = "gemini_models_get"
	ProviderOperationGeminiFilesCreate                  ProviderOperation = "gemini_files_create"
	ProviderOperationGeminiFilesList                    ProviderOperation = "gemini_files_list"
	ProviderOperationGeminiFilesGet                     ProviderOperation = "gemini_files_get"
	ProviderOperationGeminiFilesDelete                  ProviderOperation = "gemini_files_delete"
	ProviderOperationGeminiCachesCreate                 ProviderOperation = "gemini_caches_create"
	ProviderOperationGeminiCachesList                   ProviderOperation = "gemini_caches_list"
	ProviderOperationGeminiCachesGet                    ProviderOperation = "gemini_caches_get"
	ProviderOperationGeminiCachesUpdate                 ProviderOperation = "gemini_caches_update"
	ProviderOperationGeminiCachesDelete                 ProviderOperation = "gemini_caches_delete"
	ProviderOperationGeminiBatchesCreate                ProviderOperation = "gemini_batches_create"
	ProviderOperationGeminiBatchesList                  ProviderOperation = "gemini_batches_list"
	ProviderOperationGeminiBatchesGet                   ProviderOperation = "gemini_batches_get"
	ProviderOperationGeminiBatchesCancel                ProviderOperation = "gemini_batches_cancel"
	ProviderOperationMistralFIM                         ProviderOperation = "mistral_fim"
	ProviderOperationMistralChatModerations             ProviderOperation = "mistral_chat_moderations"
	ProviderOperationMistralOCR                         ProviderOperation = "mistral_ocr"
	ProviderOperationMistralFilesCreate                 ProviderOperation = "mistral_files_create"
	ProviderOperationMistralFilesList                   ProviderOperation = "mistral_files_list"
	ProviderOperationMistralFilesGet                    ProviderOperation = "mistral_files_get"
	ProviderOperationMistralFilesDelete                 ProviderOperation = "mistral_files_delete"
	ProviderOperationMistralFilesContent                ProviderOperation = "mistral_files_content"
	ProviderOperationMistralFilesURL                    ProviderOperation = "mistral_files_url"
	ProviderOperationMistralBatchCreate                 ProviderOperation = "mistral_batch_create"
	ProviderOperationMistralBatchList                   ProviderOperation = "mistral_batch_list"
	ProviderOperationMistralBatchGet                    ProviderOperation = "mistral_batch_get"
	ProviderOperationMistralBatchCancel                 ProviderOperation = "mistral_batch_cancel"
	ProviderOperationMistralBatchDelete                 ProviderOperation = "mistral_batch_delete"
	ProviderOperationMistralAgentsCreate                ProviderOperation = "mistral_agents_create"
	ProviderOperationMistralAgentsList                  ProviderOperation = "mistral_agents_list"
	ProviderOperationMistralAgentsPages                 ProviderOperation = "mistral_agents_pages"
	ProviderOperationMistralAgentsGet                   ProviderOperation = "mistral_agents_get"
	ProviderOperationMistralAgentsUpdate                ProviderOperation = "mistral_agents_update"
	ProviderOperationMistralAgentsDelete                ProviderOperation = "mistral_agents_delete"
	ProviderOperationMistralConversationsCreate         ProviderOperation = "mistral_conversations_create"
	ProviderOperationMistralConversationsList           ProviderOperation = "mistral_conversations_list"
	ProviderOperationMistralConversationsGet            ProviderOperation = "mistral_conversations_get"
	ProviderOperationMistralConversationsAppend         ProviderOperation = "mistral_conversations_append"
	ProviderOperationMistralConversationsDelete         ProviderOperation = "mistral_conversations_delete"
	ProviderOperationMistralConversationsHistory        ProviderOperation = "mistral_conversations_history"
	ProviderOperationMistralConversationsMessages       ProviderOperation = "mistral_conversations_messages"
	ProviderOperationMistralConversationsRestart        ProviderOperation = "mistral_conversations_restart"
)

// ProviderOperationInput carries only identifiers and pagination fields used
// by fixed provider operations. It never accepts an upstream path or method.
type ProviderOperationInput struct {
	ResourceID          string   `json:"resource_id,omitempty"`
	ParentResourceID    string   `json:"parent_resource_id,omitempty"`
	SecondaryResourceID string   `json:"secondary_resource_id,omitempty"`
	BeforeID            string   `json:"before_id,omitempty"`
	AfterID             string   `json:"after_id,omitempty"`
	Limit               int      `json:"limit,omitempty"`
	Order               string   `json:"order,omitempty"`
	Purpose             string   `json:"purpose,omitempty"`
	PaginationToken     string   `json:"pagination_token,omitempty"`
	SortBy              string   `json:"sort_by,omitempty"`
	Filter              string   `json:"filter,omitempty"`
	Include             []string `json:"include,omitempty"`
	IncludeObfuscation  *bool    `json:"include_obfuscation,omitempty"`
	StartingAfter       *int64   `json:"starting_after,omitempty"`
	Stream              *bool    `json:"stream,omitempty"`
	Page                *int     `json:"page,omitempty"`
	PageSize            *int     `json:"page_size,omitempty"`
	CreatedByMe         *bool    `json:"created_by_me,omitempty"`
	OrderBy             string   `json:"order_by,omitempty"`
	Status              string   `json:"status,omitempty"`
	Model               string   `json:"model,omitempty"`
}

type AnthropicNativeResourceMapping struct {
	PublicID          string              `json:"public_id"`
	TenantID          string              `json:"tenant_id"`
	ProviderAccountID string              `json:"provider_account_id"`
	UpstreamID        string              `json:"-"`
	Status            string              `json:"status"`
	ReservationID     string              `json:"reservation_id,omitempty"`
	Reservation       *BudgetReservation  `json:"reservation,omitempty"`
	PricingRevision   string              `json:"pricing_revision,omitempty"`
	EstimatedUsage    *UsageCostBreakdown `json:"estimated_usage,omitempty"`
}

// ProviderNativeResourceMapping is the provider-neutral identity and billing
// binding for stateful upstream resources.
type ProviderNativeResourceMapping struct {
	PublicID          string                 `json:"public_id"`
	TenantID          string                 `json:"tenant_id"`
	Provider          string                 `json:"provider"`
	ProviderAccountID string                 `json:"provider_account_id"`
	ResourceKind      string                 `json:"resource_kind"`
	UpstreamID        string                 `json:"-"`
	Status            string                 `json:"status"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	ReservationID     string                 `json:"reservation_id,omitempty"`
	Reservation       *BudgetReservation     `json:"reservation,omitempty"`
	PricingRevision   string                 `json:"pricing_revision,omitempty"`
	EstimatedUsage    *UsageCostBreakdown    `json:"estimated_usage,omitempty"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty"`
}

type TunnelRequest struct {
	RequestID            string                       `json:"request_id"`
	TenantID             string                       `json:"tenant_id"`
	TraceID              string                       `json:"trace_id"`
	Stream               bool                         `json:"stream"`
	Operation            ProviderOperation            `json:"provider_operation,omitempty"`
	OperationInput       ProviderOperationInput       `json:"provider_operation_input,omitempty"`
	Deployment           Deployment                   `json:"deployment"`
	ModelRuntimeProfiles []ModelRuntimeProfileBinding `json:"model_runtime_profiles,omitempty"`
	Headers              map[string]string            `json:"headers,omitempty"`
	Body                 []byte                       `json:"body"`
	RequestMetadata      map[string]interface{}       `json:"request_metadata,omitempty"`
}

// TunnelResponse is the final provider response.
type TunnelResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       []byte            `json:"body"`
}

type TunnelRawResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       io.ReadCloser
}

type TunnelWebSocketFrame struct {
	Type int    `json:"type"`
	Data []byte `json:"data,omitempty"`
}

type TunnelWebSocket interface {
	Read(ctx context.Context) (TunnelWebSocketFrame, error)
	Write(ctx context.Context, frame TunnelWebSocketFrame) error
	Close(code int, reason string) error
}

type TunnelWebSocketTransport interface {
	OpenWebSocket(ctx context.Context, req TunnelRequest) (TunnelWebSocket, error)
}

// TunnelChunk is a streaming chunk flowing through the tunnel path.
type TunnelChunk struct {
	Data   []byte `json:"data"`
	IsDone bool   `json:"is_done"`
}

// TunnelStream is the streaming transport contract.
type TunnelStream interface {
	Chunks() <-chan TunnelChunk
	Errors() <-chan error
	Close() error
	Headers() map[string]string
	StatusCode() int
}

// RequestLogEntry records the normalized request lifecycle.
type RequestLogEntry struct {
	LogID                    string                 `json:"log_id,omitempty"`
	RequestID                string                 `json:"request_id"`
	TraceID                  string                 `json:"trace_id,omitempty"`
	TokenHash                string                 `json:"token_hash,omitempty"`
	TenantID                 string                 `json:"tenant_id"`
	TeamID                   string                 `json:"team_id,omitempty"`
	ActorType                string                 `json:"actor_type,omitempty"`
	ActorID                  string                 `json:"actor_id,omitempty"`
	Model                    string                 `json:"model"`
	Provider                 string                 `json:"provider,omitempty"`
	ProviderModel            string                 `json:"provider_model,omitempty"`
	VirtualModel             string                 `json:"virtual_model,omitempty"`
	Route                    string                 `json:"route,omitempty"`
	RequestType              string                 `json:"request_type,omitempty"`
	Status                   string                 `json:"status"`
	StatusCode               int                    `json:"status_code,omitempty"`
	ErrorType                string                 `json:"error_type,omitempty"`
	CacheStatus              string                 `json:"cache_status,omitempty"`
	CacheMode                string                 `json:"cache_mode,omitempty"`
	RetryCount               int                    `json:"retry_count,omitempty"`
	FallbackCount            int                    `json:"fallback_count,omitempty"`
	FallbackStatus           string                 `json:"fallback_status,omitempty"`
	LoadBalanceStatus        string                 `json:"load_balance_status,omitempty"`
	CostUSD                  float64                `json:"cost_usd"`
	CostIRR                  float64                `json:"cost_irr,omitempty"`
	InputTokens              int                    `json:"input_tokens,omitempty"`
	OutputTokens             int                    `json:"output_tokens,omitempty"`
	TotalTokens              int                    `json:"total_tokens,omitempty"`
	ResponseTimeMS           int                    `json:"response_time_ms"`
	StartedAt                time.Time              `json:"started_at,omitempty"`
	RequestBody              map[string]interface{} `json:"request_body,omitempty"`
	ResponseBody             map[string]interface{} `json:"response_body,omitempty"`
	Metadata                 map[string]interface{} `json:"metadata,omitempty"`
	RequestPayloadTruncated  bool                   `json:"request_payload_truncated,omitempty"`
	ResponsePayloadTruncated bool                   `json:"response_payload_truncated,omitempty"`
}

// VirtualModel defines a tenant-scoped alias that routes to concrete provider targets.
type VirtualModel struct {
	ID              string                             `json:"id"`
	TenantID        string                             `json:"tenant_id"`
	Alias           string                             `json:"alias"`
	RoutingStrategy string                             `json:"routing_strategy"`
	Enabled         bool                               `json:"enabled"`
	ExposedModes    []string                           `json:"exposed_modes"`
	Version         int                                `json:"version"`
	Targets         []VirtualModelTarget               `json:"targets"`
	Access          TenantProviderAccountAccessControl `json:"access"`
	CreatedBy       string                             `json:"created_by,omitempty"`
	UpdatedBy       string                             `json:"updated_by,omitempty"`
	CreatedAt       time.Time                          `json:"created_at"`
	UpdatedAt       time.Time                          `json:"updated_at"`
}

// VirtualModelSnapshot pins one virtual model alias to a static list of target FQNs for PAT enforcement.
type VirtualModelSnapshot struct {
	Alias          string   `json:"alias"`
	VirtualModelID string   `json:"virtual_model_id"`
	Version        int      `json:"version"`
	TargetFQNs     []string `json:"target_fqns"`
	ExposedModes   []string `json:"exposed_modes,omitempty"`
}

type VirtualModelTarget struct {
	ID       string             `json:"id"`
	Order    int                `json:"order"`
	Weight   float64            `json:"weight,omitempty"`
	Provider string             `json:"provider"`
	Model    string             `json:"model"`
	Policy   VirtualModelPolicy `json:"policy"`
}

type VirtualModelPolicy struct {
	TimeoutSec float64                    `json:"timeout_sec,omitempty"`
	Retry      VirtualModelRetryPolicy    `json:"retry"`
	Fallback   VirtualModelFallbackPolicy `json:"fallback"`
}

type VirtualModelRetryPolicy struct {
	Attempts    int   `json:"attempts"`
	DelayMS     int   `json:"delay_ms"`
	StatusCodes []int `json:"status_codes,omitempty"`
}

type VirtualModelFallbackPolicy struct {
	StatusCodes []int `json:"status_codes,omitempty"`
	Candidate   bool  `json:"candidate"`
}

type CreateVirtualModelRequest struct {
	Alias           string                             `json:"alias"`
	RoutingStrategy string                             `json:"routing_strategy"`
	Enabled         *bool                              `json:"enabled,omitempty"`
	ExposedModes    []string                           `json:"exposed_modes,omitempty"`
	Targets         []VirtualModelTarget               `json:"targets"`
	Access          TenantProviderAccountAccessControl `json:"access"`
}

type UpdateVirtualModelRequest struct {
	Alias           *string                             `json:"alias,omitempty"`
	RoutingStrategy *string                             `json:"routing_strategy,omitempty"`
	Enabled         *bool                               `json:"enabled,omitempty"`
	ExposedModes    *[]string                           `json:"exposed_modes,omitempty"`
	Targets         []VirtualModelTarget                `json:"targets,omitempty"`
	Access          *TenantProviderAccountAccessControl `json:"access,omitempty"`
}

// TenantSeatUsage reports how many of the plan's user seats a tenant uses.
// Seats are active members plus pending, unexpired invitations. Limit is nil
// when the plan has no user cap.
type TenantSeatUsage struct {
	Used               int  `json:"used"`
	ActiveMembers      int  `json:"active_members"`
	PendingInvitations int  `json:"pending_invitations"`
	Limit              *int `json:"limit"`
}
