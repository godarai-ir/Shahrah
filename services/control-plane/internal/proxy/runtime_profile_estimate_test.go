package proxy

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestRuntimeProfilesFromEstimateUsesRegistryForFirstClassProviders(t *testing.T) {
	tests := []struct {
		provider string
		model    string
		required bool
	}{
		{provider: "openai", model: "gpt-5", required: false},
		{provider: "anthropic", model: "claude-opus-5", required: true},
		{provider: "xai", model: "grok-4.5", required: true},
		{provider: "deepseek", model: "deepseek-v4-flash", required: true},
		{provider: "moonshot", model: "kimi-k3", required: true},
		{provider: "gemini", model: "gemini-3.5-flash", required: true},
		{provider: "mistral", model: "mistral-large-2607", required: true},
	}

	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			if !test.required {
				profiles, err := runtimeProfilesFromEstimate(
					test.provider,
					test.model,
					platform.UsageCostBreakdown{},
				)
				if err != nil || len(profiles) != 0 {
					t.Fatalf("unexpected optional runtime profile result: %+v, %v", profiles, err)
				}

				return
			}

			binding := runtimeProfileBindingForProvider(test.provider, test.model)
			profiles, err := runtimeProfilesFromEstimate(
				test.provider,
				test.model,
				platform.UsageCostBreakdown{
					PricingRevision:     binding.CatalogRevision,
					ModelRuntimeProfile: &binding,
				},
			)
			if err != nil {
				t.Fatal(err)
			}

			if len(profiles) != 1 || !profiles[0].MatchesModel(test.model) {
				t.Fatalf("unexpected runtime profile result: %+v", profiles)
			}
		})
	}
}

func TestRuntimeProfilesFromEstimateFailsClosedForEveryRequiredProvider(t *testing.T) {
	for _, provider := range []string{
		"anthropic",
		"xai",
		"deepseek",
		"moonshot",
		"google-gemini",
		"mistral-ai",
	} {
		t.Run(provider, func(t *testing.T) {
			_, err := runtimeProfilesFromEstimate(
				provider,
				"model",
				platform.UsageCostBreakdown{},
			)
			if err == nil {
				t.Fatal("expected missing runtime profile to fail closed")
			}
		})
	}
}

func runtimeProfileBindingForProvider(
	provider string,
	model string,
) platform.ModelRuntimeProfileBinding {
	binding := deepSeekProxyRuntimeProfile("deepseek-v4-flash")
	binding.Provider = provider
	binding.CatalogRevision = provider + "-test-revision"
	binding.CanonicalID = provider + "/" + model
	binding.Slug = model

	return binding
}
