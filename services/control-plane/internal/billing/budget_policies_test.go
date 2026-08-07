package billing

import (
	"context"
	"errors"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"go.uber.org/zap"
)

func newPolicyTestService(c *cache.MemoryCache, runtimeProfiles RuntimeProfileStore) *Service {
	return NewServiceWithDependencies(Dependencies{
		Cache:           c,
		BudgetWindows:   c,
		Wallets:         &budgetWindowWalletStore{wallet: &platform.Wallet{TenantID: "tenant-1", BalanceIRR: 1000}},
		RuntimeProfiles: runtimeProfiles,
	}, 850000, "test")
}

func TestReserveBudgetEnforcesScopedTeamPolicyWindow(t *testing.T) {
	mem, err := testcache.New(t, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMemoryCache returned error: %v", err)
	}

	service := newPolicyTestService(mem, staticRuntimeProfiles{profile: &runtimeprofile.Profile{
		TenantID: "tenant-1",
		BudgetRules: []platform.BudgetLimitRule{{
			RuleID:              "team-monthly",
			Enabled:             true,
			LimitTo:             50,
			Unit:                platform.BudgetLimitCostPerMonth,
			BudgetAppliesPer:    stringPtr(string(platform.BudgetAppliesPerTeam)),
			BlockOnBudgetExceed: true,
		}},
	}})

	auth := &platform.AuthContext{TenantID: "tenant-1", TeamID: "team-a", TokenHash: "hash-1"}

	first, err := service.ReserveBudget(context.Background(), auth, platform.UsageCostBreakdown{Model: "gpt-test", TotalCostIRR: 40}, "req-1")
	if err != nil {
		t.Fatalf("first ReserveBudget returned error: %v", err)
	}

	if len(first.Targets) != 2 {
		t.Fatalf("expected tenant credit plus one policy target, got %+v", first.Targets)
	}

	if first.Targets[1].RuleID != "team-monthly" || first.Targets[1].Scope != string(platform.BudgetAppliesPerTeam) || first.Targets[1].ScopeValue != "team-a" {
		t.Fatalf("unexpected policy target: %+v", first.Targets[1])
	}

	_, err = service.ReserveBudget(context.Background(), auth, platform.UsageCostBreakdown{Model: "gpt-test", TotalCostIRR: 20}, "req-2")
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("expected ErrBudgetExceeded for exhausted team policy, got %v", err)
	}
}

func TestReserveBudgetIgnoresNonMatchingPolicyConditions(t *testing.T) {
	mem, err := testcache.New(t, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMemoryCache returned error: %v", err)
	}

	service := newPolicyTestService(mem, staticRuntimeProfiles{profile: &runtimeprofile.Profile{
		TenantID: "tenant-1",
		BudgetRules: []platform.BudgetLimitRule{{
			RuleID:              "expensive-model",
			Enabled:             true,
			LimitTo:             10,
			Unit:                platform.BudgetLimitCostPerDay,
			BudgetAppliesPer:    stringPtr(string(platform.BudgetAppliesPerModel)),
			BlockOnBudgetExceed: true,
			Conditions: map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{"field": "model", "operator": "in", "values": []interface{}{"gpt-expensive"}},
				},
			},
		}},
	}})

	reservation, err := service.ReserveBudget(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, platform.UsageCostBreakdown{Model: "gpt-cheap", TotalCostIRR: 100}, "req-1")
	if err != nil {
		t.Fatalf("ReserveBudget returned error: %v", err)
	}

	if len(reservation.Targets) != 1 {
		t.Fatalf("expected only tenant credit target for nonmatching policy, got %+v", reservation.Targets)
	}
}

func TestReserveBudgetMatchesMetadataAndRequestTypeConditions(t *testing.T) {
	mem, err := testcache.New(t, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMemoryCache returned error: %v", err)
	}

	service := newPolicyTestService(mem, staticRuntimeProfiles{profile: &runtimeprofile.Profile{
		TenantID: "tenant-1",
		BudgetRules: []platform.BudgetLimitRule{{
			RuleID:              "project-chat",
			Enabled:             true,
			LimitTo:             30,
			Unit:                platform.BudgetLimitCostPerDay,
			BudgetAppliesPer:    stringPtr("metadata.project_id"),
			BlockOnBudgetExceed: true,
			Conditions: map[string]interface{}{
				"request_type": "chat_completions",
				"conditions": []interface{}{
					map[string]interface{}{"field": "metadata", "operator": "equals", "metadataKey": "project_id", "value": "alpha"},
				},
			},
		}},
	}})

	breakdown := platform.UsageCostBreakdown{
		Model:        "gpt-test",
		RequestType:  "chat_completions",
		TotalCostIRR: 20,
		Metadata:     map[string]interface{}{"project_id": "alpha"},
	}

	reservation, err := service.ReserveBudget(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, breakdown, "req-1")
	if err != nil {
		t.Fatalf("ReserveBudget returned error: %v", err)
	}

	if len(reservation.Targets) != 2 || reservation.Targets[1].Scope != "metadata.project_id" || reservation.Targets[1].ScopeValue != "alpha" {
		t.Fatalf("unexpected metadata policy target: %+v", reservation.Targets)
	}

	_, err = service.ReserveBudget(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, breakdown, "req-2")
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("expected metadata policy to be exhausted, got %v", err)
	}
}

func TestReserveBudgetRollsBackAlreadyReservedPolicyTargetsOnFailure(t *testing.T) {
	mem, err := testcache.New(t, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMemoryCache returned error: %v", err)
	}

	service := newPolicyTestService(mem, staticRuntimeProfiles{profile: &runtimeprofile.Profile{
		TenantID: "tenant-1",
		BudgetRules: []platform.BudgetLimitRule{
			{
				RuleID:              "model-monthly",
				Enabled:             true,
				Priority:            1,
				LimitTo:             100,
				Unit:                platform.BudgetLimitCostPerMonth,
				BudgetAppliesPer:    stringPtr(string(platform.BudgetAppliesPerModel)),
				BlockOnBudgetExceed: true,
			},
			{
				RuleID:              "expensive-daily",
				Enabled:             true,
				Priority:            2,
				LimitTo:             10,
				Unit:                platform.BudgetLimitCostPerDay,
				BudgetAppliesPer:    stringPtr(string(platform.BudgetAppliesPerModel)),
				BlockOnBudgetExceed: true,
				Conditions: map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{"field": "model", "operator": "in", "values": []interface{}{"gpt-expensive"}},
					},
				},
			},
		},
	}})

	auth := &platform.AuthContext{TenantID: "tenant-1"}

	_, err = service.ReserveBudget(context.Background(), auth, platform.UsageCostBreakdown{Model: "gpt-expensive", TotalCostIRR: 20}, "req-fail")
	if !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("expected ErrBudgetExceeded for second policy, got %v", err)
	}

	reservation, err := service.ReserveBudget(context.Background(), auth, platform.UsageCostBreakdown{Model: "gpt-cheap", TotalCostIRR: 90}, "req-after-rollback")
	if err != nil {
		t.Fatalf("expected rollback to free first policy reservation, got %v", err)
	}

	if len(reservation.Targets) != 2 {
		t.Fatalf("expected tenant credit plus monthly model policy, got %+v", reservation.Targets)
	}
}

type staticRuntimeProfiles struct {
	profile *runtimeprofile.Profile
	err     error
}

func (s staticRuntimeProfiles) Get(context.Context, string) (*runtimeprofile.Profile, error) {
	return s.profile, s.err
}

func stringPtr(v string) *string {
	return &v
}
