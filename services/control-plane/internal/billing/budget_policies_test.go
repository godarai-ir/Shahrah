package billing

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/stretchr/testify/require"
)

// creditRule is a blocking credit budget rule.
func creditRule(ruleID string, unit platform.BudgetLimitUnit, appliesPer string, limitCredits int64, conditions map[string]interface{}) platform.BudgetLimitRule {
	limitMC := limitCredits * credits.MillicreditsPerCredit

	return platform.BudgetLimitRule{
		RuleID:           ruleID,
		Enabled:          true,
		Unit:             unit,
		BudgetAppliesPer: stringPtr(appliesPer),
		LimitMC:          &limitMC,
		OnExceed:         platform.BudgetOnExceedBlock,
		Conditions:       conditions,
	}
}

func modelBreakdown(model string, creditsAmount int) platform.UsageCostBreakdown {
	breakdown := creditBreakdown(creditsAmount)
	breakdown.Model = model

	return breakdown
}

func TestReserveBudgetEnforcesScopedTeamPolicyWindow(t *testing.T) {
	h := newCreditTestHarness(t, 1_000_000,
		creditRule("team-monthly", platform.BudgetLimitCostPerMonth, string(platform.BudgetAppliesPerTeam), 50, nil))
	auth := &platform.AuthContext{TenantID: "tenant-1", TeamID: "team-a", TokenHash: "hash-1"}

	first, err := h.service.ReserveBudget(context.Background(), auth, creditBreakdown(40), "req-1")
	require.NoError(t, err)
	require.Len(t, first.Targets, 2, "the credit pool plus one rule window")
	require.Equal(t, "team-monthly", first.Targets[1].RuleID)
	require.Equal(t, string(platform.BudgetAppliesPerTeam), first.Targets[1].Scope)
	require.Equal(t, "team-a", first.Targets[1].ScopeValue)

	_, err = h.service.ReserveBudget(context.Background(), auth, creditBreakdown(20), "req-2")
	require.ErrorIs(t, err, ErrBudgetExceeded)
	require.Equal(t, string(platform.BudgetAppliesPerTeam), BudgetErrorScope(err))
}

func TestReserveBudgetIgnoresNonMatchingPolicyConditions(t *testing.T) {
	h := newCreditTestHarness(t, 1_000_000, creditRule("expensive-model", platform.BudgetLimitCostPerDay, string(platform.BudgetAppliesPerModel), 10,
		map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{"field": "model", "operator": "in", "values": []interface{}{"gpt-expensive"}},
			},
		}))

	reservation, err := h.service.ReserveBudget(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, modelBreakdown("gpt-cheap", 100), "req-1")
	require.NoError(t, err)
	require.Len(t, reservation.Targets, 1, "only the credit pool for a nonmatching rule")
}

func TestReserveBudgetMatchesMetadataAndRequestTypeConditions(t *testing.T) {
	h := newCreditTestHarness(t, 1_000_000, creditRule("project-chat", platform.BudgetLimitCostPerDay, "metadata.project_id", 30,
		map[string]interface{}{
			"request_type": "chat_completions",
			"conditions": []interface{}{
				map[string]interface{}{"field": "metadata", "operator": "equals", "metadataKey": "project_id", "value": "alpha"},
			},
		}))

	breakdown := creditBreakdown(20)
	breakdown.RequestType = "chat_completions"
	breakdown.Metadata = map[string]interface{}{"project_id": "alpha"}

	reservation, err := h.service.ReserveBudget(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, breakdown, "req-1")
	require.NoError(t, err)
	require.Len(t, reservation.Targets, 2)
	require.Equal(t, "metadata.project_id", reservation.Targets[1].Scope)
	require.Equal(t, "alpha", reservation.Targets[1].ScopeValue)

	_, err = h.service.ReserveBudget(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, breakdown, "req-2")
	require.ErrorIs(t, err, ErrBudgetExceeded, "the metadata rule is exhausted")
}

func TestReserveBudgetRollsBackAlreadyReservedPolicyTargetsOnFailure(t *testing.T) {
	monthly := creditRule("model-monthly", platform.BudgetLimitCostPerMonth, string(platform.BudgetAppliesPerModel), 100, nil)
	monthly.Priority = 1
	daily := creditRule("expensive-daily", platform.BudgetLimitCostPerDay, string(platform.BudgetAppliesPerModel), 10,
		map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{"field": "model", "operator": "in", "values": []interface{}{"gpt-expensive"}},
			},
		})
	daily.Priority = 2

	h := newCreditTestHarness(t, 1_000_000, monthly, daily)
	auth := &platform.AuthContext{TenantID: "tenant-1"}

	_, err := h.service.ReserveBudget(context.Background(), auth, modelBreakdown("gpt-expensive", 20), "req-fail")
	require.ErrorIs(t, err, ErrBudgetExceeded, "the daily rule refuses the request")

	reservation, err := h.service.ReserveBudget(context.Background(), auth, modelBreakdown("gpt-cheap", 90), "req-after-rollback")
	require.NoError(t, err, "the refused request released its monthly reservation")
	require.Len(t, reservation.Targets, 2, "the credit pool plus the monthly model rule")
}

type staticRuntimeProfiles struct {
	profile *runtimeprofile.Profile
	err     error
}

func (s staticRuntimeProfiles) GetSnapshot(context.Context, string) (*runtimeprofile.Profile, error) {
	return s.profile, s.err
}

func stringPtr(v string) *string {
	return &v
}
