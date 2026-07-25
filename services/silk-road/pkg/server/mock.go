package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type mockConfig struct {
	Latency              time.Duration
	ErrorRate            float64
	ErrorStatus          int
	AllowedTenantIDs     map[string]struct{}
	AllowedDeploymentIDs map[string]struct{}
}

type mockAdapter struct {
	cfg mockConfig
}

func mockProviderEnabled() bool {
	return parseBoolEnv("SILK_ROAD_ENABLE_MOCK_PROVIDER")
}

func loadMockConfigFromEnv() mockConfig {
	return mockConfig{
		Latency:              time.Duration(intEnv("SILK_ROAD_MOCK_LATENCY_MS", 50)) * time.Millisecond,
		ErrorRate:            floatEnv("SILK_ROAD_MOCK_ERROR_RATE", 0),
		ErrorStatus:          intEnv("SILK_ROAD_MOCK_ERROR_STATUS", http.StatusTooManyRequests),
		AllowedTenantIDs:     csvSetEnv("SILK_ROAD_MOCK_ALLOWED_TENANT_IDS"),
		AllowedDeploymentIDs: csvSetEnv("SILK_ROAD_MOCK_ALLOWED_DEPLOYMENT_IDS"),
	}
}

func newMockAdapter(cfg mockConfig) providerAdapter {
	if cfg.ErrorStatus < 400 || cfg.ErrorStatus > 599 {
		cfg.ErrorStatus = http.StatusTooManyRequests
	}
	if cfg.ErrorRate < 0 {
		cfg.ErrorRate = 0
	}
	if cfg.ErrorRate > 1 {
		cfg.ErrorRate = 1
	}

	return &mockAdapter{cfg: cfg}
}

func (a *mockAdapter) Dispatch(ctx context.Context, _ *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	if !a.requestAllowed(payload) {
		return jsonAdapterResponse(http.StatusForbidden, map[string]interface{}{
			"error": map[string]interface{}{
				"message": "mock provider is not authorized for this route",
				"type":    "permission_error",
				"code":    "mock_provider_access_denied",
			},
		})
	}

	if a.cfg.Latency > 0 {
		timer := time.NewTimer(a.cfg.Latency)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	if a.cfg.ErrorRate > 0 && rand.Float64() < a.cfg.ErrorRate {
		return jsonAdapterResponse(a.cfg.ErrorStatus, map[string]interface{}{
			"error": map[string]interface{}{
				"message": "mock provider injected error",
				"type":    "mock_provider_error",
				"code":    "mock_error",
			},
		})
	}

	path := openAICompatiblePath(payload)
	switch path {
	case "/responses":
		return jsonAdapterResponse(http.StatusOK, mockResponseBody(payload))
	case "/embeddings":
		return jsonAdapterResponse(http.StatusOK, mockEmbeddingsBody(payload))
	case "/moderations":
		return jsonAdapterResponse(http.StatusOK, mockModerationsBody(payload))
	case "/rerank", "/v2/rerank":
		return jsonAdapterResponse(http.StatusOK, mockRerankBody(payload))
	case "/images/generations", "/images/edits", "/images/variations":
		return jsonAdapterResponse(http.StatusOK, mockImagesBody(payload))
	default:
		return jsonAdapterResponse(http.StatusOK, mockChatBody(payload))
	}
}

func (a *mockAdapter) requestAllowed(payload platform.TunnelRequest) bool {
	tenantID := strings.TrimSpace(payload.TenantID)
	deploymentID := strings.TrimSpace(payload.Deployment.ID)
	if tenantID == "" || deploymentID == "" {
		return false
	}

	if _, ok := a.cfg.AllowedTenantIDs[tenantID]; !ok {
		if _, wildcard := a.cfg.AllowedTenantIDs["*"]; !wildcard {
			return false
		}
	}

	_, ok := a.cfg.AllowedDeploymentIDs[deploymentID]
	return ok
}

func jsonAdapterResponse(status int, body map[string]interface{}) (*adapterResponse, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("X-Shahrah-Mock-Provider", "true")

	return &adapterResponse{
		StatusCode: status,
		Headers:    headers,
		Body:       ioNopCloser(raw),
	}, nil
}

func mockChatBody(payload platform.TunnelRequest) map[string]interface{} {
	model := mockRequestModel(payload)
	now := time.Now().Unix()

	return map[string]interface{}{
		"id":      "chatcmpl_mock_" + strconv.FormatInt(now, 10),
		"object":  "chat.completion",
		"created": now,
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Mock provider response for Shahrah load testing.",
				},
				"finish_reason": "stop",
			},
		},
		"usage": mockUsage(),
	}
}

