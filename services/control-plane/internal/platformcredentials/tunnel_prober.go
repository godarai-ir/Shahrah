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

// TunnelCredentialProber dispatches provider-specific probes through Silk Road.
type TunnelCredentialProber struct {
	registry *CredentialProberRegistry
}

// NewTunnelCredentialProber creates the production provider-prober registry.
func NewTunnelCredentialProber(tunnel platform.TunnelTransport) *TunnelCredentialProber {
	generic := &genericTunnelCredentialProber{tunnel: tunnel}
	registry := NewCredentialProberRegistry(generic)
	registry.Register("anthropic", &anthropicTunnelCredentialProber{tunnel: tunnel})
	registry.Register("xai", &xaiTunnelCredentialProber{tunnel: tunnel})

	return &TunnelCredentialProber{registry: registry}
}

// Register installs or replaces a provider-specific tunnel prober.
func (p *TunnelCredentialProber) Register(provider string, prober CredentialProber) {
	if p == nil || p.registry == nil {
		return
	}

	p.registry.Register(provider, prober)
}

// Probe validates a credential with its provider-specific implementation.
func (p *TunnelCredentialProber) Probe(
	ctx context.Context,
	record EncryptedCredential,
	secrets map[string]string,
) ProbeResult {
	if p == nil || p.registry == nil {
		return ProbeResult{
			Category: "network",
			Summary:  "Provider validation could not reach Silk Road.",
		}
	}

	return p.registry.Probe(ctx, record, secrets)
}

type genericTunnelCredentialProber struct {
	tunnel platform.TunnelTransport
}

func (p *genericTunnelCredentialProber) Probe(
	ctx context.Context,
	record EncryptedCredential,
	secrets map[string]string,
) ProbeResult {
	if p == nil || p.tunnel == nil {
		return ProbeResult{Category: "network", Summary: "Provider validation could not reach Silk Road."}
	}

	if _, ok := documentedProbe(record); !ok {
		return unsupportedProbeResult()
	}

	deployment, ok := credentialProbeDeployment(record, secrets)
	if !ok {
		return unsupportedProbeResult()
	}

	response, err := p.tunnel.Proxy(ctx, platform.TunnelRequest{
		Operation:  platform.ProviderOperationCredentialProbe,
		Deployment: deployment,
		Headers:    map[string]string{"Accept": "application/json"},
	})
	if err != nil {
		return ProbeResult{Category: "network", Summary: "Provider validation could not reach Silk Road."}
	}

	return genericProbeResult(response.StatusCode)
}

type xaiTunnelCredentialProber struct {
	tunnel platform.TunnelTransport
}

func (p *xaiTunnelCredentialProber) Probe(
	ctx context.Context,
	record EncryptedCredential,
	secrets map[string]string,
) ProbeResult {
	if p == nil || p.tunnel == nil {
		return ProbeResult{Category: "network", Summary: "xAI validation could not reach Silk Road."}
	}

	deployment, ok := credentialProbeDeployment(record, secrets)
	if !ok {
		return unsupportedProbeResult()
	}

	response, err := p.tunnel.Proxy(ctx, platform.TunnelRequest{
		Operation:  platform.ProviderOperationCredentialProbe,
		Deployment: deployment,
		Headers:    map[string]string{"Accept": "application/json"},
	})
	if err != nil {
		return ProbeResult{Category: "network", Summary: "xAI validation could not reach Silk Road."}
	}

	return classifyXAIAPIKeyProbe(response.StatusCode, response.Body, record.ModelHint)
}

type anthropicTunnelCredentialProber struct {
	tunnel platform.TunnelTransport
}

