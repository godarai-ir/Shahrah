package platformcredentials

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type TunnelCredentialProber struct {
	tunnel platform.TunnelTransport
}

type ProbeMetadata struct {
	WIFExpiresAt   *time.Time
	WIFRefreshedAt *time.Time
	WIFScopes      []string
}

type DetailedCredentialProber interface {
	ProbeDetailed(
		context.Context,
		EncryptedCredential,
		map[string]string,
	) (bool, string, string, ProbeMetadata)
}

func NewTunnelCredentialProber(tunnel platform.TunnelTransport) *TunnelCredentialProber {
	return &TunnelCredentialProber{tunnel: tunnel}
}

func (p *TunnelCredentialProber) Probe(ctx context.Context, record EncryptedCredential, secrets map[string]string) (bool, string, string) {
	success, category, summary, _ := p.ProbeDetailed(ctx, record, secrets)
	return success, category, summary
}

func (p *TunnelCredentialProber) ProbeDetailed(
	ctx context.Context,
	record EncryptedCredential,
	secrets map[string]string,
) (bool, string, string, ProbeMetadata) {
	if p == nil || p.tunnel == nil || normalizeProvider(record.ProviderSlug) != "anthropic" {
		success, category, summary := NewHTTPProber(nil).Probe(
			ctx,
			record,
			secrets,
		)
		return success, category, summary, ProbeMetadata{}
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
		return false, "unsupported", "Anthropic authentication mode is not supported.", ProbeMetadata{}
	}

	models, headers, category, summary := p.call(ctx, deployment, platform.ProviderOperationAnthropicModelsList, nil)
	metadata := anthropicWIFProbeMetadata(headers)
	if category != "ok" {
		return false, category, summary, metadata
	}
	var modelList struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(models, &modelList); err != nil || len(modelList.Data) == 0 ||
		strings.TrimSpace(modelList.Data[0].ID) == "" {
		return false, "provider", "Anthropic Models API returned no probeable model.", metadata
	}
	countBody, _ := json.Marshal(map[string]interface{}{
		"model": modelList.Data[0].ID,
		"messages": []map[string]interface{}{{
			"role": "user", "content": "credential scope probe",
		}},
	})
	if _, _, category, summary = p.call(ctx, deployment, platform.ProviderOperationAnthropicCountTokens, countBody); category != "ok" {
		if category == "forbidden" {
			return false, "missing_scope", "Credential is missing workspace:inference.", metadata
		}
		return false, category, summary, metadata
	}
	if _, _, category, summary = p.call(ctx, deployment, platform.ProviderOperationAnthropicFilesList, nil); category != "ok" {
		if category == "forbidden" {
			return false, "missing_scope", "Credential is missing workspace:developer.", metadata
		}
		return false, category, summary, metadata
	}
	return true, "ok", "Anthropic credential accepted; Models, workspace:inference, and workspace:developer capabilities are available.", metadata
}

func (p *TunnelCredentialProber) call(ctx context.Context, deployment platform.Deployment, operation platform.ProviderOperation, body []byte) ([]byte, map[string]string, string, string) {
	response, err := p.tunnel.Proxy(ctx, platform.TunnelRequest{
		Operation: operation, Deployment: deployment, Body: body,
		Headers: map[string]string{"content-type": "application/json"},
	})
	if err != nil {
		return nil, nil, "network", "Anthropic validation could not reach Silk Road."
	}
	switch {
	case response.StatusCode >= 200 && response.StatusCode < 300:
		return response.Body, response.Headers, "ok", ""
	case response.StatusCode == http.StatusUnauthorized:
		return nil, response.Headers, "invalid_credential", "Anthropic rejected the credential."
	case response.StatusCode == http.StatusForbidden:
		return nil, response.Headers, "forbidden", "Anthropic credential lacks permission for a required capability."
	case response.StatusCode == http.StatusTooManyRequests || response.StatusCode == http.StatusPaymentRequired:
		return response.Body, response.Headers, "ok", ""
	default:
		return nil, response.Headers, "provider", fmt.Sprintf("Anthropic validation returned HTTP %d.", response.StatusCode)
	}
}

func anthropicWIFProbeMetadata(headers map[string]string) ProbeMetadata {
	metadata := ProbeMetadata{}
	for name, value := range headers {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "x-godarai-anthropic-wif-expires-at":
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				metadata.WIFExpiresAt = &parsed
			}
		case "x-godarai-anthropic-wif-refreshed-at":
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				metadata.WIFRefreshedAt = &parsed
			}
		case "x-godarai-anthropic-wif-scopes":
			metadata.WIFScopes = strings.Fields(value)
		}
	}
	return metadata
}
