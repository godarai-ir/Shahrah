package video

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type ProviderResult struct {
	ProviderJobID  string
	Status         string
	ActualUsage    []platform.UsageMeasure
	Artifacts      []platform.VideoArtifact
	Raw            map[string]interface{}
	PollAfter      time.Duration
	RequestIsExact bool
}

type ProviderRejectedError struct {
	StatusCode int
	Message    string
}

func (e *ProviderRejectedError) Error() string {
	return fmt.Sprintf("video provider returned status %d: %s", e.StatusCode, e.Message)
}

func IsProviderRejection(err error) bool {
	var rejected *ProviderRejectedError
	return errors.As(err, &rejected)
}

type ProviderAdapter interface {
	Submit(ctx context.Context, requestID, tenantID string, dep platform.Deployment, req platform.VideoCreateRequest, inputURL string) (ProviderResult, error)
	GetStatus(ctx context.Context, requestID, tenantID string, dep platform.Deployment, providerJobID string) (ProviderResult, error)
	Cancel(ctx context.Context, requestID, tenantID string, dep platform.Deployment, providerJobID string) (ProviderResult, error)
	GetActualUsage(ctx context.Context, requestID, tenantID string, dep platform.Deployment, providerJobID string) ([]platform.UsageMeasure, error)
	GetOutputs(ctx context.Context, requestID, tenantID string, dep platform.Deployment, providerJobID string) ([]platform.VideoArtifact, error)
}

type AdapterRegistry struct {
	tunnel  platform.TunnelTransport
	enabled map[string]struct{}
}

func NewAdapterRegistry(tunnel platform.TunnelTransport) *AdapterRegistry {
	var enabled map[string]struct{}
	if raw := strings.TrimSpace(os.Getenv("VIDEO_ENABLED_PROVIDERS")); raw != "" {
		enabled = map[string]struct{}{}
		for _, provider := range strings.Split(raw, ",") {
			if normalized := normalizeProvider(provider); normalized != "" {
				enabled[normalized] = struct{}{}
			}
		}
	}
	return &AdapterRegistry{tunnel: tunnel, enabled: enabled}
}

func (r *AdapterRegistry) Adapter(provider string) (ProviderAdapter, error) {
	if r == nil || r.tunnel == nil {
		return nil, errors.New("video provider transport unavailable")
	}
	normalized := normalizeProvider(provider)
	if r.enabled != nil {
		if _, ok := r.enabled[normalized]; !ok {
			return nil, fmt.Errorf("video provider %q is disabled by rollout policy", provider)
		}
	}
	switch normalized {
	case "openai", "azure", "azure-open-ai", "xai", "deepinfra", "togetherai", "together-ai",
		"google-gemini", "gemini", "google-vertex", "vertex", "vertex-ai", "vertexai",
		"aws-bedrock", "bedrock":
		return &tunnelVideoAdapter{tunnel: r.tunnel, provider: normalizeProvider(provider)}, nil
	default:
		return nil, fmt.Errorf("video provider %q is not supported", provider)
	}
}

type tunnelVideoAdapter struct {
	tunnel   platform.TunnelTransport
	provider string
}

func (a *tunnelVideoAdapter) Submit(ctx context.Context, requestID, tenantID string, dep platform.Deployment, req platform.VideoCreateRequest, inputURL string) (ProviderResult, error) {
	endpoint, payload, exact, err := a.submitPayload(dep, req, inputURL)
	if err != nil {
		return ProviderResult{}, err
	}
	if (a.provider == "aws-bedrock" || a.provider == "bedrock") && strings.TrimSpace(requestID) != "" {
		if body, ok := payload.(map[string]interface{}); ok {
			body["clientRequestToken"] = requestID
		}
	}
	result, err := a.call(ctx, requestID, tenantID, dep, http.MethodPost, endpoint, payload)
	result.RequestIsExact = exact
	return result, err
}

