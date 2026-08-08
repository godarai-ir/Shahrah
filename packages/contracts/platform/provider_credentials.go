package platform

import "time"

const (
	CredentialSourcePlatformManaged   = "platform_managed"
	CredentialSourceEnvBreakGlass     = "platform_env_break_glass"
	ProviderCredentialModeEnv         = "env"
	ProviderCredentialModeStore       = "platform_store"
	ProviderCredentialModeBreakGlass  = "platform_store_with_env_break_glass"
	ProviderCredentialEnvironmentProd = "production"
)

// ResolvedProviderAuth is the typed provider authentication material carried
// only on the internal provider-egress dispatch. It must never be used in metadata,
// logs, traces, audits, analytics, or public/admin responses.
type ResolvedProviderAuth struct {
	Mode   string `json:"mode"`
	APIKey string `json:"api_key,omitempty"`
	// OAuthAccessToken is an already-refreshed, short-lived token. It is only
	// carried in the internal provider-egress request and must never be persisted.
	OAuthAccessToken string                    `json:"oauth_access_token,omitempty"`
	AnthropicWIF     *ResolvedAnthropicWIFAuth `json:"anthropic_wif,omitempty"`
}

// ResolvedAnthropicWIFAuth contains identifiers only. Silk Road resolves the
// allowlisted token source and performs token exchange in the Europe runtime.
// Neither an OIDC assertion nor an exchanged access token is serialized back.
type ResolvedAnthropicWIFAuth struct {
	FederationRuleID string `json:"federation_rule_id"`
	OrganizationID   string `json:"organization_id"`
	ServiceAccountID string `json:"service_account_id"`
	WorkspaceID      string `json:"workspace_id,omitempty"`
	TokenSourceID    string `json:"token_source_id"`
}

type ProviderConnectionField struct {
	Key              string                 `json:"key"`
	Label            string                 `json:"label"`
	Type             string                 `json:"type"`
	Secret           bool                   `json:"secret"`
	Required         bool                   `json:"required,omitempty"`
	RequiredWhen     map[string]interface{} `json:"required_when,omitempty"`
	Validation       map[string]interface{} `json:"validation,omitempty"`
	Scope            string                 `json:"scope"`
	RuntimeMapping   string                 `json:"runtime_mapping,omitempty"`
	HelpText         string                 `json:"help_text,omitempty"`
	DocumentationURL string                 `json:"documentation_url,omitempty"`
	RedactionPolicy  string                 `json:"redaction_policy,omitempty"`
}

type ProviderConnectionSchema struct {
	ProviderSlug string                    `json:"provider_slug"`
	AuthMode     string                    `json:"auth_mode"`
	Version      int                       `json:"version"`
	RuntimeReady bool                      `json:"runtime_ready"`
	Fields       []ProviderConnectionField `json:"fields"`
}

type PlatformProviderConnectionScope struct {
	ID                string   `json:"id,omitempty"`
	ModelPattern      string   `json:"model_pattern,omitempty"`
	DeploymentPattern string   `json:"deployment_pattern,omitempty"`
	Capabilities      []string `json:"capabilities,omitempty"`
	Regions           []string `json:"regions,omitempty"`
	Enabled           bool     `json:"enabled"`
}