func (p *anthropicTunnelCredentialProber) Probe(
	ctx context.Context,
	record EncryptedCredential,
	secrets map[string]string,
) ProbeResult {
	if p == nil || p.tunnel == nil {
		return ProbeResult{Category: "network", Summary: "Anthropic validation could not reach Silk Road."}
	}

	deployment := platform.Deployment{
		Provider:            "anthropic",
		APIBase:             record.BaseURL,
		APIVersion:          record.APIVersion,
		ConnectionID:        record.ConnectionID,
		CredentialVersionID: record.CredentialVersionID,
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
		return ProbeResult{Category: "unsupported", Summary: "Anthropic authentication mode is not supported."}
	}

	models, headers, category, summary := p.call(ctx, deployment, platform.ProviderOperationAnthropicModelsList, nil)
	metadata := anthropicWIFProbeMetadata(headers)

	if category != "ok" {
		return ProbeResult{
			Authenticated: category != "invalid_credential" && category != "network",
			Category:      category,
			Summary:       summary,
			Metadata:      metadata,
		}
	}

	var modelList struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(models, &modelList); err != nil || len(modelList.Data) == 0 ||
		strings.TrimSpace(modelList.Data[0].ID) == "" {
		return ProbeResult{
			Authenticated: true,
			Category:      "provider",
			Summary:       "Anthropic Models API returned no probeable model.",
			Metadata:      metadata,
		}
	}

	countBody, _ := json.Marshal(map[string]interface{}{ //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
		"model": modelList.Data[0].ID,
		"messages": []map[string]interface{}{{
			"role": "user", "content": "credential scope probe",
		}},
	})
	if _, _, category, summary = p.call(ctx, deployment, platform.ProviderOperationAnthropicCountTokens, countBody); category != "ok" {
		if category == "forbidden" {
			category = "missing_scope"
			summary = "Credential is missing workspace:inference."
		}

		return ProbeResult{Authenticated: true, Category: category, Summary: summary, Metadata: metadata}
	}

	if _, _, category, summary = p.call(ctx, deployment, platform.ProviderOperationAnthropicFilesList, nil); category != "ok" {
		if category == "forbidden" {
			category = "missing_scope"
			summary = "Credential is missing workspace:developer."
		}

		return ProbeResult{Authenticated: true, Category: category, Summary: summary, Metadata: metadata}
	}

	return ProbeResult{
		Authenticated: true,
		Ready:         true,
		Category:      "ok",
		Summary:       "Anthropic credential accepted; Models, workspace:inference, and workspace:developer capabilities are available.",
		Metadata:      metadata,
	}
}

func (p *anthropicTunnelCredentialProber) call(
	ctx context.Context,
	deployment platform.Deployment,
	operation platform.ProviderOperation,
	body []byte,
) ([]byte, map[string]string, string, string) {
	response, err := p.tunnel.Proxy(ctx, platform.TunnelRequest{
		Operation: operation, Deployment: deployment, Body: body,
		Headers: map[string]string{"content-type": "application/json"},
	})
	if err != nil {
		return nil, nil, "network", "Anthropic validation could not reach Silk Road."
	}

	switch {
	case response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices:
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

func genericProbeResult(status int) ProbeResult {
	switch {
	case status >= http.StatusOK && status < http.StatusMultipleChoices:
		return ProbeResult{Authenticated: true, Ready: true, Category: "ok", Summary: "Provider accepted the credential."}
	case status == http.StatusUnauthorized:
		return ProbeResult{Category: "invalid_credential", Summary: "Provider rejected the credential."}
	case status == http.StatusForbidden:
		return ProbeResult{Authenticated: true, Category: "forbidden", Summary: "The credential authenticated but lacks permission for the validation endpoint."}
	case status == http.StatusTooManyRequests || status == http.StatusPaymentRequired:
		return ProbeResult{Authenticated: true, Ready: true, Category: "quota_exhausted", Summary: "Credential authenticated; quota or billing prevented the probe."}
	case status == http.StatusNotFound:
		return ProbeResult{Category: "deployment_not_found", Summary: "The configured provider endpoint was not found."}
	default:
		return ProbeResult{Category: "provider", Summary: fmt.Sprintf("Provider validation returned HTTP %d.", status)}
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
