package admin

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestValidateCreateRateLimitRuleRequest(t *testing.T) {
	err := validateCreateRateLimitRuleRequest(platform.CreateRateLimitRuleRequest{})
	if err == nil {
		t.Fatalf("expected error for empty request")
	}

	err = validateCreateRateLimitRuleRequest(platform.CreateRateLimitRuleRequest{
		RuleID: "rule-1",
		Limit:  map[string]interface{}{"max": 10},
	})
	if err != nil {
		t.Fatalf("expected valid request, got %v", err)
	}
}

func TestValidateUpdateRateLimitRuleRequest(t *testing.T) {
	priority := -1

	err := validateUpdateRateLimitRuleRequest(platform.UpdateRateLimitRuleRequest{
		Priority: &priority,
	})
	if err == nil {
		t.Fatalf("expected error for negative priority")
	}
}