func (a *tunnelVideoAdapter) GetStatus(ctx context.Context, requestID, tenantID string, dep platform.Deployment, providerJobID string) (ProviderResult, error) {
	switch a.provider {
	case "google-vertex", "vertex", "vertex-ai", "vertexai":
		if strings.TrimSpace(dep.ProjectID) == "" || strings.TrimSpace(dep.RegionName) == "" {
			return ProviderResult{}, errors.New("vertex video requires project_id and region")
		}
		endpoint := fmt.Sprintf("/v1/projects/%s/locations/%s/publishers/google/models/%s:fetchPredictOperation",
			url.PathEscape(dep.ProjectID), url.PathEscape(dep.RegionName), url.PathEscape(dep.Model))
		return a.call(ctx, requestID, tenantID, dep, http.MethodPost, endpoint, map[string]interface{}{
			"operationName": providerJobID,
		})
	}
	endpoint := a.statusPath(dep, providerJobID)
	return a.call(ctx, requestID, tenantID, dep, http.MethodGet, endpoint, nil)
}

func (a *tunnelVideoAdapter) GetActualUsage(ctx context.Context, requestID, tenantID string, dep platform.Deployment, providerJobID string) ([]platform.UsageMeasure, error) {
	result, err := a.GetStatus(ctx, requestID, tenantID, dep, providerJobID)
	return result.ActualUsage, err
}

func (a *tunnelVideoAdapter) GetOutputs(ctx context.Context, requestID, tenantID string, dep platform.Deployment, providerJobID string) ([]platform.VideoArtifact, error) {
	result, err := a.GetStatus(ctx, requestID, tenantID, dep, providerJobID)
	return result.Artifacts, err
}

func (a *tunnelVideoAdapter) Cancel(ctx context.Context, requestID, tenantID string, dep platform.Deployment, providerJobID string) (ProviderResult, error) {
	method := http.MethodDelete
	endpoint := a.statusPath(dep, providerJobID)
	var payload interface{}
	switch a.provider {
	case "aws-bedrock", "bedrock":
		method = http.MethodPost
		endpoint += "/stop"
	case "google-gemini", "gemini", "google-vertex", "vertex", "vertex-ai", "vertexai":
		method = http.MethodPost
		endpoint += ":cancel"
	}
	return a.call(ctx, requestID, tenantID, dep, method, endpoint, payload)
}

