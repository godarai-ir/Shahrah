package runtimeprofile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	now := c.now().UTC()
	profile := &Profile{
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
