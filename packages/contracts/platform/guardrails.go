package platform

import "time"

type GuardrailOperation string

const (
	GuardrailOperationValidate GuardrailOperation = "validate"
	GuardrailOperationMutate   GuardrailOperation = "mutate"
)

type GuardrailEnforcement string

const (
	GuardrailEnforcementAudit              GuardrailEnforcement = "audit"
	GuardrailEnforcementEnforce            GuardrailEnforcement = "enforce"
	GuardrailEnforcementEnforceIgnoreError GuardrailEnforcement = "enforce_ignore_error"
)

type TenantGuardrail struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	Slug        string                 `json:"slug"`
	Selector    string                 `json:"selector"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"`
	Operation   GuardrailOperation     `json:"operation"`
	Enforcement GuardrailEnforcement   `json:"enforcement"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	CreatedBy   string                 `json:"created_by,omitempty"`
	UpdatedBy   string                 `json:"updated_by,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type CreateTenantGuardrailRequest struct {
	Slug        string                 `json:"slug"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"`
	Operation   GuardrailOperation     `json:"operation"`
	Enforcement GuardrailEnforcement   `json:"enforcement"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Config      map[string]interface{} `json:"config"`
}

type UpdateTenantGuardrailRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	Type        *string                `json:"type,omitempty"`
	Operation   *GuardrailOperation    `json:"operation,omitempty"`
	Enforcement *GuardrailEnforcement  `json:"enforcement,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

type TenantGuardrailListResponse struct {
	Data       []TenantGuardrail `json:"data"`
	Total      int               `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
	NextOffset *int              `json:"next_offset,omitempty"`
}