func (a *tunnelVideoAdapter) submitPayload(dep platform.Deployment, req platform.VideoCreateRequest, inputURL string) (string, interface{}, bool, error) {
	count := req.Count
	if count <= 0 {
		count = 1
	}
	switch a.provider {
	case "xai":
		if count != 1 {
			return "", nil, false, errors.New("xAI video generation supports n=1 per provider job")
		}
		body := map[string]interface{}{"model": dep.Model, "prompt": req.Prompt, "duration": req.Seconds}
		putOptional(body, "resolution", req.Resolution)
		putOptional(body, "aspect_ratio", req.AspectRatio)
		if strings.TrimSpace(inputURL) != "" {
			body["image"] = map[string]interface{}{"url": inputURL}
		}
		copyProviderOptions(body, req.ProviderOptions)
		return "/videos/generations", body, true, nil
	case "deepinfra":
		body := map[string]interface{}{"prompt": req.Prompt, "duration": req.Seconds, "num_outputs": count}
		putOptional(body, "resolution", req.Resolution)
		putOptional(body, "aspect_ratio", req.AspectRatio)
		putOptional(body, "image_url", inputURL)
		copyProviderOptions(body, req.ProviderOptions)
		return "/v1/inference/" + url.PathEscape(dep.Model), body, false, nil
	case "google-gemini", "gemini":
		instance := map[string]interface{}{"prompt": req.Prompt}
		putOptional(instance, "image", inputURL)
		parameters := map[string]interface{}{"durationSeconds": req.Seconds, "sampleCount": count}
		putOptional(parameters, "resolution", req.Resolution)
		putOptional(parameters, "aspectRatio", req.AspectRatio)
		putOptional(parameters, "storageUri", firstNestedString(dep.ModelInfo, "video_output_uri"))
		copyProviderOptions(parameters, req.ProviderOptions)
		return "/models/" + url.PathEscape(dep.Model) + ":predictLongRunning",
			map[string]interface{}{"instances": []interface{}{instance}, "parameters": parameters}, true, nil
	case "google-vertex", "vertex", "vertex-ai", "vertexai":
		if strings.TrimSpace(dep.ProjectID) == "" || strings.TrimSpace(dep.RegionName) == "" {
			return "", nil, false, errors.New("vertex video requires project_id and region")
		}
		instance := map[string]interface{}{"prompt": req.Prompt}
		putOptional(instance, "image", inputURL)
		parameters := map[string]interface{}{"durationSeconds": req.Seconds, "sampleCount": count}
		putOptional(parameters, "resolution", req.Resolution)
		putOptional(parameters, "aspectRatio", req.AspectRatio)
		copyProviderOptions(parameters, req.ProviderOptions)
		endpoint := fmt.Sprintf("/v1/projects/%s/locations/%s/publishers/google/models/%s:predictLongRunning",
			url.PathEscape(dep.ProjectID), url.PathEscape(dep.RegionName), url.PathEscape(dep.Model))
		return endpoint, map[string]interface{}{"instances": []interface{}{instance}, "parameters": parameters}, true, nil
	case "aws-bedrock", "bedrock":
		if count != 1 {
			return "", nil, false, errors.New("bedrock video generation supports n=1 per provider job")
		}
		if strings.TrimSpace(inputURL) != "" {
			return "", nil, false, errors.New("bedrock image-to-video requires a provider-accessible S3 input and is not enabled for managed HTTP media")
		}
		outputS3URI := firstNestedString(dep.ModelInfo, "video_output_s3_uri")
		if outputS3URI == "" {
			return "", nil, false, errors.New("bedrock video requires video_output_s3_uri in the provider account configuration")
		}
		body := map[string]interface{}{
			"modelId": dep.Model,
			"modelInput": map[string]interface{}{
				"taskType": "TEXT_VIDEO", "textToVideoParams": map[string]interface{}{"text": req.Prompt},
				"videoGenerationConfig": map[string]interface{}{"durationSeconds": req.Seconds, "fps": 24, "dimension": "1280x720"},
			},
			"outputDataConfig": map[string]interface{}{
				"s3OutputDataConfig": map[string]interface{}{"s3Uri": outputS3URI},
			},
		}
		return "/model/" + url.PathEscape(dep.Model) + "/async-invoke", body, true, nil
	case "togetherai", "together-ai":
		body := map[string]interface{}{"model": dep.Model, "prompt": req.Prompt, "seconds": strconv.Itoa(req.Seconds), "n": count}
		putOptional(body, "size", req.Size)
		putOptional(body, "resolution", req.Resolution)
		putOptional(body, "aspect_ratio", req.AspectRatio)
		putOptional(body, "input_reference", inputURL)
		if req.Seed != nil {
			body["seed"] = *req.Seed
		}
		copyProviderOptions(body, req.ProviderOptions)
		return "/videos", body, true, nil
	default:
		if count != 1 {
			return "", nil, false, errors.New("this provider supports n=1 per video job")
		}
		body := map[string]interface{}{"model": dep.Model, "prompt": req.Prompt, "seconds": strconv.Itoa(req.Seconds)}
		putOptional(body, "size", req.Size)
		putOptional(body, "resolution", req.Resolution)
		putOptional(body, "aspect_ratio", req.AspectRatio)
		putOptional(body, "input_reference", inputURL)
		if req.Seed != nil {
			body["seed"] = *req.Seed
		}
		copyProviderOptions(body, req.ProviderOptions)
		return "/videos", body, true, nil
	}
}

