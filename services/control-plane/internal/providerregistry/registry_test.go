package providerregistry

import "testing"

func TestDefinitionsCoverFirstClassProviders(t *testing.T) {
	want := map[string]bool{
		"openai":        false,
		"anthropic":     true,
		"xai":           true,
		"deepseek":      true,
		"moonshot":      true,
		"google-gemini": true,
		"mistral-ai":    true,
	}

	got := make(map[string]bool, len(definitions))
	for _, definition := range Definitions() {
		if definition.PricingSnapshot == "" {
			t.Fatalf("provider %q has no pricing snapshot", definition.Slug)
		}

		got[definition.Slug] = definition.DispatchRuntimeProfileRequired
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d providers, got %d", len(want), len(got))
	}

	for provider, runtimeProfileRequired := range want {
		if got[provider] != runtimeProfileRequired {
			t.Fatalf("provider %q runtime-profile requirement = %t, want %t", provider, got[provider], runtimeProfileRequired)
		}
	}
}

func TestFromModelIdentifierResolvesEveryFirstClassProvider(t *testing.T) {
	tests := map[string]string{
		"openai:primary:gpt-5":               "openai",
		"anthropic:primary:claude-opus-5":    "anthropic",
		"xai:primary:grok-4.5":               "xai",
		"deepseek:primary:deepseek-v4-flash": "deepseek",
		"moonshot:primary:kimi-k3":           "moonshot",
		"gemini/gemini-3.5-flash":            "google-gemini",
		"mistral:primary:mistral-large-2607": "mistral-ai",
		"claude-sonnet-5":                    "anthropic",
		"voxtral-mini-2602":                  "mistral-ai",
	}

	for model, want := range tests {
		t.Run(model, func(t *testing.T) {
			definition, ok := FromModelIdentifier(model)
			if !ok || definition.Slug != want {
				t.Fatalf("FromModelIdentifier(%q) = %q, %t; want %q", model, definition.Slug, ok, want)
			}
		})
	}
}

func TestAliasesResolveToCanonicalProviders(t *testing.T) {
	for alias, want := range map[string]string{
		"gemini":  "google-gemini",
		"mistral": "mistral-ai",
	} {
		definition, ok := Lookup(alias)
		if !ok || definition.Slug != want {
			t.Fatalf("Lookup(%q) = %q, %t; want %q", alias, definition.Slug, ok, want)
		}
	}
}
