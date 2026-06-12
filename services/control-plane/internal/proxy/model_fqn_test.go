package proxy

import "testing"

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

