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

func TestBillingEstimateModelPreservesFirstClassProviderFQN(t *testing.T) {
	models := []string{
		"openai:production:gpt-5",
		"anthropic:production:claude-opus-5",
		"xai:production:grok-4.5",
		"deepseek:production:deepseek-v4-flash",
		"moonshot:production:kimi-k3",
		"google-gemini:production:gemini-3.1-pro-preview",
		"mistral-ai:production:mistral-large-2607",
	}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			if got := billingEstimateModel(model); got != model {
				t.Fatalf("expected provider FQN %q to be preserved, got %q", model, got)
			}
		})
	}
}

func TestBillingEstimateModelStripsUnsupportedProviderFQN(t *testing.T) {
	got := billingEstimateModel("legacy:production:legacy-chat")
	if got != "legacy-chat" {
		t.Fatalf("expected unsupported provider FQN to use its official model, got %q", got)
	}
}
