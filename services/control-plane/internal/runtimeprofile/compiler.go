package runtimeprofile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type SourceRepository interface {
	GetTenantByID(ctx context.Context, tenantID string) (*platform.Tenant, error)
	ListRuntimeAuthContexts(ctx context.Context, tenantID string) ([]platform.AuthContext, error)
	ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error)
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
	ListEnabledTenantGuardrails(ctx context.Context, tenantID string) ([]platform.TenantGuardrail, error)
	ResolveEffectiveEntitlementLimits(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementLimits, error)
	ListRateLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.RateLimitRule, int, error)
	ListBudgetLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.BudgetLimitRule, int, error)
}

// SeatSource is implemented by sources that know the tenant's member seats.
// It is optional so profiles compile without seats in tests and tools.
type SeatSource interface {
	GetSeatProfile(ctx context.Context, tenantID string) (*platform.SeatProfile, error)
}

type Compiler struct {
	repo SourceRepository
	now  func() time.Time
}

func NewCompiler(repo SourceRepository) *Compiler {
	return &Compiler{repo: repo, now: time.Now}
}

func (c *Compiler) Compile(ctx context.Context, tenantID string, version int64) (*Profile, error) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, ErrProfileUnavailable
	}

	if version <= 0 {
		version = 1
	}

	tenant, err := c.repo.GetTenantByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	authContexts, err := c.repo.ListRuntimeAuthContexts(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	virtualModels, err := c.repo.ListVirtualModels(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	deployments, err := c.repo.ListDeployments(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	guardrails, err := c.repo.ListEnabledTenantGuardrails(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	entitlements, err := c.repo.ResolveEffectiveEntitlementLimits(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("resolve runtime entitlements: %w", err)
	}

	rateRules, _, err := c.repo.ListRateLimitRules(ctx, tenantID, "", 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("list runtime rate-limit rules: %w", err)
	}

	budgetRules, _, err := c.repo.ListBudgetLimitRules(ctx, tenantID, "", 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("list runtime budget-limit rules: %w", err)
	}

	var seats *platform.SeatProfile

	if seatSource, ok := c.repo.(SeatSource); ok {
		seats, err = seatSource.GetSeatProfile(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("load runtime seats: %w", err)
		}
	}

	sort.SliceStable(rateRules, func(i, j int) bool {
		if rateRules[i].Priority == rateRules[j].Priority {
			return runtimeRateLimitRuleID(rateRules[i]) < runtimeRateLimitRuleID(rateRules[j])
		}

		return rateRules[i].Priority < rateRules[j].Priority
	})
	sort.SliceStable(budgetRules, func(i, j int) bool {
		if budgetRules[i].Priority == budgetRules[j].Priority {
			return runtimeBudgetRuleID(budgetRules[i]) < runtimeBudgetRuleID(budgetRules[j])
		}

		return budgetRules[i].Priority < budgetRules[j].Priority
	})

	now := c.now().UTC()
	profile := &Profile{
		SchemaVersion:   ProfileSchemaVersion,
		TenantID:        tenantID,
		Version:         version,
		CompiledAt:      now,
		SourceUpdatedAt: now,
		Tenant:          *tenant,
		AuthByTokenHash: make(map[string]platform.AuthContext, len(authContexts)),
		VirtualModels:   make(map[string]platform.VirtualModel, len(virtualModels)),
		Deployments:     make([]platform.Deployment, 0, len(deployments)),
		Guardrails:      make(map[string]platform.TenantGuardrail, len(guardrails)),
		Entitlements:    entitlements,
		RateLimitRules:  rateRules,
		BudgetRules:     budgetRules,
		Seats:           seats,
		ProviderAPIKeys: make(map[string]string),
	}

	for _, authContext := range authContexts {
		tokenHash := strings.TrimSpace(authContext.TokenHash)
		if tokenHash == "" {
			continue
		}

		profile.AuthByTokenHash[tokenHash] = cloneAuthContext(authContext)
	}

	for _, vm := range virtualModels {
		if !vm.Enabled {
			continue
		}

		alias := strings.TrimSpace(vm.Alias)
		if alias == "" {
			continue
		}

		profile.VirtualModels[alias] = cloneVirtualModel(vm)
	}

	for _, dep := range deployments {
		modelName := strings.TrimSpace(dep.ModelName)
		if modelName == "" {
			continue
		}

		apiKey := strings.TrimSpace(dep.APIKey)
		dep.APIKey = ""

		if apiKey != "" {
			profile.ProviderAPIKeys[modelName] = apiKey
		}

		profile.Deployments = append(profile.Deployments, dep)
	}

	for _, guardrail := range guardrails {
		slug := strings.TrimSpace(guardrail.Slug)
		if slug == "" || !guardrail.Enabled {
			continue
		}

		profile.Guardrails[slug] = cloneGuardrail(guardrail)
	}

	profile.Checksum = checksumProfile(profile)

	return profile, nil
}

func sortProfilePolicies(profile *Profile) {
	if profile == nil {
		return
	}

	sort.SliceStable(profile.RateLimitRules, func(i, j int) bool {
		if profile.RateLimitRules[i].Priority == profile.RateLimitRules[j].Priority {
			return runtimeRateLimitRuleID(profile.RateLimitRules[i]) < runtimeRateLimitRuleID(profile.RateLimitRules[j])
		}

		return profile.RateLimitRules[i].Priority < profile.RateLimitRules[j].Priority
	})
	sort.SliceStable(profile.BudgetRules, func(i, j int) bool {
		if profile.BudgetRules[i].Priority == profile.BudgetRules[j].Priority {
			return runtimeBudgetRuleID(profile.BudgetRules[i]) < runtimeBudgetRuleID(profile.BudgetRules[j])
		}

		return profile.BudgetRules[i].Priority < profile.BudgetRules[j].Priority
	})
}

func runtimeRateLimitRuleID(rule platform.RateLimitRule) string {
	if strings.TrimSpace(rule.RuleID) != "" {
		return strings.TrimSpace(rule.RuleID)
	}

	return strings.TrimSpace(rule.ID)
}

func runtimeBudgetRuleID(rule platform.BudgetLimitRule) string {
	if strings.TrimSpace(rule.RuleID) != "" {
		return strings.TrimSpace(rule.RuleID)
	}

	return strings.TrimSpace(rule.ID)
}

func checksumProfile(profile *Profile) string {
	if profile == nil {
		return ""
	}

	redacted := cloneProfile(profile)
	redacted.ProviderAPIKeys = nil

	raw, err := json.Marshal(redacted)
	if err != nil {
		return ""
	}

	sum := sha256.Sum256(raw)

	return hex.EncodeToString(sum[:])
}
