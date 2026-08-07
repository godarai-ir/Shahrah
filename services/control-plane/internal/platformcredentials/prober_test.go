package platformcredentials

import (
	"context"
	"strings"
	"testing"
)

type staticCredentialProber struct {
	result ProbeResult
	calls  int
}

func (p *staticCredentialProber) Probe(
	context.Context,
	EncryptedCredential,
	map[string]string,
) ProbeResult {
	p.calls++

	return p.result
}

func TestCredentialProberRegistryDispatchesByCanonicalProvider(t *testing.T) {
	t.Parallel()

	fallback := &staticCredentialProber{result: ProbeResult{Category: "fallback"}}
	xai := &staticCredentialProber{result: ProbeResult{Authenticated: true, Ready: true, Category: "xai"}}
	registry := NewCredentialProberRegistry(fallback)
	registry.Register(" XAI ", xai)

	result := registry.Probe(context.Background(), EncryptedCredential{
		Binding: Binding{ProviderSlug: "xai"},
	}, nil)

	if result.Category != "xai" || xai.calls != 1 || fallback.calls != 0 {
		t.Fatalf("unexpected dedicated dispatch: result=%#v xai_calls=%d fallback_calls=%d", result, xai.calls, fallback.calls)
	}

	result = registry.Probe(context.Background(), EncryptedCredential{
		Binding: Binding{ProviderSlug: "openai"},
	}, nil)

	if result.Category != "fallback" || fallback.calls != 1 {
		t.Fatalf("unexpected fallback dispatch: result=%#v fallback_calls=%d", result, fallback.calls)
	}
}

func TestClassifyXAIAPIKeyProbeRequiresSelectedModelPermission(t *testing.T) {
	t.Parallel()

	result := classifyXAIAPIKeyProbe(200, []byte(`{
		"api_key_disabled": false,
		"api_key_blocked": false,
		"team_blocked": false,
		"acls": ["api-key:endpoint:chat", "api-key:model:grok-4.5"]
	}`), "xai:account:grok-4.3")

	if !result.Authenticated || result.Ready || result.Category != "missing_scope" {
		t.Fatalf("unexpected result: %#v", result)
	}

	if !strings.Contains(result.Summary, "api-key:model:grok-4.3") {
		t.Fatalf("missing model permission was not reported: %q", result.Summary)
	}
}
