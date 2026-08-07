package bootstrap

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/policies"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/virtualmodels"
)

// runtimeProfileSources composes narrow feature repositories for profile compilation.
type runtimeProfileSources struct {
	tenants  *storage.TenantRepository
	auth     *storage.RuntimeProfileRepository
	models   *virtualmodels.Service
	policies *policies.Service
	plans    *storage.PlanRepository
}

func newRuntimeProfileSources(
	store *storage.Store,
	models *virtualmodels.Service,
	policyApplication *policies.Service,
) runtimeProfileSources {
	return runtimeProfileSources{
		tenants:  store.TenantRepository,
		auth:     store.RuntimeProfileRepository,
		models:   models,
		policies: policyApplication,
		plans:    store.PlanRepository,
	}
}

func (s runtimeProfileSources) GetTenantByID(ctx context.Context, tenantID string) (*platform.Tenant, error) {
	return s.tenants.GetTenantByID(ctx, tenantID)
}

func (s runtimeProfileSources) ListRuntimeAuthContexts(ctx context.Context, tenantID string) ([]platform.AuthContext, error) {
	return s.auth.ListRuntimeAuthContexts(ctx, tenantID)
}

func (s runtimeProfileSources) ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error) {
	return s.models.ListVirtualModels(ctx, tenantID)
}

func (s runtimeProfileSources) ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	return s.models.ListDeployments(ctx, tenantID)
}

func (s runtimeProfileSources) ListEnabledTenantGuardrails(ctx context.Context, tenantID string) ([]platform.TenantGuardrail, error) {
	return s.policies.ListEnabledTenantGuardrails(ctx, tenantID)
}

func (s runtimeProfileSources) ResolveEffectiveEntitlementLimits(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementLimits, error) {
	return s.plans.ResolveEffectiveEntitlementLimits(ctx, tenantID)
}

func (s runtimeProfileSources) ListRateLimitRules(
	ctx context.Context,
	tenantID string,
	query string,
	limit int,
	offset int,
) ([]platform.RateLimitRule, int, error) {
	return s.policies.ListRateLimitRules(ctx, tenantID, query, limit, offset)
}

func (s runtimeProfileSources) ListBudgetLimitRules(
	ctx context.Context,
	tenantID string,
	query string,
	limit int,
	offset int,
) ([]platform.BudgetLimitRule, int, error) {
	return s.policies.ListBudgetLimitRules(ctx, tenantID, query, limit, offset)
}