type PlatformProviderCredentialVersion struct {
	ID          string     `json:"id"`
	Version     int64      `json:"version"`
	Backend     string     `json:"secret_backend"`
	Configured  bool       `json:"configured"`
	DisplayHint string     `json:"display_hint,omitempty"`
	Status      string     `json:"status"`
	ValidFrom   *time.Time `json:"valid_from,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	RetireAt    *time.Time `json:"retire_at,omitempty"`
	CreatedBy   string     `json:"created_by,omitempty"`
	ActivatedBy string     `json:"activated_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type PlatformProviderConnection struct {
	ID                        string                              `json:"id"`
	ProviderSlug              string                              `json:"provider_slug"`
	Name                      string                              `json:"name"`
	Environment               string                              `json:"environment"`
	AuthMode                  string                              `json:"auth_mode"`
	SchemaVersion             int                                 `json:"schema_version"`
	Config                    map[string]interface{}              `json:"config"`
	Status                    string                              `json:"status"`
	Region                    string                              `json:"region,omitempty"`
	Priority                  int                                 `json:"priority"`
	Weight                    float64                             `json:"weight"`
	IsPrimary                 bool                                `json:"is_primary"`
	ActiveCredentialVersionID string                              `json:"active_credential_version_id,omitempty"`
	ActiveVersion             *PlatformProviderCredentialVersion  `json:"active_version,omitempty"`
	Versions                  []PlatformProviderCredentialVersion `json:"credential_versions,omitempty"`
	Scopes                    []PlatformProviderConnectionScope   `json:"scopes,omitempty"`
	LastTestStatus            string                              `json:"last_test_status,omitempty"`
	LastTestCategory          string                              `json:"last_test_category,omitempty"`
	LastTestSummary           string                              `json:"last_test_summary,omitempty"`
	LastTestedAt              *time.Time                          `json:"last_tested_at,omitempty"`
	LastSuccessAt             *time.Time                          `json:"last_success_at,omitempty"`
	LastRotationAt            *time.Time                          `json:"last_rotation_at,omitempty"`
	Revision                  int64                               `json:"revision"`
	CreatedBy                 string                              `json:"created_by,omitempty"`
	UpdatedBy                 string                              `json:"updated_by,omitempty"`
	CreatedAt                 time.Time                           `json:"created_at"`
	UpdatedAt                 time.Time                           `json:"updated_at"`
}

type CreatePlatformProviderConnectionRequest struct {
	ProviderSlug  string                            `json:"provider_slug"`
	Name          string                            `json:"name"`
	Environment   string                            `json:"environment"`
	AuthMode      string                            `json:"auth_mode"`
	SchemaVersion int                               `json:"schema_version"`
	Config        map[string]interface{}            `json:"config"`
	Region        string                            `json:"region,omitempty"`
	Priority      int                               `json:"priority,omitempty"`
	Weight        float64                           `json:"weight,omitempty"`
	IsPrimary     bool                              `json:"is_primary,omitempty"`
	Scopes        []PlatformProviderConnectionScope `json:"scopes,omitempty"`
}

type UpdatePlatformProviderConnectionRequest struct {
	Name      *string                            `json:"name,omitempty"`
	Config    map[string]interface{}             `json:"config,omitempty"`
	Region    *string                            `json:"region,omitempty"`
	Priority  *int                               `json:"priority,omitempty"`
	Weight    *float64                           `json:"weight,omitempty"`
	IsPrimary *bool                              `json:"is_primary,omitempty"`
	Scopes    *[]PlatformProviderConnectionScope `json:"scopes,omitempty"`
	Revision  int64                              `json:"revision"`
}

type CreatePlatformProviderCredentialVersionRequest struct {
	Secrets   map[string]string `json:"secrets"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
}

// ProviderCredentialProbeDetails contains sanitized credential readiness data.
type ProviderCredentialProbeDetails struct {
	Authenticated      bool     `json:"authenticated"`
	Ready              bool     `json:"ready"`
	Permissions        []string `json:"permissions,omitempty"`
	CredentialDisabled *bool    `json:"credential_disabled,omitempty"`
	CredentialBlocked  *bool    `json:"credential_blocked,omitempty"`
	AccountBlocked     *bool    `json:"account_blocked,omitempty"`
}

type PlatformProviderCredentialTestResult struct {
	Success       bool                            `json:"success"`
	Authenticated bool                            `json:"authenticated"`
	Ready         bool                            `json:"ready"`
	Category      string                          `json:"category"`
	Summary       string                          `json:"summary"`
	TestedAt      time.Time                       `json:"tested_at"`
	Version       int64                           `json:"version"`
	Validated     bool                            `json:"validated"`
	Details       *ProviderCredentialProbeDetails `json:"details,omitempty"`
}

type PlatformProviderCredentialActionRequest struct {
	Revision int64  `json:"revision"`
	Reason   string `json:"reason,omitempty"`
}
