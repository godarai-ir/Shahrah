package platform

import (
	"context"
	"time"
)

// Tenant represents the root tenant boundary for Shahrah.
type Tenant struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Domain    string    `json:"domain"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthContext is the effective request auth context.
type AuthContext struct {
	TenantID       string     `json:"tenant_id"`
	OrganizationID string     `json:"organization_id,omitempty"`
	TeamID         string     `json:"team_id,omitempty"`
	KeyID          string     `json:"key_id"`
	KeyAlias       string     `json:"key_alias,omitempty"`
	Role           string     `json:"role"`
	AllowedModels  []string   `json:"allowed_models,omitempty"`
	Spend          float64    `json:"spend"`
	MaxBudget      *float64   `json:"max_budget,omitempty"`
	RPMLimit       *int64     `json:"rpm_limit,omitempty"`
	TPMLimit       *int64     `json:"tpm_limit,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	Blocked        bool       `json:"blocked"`
	TokenHash      string     `json:"token_hash,omitempty"`
}

// Wallet tracks tenant balance in IRR.
type Wallet struct {
	ID         string  `json:"id"`
	TenantID   string  `json:"tenant_id"`
	BalanceIRR float64 `json:"balance_irr"`
	PendingIRR float64 `json:"pending_irr"`
}

// Transaction records wallet changes.
type Transaction struct {
	ID           string                 `json:"id"`
	WalletID     string                 `json:"wallet_id"`
	TenantID     string                 `json:"tenant_id"`
	Type         string                 `json:"type"`
	AmountIRR    float64                `json:"amount_irr"`
	AmountUSD    *float64               `json:"amount_usd,omitempty"`
	ExchangeRate *float64               `json:"exchange_rate,omitempty"`
	Status       string                 `json:"status"`
	Description  string                 `json:"description,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
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
	OrganizationID  string                 `json:"organization_id,omitempty"`
	TeamID          string                 `json:"team_id,omitempty"`
	Model           string                 `json:"model"`
	InputTokens     int                    `json:"input_tokens"`
	OutputTokens    int                    `json:"output_tokens"`
	TotalTokens     int                    `json:"total_tokens"`
	CostUSD         float64                `json:"cost_usd"`
	CostIRR         float64                `json:"cost_irr"`
	ResponseTimeMS  int                    `json:"response_time_ms"`
	Status          string                 `json:"status"`
	RequestMetadata map[string]interface{} `json:"request_metadata,omitempty"`
	CompletionStart time.Time              `json:"completion_start_time"`
	CompletionEnd   time.Time              `json:"completion_end_time"`
}

// UsageCostBreakdown captures estimated or final cost.
type UsageCostBreakdown struct {
	Model         string  `json:"model"`
	InputTokens   int     `json:"input_tokens"`
	OutputTokens  int     `json:"output_tokens"`
	TotalTokens   int     `json:"total_tokens"`
	InputCostUSD  float64 `json:"input_cost_usd"`
	OutputCostUSD float64 `json:"output_cost_usd"`
	TotalCostUSD  float64 `json:"total_cost_usd"`
	ExchangeRate  float64 `json:"exchange_rate"`
	TotalCostIRR  float64 `json:"total_cost_irr"`
}

// Deployment represents a configured provider target.
type Deployment struct {
	ID         string                 `json:"id"`
	TenantID   string                 `json:"tenant_id,omitempty"`
	ModelName  string                 `json:"model_name"`
	Provider   string                 `json:"provider"`
	Model      string                 `json:"model"`
	APIBase    string                 `json:"api_base,omitempty"`
	APIKey     string                 `json:"api_key,omitempty"`
	APIVersion string                 `json:"api_version,omitempty"`
	RegionName string                 `json:"region_name,omitempty"`
	RPM        int64                  `json:"rpm,omitempty"`
	TPM        int64                  `json:"tpm,omitempty"`
	Weight     float64                `json:"weight,omitempty"`
	TimeoutSec float64                `json:"timeout,omitempty"`
	ModelInfo  map[string]interface{} `json:"model_info,omitempty"`
}

// TenantStore resolves tenants for requests.
type TenantStore interface {
	GetTenantByHost(ctx context.Context, host string) (*Tenant, error)
	EnsureTenant(ctx context.Context, slug string) (*Tenant, error)
}

// KeyStore resolves and persists virtual keys.
type KeyStore interface {
	GetKeyByHash(ctx context.Context, hash string) (*AuthContext, error)
	CreateKey(ctx context.Context, params CreateKeyParams) (*CreatedKey, error)
	ListKeys(ctx context.Context, tenantID string) ([]CreatedKey, error)
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

// ProviderForwarder forwards provider requests on the Europe side.
type ProviderForwarder interface {
	Forward(ctx context.Context, req TunnelRequest) (*TunnelResponse, error)
	ForwardStream(ctx context.Context, req TunnelRequest) (TunnelStream, error)
}

// CreateKeyParams are the inputs for creating a virtual key.
type CreateKeyParams struct {
	TenantID  string   `json:"tenant_id"`
	KeyAlias  string   `json:"key_alias"`
	Role      string   `json:"role"`
	Models    []string `json:"models"`
	MaxBudget *float64 `json:"max_budget,omitempty"`
	RPMLimit  *int64   `json:"rpm_limit,omitempty"`
	TPMLimit  *int64   `json:"tpm_limit,omitempty"`
	TeamID    string   `json:"team_id,omitempty"`
	OrgID     string   `json:"organization_id,omitempty"`
	ExpiresIn string   `json:"expires_in,omitempty"`
}

// CreatedKey is the persisted key metadata returned by admin APIs.
type CreatedKey struct {
	ID        string     `json:"id"`
	Key       string     `json:"key,omitempty"`
	KeyAlias  string     `json:"key_alias"`
	MaskedKey string     `json:"key_name,omitempty"`
	TenantID  string     `json:"tenant_id"`
	Role      string     `json:"role"`
	Models    []string   `json:"models"`
	MaxBudget *float64   `json:"max_budget,omitempty"`
	RPMLimit  *int64     `json:"rpm_limit,omitempty"`
	TPMLimit  *int64     `json:"tpm_limit,omitempty"`
	TeamID    string     `json:"team_id,omitempty"`
	OrgID     string     `json:"organization_id,omitempty"`
	Spend     float64    `json:"spend"`
	Blocked   bool       `json:"blocked"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	TokenHash string     `json:"token_hash,omitempty"`
}

// TunnelRequest is the internal request sent over the Shahrah tunnel path.
type TunnelRequest struct {
	RequestID       string                 `json:"request_id"`
	TenantID        string                 `json:"tenant_id"`
	TraceID         string                 `json:"trace_id"`
	Stream          bool                   `json:"stream"`
	Deployment      Deployment             `json:"deployment"`
	Headers         map[string]string      `json:"headers,omitempty"`
	Body            []byte                 `json:"body"`
	RequestMetadata map[string]interface{} `json:"request_metadata,omitempty"`
}

// TunnelResponse is the final provider response.
type TunnelResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       []byte            `json:"body"`
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
}

// RequestLogEntry records the normalized request lifecycle.
type RequestLogEntry struct {
	RequestID      string                 `json:"request_id"`
	TokenHash      string                 `json:"token_hash,omitempty"`
	TenantID       string                 `json:"tenant_id"`
	Model          string                 `json:"model"`
	Status         string                 `json:"status"`
	CostUSD        float64                `json:"cost_usd"`
	ResponseTimeMS int                    `json:"response_time_ms"`
	RequestBody    map[string]interface{} `json:"request_body,omitempty"`
	ResponseBody   map[string]interface{} `json:"response_body,omitempty"`
}