func mockResponseBody(payload platform.TunnelRequest) map[string]interface{} {
	model := mockRequestModel(payload)
	now := time.Now().Unix()

	return map[string]interface{}{
		"id":         "resp_mock_" + strconv.FormatInt(now, 10),
		"object":     "response",
		"created_at": now,
		"status":     "completed",
		"model":      model,
		"output": []map[string]interface{}{
			{
				"id":     "msg_mock",
				"type":   "message",
				"status": "completed",
				"role":   "assistant",
				"content": []map[string]interface{}{
					{
						"type": "output_text",
						"text": "Mock provider response for Shahrah load testing.",
					},
				},
			},
		},
		"usage": map[string]interface{}{
			"input_tokens":  16,
			"output_tokens": 12,
			"total_tokens":  28,
		},
	}
}

func mockEmbeddingsBody(payload platform.TunnelRequest) map[string]interface{} {
	model := mockRequestModel(payload)
	count := mockEmbeddingInputCount(payload)
	data := make([]map[string]interface{}, 0, count)
	for i := 0; i < count; i++ {
		data = append(data, map[string]interface{}{
			"object":    "embedding",
			"index":     i,
			"embedding": []float64{0.01, 0.02, 0.03, float64(i) / 100},
		})
	}

	return map[string]interface{}{
		"object": "list",
		"data":   data,
		"model":  model,
		"usage": map[string]interface{}{
			"prompt_tokens": count * 8,
			"total_tokens":  count * 8,
		},
	}
}

func mockModerationsBody(payload platform.TunnelRequest) map[string]interface{} {
	model := mockRequestModel(payload)
	return map[string]interface{}{
		"id":    "modr_mock",
		"model": model,
		"results": []map[string]interface{}{
			{
				"flagged": false,
				"categories": map[string]bool{
					"hate":     false,
					"violence": false,
				},
				"category_scores": map[string]float64{
					"hate":     0,
					"violence": 0,
				},
			},
		},
	}
}

func mockRerankBody(payload platform.TunnelRequest) map[string]interface{} {
	count := mockRerankDocumentCount(payload)
	results := make([]map[string]interface{}, 0, count)
	for i := 0; i < count; i++ {
		results = append(results, map[string]interface{}{
			"index":           i,
			"relevance_score": 1.0 / float64(i+1),
		})
	}

	return map[string]interface{}{
		"id":      "rerank_mock",
		"results": results,
		"usage": map[string]interface{}{
			"prompt_tokens": count * 16,
			"total_tokens":  count * 16,
		},
	}
}

func mockImagesBody(payload platform.TunnelRequest) map[string]interface{} {
	return map[string]interface{}{
		"created": time.Now().Unix(),
		"data": []map[string]interface{}{
			{
				"url":            "https://example.invalid/shahrah/mock-image.png",
				"revised_prompt": "Mock image response for Shahrah load testing.",
			},
		},
		"usage": map[string]interface{}{
			"cost_in_usd_ticks": 0,
			"total_tokens":      0,
		},
	}
}

func mockRequestModel(payload platform.TunnelRequest) string {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload.Body, &raw); err == nil {
		if model, ok := raw["model"].(string); ok && strings.TrimSpace(model) != "" {
			return strings.TrimSpace(model)
		}
	}
	if strings.TrimSpace(payload.Deployment.Model) != "" {
		return strings.TrimSpace(payload.Deployment.Model)
	}
	return "mock-model"
}

func mockEmbeddingInputCount(payload platform.TunnelRequest) int {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload.Body, &raw); err != nil {
		return 1
	}

	switch input := raw["input"].(type) {
	case []interface{}:
		if len(input) > 0 {
			return len(input)
		}
	case string:
		if strings.TrimSpace(input) != "" {
			return 1
		}
	}

	return 1
}

func mockRerankDocumentCount(payload platform.TunnelRequest) int {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload.Body, &raw); err != nil {
		return 1
	}

	docs, _ := raw["documents"].([]interface{})
	topN := len(docs)
	if topN == 0 {
		topN = 1
	}
	if rawTopN, ok := raw["top_n"].(float64); ok && int(rawTopN) > 0 && int(rawTopN) < topN {
		topN = int(rawTopN)
	}
	return topN
}

func mockUsage() map[string]interface{} {
	return map[string]interface{}{
		"prompt_tokens":     16,
		"completion_tokens": 12,
		"total_tokens":      28,
	}
}

func parseBoolEnv(key string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(key))) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func intEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func floatEnv(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}

func csvSetEnv(key string) map[string]struct{} {
	values := make(map[string]struct{})
	for _, item := range strings.Split(os.Getenv(key), ",") {
		if normalized := strings.TrimSpace(item); normalized != "" {
			values[normalized] = struct{}{}
		}
	}
	return values
}

func ioNopCloser(raw []byte) io.ReadCloser {
	return io.NopCloser(bytes.NewReader(raw))
}
