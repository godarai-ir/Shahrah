package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	if path == audioTranscriptionsPath {
		return cohereAudioTranscriptionsDispatch(ctx, client, payload)
	}

	if path != "/v2/rerank" && path != "/rerank" {
		return nil, fmt.Errorf("cohere adapter only supports rerank and audio transcriptions")
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

func (a *cohereAdapter) DispatchRaw(
	ctx context.Context,
	client *http.Client,
	payload platform.TunnelRequest,
	body io.Reader,
	contentLength int64,
) (*adapterResponse, error) {
	if openAICompatiblePath(payload) == audioTranscriptionsPath {
		return cohereAudioTranscriptionsDispatchRaw(ctx, client, payload, body)
	}

	buffered, err := readBufferedProviderBody(body, contentLength)
	if err != nil {
		return nil, err
	}

	payload.Body = buffered

	return a.Dispatch(ctx, client, payload)
}
