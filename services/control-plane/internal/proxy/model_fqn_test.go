package proxy

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestNormalizeModelIdentifier_CanonicalizesLegacyProviderScopedFQN(t *testing.T) {
	got := normalizeModelIdentifier("openai:openai-daring-lotus:openai/gpt-4o-mini")

	want := "openai:openai-daring-lotus:gpt-4o-mini"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestProviderModelID_ReturnsCanonicalOfficialModel(t *testing.T) {
	got := providerModelID("openai:openai-daring-lotus:openai/gpt-4o-mini")
	if got != "gpt-4o-mini" {
		t.Fatalf("expected canonical official model, got %q", got)
	}
}

func TestBillingModelForDeployment_UsesProviderAndWildcardFallback(t *testing.T) {
	got := billingModelForDeployment("openai:default:gpt-image-1", platform.Deployment{Provider: "openai", Model: "*"})
	if got != "openai/gpt-image-1" {
		t.Fatalf("expected provider-qualified billing model, got %q", got)
	}
}

func TestBillingEstimateModel_PreservesGoogleGeminiFQN(t *testing.T) {
	got := billingEstimateModel(
		"google-gemini:production:gemini-3.1-pro-preview",
	)
	if got != "google-gemini:production:gemini-3.1-pro-preview" {
		t.Fatalf("expected Gemini FQN to be preserved, got %q", got)
	}
}

func TestBillingEstimateModel_StripsNonCatalogProviderFQN(t *testing.T) {
	got := billingEstimateModel("openai:production:gpt-4o-mini")
	if got != "gpt-4o-mini" {
		t.Fatalf("expected provider model id, got %q", got)
	}
}
