package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type cohereAdapter struct{}

func newCohereAdapter() providerAdapter {
	return &cohereAdapter{}
}

func (a *cohereAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	path := openAICompatiblePath(payload)
	if path != "/v2/rerank" && path != "/rerank" {
		return nil, fmt.Errorf("cohere adapter only supports rerank")
	}

	target, err := resolveBaseURL(payload.Deployment.APIBase, "https://api.cohere.com")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(target, "/v2/rerank"), bytes.NewReader(payload.Body))
	if err != nil {
		return nil, err
	}
	if contentType := strings.TrimSpace(payload.Headers["Content-Type"]); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	if apiKey := deploymentAPIKey(payload.Deployment); apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	applyDocumentedProviderHeaders(req, payload.Deployment)

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
