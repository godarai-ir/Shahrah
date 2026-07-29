package platformcredentials

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type TunnelCredentialProber struct {
	tunnel platform.TunnelTransport
}

func NewTunnelCredentialProber(tunnel platform.TunnelTransport) *TunnelCredentialProber {
	return &TunnelCredentialProber{tunnel: tunnel}
}

func (p *TunnelCredentialProber) Probe(ctx context.Context, record EncryptedCredential, secrets map[string]string) (bool, string, string) {
	if p == nil || p.tunnel == nil || normalizeProvider(record.ProviderSlug) != "anthropic" {
		return NewHTTPProber(nil).Probe(ctx, record, secrets)
	}
	deployment := platform.Deployment{
		Provider: "anthropic", APIBase: record.BaseURL, APIVersion: record.APIVersion,
		ConnectionID: record.ConnectionID, CredentialVersionID: record.CredentialVersionID,
		ProviderHeaders: map[string]string{
			"Anthropic-Beta-Allowlist": "files-api-2025-04-14",
		},
	}
	switch record.AuthMode {
	case "api_key":
		deployment.ProviderAuth = &platform.ResolvedProviderAuth{
			Mode: string(platform.ProviderAuthModeAPIKey), APIKey: secrets["api_key"],
		}
	case string(platform.ProviderAuthModeAnthropicWIF):
		deployment.ProviderAuth = &platform.ResolvedProviderAuth{
			Mode: string(platform.ProviderAuthModeAnthropicWIF),
			AnthropicWIF: &platform.ResolvedAnthropicWIFAuth{
				FederationRuleID: record.FederationRuleID,
				OrganizationID:   record.OrganizationID,
				ServiceAccountID: record.ServiceAccountID,
				WorkspaceID:      record.WorkspaceID,
				TokenSourceID:    record.TokenSourceID,
			},
		}
	default:
		return false, "unsupported", "Anthropic authentication mode is not supported."
	}

	models, category, summary := p.call(ctx, deployment, platform.ProviderOperationAnthropicModelsList, nil)
	if category != "ok" {
		return false, category, summary
	}
	var modelList struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(models, &modelList); err != nil || len(modelList.Data) == 0 ||
		strings.TrimSpace(modelList.Data[0].ID) == "" {
		return false, "provider", "Anthropic Models API returned no probeable model."
	}
	countBody, _ := json.Marshal(map[string]interface{}{
		"model": modelList.Data[0].ID,
		"messages": []map[string]interface{}{{
			"role": "user", "content": "credential scope probe",
		}},
	})
	if _, category, summary = p.call(ctx, deployment, platform.ProviderOperationAnthropicCountTokens, countBody); category != "ok" {
		if category == "forbidden" {
			return false, "missing_scope", "Credential is missing workspace:inference."
		}
		return false, category, summary
	}
	if _, category, summary = p.call(ctx, deployment, platform.ProviderOperationAnthropicFilesList, nil); category != "ok" {
		if category == "forbidden" {
			return false, "missing_scope", "Credential is missing workspace:developer."
		}
		return false, category, summary
	}
	return true, "ok", "Anthropic credential accepted; Models, workspace:inference, and workspace:developer capabilities are available."
}

func (p *TunnelCredentialProber) call(ctx context.Context, deployment platform.Deployment, operation platform.ProviderOperation, body []byte) ([]byte, string, string) {
	response, err := p.tunnel.Proxy(ctx, platform.TunnelRequest{
		Operation: operation, Deployment: deployment, Body: body,
		Headers: map[string]string{"content-type": "application/json"},
	})
	if err != nil {
		return nil, "network", "Anthropic validation could not reach Silk Road."
	}
	switch {
	case response.StatusCode >= 200 && response.StatusCode < 300:
		return response.Body, "ok", ""
	case response.StatusCode == http.StatusUnauthorized:
		return nil, "invalid_credential", "Anthropic rejected the credential."
	case response.StatusCode == http.StatusForbidden:
		return nil, "forbidden", "Anthropic credential lacks permission for a required capability."
	case response.StatusCode == http.StatusTooManyRequests || response.StatusCode == http.StatusPaymentRequired:
		return response.Body, "ok", ""
	default:
		return nil, "provider", fmt.Sprintf("Anthropic validation returned HTTP %d.", response.StatusCode)
	}
}
