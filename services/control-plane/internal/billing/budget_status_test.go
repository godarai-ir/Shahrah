package billing

import (
	"context"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
)

func TestBudgetStatusReportsTheCallersRules(t *testing.T) {
	userRule := creditUserRule(platform.BudgetOnExceedBlock, 10)
	userRule.Name = "Daily per user"

	otherTeam := creditUserRule(platform.BudgetOnExceedAlert, 5)
	otherTeam.RuleID = "other-team"
	otherTeam.Conditions = map[string]interface{}{"team_id": []interface{}{"team-b"}}

	perModel := creditUserRule(platform.BudgetOnExceedBlock, 5)
	perModel.RuleID = "per-model"
	perModel.BudgetAppliesPer = stringPtr(string(platform.BudgetAppliesPerModel))

	modelConditioned := creditUserRule(platform.BudgetOnExceedAlert, 20)
	modelConditioned.RuleID = "gpt-only"
	modelConditioned.Conditions = map[string]interface{}{"conditions": []interface{}{
		map[string]interface{}{"field": "model", "operator": "in", "values": []interface{}{"gpt-test"}},
	}}

	h := newCreditTestHarness(t, 50_000, userRule, otherTeam, perModel, modelConditioned)
	ctx := context.Background()
	caller := &platform.AuthContext{TenantID: "tenant-1", UserID: "user-1", TeamID: "team-a", TokenHash: "hash-1"}

	served, err := h.service.ReserveBudget(ctx, caller, creditBreakdown(3), "req-1")
	require.NoError(t, err)
	require.NoError(t, h.service.CommitBudget(ctx, served, creditBreakdown(3)))

	_, err = h.service.ReserveBudget(ctx, caller, creditBreakdown(2), "req-2")
	require.NoError(t, err)

	status, err := h.service.BudgetStatus(ctx, caller, false)
	require.NoError(t, err)
	require.Equal(t, "credit", status.Unit)
	require.Nil(t, status.TenantBalance, "the balance is shown only with billing read")

	rules := map[string]platform.BudgetRuleStatus{}
	for _, rule := range status.Rules {
		rules[rule.RuleID] = rule
	}

	require.Len(t, rules, 2, "another team's rule and the per-model rule are left out")
	require.Equal(t, platform.BudgetRuleStatus{
		RuleID: "user-daily", Name: "Daily per user", Scope: "user", Limit: "10.00", Spent: "3.00", Remaining: "5.00",
		WindowEnd: rules["user-daily"].WindowEnd, OnExceed: "block",
	}, rules["user-daily"], "remaining subtracts the 2 credits still reserved")
	require.Equal(t, "17.00", rules["gpt-only"].Remaining, "request conditions cannot be known, so the rule is reported")

	withBalance, err := h.service.BudgetStatus(ctx, caller, true)
	require.NoError(t, err)
	require.Equal(t, "45.00", withBalance.TenantBalance.Remaining)
	require.Equal(t, "2.00", withBalance.TenantBalance.Reserved)
}

func TestBudgetStatusBeforeAnyRequestReadsTheBalance(t *testing.T) {
	h := newCreditTestHarness(t, 7_000)

	status, err := h.service.BudgetStatus(context.Background(), creditAuth, true)
	require.NoError(t, err)
	require.Empty(t, status.Rules)
	require.Equal(t, "7.00", status.TenantBalance.Remaining)
	require.Equal(t, "0.00", status.TenantBalance.Reserved)
}
