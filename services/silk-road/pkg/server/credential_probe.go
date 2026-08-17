package server

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type credentialProbeSpec struct {
	url         string
	auth        string
	method      string
	contentType string
	body        []byte
}

func dispatchCredentialProbe(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
) (*adapterResponse, error) {
	spec, err := resolveCredentialProbe(payload.Deployment)
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = &http.Client{}
	}

	method := spec.method
	if method == "" {
		method = http.MethodGet
	}

	req, err := http.NewRequestWithContext(ctx, method, spec.url, bytes.NewReader(spec.body))
	if err != nil {
		return nil, fmt.Errorf("build credential probe: %w", err)
	}

	secret := deploymentAPIKey(payload.Deployment)
	if secret == "" {
		return nil, fmt.Errorf("credential probe auth material is unavailable")
	}

	switch spec.auth {
	case "api-key":
		req.Header.Set("api-key", secret)
	case "x-goog-api-key":
		req.Header.Set("x-goog-api-key", secret)
	case "bearer":
		req.Header.Set("Authorization", "Bearer "+secret)
	default:
		return nil, fmt.Errorf("credential probe auth mode is unsupported")
	}

	req.Header.Set("Accept", "application/json")

	if spec.contentType != "" {
		req.Header.Set("Content-Type", spec.contentType)
	}

	applyDocumentedProviderHeaders(req, payload.Deployment)

	resp, err := client.Do(req) //nolint:bodyclose // The server handler owns and closes the body transferred through adapterResponse.
	if err != nil {
		return nil, err
	}

	return &adapterResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       resp.Body,
	}, nil
}

func resolveCredentialProbe(deployment platform.Deployment) (credentialProbeSpec, error) {
	provider := normalizeCredentialProbeProvider(deployment.Provider)
	base := strings.TrimRight(strings.TrimSpace(deployment.APIBase), "/")

	var spec credentialProbeSpec

	switch provider {
	case "azure-open-ai":
		if base == "" {
			return credentialProbeSpec{}, fmt.Errorf("provider api_base is required")
		}

		if strings.Contains(strings.ToLower(base), "/openai/v1") {
			spec = credentialProbeSpec{url: base + "/models", auth: "api-key"}
		} else {
			version := strings.TrimSpace(deployment.APIVersion)
			if version == "" {
				version = "2024-10-21"
			}

			spec = credentialProbeSpec{
				url:  base + "/openai/models?api-version=" + url.QueryEscape(version),
				auth: "api-key",
			}
		}
	case "cerebras":
		spec = credentialModelsProbe(base, "https://api.cerebras.ai/v1")
	case "cohere":
		if base == "" {
			base = "https://api.cohere.com"
		}

		spec = credentialProbeSpec{url: base + "/v1/models", auth: "bearer"}
	case "databricks":
		parsed, err := url.Parse(base)
		if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
			return credentialProbeSpec{}, fmt.Errorf("provider api_base is invalid")
		}

		parsed.Path = "/api/2.0/serving-endpoints"
		parsed.RawPath, parsed.RawQuery, parsed.Fragment = "", "", ""
		spec = credentialProbeSpec{url: parsed.String(), auth: "bearer"}
	case "deepinfra":
		spec = credentialModelsProbe(base, "https://api.deepinfra.com/v1/openai")
	case "deepseek":
		spec = credentialModelsProbe(base, "https://api.deepseek.com")
	case "zai":
		if base == "" {
			base = "https://api.z.ai/api/paas/v4"
		}

		if err := validateZAIBaseURL(base); err != nil {
			return credentialProbeSpec{}, err
		}

		spec = credentialProbeSpec{
			url:         strings.TrimRight(base, "/") + "/chat/completions",
			auth:        "bearer",
			method:      http.MethodPost,
			contentType: "application/json",
			body:        []byte(`{"model":"glm-4.7-flashx","messages":[{"role":"user","content":"ping"}],"thinking":{"type":"disabled"},"do_sample":false,"max_tokens":1}`),
		}
	case "google-gemini":
		if base == "" {
			base = "https://generativelanguage.googleapis.com"
		}

		lower := strings.ToLower(base)
		for _, suffix := range []string{"/v1beta/openai", "/v1/openai", "/v1beta", "/v1"} {
			if strings.HasSuffix(lower, suffix) {
				base = base[:len(base)-len(suffix)]
				break
			}
		}

		spec = credentialProbeSpec{url: base + "/v1beta/models?pageSize=1", auth: "x-goog-api-key"}
	case "groq":
		spec = credentialModelsProbe(base, "https://api.groq.com/openai/v1")
	case "mistral-ai":
		spec = credentialModelsProbe(base, "https://api.mistral.ai/v1")
	case "moonshot":
		spec = credentialModelsProbe(base, "https://api.moonshot.ai/v1")
	case "openai":
		spec = credentialModelsProbe(base, "https://api.openai.com/v1")
	case "openrouter":
		spec = credentialModelsProbe(base, "https://openrouter.ai/api/v1")
	case "sambanova":
		if base == "" {
			return credentialProbeSpec{}, fmt.Errorf("provider api_base is required")
		}

		spec = credentialModelsProbe(base, "")
	case "together-ai":
		spec = credentialModelsProbe(base, "https://api.together.xyz/v1")
	case "xai":
		if base == "" {
			base = "https://api.x.ai/v1"
		}

		spec = credentialProbeSpec{
			url:  strings.TrimRight(base, "/") + "/api-key",
			auth: "bearer",
		}
	default:
		return credentialProbeSpec{}, fmt.Errorf("provider credential probe is unsupported")
	}

	if !safeCredentialProbeURL(provider, spec.url) {
		return credentialProbeSpec{}, fmt.Errorf("provider credential probe URL is not allowed")
	}

	return spec, nil
}

func credentialModelsProbe(base, fallback string) credentialProbeSpec {
	if base == "" {
		base = fallback
	}

	return credentialProbeSpec{
		url:  strings.TrimRight(base, "/") + "/models",
		auth: "bearer",
	}
}

func normalizeCredentialProbeProvider(provider string) string {
	switch normalizeProvider(provider) {
	case "azure":
		return "azure-open-ai"
	case "gemini":
		return "google-gemini"
	case "mistral":
		return "mistral-ai"
	case "togetherai":
		return "together-ai"
	default:
		return normalizeProvider(provider)
	}
}

func safeCredentialProbeURL(provider, raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	if net.ParseIP(host) != nil || host == "localhost" ||
		strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return false
	}

	exact := map[string]string{
		"cerebras": "api.cerebras.ai", "cohere": "api.cohere.com",
		"deepinfra": "api.deepinfra.com", "deepseek": "api.deepseek.com",
		"zai":           "api.z.ai",
		"google-gemini": "generativelanguage.googleapis.com", "groq": "api.groq.com",
		"mistral-ai": "api.mistral.ai", "moonshot": "api.moonshot.ai", "openai": "api.openai.com",
		"openrouter": "openrouter.ai", "together-ai": "api.together.xyz", "xai": "api.x.ai",
	}
	if expected := exact[provider]; expected != "" {
		return host == expected
	}

	switch provider {
	case "azure-open-ai":
		return strings.HasSuffix(host, ".openai.azure.com") || strings.HasSuffix(host, ".services.ai.azure.com")
	case "databricks":
		return strings.HasSuffix(host, ".cloud.databricks.com") ||
			strings.HasSuffix(host, ".azuredatabricks.net") || strings.HasSuffix(host, ".gcp.databricks.com")
	case "sambanova":
		return true
	default:
		return false
	}
}
