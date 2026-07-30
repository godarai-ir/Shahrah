package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type openAICompatibleAdapter struct {
	defaultBaseURL         string
	authHeader             string
	apiVersionQuery        bool
	azureDeploymentRouting bool
	baseURLNormalizer      func(string) string
}

func newOpenAICompatibleAdapter(defaultBaseURL string) providerAdapter {
	return newOpenAICompatibleAdapterWithAuth(defaultBaseURL, "Authorization", false)

}

func newOpenAICompatibleAdapterWithAuth(defaultBaseURL, authHeader string, apiVersionQuery bool) providerAdapter {
	return &openAICompatibleAdapter{defaultBaseURL: strings.TrimSpace(defaultBaseURL), authHeader: authHeader, apiVersionQuery: apiVersionQuery}
}

func newAzureOpenAIAdapter() providerAdapter {
	return &openAICompatibleAdapter{authHeader: "api-key", apiVersionQuery: true, azureDeploymentRouting: true}
}

func (a *openAICompatibleAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	target, err := resolveBaseURL(payload.Deployment.APIBase, a.defaultBaseURL)
	if err != nil {
		return nil, err
	}
	if a.baseURLNormalizer != nil {
		target = a.baseURLNormalizer(target)
	}

	if a.azureDeploymentRouting && !strings.Contains(strings.ToLower(target), "/openai/v1") && !strings.Contains(strings.ToLower(target), "/openai/deployments/") {
		deployment := strings.TrimSpace(payload.Deployment.Model)
		if deployment == "" {
			return nil, fmt.Errorf("azure openai deployment model is required")
		}
		target = joinURL(target, "/openai/deployments/"+url.PathEscape(deployment))
	}
	method, operationPath, err := resolveOpenAIOperation(payload)
	if err != nil {
		return nil, err
	}
	endpoint := joinURL(target, operationPath)
	apiVersion := strings.TrimSpace(payload.Deployment.APIVersion)
	if a.azureDeploymentRouting && !strings.Contains(strings.ToLower(target), "/openai/v1") && apiVersion == "" {
		apiVersion = "2024-10-21"
	}
	if a.apiVersionQuery && !strings.Contains(strings.ToLower(target), "/openai/v1") && apiVersion != "" {
		separator := "?"
		if strings.Contains(endpoint, "?") {
			separator = "&"
		}
		endpoint += separator + "api-version=" + url.QueryEscape(apiVersion)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(payload.Body))
	if err != nil {
		return nil, err
	}
	if contentType := strings.TrimSpace(payload.Headers["Content-Type"]); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	if clientRequestID := strings.TrimSpace(payload.Headers["X-Client-Request-Id"]); clientRequestID != "" {
		req.Header.Set("X-Client-Request-Id", clientRequestID)
	}
	if apiKey := deploymentAPIKey(payload.Deployment); apiKey != "" {
		if strings.EqualFold(a.authHeader, "api-key") {
			req.Header.Set("api-key", apiKey)
		} else {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}
	applyDocumentedProviderHeaders(req, payload.Deployment)
	if payload.Stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return &adapterResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       resp.Body,
	}, nil
}

func resolveBaseURL(explicit string, fallback string) (string, error) {
	if target := strings.TrimSpace(explicit); target != "" {
		return strings.TrimRight(target, "/"), nil
	}
	if target := strings.TrimSpace(fallback); target != "" {
		return strings.TrimRight(target, "/"), nil
	}
	return "", fmt.Errorf("provider api_base is required")
}

// openAICompatiblePath remains for provider-specific adapters that only need
// the historical fixed endpoint subset. Unsupported metadata is converted to
// an inert path rather than being interpreted as an arbitrary upstream path.
func openAICompatiblePath(payload platform.TunnelRequest) string {
	_, path, err := resolveLegacyOpenAICompatibleOperation(payload.RequestMetadata)
	if err != nil {
		return "/invalid-provider-operation"
	}
	return path
}

func joinURL(base string, path string) string {
	return strings.TrimRight(base, "/") + path
}
