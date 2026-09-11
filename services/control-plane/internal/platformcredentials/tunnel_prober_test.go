package platformcredentials

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type metadataProbeTunnel struct {
	calls []platform.ProviderOperation
}

type recordingProbeTunnel struct {
	request  platform.TunnelRequest
	response platform.TunnelResponse
}

func (t *recordingProbeTunnel) Proxy(_ context.Context, request platform.TunnelRequest) (*platform.TunnelResponse, error) {
	t.request = request
	return &t.response, nil
}

func (*recordingProbeTunnel) Stream(context.Context, platform.TunnelRequest) (platform.TunnelStream, error) {
	return nil, errors.New("not implemented")
}

func (*recordingProbeTunnel) Healthy(context.Context) error { return nil }

func (t *metadataProbeTunnel) Proxy(
	_ context.Context,
	request platform.TunnelRequest,
) (*platform.TunnelResponse, error) {
	t.calls = append(t.calls, request.Operation)
	headers := map[string]string{
		"X-GodarAI-Anthropic-WIF-Expires-At":   "2026-07-29T12:00:00Z",
		"X-GodarAI-Anthropic-WIF-Refreshed-At": "2026-07-29T11:00:00Z",
		"X-GodarAI-Anthropic-WIF-Scopes":       "workspace:developer workspace:inference",
	}

	switch request.Operation {
	case platform.ProviderOperationAnthropicModelsList:
		return &platform.TunnelResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       []byte(`{"data":[{"id":"claude-sonnet-4-6"}]}`),
		}, nil
	case platform.ProviderOperationAnthropicCountTokens,
		platform.ProviderOperationAnthropicFilesList:
		return &platform.TunnelResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       []byte(`{}`),
		}, nil
	default:
		return nil, errors.New("unexpected operation")
	}
}

func (t *metadataProbeTunnel) Stream(
	context.Context,
	platform.TunnelRequest,
) (platform.TunnelStream, error) {
	return nil, errors.New("not implemented")
}

func (t *metadataProbeTunnel) Healthy(context.Context) error { return nil }

