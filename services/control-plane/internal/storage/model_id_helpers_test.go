package storage

import "testing"

func TestNormalizeModelFQN_StripsRedundantProviderPrefix(t *testing.T) {
	got := normalizeModelFQN("mistral-ai:mistral-ai-eager-harbor:mistral-ai/ministral-8b-2512")
	want := "mistral-ai:mistral-ai-eager-harbor:ministral-8b-2512"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRuntimePolicyByModel_NormalizesVirtualTargetFQN(t *testing.T) {
	policies := []runtimeAllowedModelPolicy{
		{
			Model: "mistral-ai:mistral-ai-eager-harbor:ministral-8b-2512",
			Types: []string{"chat"},
		},
	}

	policy, ok := runtimePolicyByModel(policies, "mistral-ai:mistral-ai-eager-harbor:mistral-ai/ministral-8b-2512")
	if !ok {
		t.Fatal("expected policy to match normalized target FQN")
	}
	if len(policy.Types) != 1 || policy.Types[0] != "chat" {
		t.Fatalf("expected chat policy, got %#v", policy.Types)
	}
}
