package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type openAICompatibleAdapter struct {
	defaultBaseURL string
}

func newOpenAICompatibleAdapter(defaultBaseURL string) providerAdapter {
	return &openAICompatibleAdapter{defaultBaseURL: strings.TrimSpace(defaultBaseURL)}
}

func (a *openAICompatibleAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	target, err := resolveBaseURL(payload.Deployment.APIBase, a.defaultBaseURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(target, openAICompatiblePath(payload)), bytes.NewReader(payload.Body))
	if err != nil {
		return nil, err
	}
	if contentType := strings.TrimSpace(payload.Headers["Content-Type"]); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	if payload.Deployment.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+payload.Deployment.APIKey)
	}
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

func openAICompatiblePath(payload platform.TunnelRequest) string {
	if payload.RequestMetadata != nil {
		if raw, ok := payload.RequestMetadata["provider_endpoint_path"].(string); ok {
			path := strings.TrimSpace(raw)
			if path != "" {
				if !strings.HasPrefix(path, "/") {
					return "/" + path
				}
				return path
			}
		}
	}
	return "/chat/completions"
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

func joinURL(base string, path string) string {
	return strings.TrimRight(base, "/") + path
}
