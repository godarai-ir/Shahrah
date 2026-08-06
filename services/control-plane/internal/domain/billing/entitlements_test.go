package billing

import "testing"

func TestResolveEntitlementsAppliesDomainPrecedence(t *testing.T) {
	values := []EntitlementValue{
		{Key: "requests", Value: 10, Source: EntitlementSourceDefault},
		{Key: "requests", Value: 100, Source: EntitlementSourcePlan},
		{Key: "requests", Value: 250, Source: EntitlementSourceOverride},
		{Key: "models", Value: []string{"base"}, Source: EntitlementSourceDefault},
	}

	features, overrides := ResolveEntitlements(values)
	if features["requests"] != 250 {
		t.Fatalf("expected override value, got %v", features["requests"])
	}

	if overrides["requests"] != 250 {
		t.Fatalf("expected override metadata, got %v", overrides["requests"])
	}

	if _, exists := overrides["models"]; exists {
		t.Fatal("default entitlement must not be reported as an override")
	}
}
