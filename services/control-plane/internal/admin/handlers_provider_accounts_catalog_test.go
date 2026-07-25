package admin

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestBuildTenantProviderAccountModelCatalogAggregatesAvailability(t *testing.T) {
	inputPrice := 2.5
	outputPrice := 10.0
	records := []platform.TenantProviderAccountCatalogRecord{
		{
			Account: platform.TenantProviderAccount{
				ID:                            "account-1",
				Provider:                      "openai",
				Name:                          "primary",
				AuthMode:                      platform.ProviderAuthModeAPIKey,
				HasCredential:                 true,
				UsePlatformManagedCredentials: false,
			},
			ManagerCount: 1,
			Model:        "gpt-4.1",
			Types:        []platform.ModelType{platform.ModelTypeChat},
			PricingMode:  platform.PricingModePublic,
		},
		{
			Account: platform.TenantProviderAccount{
				ID:                            "account-2",
				Provider:                      "openai",
				Name:                          "managed",
				AuthMode:                      platform.ProviderAuthModeAPIKey,
				UsePlatformManagedCredentials: true,
			},
			ManagerCount: 1,
			Model:        "gpt-4.1",
			Types:        []platform.ModelType{platform.ModelTypeChat},
			PricingMode:  platform.PricingModePublic,
		},
	}
	catalog := tenantProviderCatalogLookup{
		"openai": {
			"gpt-4.1": {
				Provider:    "openai",
				Model:       "gpt-4.1",
				DisplayName: "GPT-4.1",
				Types:       []platform.ModelType{platform.ModelTypeChat},
				PublicPricing: &platform.ModelPrice{
					InputUSDPer1M:  &inputPrice,
					OutputUSDPer1M: &outputPrice,
				},
				IsDeprecated:     true,
				ReplacementModel: "gpt-5",
			},
		},
	}

	items := buildTenantProviderAccountModelCatalog(records, catalog, tenantProviderModelCatalogQuery{
		Sort:  "name",
		Order: "asc",
	})

	if len(items) != 1 {
		t.Fatalf("expected one aggregated model, got %d", len(items))
	}

	item := items[0]
	if got := len(item.Availability); got != 2 {
		t.Fatalf("expected two availability paths, got %d", got)
	}

	if item.DisplayName != "GPT-4.1" || !item.IsDeprecated || item.ReplacementModel != "gpt-5" {
		t.Fatalf("catalog metadata was not preserved: %#v", item)
	}

	if item.Availability[0].FQN != "openai:managed:gpt-4.1" ||
		item.Availability[1].FQN != "openai:primary:gpt-4.1" {
		t.Fatalf("unexpected FQNs: %#v", item.Availability)
	}

	if item.Availability[0].CredentialSource != platform.CredentialSourceGodarManagedKey {
		t.Fatalf("expected managed credential source, got %q", item.Availability[0].CredentialSource)
	}

	if item.Availability[1].CredentialSource != platform.CredentialSourceUserBYOK {
		t.Fatalf("expected BYOK credential source, got %q", item.Availability[1].CredentialSource)
	}

	for _, availability := range item.Availability {
		if !availability.PricingResolved || availability.Readiness != platform.ProviderAccountReadinessReady {
			t.Fatalf("expected ready, priced availability: %#v", availability)
		}
	}
}

func TestAccountReadinessPrecedence(t *testing.T) {
	account := platform.TenantProviderAccount{
		AuthMode: platform.ProviderAuthModeAPIKey,
	}

	summary := platform.TenantProviderAccountSummary{}
	if got := accountReadiness(account, summary); got != platform.ProviderAccountReadinessNeedsCredentials {
		t.Fatalf("expected credentials readiness first, got %q", got)
	}

	account.UsePlatformManagedCredentials = true
	if got := accountReadiness(account, summary); got != platform.ProviderAccountReadinessNeedsModels {
		t.Fatalf("expected needs models, got %q", got)
	}

	summary.ModelCount = 1
	if got := accountReadiness(account, summary); got != platform.ProviderAccountReadinessNeedsAccess {
		t.Fatalf("expected needs access, got %q", got)
	}

	summary.ManagerCount = 1

	summary.PricingUnresolvedCount = 1
	if got := accountReadiness(account, summary); got != platform.ProviderAccountReadinessPricingIncomplete {
		t.Fatalf("expected pricing incomplete, got %q", got)
	}

	summary.PricingUnresolvedCount = 0
	if got := accountReadiness(account, summary); got != platform.ProviderAccountReadinessReady {
		t.Fatalf("expected ready, got %q", got)
	}
}
