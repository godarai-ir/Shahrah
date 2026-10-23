package admin

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestValidateCreateBudgetLimitRuleRequest(t *testing.T) {
	err := validateCreateBudgetLimitRuleRequest(platform.CreateBudgetLimitRuleRequest{})
	if err == nil {
		t.Fatalf("expected error for empty request")
	}

	err = validateCreateBudgetLimitRuleRequest(platform.CreateBudgetLimitRuleRequest{
		RuleID:   "budget-rule-1",
		LimitTo:  100,
		Unit:     platform.BudgetLimitCostPerMonth,
		Alerts:   map[string]interface{}{"thresholds": []int{75, 90, 100}},
		Enabled:  boolPtr(true),
		Priority: intPtr(10),
	})
	if err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestValidateUpdateBudgetLimitRuleRequest(t *testing.T) {
	priority := -1

	err := validateUpdateBudgetLimitRuleRequest(platform.UpdateBudgetLimitRuleRequest{
		Priority: &priority,
	})
	if err == nil {
		t.Fatalf("expected error for negative priority")
	}
}

func TestValidateBudgetAppliesPer(t *testing.T) {
	validValues := []string{
		string(platform.BudgetAppliesPerTenant),
		string(platform.BudgetAppliesPerUser),
		string(platform.BudgetAppliesPerTeam),
		string(platform.BudgetAppliesPerModel),
		string(platform.BudgetAppliesPerVirtualAccount),
		"metadata.project_id",
	}
	for _, value := range validValues {
		err := validateCreateBudgetLimitRuleRequest(platform.CreateBudgetLimitRuleRequest{
			RuleID:           "rule",
			LimitTo:          10,
			Unit:             platform.BudgetLimitCostPerDay,
			BudgetAppliesPer: stringPtr(value),
		})
		if err != nil {
			t.Fatalf("expected %s applies per to be valid, got %v", value, err)
		}
	}
}

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }
func stringPtr(v string) *string {
	return &v
}
