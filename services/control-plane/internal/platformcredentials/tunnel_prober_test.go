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