func (a *tunnelVideoAdapter) statusPath(dep platform.Deployment, providerJobID string) string {
	escaped := url.PathEscape(strings.TrimSpace(providerJobID))
	switch a.provider {
	case "deepinfra":
		return "/v1/inference/status/" + escaped
	case "google-gemini", "gemini":
		if strings.HasPrefix(strings.TrimSpace(providerJobID), "operations/") {
			return "/" + strings.TrimLeft(strings.TrimSpace(providerJobID), "/")
		}
		return "/operations/" + escaped
	case "google-vertex", "vertex", "vertex-ai", "vertexai":
		return "/v1/" + strings.TrimPrefix(providerJobID, "/")
	case "aws-bedrock", "bedrock":
		return "/async-invoke/" + escaped
	default:
		return "/videos/" + escaped
	}
}

func (a *tunnelVideoAdapter) call(ctx context.Context, requestID, tenantID string, dep platform.Deployment, method, endpoint string, payload interface{}) (ProviderResult, error) {
	body := []byte(nil)
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return ProviderResult{}, err
		}
	}
	response, err := a.tunnel.Proxy(ctx, platform.TunnelRequest{
		RequestID: requestID, TraceID: requestID, TenantID: tenantID, Deployment: dep,
		Headers: map[string]string{"Content-Type": "application/json"}, Body: body,
		RequestMetadata: map[string]interface{}{
			"provider_operation":     "video",
			"provider_endpoint_path": endpoint,
			"provider_http_method":   method,
		},
	})
	if err != nil {
		return ProviderResult{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ProviderResult{}, &ProviderRejectedError{
			StatusCode: response.StatusCode,
			Message:    safeProviderError(response.Body),
		}
	}
	return parseProviderResult(response.Body), nil
}

func parseProviderResult(body []byte) ProviderResult {
	raw := map[string]interface{}{}
	_ = json.Unmarshal(body, &raw)
	result := ProviderResult{Raw: raw, PollAfter: 3 * time.Second}
	result.ProviderJobID = firstNestedString(raw, "id", "name", "job_id", "request_id", "invocationArn", "operation.name")
	result.Status = normalizeProviderStatus(firstNestedString(raw, "status", "state", "phase", "operation.status"))
	result.Artifacts = extractArtifacts(raw)
	result.ActualUsage = extractVideoUsage(raw, result.Artifacts)
	if len(result.Artifacts) > 0 && (result.Status == "" || result.Status == platform.VideoGenerationQueued) {
		result.Status = platform.VideoGenerationSucceeded
	}
	if result.Status == "" {
		result.Status = platform.VideoGenerationQueued
	}
	return result
}

// ParseWebhookResult normalizes a verified provider webhook into the same
// result shape consumed by the polling worker.
func ParseWebhookResult(body []byte) ProviderResult {
	return parseProviderResult(body)
}

func extractArtifacts(raw map[string]interface{}) []platform.VideoArtifact {
	urls := make([]string, 0)
	collectURLs(raw, "", &urls)
	seen := map[string]bool{}
	artifacts := make([]platform.VideoArtifact, 0, len(urls))
	for _, outputURL := range urls {
		if seen[outputURL] {
			continue
		}
		if parsed, err := url.Parse(outputURL); err == nil && parsed.Scheme == "s3" && pathExtension(parsed.Path) == "" {
			outputURL = strings.TrimRight(outputURL, "/") + "/output.mp4"
		}
		seen[outputURL] = true
		artifacts = append(artifacts, platform.VideoArtifact{
			ID: strconv.Itoa(len(artifacts) + 1), ProviderURL: outputURL, ContentType: "video/mp4",
		})
	}
	return artifacts
}

func collectURLs(value interface{}, parentKey string, out *[]string) {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, child := range typed {
			if stringValue, ok := child.(string); ok && isVideoURLKey(key, parentKey) && isAbsoluteHTTPURL(stringValue) {
				*out = append(*out, stringValue)
			}
			collectURLs(child, key, out)
		}
	case []interface{}:
		for _, child := range typed {
			collectURLs(child, parentKey, out)
		}
	}
}

