package runtimeprofile

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestCompilerBuildsProfileAndRedactsProviderSecrets(t *testing.T) {
	repo := &fakeSourceRepo{
		tenant: platform.Tenant{ID: "tenant-1", Slug: "acme", Domain: "acme.localhost", Status: "active", CreatedAt: time.Now()},
		auth: []platform.AuthContext{{
			TenantID: "tenant-1", KeyID: "key-1", TokenHash: "hash-1",
			AllowedModels: []string{"openai:main:gpt-4o"},
		}},
		virtualModels: []platform.VirtualModel{{
			ID: "vm-1", TenantID: "tenant-1", Alias: "fast-chat", Enabled: true, Version: 3,
			Targets: []platform.VirtualModelTarget{{Model: "openai:main:gpt-4o"}},
		}},
		deployments: []platform.Deployment{{
			ID: "dep-1", TenantID: "tenant-1", ModelName: "openai:main:gpt-4o",
			Provider: "openai", Model: "gpt-4o", APIKey: "sk-secret",
		}},
		guardrails: []platform.TenantGuardrail{{TenantID: "tenant-1", Slug: "mask-card", Enabled: true}},
		entitlements: &platform.EffectiveEntitlementLimits{
			TenantID: "tenant-1",
			Subscription: &platform.TenantSubscription{
				TenantID: "tenant-1",
				Status:   "active",
			},
			Features: map[string]interface{}{platform.FeatureUsersMax: float64(10)},
		},
		rateRules:   []platform.RateLimitRule{{RuleID: "rpm", Enabled: true}},
		budgetRules: []platform.BudgetLimitRule{{RuleID: "daily", Enabled: true}},
	}

	compiler := NewCompiler(repo)

	profile, err := compiler.Compile(context.Background(), "tenant-1", 7)
	if err != nil {
		t.Fatalf("Compile returned error: %v", err)
	}

	if profile.Version != 7 || profile.Tenant.Slug != "acme" {
		t.Fatalf("unexpected profile metadata: %+v", profile)
	}

	if _, ok := profile.AuthByTokenHash["hash-1"]; !ok {
		t.Fatal("expected auth context by token hash")
	}

	if vm := profile.VirtualModels["fast-chat"]; vm.Version != 3 {
		t.Fatalf("expected virtual model version 3, got %+v", vm)
	}

	if len(profile.Deployments) != 1 || profile.Deployments[0].APIKey != "" {
		t.Fatalf("expected deployment api key to be redacted, got %+v", profile.Deployments)
	}

	if got := profile.ProviderAPIKeys["openai:main:gpt-4o"]; got != "sk-secret" {
		t.Fatalf("expected in-memory provider secret, got %q", got)
	}

	raw, err := json.Marshal(profile)
	if err != nil {
		t.Fatalf("Marshal returned error: %v", err)
	}

	if strings.Contains(string(raw), "sk-secret") {
		t.Fatalf("profile JSON leaked provider secret: %s", raw)
	}

	if profile.Checksum == "" {
		t.Fatal("expected checksum")
	}
}

type fakeSourceRepo struct {
	tenant        platform.Tenant
	auth          []platform.AuthContext
	virtualModels []platform.VirtualModel
	deployments   []platform.Deployment
	guardrails    []platform.TenantGuardrail
	entitlements  *platform.EffectiveEntitlementLimits
	rateRules     []platform.RateLimitRule
	budgetRules   []platform.BudgetLimitRule
}

func (r *fakeSourceRepo) GetTenantByID(context.Context, string) (*platform.Tenant, error) {
	out := r.tenant
	return &out, nil
}

func (r *fakeSourceRepo) ListRuntimeAuthContexts(context.Context, string) ([]platform.AuthContext, error) {
	return append([]platform.AuthContext(nil), r.auth...), nil
}

func (r *fakeSourceRepo) ListVirtualModels(context.Context, string) ([]platform.VirtualModel, error) {
	return append([]platform.VirtualModel(nil), r.virtualModels...), nil
}

func (r *fakeSourceRepo) ListDeployments(context.Context, string) ([]platform.Deployment, error) {
	return append([]platform.Deployment(nil), r.deployments...), nil
}

func (r *fakeSourceRepo) ListEnabledTenantGuardrails(context.Context, string) ([]platform.TenantGuardrail, error) {
	return append([]platform.TenantGuardrail(nil), r.guardrails...), nil
}

func (r *fakeSourceRepo) ResolveEffectiveEntitlementLimits(context.Context, string) (*platform.EffectiveEntitlementLimits, error) {
	return r.entitlements, nil
}

func (r *fakeSourceRepo) ListRateLimitRules(context.Context, string, string, int, int) ([]platform.RateLimitRule, int, error) {
	return append([]platform.RateLimitRule(nil), r.rateRules...), len(r.rateRules), nil
}

func (r *fakeSourceRepo) ListBudgetLimitRules(context.Context, string, string, int, int) ([]platform.BudgetLimitRule, int, error) {
	return append([]platform.BudgetLimitRule(nil), r.budgetRules...), len(r.budgetRules), nil
}
