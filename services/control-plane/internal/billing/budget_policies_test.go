package billing

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
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
	require.NoError(t, h.service.CommitBudget(context.Background(), first, creditBreakdown(50)))

	_, err = h.service.ReserveBudget(context.Background(), auth, creditBreakdown(1), "req-2")
	require.ErrorIs(t, err, ErrBudgetExceeded)
	require.Equal(t, string(platform.BudgetAppliesPerTeam), BudgetErrorScope(err))

	otherTeam := &platform.AuthContext{TenantID: "tenant-1", TeamID: "team-b", TokenHash: "hash-2"}
	_, err = h.service.ReserveBudget(context.Background(), otherTeam, creditBreakdown(1), "req-3")
	require.NoError(t, err, "each team has its own window")
}

func TestBudgetPolicyWeekStartsOnSaturdayInTehran(t *testing.T) {
	// Saturday 2026-10-17 00:00 in Tehran (UTC+03:30) is Friday 20:30 UTC.
	for _, tc := range []struct {
		name string
		now  time.Time
	}{
		{name: "saturday midnight tehran", now: time.Date(2026, time.October, 16, 20, 30, 0, 0, time.UTC)},
		{name: "monday", now: time.Date(2026, time.October, 19, 12, 0, 0, 0, time.UTC)},
		{name: "friday last second tehran", now: time.Date(2026, time.October, 23, 20, 29, 59, 0, time.UTC)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			windowID, windowEnd, err := budgetPolicyWindowForTime(platform.BudgetLimitCostPerWeek, tc.now)
			require.NoError(t, err)
			require.Equal(t, "W20261017", windowID)
			require.Equal(t, time.Date(2026, time.October, 23, 20, 30, 0, 0, time.UTC), windowEnd)
		})
	}

	before, _, err := budgetPolicyWindowForTime(platform.BudgetLimitCostPerWeek, time.Date(2026, time.October, 16, 20, 29, 59, 0, time.UTC))
	require.NoError(t, err)
	require.Equal(t, "W20261010", before, "Friday evening UTC before Tehran midnight is still the previous week")

	next, _, err := budgetPolicyWindowForTime(platform.BudgetLimitCostPerWeek, time.Date(2026, time.October, 23, 20, 30, 0, 0, time.UTC))
	require.NoError(t, err)
	require.Equal(t, "W20261024", next, "the next week starts on the following Tehran Saturday")
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
	require.NoError(t, h.service.CommitBudget(context.Background(), reservation, creditBreakdown(30)))

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
	ctx := context.Background()

	served, err := h.service.ReserveBudget(ctx, auth, modelBreakdown("gpt-expensive", 10), "req-served")
	require.NoError(t, err)
	require.NoError(t, h.service.CommitBudget(ctx, served, modelBreakdown("gpt-expensive", 10)))

	refused, err := h.service.ReserveBudget(ctx, auth, modelBreakdown("gpt-expensive", 1), "req-fail")
	require.ErrorIs(t, err, ErrBudgetExceeded, "the daily rule refuses the request")
	require.Nil(t, refused)

	monthlyKey := served.Targets[1].WindowKey
	result, err := h.cache.CommitBudget(ctx, cache.BudgetCommitRequest{
		WindowKey:      monthlyKey,
		ReservationKey: credits.PolicyReservationKey("tenant-1", budgetReservationID("req-fail"), shortHash(monthlyKey)),
		ActualIRR:      1,
	})
	require.NoError(t, err)
	require.Equal(t, "released", result.State, "the refused request released its monthly reservation")
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