func TestTunnelCredentialProberReturnsSanitizedWIFMetadata(t *testing.T) {
	t.Parallel()

	tunnel := &metadataProbeTunnel{}
	prober := NewTunnelCredentialProber(tunnel)

	result := prober.Probe(
		context.Background(),
		EncryptedCredential{
			Binding: Binding{
				ProviderSlug: "anthropic",
				AuthMode:     string(platform.ProviderAuthModeAnthropicWIF),
				ConnectionID: "connection-1",
			},
			FederationRuleID: "rule-1",
			OrganizationID:   "org-1",
			ServiceAccountID: "service-1",
			WorkspaceID:      "workspace-1",
			TokenSourceID:    "source-1",
		},
		nil,
	)
	if !result.Ready || result.Category != "ok" {
		t.Fatalf("probe failed: ready=%v category=%q", result.Ready, result.Category)
	}

	if len(tunnel.calls) != 3 {
		t.Fatalf("expected three capability probes, got %d", len(tunnel.calls))
	}

	if result.Metadata.WIFExpiresAt == nil ||
		!result.Metadata.WIFExpiresAt.Equal(time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected expiry: %#v", result.Metadata.WIFExpiresAt)
	}

	if result.Metadata.WIFRefreshedAt == nil ||
		!result.Metadata.WIFRefreshedAt.Equal(time.Date(2026, 7, 29, 11, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected refresh time: %#v", result.Metadata.WIFRefreshedAt)
	}

	if len(result.Metadata.WIFScopes) != 2 ||
		result.Metadata.WIFScopes[0] != "workspace:developer" ||
		result.Metadata.WIFScopes[1] != "workspace:inference" {
		t.Fatalf("unexpected scopes: %#v", result.Metadata.WIFScopes)
	}
}

func TestTunnelCredentialProberRoutesGeminiProbeThroughSilkRoad(t *testing.T) {
	t.Parallel()

	tunnel := &recordingProbeTunnel{response: platform.TunnelResponse{StatusCode: 200}}
	prober := NewTunnelCredentialProber(tunnel)
	result := prober.Probe(context.Background(), EncryptedCredential{
		Binding: Binding{
			ProviderSlug:        "google-gemini",
			AuthMode:            "api_key",
			ConnectionID:        "connection-1",
			CredentialVersionID: "credential-version-1",
		},
		BaseURL: "https://generativelanguage.googleapis.com",
	}, map[string]string{"api_key": "gemini-secret"})

	if !result.Ready || result.Category != "ok" {
		t.Fatalf("probe failed: ready=%v category=%q", result.Ready, result.Category)
	}

	if tunnel.request.Operation != platform.ProviderOperationCredentialProbe {
		t.Fatalf("operation = %q", tunnel.request.Operation)
	}

	if tunnel.request.Deployment.Provider != "google-gemini" ||
		tunnel.request.Deployment.ProviderAuth == nil ||
		tunnel.request.Deployment.ProviderAuth.APIKey != "gemini-secret" {
		t.Fatalf("unexpected deployment: %#v", tunnel.request.Deployment)
	}
}

func TestTunnelCredentialProberFailsClosedWithoutTunnel(t *testing.T) {
	t.Parallel()

	result := NewTunnelCredentialProber(nil).Probe(context.Background(), EncryptedCredential{
		Binding: Binding{ProviderSlug: "google-gemini", AuthMode: "api_key"},
	}, map[string]string{"api_key": "gemini-secret"})
	if result.Ready || result.Category != "network" {
		t.Fatalf("ready=%v category=%q, want false/network", result.Ready, result.Category)
	}
}

func TestTunnelCredentialProberUsesDedicatedXAIKeyStatusProbe(t *testing.T) {
	t.Parallel()

	tunnel := &recordingProbeTunnel{response: platform.TunnelResponse{
		StatusCode: http.StatusOK,
		Body: []byte(`{
			"api_key_disabled": false,
			"api_key_blocked": false,
			"team_blocked": true,
			"acls": ["api-key:model:*", "api-key:endpoint:*"]
		}`),
	}}
	prober := NewTunnelCredentialProber(tunnel)

	result := prober.Probe(context.Background(), EncryptedCredential{
		Binding:   Binding{ProviderSlug: "xai", AuthMode: "api_key"},
		ModelHint: "grok-4.3",
	}, map[string]string{"api_key": "xai-secret"})

	if !result.Authenticated || result.Ready || result.Category != "team_blocked" {
		t.Fatalf("unexpected result: %#v", result)
	}

	if result.Metadata.AccountBlocked == nil || !*result.Metadata.AccountBlocked {
		t.Fatalf("expected blocked team metadata: %#v", result.Metadata)
	}

	if tunnel.request.Operation != platform.ProviderOperationCredentialProbe {
		t.Fatalf("operation = %q", tunnel.request.Operation)
	}
}

func TestTunnelCredentialProberRoutesZAIBillableProbeThroughSilkRoad(t *testing.T) {
	t.Parallel()

	tunnel := &recordingProbeTunnel{response: platform.TunnelResponse{StatusCode: http.StatusOK}}
	prober := NewTunnelCredentialProber(tunnel)

	result := prober.Probe(context.Background(), EncryptedCredential{
		Binding: Binding{
			ProviderSlug:        "zai",
			AuthMode:            "api_key",
			ConnectionID:        "connection-1",
			CredentialVersionID: "credential-version-1",
		},
	}, map[string]string{"api_key": "zai-secret"})

	if !result.Ready || result.Category != "ok" {
		t.Fatalf("probe failed: ready=%v category=%q summary=%q", result.Ready, result.Category, result.Summary)
	}

	if tunnel.request.Operation != platform.ProviderOperationCredentialProbe {
		t.Fatalf("operation = %q", tunnel.request.Operation)
	}

	if tunnel.request.Deployment.Provider != "zai" ||
		tunnel.request.Deployment.ProviderAuth == nil ||
		tunnel.request.Deployment.ProviderAuth.APIKey != "zai-secret" {
		t.Fatalf("unexpected deployment: %#v", tunnel.request.Deployment)
	}
}

func TestTunnelCredentialProberClassifiesZAIRejectedCredential(t *testing.T) {
	t.Parallel()

	tunnel := &recordingProbeTunnel{response: platform.TunnelResponse{StatusCode: http.StatusUnauthorized}}
	prober := NewTunnelCredentialProber(tunnel)

	result := prober.Probe(context.Background(), EncryptedCredential{
		Binding: Binding{ProviderSlug: "zai", AuthMode: "api_key"},
	}, map[string]string{"api_key": "zai-secret"})

	if result.Ready || result.Authenticated || result.Category != "invalid_credential" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
