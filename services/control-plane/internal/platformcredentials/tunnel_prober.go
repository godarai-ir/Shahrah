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
	if p == nil || p.tunnel == nil {
		return false, "network", "Provider validation could not reach Silk Road.", ProbeMetadata{}
	}

	if normalizeProvider(record.ProviderSlug) != "anthropic" {
		success, category, summary := p.probeProvider(ctx, record, secrets)

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

	countBody, _ := json.Marshal(map[string]interface{}{ //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
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

func (p *TunnelCredentialProber) probeProvider(
	ctx context.Context,
	record EncryptedCredential,
	secrets map[string]string,
) (bool, string, string) {
	if _, ok := documentedProbe(record); !ok {
		return false, "unsupported", "This provider authentication mode has no implemented non-billable validation probe."
	}

	deployment, ok := credentialProbeDeployment(record, secrets)
	if !ok {
		return false, "unsupported", "This provider authentication mode has no implemented non-billable validation probe."
	}

	response, err := p.tunnel.Proxy(ctx, platform.TunnelRequest{
		Operation:  platform.ProviderOperationCredentialProbe,
		Deployment: deployment,
		Headers:    map[string]string{"Accept": "application/json"},
	})
	if err != nil {
		return false, "network", "Provider validation could not reach Silk Road."
	}

	switch {
	case response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices:
		return true, "ok", "Provider accepted the credential."
	case response.StatusCode == http.StatusUnauthorized:
		return false, "invalid_credential", "Provider rejected the credential."
	case response.StatusCode == http.StatusForbidden:
		return false, "forbidden", "The credential authenticated but lacks permission for the validation endpoint."
	case response.StatusCode == http.StatusTooManyRequests || response.StatusCode == http.StatusPaymentRequired:
		return true, "quota_exhausted", "Credential authenticated; quota or billing prevented the probe."
	case response.StatusCode == http.StatusNotFound:
		return false, "deployment_not_found", "The configured provider endpoint was not found."
	default:
		return false, "provider", fmt.Sprintf("Provider validation returned HTTP %d.", response.StatusCode)
	}
}

func credentialProbeDeployment(
	record EncryptedCredential,
	secrets map[string]string,
) (platform.Deployment, bool) {
	deployment := platform.Deployment{
		Provider:            record.ProviderSlug,
		APIBase:             record.BaseURL,
		APIVersion:          record.APIVersion,
		ConnectionID:        record.ConnectionID,
		CredentialVersionID: record.CredentialVersionID,
		ProviderHeaders:     credentialProbeHeaders(record),
	}

	switch record.AuthMode {
	case string(platform.ProviderAuthModeAPIKey):
		deployment.ProviderAuth = &platform.ResolvedProviderAuth{
			Mode:   string(platform.ProviderAuthModeAPIKey),
			APIKey: strings.TrimSpace(secrets["api_key"]),
		}
	case "personal_access_token":
		deployment.ProviderAuth = &platform.ResolvedProviderAuth{
			Mode:   "personal_access_token",
			APIKey: strings.TrimSpace(secrets["access_token"]),
		}
	default:
		return platform.Deployment{}, false
	}

	return deployment, deployment.ProviderAuth.APIKey != ""
}

func credentialProbeHeaders(record EncryptedCredential) map[string]string {
	headers := map[string]string{}

	for name, value := range map[string]string{
		"OpenAI-Organization": record.Organization,
		"OpenAI-Project":      record.Project,
		"HTTP-Referer":        record.HTTPReferer,
		"X-OpenRouter-Title":  record.AppTitle,
		"X-Client-Name":       record.ClientName,
	} {
		if value = strings.TrimSpace(value); value != "" {
			headers[name] = value
		}
	}

	return headers
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
