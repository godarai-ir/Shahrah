package runtimeprofile

import (
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const (
	RedisChannel        = "tenant_runtime:events"
	ProfileStatusActive = "active"
	ProfileStatusFailed = "failed"
)

var (
	ErrProfileUnavailable = errors.New("tenant_runtime_profile_unavailable")
	ErrAuthNotFound       = errors.New("tenant runtime auth context not found")
)

type Profile struct {
	TenantID        string                               `json:"tenant_id"`
	Version         int64                                `json:"version"`
	Checksum        string                               `json:"checksum"`
	CompiledAt      time.Time                            `json:"compiled_at"`
	SourceUpdatedAt time.Time                            `json:"source_updated_at"`
	Tenant          platform.Tenant                      `json:"tenant"`
	AuthByTokenHash map[string]platform.AuthContext      `json:"auth_by_token_hash"`
	VirtualModels   map[string]platform.VirtualModel     `json:"virtual_models"`
	Deployments     []platform.Deployment                `json:"deployments"`
	Guardrails      map[string]platform.TenantGuardrail  `json:"guardrails"`
	Entitlements    *platform.EffectiveEntitlementLimits `json:"entitlements,omitempty"`
	RateLimitRules  []platform.RateLimitRule             `json:"rate_limit_rules,omitempty"`
	BudgetRules     []platform.BudgetLimitRule           `json:"budget_rules,omitempty"`

	ProviderAPIKeys map[string]string `json:"-"`
}

type StoredProfile struct {
	TenantID        string
	Version         int64
	ProfileJSON     []byte
	Checksum        string
	CompiledAt      time.Time
	SourceUpdatedAt time.Time
	Status          string
	Error           string
}

type Event struct {
	TenantID string `json:"tenant_id"`
	Version  int64  `json:"version"`
	Reason   string `json:"reason,omitempty"`
}

func ActiveVersionKey(tenantID string) string {
	return "tenant_runtime:" + strings.TrimSpace(tenantID) + ":active_version"
}

func ProfileKey(tenantID string, version int64) string {
	return "tenant_runtime:" + strings.TrimSpace(tenantID) + ":profile:" + formatVersion(version)
}