func extractVideoUsage(raw map[string]interface{}, artifacts []platform.VideoArtifact) []platform.UsageMeasure {
	totalSeconds := firstNestedNumber(raw, "usage.output_seconds", "usage.seconds")
	duration := firstNestedNumber(raw, "duration", "duration_seconds", "video.duration")
	resolution := firstNestedString(raw, "usage.resolution", "resolution", "video.resolution")
	count := len(artifacts)
	if reported := int(firstNestedNumber(raw, "usage.output_count", "n", "count")); reported > 0 {
		count = reported
	}
	if count <= 0 {
		count = 1
	}
	if totalSeconds <= 0 && duration > 0 {
		totalSeconds = duration * float64(count)
	}
	if totalSeconds <= 0 {
		return nil
	}
	return []platform.UsageMeasure{{
		Usage: "output", Unit: "second",
		Amount:     strconv.FormatFloat(totalSeconds, 'f', -1, 64),
		Resolution: resolution, ServiceTier: "standard",
		Qualifiers: map[string]interface{}{"output_count": count},
	}}
}

func normalizeProviderStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "queued", "pending", "submitted", "starting":
		return platform.VideoGenerationQueued
	case "running", "processing", "in_progress", "inprogress":
		return platform.VideoGenerationRunning
	case "succeeded", "completed", "complete", "done", "success":
		return platform.VideoGenerationSucceeded
	case "failed", "error", "expired":
		return platform.VideoGenerationFailed
	case "cancelled", "canceled":
		return platform.VideoGenerationCancelled
	default:
		return ""
	}
}

func firstNestedString(raw map[string]interface{}, paths ...string) string {
	for _, candidate := range paths {
		if value := nestedValue(raw, candidate); value != nil {
			if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

func firstNestedNumber(raw map[string]interface{}, paths ...string) float64 {
	for _, candidate := range paths {
		switch value := nestedValue(raw, candidate).(type) {
		case float64:
			return value
		case json.Number:
			number, _ := value.Float64()
			return number
		case string:
			number, _ := strconv.ParseFloat(value, 64)
			if number != 0 {
				return number
			}
		}
	}
	return 0
}

func nestedValue(raw map[string]interface{}, path string) interface{} {
	var current interface{} = raw
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]interface{})
		if !ok {
			return nil
		}
		current = object[part]
	}
	return current
}

func putOptional(target map[string]interface{}, key, value string) {
	if strings.TrimSpace(value) != "" {
		target[key] = strings.TrimSpace(value)
	}
}

func copyProviderOptions(target map[string]interface{}, options map[string]interface{}) {
	allowed := map[string]bool{"negative_prompt": true, "style": true, "fps": true, "guidance_scale": true, "motion_bucket_id": true}
	for key, value := range options {
		if allowed[key] {
			target[key] = value
		}
	}
}

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

func isVideoURLKey(key, parentKey string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "url" || normalized == "uri" {
		switch strings.ToLower(strings.TrimSpace(parentKey)) {
		case "video", "videos", "output", "outputs", "data", "result", "generatedsamples", "generated_videos", "fullvideo":
			return true
		default:
			return false
		}
	}
	if normalized == "s3uri" || normalized == "s3_uri" {
		return strings.EqualFold(strings.TrimSpace(parentKey), "s3OutputDataConfig")
	}
	return normalized == "video_url" || normalized == "output_url" || normalized == "output_uri"
}

func isAbsoluteHTTPURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http" || parsed.Scheme == "s3")
}

func pathExtension(value string) string {
	lastSlash := strings.LastIndex(value, "/")
	lastDot := strings.LastIndex(value, ".")
	if lastDot <= lastSlash {
		return ""
	}
	return value[lastDot:]
}

func safeProviderError(body []byte) string {
	var raw map[string]interface{}
	if json.Unmarshal(body, &raw) == nil {
		if message := firstNestedString(raw, "error.message", "message", "detail"); message != "" {
			return message
		}
	}
	return http.StatusText(http.StatusBadGateway)
}
