package platformcredentials

import (
	"context"
	"errors"
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

	ok, category, _, metadata := prober.ProbeDetailed(
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
	if !ok || category != "ok" {
		t.Fatalf("probe failed: ok=%v category=%q", ok, category)
	}

	if len(tunnel.calls) != 3 {
		t.Fatalf("expected three capability probes, got %d", len(tunnel.calls))
	}

	if metadata.WIFExpiresAt == nil ||
		!metadata.WIFExpiresAt.Equal(time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected expiry: %#v", metadata.WIFExpiresAt)
	}

	if metadata.WIFRefreshedAt == nil ||
		!metadata.WIFRefreshedAt.Equal(time.Date(2026, 7, 29, 11, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected refresh time: %#v", metadata.WIFRefreshedAt)
	}

	if len(metadata.WIFScopes) != 2 ||
		metadata.WIFScopes[0] != "workspace:developer" ||
		metadata.WIFScopes[1] != "workspace:inference" {
		t.Fatalf("unexpected scopes: %#v", metadata.WIFScopes)
	}
}

func TestTunnelCredentialProberRoutesGeminiProbeThroughSilkRoad(t *testing.T) {
	t.Parallel()

	tunnel := &recordingProbeTunnel{response: platform.TunnelResponse{StatusCode: 200}}
	prober := NewTunnelCredentialProber(tunnel)
	ok, category, _ := prober.Probe(context.Background(), EncryptedCredential{
		Binding: Binding{
			ProviderSlug:        "google-gemini",
			AuthMode:            "api_key",
			ConnectionID:        "connection-1",
			CredentialVersionID: "credential-version-1",
		},
		BaseURL: "https://generativelanguage.googleapis.com",
	}, map[string]string{"api_key": "gemini-secret"})

	if !ok || category != "ok" {
		t.Fatalf("probe failed: ok=%v category=%q", ok, category)
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

	ok, category, _ := NewTunnelCredentialProber(nil).Probe(context.Background(), EncryptedCredential{
		Binding: Binding{ProviderSlug: "google-gemini", AuthMode: "api_key"},
	}, map[string]string{"api_key": "gemini-secret"})
	if ok || category != "network" {
		t.Fatalf("ok=%v category=%q, want false/network", ok, category)
	}
}
