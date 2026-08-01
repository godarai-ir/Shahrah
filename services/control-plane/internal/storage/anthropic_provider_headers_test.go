package storage

import (
	"testing"
)

func TestAnthropicProviderHeadersUsesConfiguredBetaAllowlist(t *testing.T) {
	headers := anthropicProviderHeaders(
		"anthropic",
		map[string]interface{}{
			"beta_allowlist": []interface{}{
				"files-api-2025-04-14",
				"output-300k-2026-03-24",
				"output-300k-2026-03-24",
			},
		},
	)

	got := headers["Anthropic-Beta-Allowlist"]

	want := "files-api-2025-04-14,output-300k-2026-03-24"
	if got != want {
		t.Fatalf("unexpected Anthropic beta allowlist %q", got)
	}
}

func TestAnthropicProviderHeadersDoesNotAffectOtherProviders(t *testing.T) {
	if headers := anthropicProviderHeaders(
		"openai",
		map[string]interface{}{"beta_allowlist": "unsafe"},
	); headers != nil {
		t.Fatalf("expected no provider headers, got %#v", headers)
	}
}
