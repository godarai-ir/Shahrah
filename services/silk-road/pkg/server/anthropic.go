package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const anthropicDefaultVersion = "2023-06-01"

type anthropicAdapter struct {
	wif    *anthropicWIFManager
	wifErr error
}

type anthropicRequest struct {
	Model        string                          `json:"model"`
	Messages     []anthropicMessage              `json:"messages"`
	MaxTokens    int                             `json:"max_tokens,omitempty"`
	Temperature  *float64                        `json:"temperature,omitempty"`
	TopP         *float64                        `json:"top_p,omitempty"`
	TopK         *int                            `json:"top_k,omitempty"`
	Stop         []string                        `json:"stop_sequences,omitempty"`
	Stream       bool                            `json:"stream,omitempty"`
	System       []anthropicContentBlock         `json:"system,omitempty"`
	Thinking     *contractsopenai.ThinkingConfig `json:"thinking,omitempty"`
	Tools        []anthropicTool                 `json:"tools,omitempty"`
	ToolChoice   *anthropicToolChoice            `json:"tool_choice,omitempty"`
	Metadata     *anthropicMetadata              `json:"metadata,omitempty"`
	Output       *anthropicOutputConfig          `json:"output_config,omitempty"`
	ServiceTier  string                          `json:"service_tier,omitempty"`
	Speed        string                          `json:"speed,omitempty"`
	InferenceGeo string                          `json:"inference_geo,omitempty"`
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

type anthropicContentBlock struct {
	Type         string                 `json:"type"`
	Text         string                 `json:"text,omitempty"`
	Thinking     string                 `json:"thinking,omitempty"`
	Signature    string                 `json:"signature,omitempty"`
	Data         string                 `json:"data,omitempty"`
	Source       *anthropicSource       `json:"source,omitempty"`
	ID           string                 `json:"id,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Input        interface{}            `json:"input,omitempty"`
	ToolUseID    string                 `json:"tool_use_id,omitempty"`
	Content      interface{}            `json:"content,omitempty"`
	IsError      bool                   `json:"is_error,omitempty"`
	CacheControl map[string]interface{} `json:"cache_control,omitempty"`
	Citations    []interface{}          `json:"citations,omitempty"`
	ProviderData map[string]interface{} `json:"-"`
}

type anthropicSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
	FileID    string `json:"file_id,omitempty"`
}

type anthropicTool struct {
	Type           string                 `json:"type,omitempty"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	InputSchema    map[string]interface{} `json:"input_schema,omitempty"`
	AllowedCallers []string               `json:"allowed_callers,omitempty"`
	AllowedDomains []string               `json:"allowed_domains,omitempty"`
	BlockedDomains []string               `json:"blocked_domains,omitempty"`
	MaxUses        *int                   `json:"max_uses,omitempty"`
	UserLocation   map[string]interface{} `json:"user_location,omitempty"`
	CacheControl   map[string]interface{} `json:"cache_control,omitempty"`
	Strict         *bool                  `json:"strict,omitempty"`
	DeferLoading   *bool                  `json:"defer_loading,omitempty"`
}

type anthropicToolChoice struct {
	Type                   string `json:"type"`
	Name                   string `json:"name,omitempty"`
	DisableParallelToolUse bool   `json:"disable_parallel_tool_use,omitempty"`
}

type anthropicMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

type anthropicOutputConfig struct {
	Effort string      `json:"effort,omitempty"`
	Format interface{} `json:"format,omitempty"`
}

type anthropicResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Content    []anthropicContentBlock `json:"content"`
	Model      string                  `json:"model"`
	Usage      anthropicUsage          `json:"usage"`
	StopReason string                  `json:"stop_reason,omitempty"`
}

type anthropicUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	CacheCreation            struct {
		Ephemeral5mInputTokens int `json:"ephemeral_5m_input_tokens,omitempty"`
		Ephemeral1hInputTokens int `json:"ephemeral_1h_input_tokens,omitempty"`
	} `json:"cache_creation,omitempty"`
	ServerToolUse map[string]int `json:"server_tool_use,omitempty"`
	ServiceTier   string         `json:"service_tier,omitempty"`
	Speed         string         `json:"speed,omitempty"`
}

type anthropicStreamEvent struct {
	Type  string `json:"type"`
	Index int    `json:"index,omitempty"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
	Message struct {
		ID    string         `json:"id"`
		Model string         `json:"model"`
		Usage anthropicUsage `json:"usage"`
	} `json:"message,omitempty"`
	ContentBlock anthropicContentBlock `json:"content_block,omitempty"`
	Usage        anthropicUsage        `json:"usage,omitempty"`
	Delta        anthropicStreamDelta  `json:"delta,omitempty"`
}

type anthropicStreamDelta struct {
	Type         string                 `json:"type,omitempty"`
	Text         string                 `json:"text,omitempty"`
	Thinking     string                 `json:"thinking,omitempty"`
	Signature    string                 `json:"signature,omitempty"`
	PartialJSON  string                 `json:"partial_json,omitempty"`
	StopReason   string                 `json:"stop_reason,omitempty"`
	ProviderData map[string]interface{} `json:"-"`
}

func (block *anthropicContentBlock) UnmarshalJSON(data []byte) error {
	type alias anthropicContentBlock
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*block = anthropicContentBlock(decoded)
	block.ProviderData = raw
	return nil
}

func (delta *anthropicStreamDelta) UnmarshalJSON(data []byte) error {
	type alias anthropicStreamDelta
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*delta = anthropicStreamDelta(decoded)
	delta.ProviderData = raw
	return nil
}

func newAnthropicAdapter() providerAdapter {
	manager, err := newAnthropicWIFManagerFromEnv()
	return &anthropicAdapter{wif: manager, wifErr: err}
}

func (a *anthropicAdapter) Dispatch(ctx context.Context, client *http.Client, payload platform.TunnelRequest) (*adapterResponse, error) {
	target, err := resolveBaseURL(payload.Deployment.APIBase, "https://api.anthropic.com")
	if err != nil {
		return nil, err
	}
	if payload.Operation != "" && payload.Operation != platform.ProviderOperationChatCompletions {
		return a.dispatchNative(ctx, client, target, payload)
	}

	var openAIReq contractsopenai.ChatCompletionRequest
	if err := json.Unmarshal(payload.Body, &openAIReq); err != nil {
		return nil, fmt.Errorf("decode openai request: %w", err)
	}

	runtimeProfile, err := anthropicRuntimeProfileForModel(
		payload,
		openAIReq.Model,
		"messages",
	)
	if err != nil {
		return nil, err
	}
	if err := validateAnthropicOpenAIRequestWithProfile(
		openAIReq,
		runtimeProfile,
	); err != nil {
		return nil, err
	}
	if len(payload.Body) > runtimeProfile.Profile.Limits.MessagesBytes {
		return nil, fmt.Errorf(
			"anthropic operation payload exceeds %d bytes",
			runtimeProfile.Profile.Limits.MessagesBytes,
		)
	}
	anthropicReq, err := convertToAnthropicRequest(
		openAIReq,
		runtimeProfile.Profile,
	)
	if err != nil {
		return nil, fmt.Errorf("convert anthropic request: %w", err)
	}
	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("encode anthropic request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(target, "/v1/messages"), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := a.applyAuth(ctx, client, target, payload.Deployment, req); err != nil {
		return nil, err
	}
	req.Header.Set("anthropic-version", anthropicVersion(payload.Deployment.APIVersion))
	applyAnthropicBetaHeaders(req, payload)
	if payload.Stream {
		req.Header.Set("Accept", "text/event-stream")
	}

	resp, err := doAnthropicRequest(ctx, client, req, false)
	if err != nil {
		return nil, err
	}
	a.addWIFStatusHeaders(payload.Deployment, resp.Header)

	if resp.StatusCode >= http.StatusBadRequest {
		return &adapterResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header.Clone(),
			Body:       resp.Body,
		}, nil
	}

	if payload.Stream {
		return adaptAnthropicStreamResponse(resp, openAIReq.Model), nil
	}
	return adaptAnthropicSyncResponse(resp, openAIReq.Model)
}

func (a *anthropicAdapter) applyAuth(ctx context.Context, client *http.Client, target string, deployment platform.Deployment, req *http.Request) error {
	if deployment.ProviderAuth != nil &&
		deployment.ProviderAuth.Mode == string(platform.ProviderAuthModeAnthropicWIF) {
		if a.wifErr != nil {
			return errAnthropicWIFUnavailable
		}
		token, tokenErr := a.wif.token(
			ctx,
			client,
			target,
			anthropicWIFConnectionKey(deployment),
			deployment.ProviderAuth.AnthropicWIF,
		)
		if tokenErr != nil {
			return tokenErr
		}
		req.Header.Set("Authorization", "Bearer "+token.accessToken)
	} else {
		req.Header.Set("x-api-key", deploymentAPIKey(deployment))
	}
	return nil
}

func anthropicWIFConnectionKey(deployment platform.Deployment) string {
	if strings.TrimSpace(deployment.ConnectionID) != "" {
		return strings.TrimSpace(deployment.ConnectionID)
	}
	return strings.TrimSpace(deployment.ProviderAccountID)
}

func (a *anthropicAdapter) dispatchNative(ctx context.Context, client *http.Client, target string, payload platform.TunnelRequest) (*adapterResponse, error) {
	var err error
	payload, err = prepareAnthropicCapabilityBetaHeaders(payload)
	if err != nil {
		return nil, err
	}
	if err := validateAnthropicNativeRuntimeProfiles(payload); err != nil {
		return nil, err
	}
	method, path, maxBytes, err := anthropicOperationTarget(payload.Operation, payload.OperationInput)
	if err != nil {
		return nil, err
	}
	maxBytes = anthropicRuntimeProfilePayloadLimit(payload, maxBytes)
	if maxBytes > 0 && int64(len(payload.Body)) > maxBytes {
		return nil, fmt.Errorf("anthropic operation payload exceeds %d bytes", maxBytes)
	}
	var body io.Reader = http.NoBody
	if len(payload.Body) > 0 {
		body = bytes.NewReader(payload.Body)
	}
	req, err := http.NewRequestWithContext(ctx, method, joinURL(target, path), body)
	if err != nil {
		return nil, err
	}
	if contentType := strings.TrimSpace(headerValue(payload.Headers, "content-type")); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else if len(payload.Body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("anthropic-version", anthropicVersion(payload.Deployment.APIVersion))
	applyAnthropicBetaHeaders(req, payload)
	if payload.Stream {
		req.Header.Set("Accept", "text/event-stream")
	}
	if err := a.applyAuth(ctx, client, target, payload.Deployment, req); err != nil {
		return nil, err
	}
	resp, err := doAnthropicRequest(
		ctx,
		client,
		req,
		anthropicOperationIsIdempotent(payload.Operation),
	)
	if err != nil {
		return nil, err
	}
	a.addWIFStatusHeaders(payload.Deployment, resp.Header)
	return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: resp.Body}, nil
}

func prepareAnthropicCapabilityBetaHeaders(
	payload platform.TunnelRequest,
) (platform.TunnelRequest, error) {
	if payload.Operation != platform.ProviderOperationAnthropicBatchesCreate {
		return payload, nil
	}
	var batch struct {
		Requests []struct {
			Params struct {
				Model     string `json:"model"`
				MaxTokens int    `json:"max_tokens"`
			} `json:"params"`
		} `json:"requests"`
	}
	if err := json.Unmarshal(payload.Body, &batch); err != nil {
		return payload, fmt.Errorf("invalid Anthropic batch request")
	}
	requested := commaSeparatedSet(
		headerValue(payload.Headers, "anthropic-beta"),
	)
	accountAllowed := commaSeparatedSet(
		payload.Deployment.ProviderHeaders["Anthropic-Beta-Allowlist"] + "," +
			payload.Deployment.ProviderHeaders["anthropic-beta-allowlist"],
	)
	for _, item := range batch.Requests {
		binding, ok := platform.FindModelRuntimeProfile(
			payload.ModelRuntimeProfiles,
			item.Params.Model,
		)
		if !ok {
			return payload, fmt.Errorf(
				"Anthropic model runtime profile is missing for %s",
				item.Params.Model,
			)
		}
		required := binding.Profile.RequiredBetaHeaderForOutput(
			"message_batches",
			item.Params.MaxTokens,
		)
		if required == "" {
			continue
		}
		if _, ok := accountAllowed[required]; !ok {
			return payload, fmt.Errorf(
				"Anthropic beta capability %s is not enabled for provider account",
				required,
			)
		}
		requested[required] = struct{}{}
	}
	if len(requested) == 0 {
		return payload, nil
	}
	values := make([]string, 0, len(requested))
	for value := range requested {
		values = append(values, value)
	}
	sort.Strings(values)
	if payload.Headers == nil {
		payload.Headers = make(map[string]string)
	}
	payload.Headers["anthropic-beta"] = strings.Join(values, ",")
	return payload, nil
}

func (a *anthropicAdapter) DispatchRaw(ctx context.Context, client *http.Client, payload platform.TunnelRequest, body io.Reader, contentLength int64) (*adapterResponse, error) {
	switch payload.Operation {
	case platform.ProviderOperationAnthropicFilesCreate,
		platform.ProviderOperationAnthropicFilesContent,
		platform.ProviderOperationAnthropicBatchesResults:
	default:
		return nil, fmt.Errorf("operation %q does not support raw tunneling", payload.Operation)
	}
	target, err := resolveBaseURL(payload.Deployment.APIBase, "https://api.anthropic.com")
	if err != nil {
		return nil, err
	}
	method, path, maxBytes, err := anthropicOperationTarget(payload.Operation, payload.OperationInput)
	if err != nil {
		return nil, err
	}
	if contentLength > maxBytes && maxBytes > 0 {
		return nil, fmt.Errorf("anthropic operation payload exceeds %d bytes", maxBytes)
	}
	if body == nil {
		body = http.NoBody
	}
	if maxBytes > 0 && contentLength < 0 {
		body = &hardLimitReader{reader: body, remaining: maxBytes}
	}
	req, err := http.NewRequestWithContext(ctx, method, joinURL(target, path), body)
	if err != nil {
		return nil, err
	}
	if contentLength >= 0 {
		req.ContentLength = contentLength
	}
	if contentType := strings.TrimSpace(headerValue(payload.Headers, "content-type")); contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("anthropic-version", anthropicVersion(payload.Deployment.APIVersion))
	applyAnthropicBetaHeaders(req, payload)
	if err := a.applyAuth(ctx, client, target, payload.Deployment, req); err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	a.addWIFStatusHeaders(payload.Deployment, resp.Header)
	return &adapterResponse{StatusCode: resp.StatusCode, Headers: resp.Header.Clone(), Body: resp.Body}, nil
}

func (a *anthropicAdapter) addWIFStatusHeaders(
	deployment platform.Deployment,
	headers http.Header,
) {
	if headers == nil || a == nil || a.wif == nil ||
		deployment.ProviderAuth == nil ||
		deployment.ProviderAuth.AnthropicWIF == nil {
		return
	}
	token, ok := a.wif.status(
		anthropicWIFConnectionKey(deployment),
		deployment.ProviderAuth.AnthropicWIF,
	)
	if !ok {
		return
	}
	headers.Set(
		"X-GodarAI-Anthropic-WIF-Expires-At",
		token.expiresAt.UTC().Format(time.RFC3339),
	)
	headers.Set(
		"X-GodarAI-Anthropic-WIF-Refreshed-At",
		token.refreshedAt.UTC().Format(time.RFC3339),
	)
	if scopes := sanitizedAnthropicWIFScopes(token.scopes); len(scopes) > 0 {
		headers.Set(
			"X-GodarAI-Anthropic-WIF-Scopes",
			strings.Join(scopes, " "),
		)
	}
}

func sanitizedAnthropicWIFScopes(scopes []string) []string {
	result := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" || strings.ContainsAny(scope, "\r\n,") {
			continue
		}
		result = append(result, scope)
	}
	sort.Strings(result)
	return result
}

type hardLimitReader struct {
	reader    io.Reader
	remaining int64
	exhausted bool
}

func (r *hardLimitReader) Read(buffer []byte) (int, error) {
	if r.exhausted {
		return 0, fmt.Errorf("provider payload exceeds configured limit")
	}
	if int64(len(buffer)) > r.remaining+1 {
		buffer = buffer[:r.remaining+1]
	}
	n, err := r.reader.Read(buffer)
	if int64(n) > r.remaining {
		r.exhausted = true
		return 0, fmt.Errorf("provider payload exceeds configured limit")
	}
	r.remaining -= int64(n)
	return n, err
}

func anthropicOperationTarget(operation platform.ProviderOperation, input platform.ProviderOperationInput) (string, string, int64, error) {
	const (
		messageLimit = int64(32 * 1024 * 1024)
		batchLimit   = int64(256 * 1024 * 1024)
		fileLimit    = int64(500 * 1024 * 1024)
	)
	id := strings.TrimSpace(input.ResourceID)
	resourcePath := func(prefix, suffix string) (string, error) {
		if id == "" {
			return "", fmt.Errorf("provider operation requires resource_id")
		}
		return prefix + url.PathEscape(id) + suffix, nil
	}
	listPath := func(base string) string {
		query := url.Values{}
		if input.BeforeID != "" {
			query.Set("before_id", input.BeforeID)
		}
		if input.AfterID != "" {
			query.Set("after_id", input.AfterID)
		}
		if input.Limit > 0 {
			query.Set("limit", fmt.Sprintf("%d", input.Limit))
		}
		if len(query) == 0 {
			return base
		}
		return base + "?" + query.Encode()
	}
	switch operation {
	case platform.ProviderOperationAnthropicMessages:
		return http.MethodPost, "/v1/messages", messageLimit, nil
	case platform.ProviderOperationAnthropicCountTokens:
		return http.MethodPost, "/v1/messages/count_tokens", messageLimit, nil
	case platform.ProviderOperationAnthropicModelsList:
		return http.MethodGet, listPath("/v1/models"), 0, nil
	case platform.ProviderOperationAnthropicModelsGet:
		path, err := resourcePath("/v1/models/", "")
		return http.MethodGet, path, 0, err
	case platform.ProviderOperationAnthropicFilesCreate:
		return http.MethodPost, "/v1/files", fileLimit, nil
	case platform.ProviderOperationAnthropicFilesList:
		return http.MethodGet, listPath("/v1/files"), 0, nil
	case platform.ProviderOperationAnthropicFilesGet:
		path, err := resourcePath("/v1/files/", "")
		return http.MethodGet, path, 0, err
	case platform.ProviderOperationAnthropicFilesDelete:
		path, err := resourcePath("/v1/files/", "")
		return http.MethodDelete, path, 0, err
	case platform.ProviderOperationAnthropicFilesContent:
		path, err := resourcePath("/v1/files/", "/content")
		return http.MethodGet, path, 0, err
	case platform.ProviderOperationAnthropicBatchesCreate:
		return http.MethodPost, "/v1/messages/batches", batchLimit, nil
	case platform.ProviderOperationAnthropicBatchesList:
		return http.MethodGet, listPath("/v1/messages/batches"), 0, nil
	case platform.ProviderOperationAnthropicBatchesGet:
		path, err := resourcePath("/v1/messages/batches/", "")
		return http.MethodGet, path, 0, err
	case platform.ProviderOperationAnthropicBatchesCancel:
		path, err := resourcePath("/v1/messages/batches/", "/cancel")
		return http.MethodPost, path, 0, err
	case platform.ProviderOperationAnthropicBatchesResults:
		path, err := resourcePath("/v1/messages/batches/", "/results")
		return http.MethodGet, path, 0, err
	default:
		return "", "", 0, fmt.Errorf("unsupported typed provider operation %q", operation)
	}
}

func headerValue(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(strings.TrimSpace(key), name) {
			return value
		}
	}
	return ""
}

func applyAnthropicBetaHeaders(req *http.Request, payload platform.TunnelRequest) {
	if req == nil {
		return
	}
	allowed := commaSeparatedSet(
		payload.Deployment.ProviderHeaders["Anthropic-Beta-Allowlist"] + "," +
			payload.Deployment.ProviderHeaders["anthropic-beta-allowlist"],
	)
	requested := commaSeparatedSet(headerValue(payload.Headers, "anthropic-beta"))
	if operationNeedsFilesBeta(payload.Operation) {
		requested["files-api-2025-04-14"] = struct{}{}
	}
	enabled := make([]string, 0, len(requested))
	for value := range requested {
		if _, ok := allowed[value]; ok {
			if !anthropicRuntimeProfilesAllowBeta(
				payload.ModelRuntimeProfiles,
				value,
			) {
				continue
			}
			enabled = append(enabled, value)
		}
	}
	if len(enabled) > 0 {
		sort.Strings(enabled)
		req.Header.Set("anthropic-beta", strings.Join(enabled, ","))
	}
}

func anthropicRuntimeProfilesAllowBeta(
	profiles []platform.ModelRuntimeProfileBinding,
	header string,
) bool {
	if len(profiles) == 0 {
		return true
	}
	for _, binding := range profiles {
		if !binding.Profile.AllowsBetaHeader(header) {
			return false
		}
	}
	return true
}

func commaSeparatedSet(input string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range strings.Split(input, ",") {
		value = strings.TrimSpace(value)
		if value != "" {
			out[value] = struct{}{}
		}
	}
	return out
}

func operationNeedsFilesBeta(operation platform.ProviderOperation) bool {
	switch operation {
	case platform.ProviderOperationAnthropicFilesCreate,
		platform.ProviderOperationAnthropicFilesList,
		platform.ProviderOperationAnthropicFilesGet,
		platform.ProviderOperationAnthropicFilesDelete,
		platform.ProviderOperationAnthropicFilesContent:
		return true
	default:
		return false
	}
}

func adaptAnthropicSyncResponse(resp *http.Response, model string) (*adapterResponse, error) {
	defer resp.Body.Close()

	var anthResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthResp); err != nil {
		return nil, fmt.Errorf("decode anthropic response: %w", err)
	}
	openAIResp := convertToOpenAIResponse(anthResp, model)
	body, err := json.Marshal(openAIResp)
	if err != nil {
		return nil, fmt.Errorf("encode openai response: %w", err)
	}

	headers := resp.Header.Clone()
	headers.Set("Content-Type", "application/json")
	headers.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	return &adapterResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

func adaptAnthropicStreamResponse(resp *http.Response, model string) *adapterResponse {
	headers := resp.Header.Clone()
	headers.Set("Content-Type", "text/event-stream")
	headers.Del("Content-Length")

	pipeReader, pipeWriter := io.Pipe()
	go func() {
		defer resp.Body.Close()
		defer pipeWriter.Close()

		reader := bufio.NewReader(resp.Body)
		var currentEvent string
		streamID := ""
		streamModel := model
		completed := false
		var latestUsage anthropicUsage

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					if completed {
						_, _ = pipeWriter.Write([]byte("data: [DONE]\n\n"))
					} else {
						anthropicStreamErrors.Add(1)
						_ = pipeWriter.CloseWithError(io.ErrUnexpectedEOF)
					}
					return
				}
				_ = pipeWriter.CloseWithError(err)
				return
			}

			trimmed := strings.TrimSpace(line)
			switch {
			case trimmed == "":
				currentEvent = ""
			case strings.HasPrefix(trimmed, "event:"):
				currentEvent = strings.TrimSpace(strings.TrimPrefix(trimmed, "event:"))
			case strings.HasPrefix(trimmed, "data:"):
				data := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				if data == "" || data == "[DONE]" {
					continue
				}
				var evt anthropicStreamEvent
				if err := json.Unmarshal([]byte(data), &evt); err != nil {
					_ = pipeWriter.CloseWithError(fmt.Errorf("decode anthropic SSE event: %w", err))
					return
				}
				if evt.Message.ID != "" {
					streamID = evt.Message.ID
				}
				if evt.Message.Model != "" {
					streamModel = evt.Message.Model
				}
				switch eventType(currentEvent, evt.Type) {
				case "message_start":
					latestUsage = evt.Message.Usage
					chunk := contractsopenai.StreamChunk{
						ID: streamID, Object: "chat.completion.chunk", Created: time.Now().Unix(), Model: streamModel,
						Choices: []contractsopenai.Choice{{Index: 0, Delta: &contractsopenai.ChatMessage{Role: "assistant"}}},
					}
					if err := writeSSEChunk(pipeWriter, chunk); err != nil {
						return
					}
				case "content_block_start":
					if evt.ContentBlock.Type != "tool_use" {
						if isAnthropicServerContentBlock(
							evt.ContentBlock.Type,
						) && len(evt.ContentBlock.ProviderData) > 0 {
							chunk := contractsopenai.StreamChunk{
								ID: streamID, Object: "chat.completion.chunk",
								Created: time.Now().Unix(), Model: streamModel,
								Choices: []contractsopenai.Choice{{
									Index: evt.Index,
									Delta: &contractsopenai.ChatMessage{
										ProviderContent: []map[string]interface{}{
											evt.ContentBlock.ProviderData,
										},
									},
								}},
							}
							if err := writeSSEChunk(pipeWriter, chunk); err != nil {
								return
							}
						}
						continue
					}
					index := evt.Index
					arguments, _ := json.Marshal(evt.ContentBlock.Input)
					if string(arguments) == "null" || string(arguments) == "{}" {
						arguments = nil
					}
					chunk := contractsopenai.StreamChunk{
						ID: streamID, Object: "chat.completion.chunk", Created: time.Now().Unix(), Model: streamModel,
						Choices: []contractsopenai.Choice{{
							Index: 0,
							Delta: &contractsopenai.ChatMessage{ToolCalls: []contractsopenai.ToolCall{{
								ID: evt.ContentBlock.ID, Type: "function", Index: &index,
								Function: &contractsopenai.FunctionCall{Name: evt.ContentBlock.Name, Arguments: string(arguments)},
							}}},
						}},
					}
					if err := writeSSEChunk(pipeWriter, chunk); err != nil {
						return
					}
				case "content_block_delta":
					if evt.Delta.Text == "" && evt.Delta.Thinking == "" &&
						evt.Delta.Signature == "" &&
						evt.Delta.PartialJSON == "" &&
						len(evt.Delta.ProviderData) == 0 {
						continue
					}
					delta := &contractsopenai.ChatMessage{}
					if evt.Delta.Thinking != "" {
						delta.ReasoningContent = evt.Delta.Thinking
					}
					if evt.Delta.Signature != "" {
						delta.ThinkingBlocks = []contractsopenai.ThinkingBlock{{
							"type":      "thinking",
							"signature": evt.Delta.Signature,
						}}
					}
					if evt.Delta.Text != "" {
						delta.Content = evt.Delta.Text
					}
					if evt.Delta.PartialJSON != "" {
						index := evt.Index
						delta.ToolCalls = []contractsopenai.ToolCall{{
							Type: "function", Index: &index,
							Function: &contractsopenai.FunctionCall{Arguments: evt.Delta.PartialJSON},
						}}
					}
					if len(evt.Delta.ProviderData) > 0 &&
						evt.Delta.Type != "text_delta" &&
						evt.Delta.Type != "thinking_delta" &&
						evt.Delta.Type != "signature_delta" &&
						evt.Delta.Type != "input_json_delta" {
						delta.ProviderContent = []map[string]interface{}{
							evt.Delta.ProviderData,
						}
					}
					chunk := contractsopenai.StreamChunk{
						ID:      streamID,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   streamModel,
						Choices: []contractsopenai.Choice{{
							Index: evt.Index,
							Delta: delta,
						}},
					}
					if err := writeSSEChunk(pipeWriter, chunk); err != nil {
						return
					}
				case "message_delta":
					if evt.Usage.OutputTokens > 0 {
						latestUsage.OutputTokens = evt.Usage.OutputTokens
					}
					if evt.Usage.CacheCreationInputTokens > 0 {
						latestUsage.CacheCreationInputTokens = evt.Usage.CacheCreationInputTokens
					}
					if evt.Usage.CacheReadInputTokens > 0 {
						latestUsage.CacheReadInputTokens = evt.Usage.CacheReadInputTokens
					}
					if evt.Usage.ServerToolUse != nil {
						latestUsage.ServerToolUse = evt.Usage.ServerToolUse
					}
					if evt.Delta.StopReason == "" {
						continue
					}
					chunk := contractsopenai.StreamChunk{
						ID:      streamID,
						Object:  "chat.completion.chunk",
						Created: time.Now().Unix(),
						Model:   streamModel,
						Choices: []contractsopenai.Choice{{
							Index:        0,
							Delta:        &contractsopenai.ChatMessage{},
							FinishReason: mapAnthropicStopReason(evt.Delta.StopReason),
						}},
					}
					if err := writeSSEChunk(pipeWriter, chunk); err != nil {
						return
					}
				case "message_stop":
					completed = true
					usage := openAIUsageFromAnthropic(latestUsage)
					if usage.TotalTokens > 0 {
						chunk := contractsopenai.StreamChunk{
							ID: streamID, Object: "chat.completion.chunk", Created: time.Now().Unix(), Model: streamModel,
							Choices: []contractsopenai.Choice{}, Usage: usage,
						}
						if err := writeSSEChunk(pipeWriter, chunk); err != nil {
							return
						}
					}
				case "error":
					anthropicStreamErrors.Add(1)
					message := "Anthropic stream error"
					errorType := "api_error"
					if evt.Error != nil {
						if strings.TrimSpace(evt.Error.Message) != "" {
							message = evt.Error.Message
						}
						if strings.TrimSpace(evt.Error.Type) != "" {
							errorType = evt.Error.Type
						}
					}
					encoded, _ := json.Marshal(contractsopenai.APIError{Error: contractsopenai.ErrorInfo{
						Message: message, Type: errorType, Code: "anthropic_stream_error",
					}})
					_, _ = pipeWriter.Write([]byte("data: " + string(encoded) + "\n\n"))
					return
				}
			}
		}
	}()

	return &adapterResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       pipeReader,
	}
}

func writeSSEChunk(w io.Writer, chunk contractsopenai.StreamChunk) error {
	encoded, err := json.Marshal(chunk)
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte("data: ")); err != nil {
		return err
	}
	if _, err := w.Write(encoded); err != nil {
		return err
	}
	_, err = w.Write([]byte("\n\n"))
	return err
}

func eventType(explicit string, fallback string) string {
	if strings.TrimSpace(explicit) != "" {
		return strings.TrimSpace(explicit)
	}
	return strings.TrimSpace(fallback)
}

func anthropicVersion(version string) string {
	if strings.TrimSpace(version) == "" {
		return anthropicDefaultVersion
	}
	return strings.TrimSpace(version)
}

func convertToAnthropicRequest(
	req contractsopenai.ChatCompletionRequest,
	runtimeProfile platform.ModelRuntimeProfile,
) (anthropicRequest, error) {
	if err := validateAnthropicRequestWithRuntimeProfile(
		req,
		runtimeProfile,
	); err != nil {
		return anthropicRequest{}, err
	}
	anthReq := anthropicRequest{
		Model:       canonicalAnthropicModel(req.Model),
		MaxTokens:   4096,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        req.TopK,
		Stream:      req.Stream,
	}
	var err error
	if anthReq.Stop, err = anthropicStopSequences(req.Stop); err != nil {
		return anthropicRequest{}, err
	}
	if anthReq.Tools, err = anthropicTools(req.Tools, req.CacheControl); err != nil {
		return anthropicRequest{}, err
	}
	if anthReq.ToolChoice, err = anthropicToolChoiceFromOpenAI(req.ToolChoice, req.ParallelToolCalls); err != nil {
		return anthropicRequest{}, err
	}
	anthReq.Metadata = anthropicMetadataFromOpenAI(req.Metadata, req.User)
	anthReq.Output, err = anthropicOutputConfigFromOpenAI(
		runtimeProfile,
		req.ResponseFormat,
		req.OutputConfig,
		req.ReasoningEffort,
	)
	if err != nil {
		return anthropicRequest{}, err
	}
	anthReq.ServiceTier = strings.TrimSpace(req.ServiceTier)
	anthReq.Speed = strings.TrimSpace(req.Speed)
	anthReq.InferenceGeo = strings.TrimSpace(req.InferenceGeo)
	if err := validateAnthropicRequestControls(
		runtimeProfile,
		anthReq.ServiceTier,
		anthReq.Speed,
		anthReq.InferenceGeo,
	); err != nil {
		return anthropicRequest{}, err
	}

	for _, msg := range req.Messages {
		blocks, err := anthropicBlocksFromContent(msg.Content)
		if err != nil {
			return anthropicRequest{}, err
		}
		if strings.EqualFold(strings.TrimSpace(msg.Role), "assistant") && len(msg.ThinkingBlocks) > 0 {
			thinkingBlocks := anthropicBlocksFromThinkingBlocks(msg.ThinkingBlocks)
			blocks = append(thinkingBlocks, blocks...)
		}
		if strings.EqualFold(strings.TrimSpace(msg.Role), "assistant") && len(msg.ToolCalls) > 0 {
			toolBlocks, toolErr := anthropicBlocksFromToolCalls(msg.ToolCalls)
			if toolErr != nil {
				return anthropicRequest{}, toolErr
			}
			if len(blocks) == 1 && blocks[0].Type == "text" && strings.TrimSpace(blocks[0].Text) == "" {
				blocks = nil
			}
			blocks = append(blocks, toolBlocks...)
		}
		if msg.Role == "system" {
			for _, block := range blocks {
				if block.Type != "text" {
					return anthropicRequest{}, fmt.Errorf("system messages only support text content")
				}
				anthReq.System = append(anthReq.System, block)
			}
			continue
		}
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		if role == "tool" {
			if strings.TrimSpace(msg.ToolCallID) == "" {
				return anthropicRequest{}, fmt.Errorf("tool message requires tool_call_id")
			}
			role = "user"
			blocks = []anthropicContentBlock{{
				Type:      "tool_result",
				ToolUseID: strings.TrimSpace(msg.ToolCallID),
				Content:   anthropicToolResultContent(msg.Content),
			}}
		}
		if role != "user" && role != "assistant" {
			return anthropicRequest{}, fmt.Errorf("unsupported anthropic message role %q", msg.Role)
		}
		anthReq.Messages = append(anthReq.Messages, anthropicMessage{
			Role:    role,
			Content: blocks,
		})
	}

	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		anthReq.MaxTokens = *req.MaxTokens
	}

	anthReq.Thinking, err = anthropicThinkingConfig(
		req,
		anthReq.MaxTokens,
		runtimeProfile,
	)
	if err != nil {
		return anthropicRequest{}, err
	}

	return anthReq, nil
}

func anthropicStopSequences(value interface{}) ([]string, error) {
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil, nil
		}
		return []string{typed}, nil
	case []string:
		return append([]string(nil), typed...), nil
	case []interface{}:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("stop must contain only strings")
			}
			out = append(out, text)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("stop must be a string or string array")
	}
}

func anthropicTools(tools []contractsopenai.ToolDefinition, cacheControl map[string]interface{}) ([]anthropicTool, error) {
	out := make([]anthropicTool, 0, len(tools))
	for index, tool := range tools {
		toolType := strings.ToLower(strings.TrimSpace(tool.Type))
		if toolType == "function" {
			if tool.Function == nil {
				return nil, fmt.Errorf("tool %d requires function", index)
			}
			name := strings.TrimSpace(tool.Function.Name)
			if name == "" {
				return nil, fmt.Errorf("tool %d requires a name", index)
			}
			schema := tool.Function.Parameters
			if schema == nil {
				schema = map[string]interface{}{
					"type": "object", "properties": map[string]interface{}{},
				}
			}
			item := anthropicTool{
				Name: name, Description: tool.Function.Description,
				InputSchema: schema, Strict: tool.Function.Strict,
			}
			if index == len(tools)-1 && len(cacheControl) > 0 {
				item.CacheControl = cloneStringAnyMap(cacheControl)
			}
			out = append(out, item)
			continue
		}
		if !isSupportedAnthropicServerToolType(toolType) {
			return nil, fmt.Errorf(
				"unsupported Anthropic server tool type %q",
				tool.Type,
			)
		}
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			name = anthropicServerToolDefaultName(toolType)
		}
		item := anthropicTool{
			Type: toolType, Name: name,
			AllowedCallers: append([]string(nil), tool.AllowedCallers...),
			AllowedDomains: append([]string(nil), tool.AllowedDomains...),
			BlockedDomains: append([]string(nil), tool.BlockedDomains...),
			MaxUses:        tool.MaxUses, UserLocation: cloneStringAnyMap(tool.UserLocation),
			CacheControl: cloneStringAnyMap(tool.CacheControl),
			Strict:       tool.Strict, DeferLoading: tool.DeferLoading,
		}
		if item.CacheControl == nil && index == len(tools)-1 &&
			len(cacheControl) > 0 {
			item.CacheControl = cloneStringAnyMap(cacheControl)
		}
		out = append(out, item)
	}
	return out, nil
}

func isSupportedAnthropicServerToolType(value string) bool {
	_, ok := map[string]struct{}{
		"web_search_20260318":     {},
		"web_search_20260209":     {},
		"web_search_20250305":     {},
		"web_fetch_20260318":      {},
		"web_fetch_20260309":      {},
		"web_fetch_20260209":      {},
		"web_fetch_20250910":      {},
		"code_execution_20260521": {},
		"code_execution_20260120": {},
		"code_execution_20250825": {},
	}[strings.ToLower(strings.TrimSpace(value))]
	return ok
}

func anthropicServerToolDefaultName(toolType string) string {
	for _, name := range []string{"web_search", "web_fetch", "code_execution"} {
		if strings.HasPrefix(toolType, name+"_") {
			return name
		}
	}
	return ""
}

func anthropicToolChoiceFromOpenAI(value interface{}, parallel *bool) (*anthropicToolChoice, error) {
	disableParallel := parallel != nil && !*parallel
	if value == nil {
		if !disableParallel {
			return nil, nil
		}
		return &anthropicToolChoice{Type: "auto", DisableParallelToolUse: true}, nil
	}
	if text, ok := value.(string); ok {
		switch strings.ToLower(strings.TrimSpace(text)) {
		case "auto":
			return &anthropicToolChoice{Type: "auto", DisableParallelToolUse: disableParallel}, nil
		case "required", "any":
			return &anthropicToolChoice{Type: "any", DisableParallelToolUse: disableParallel}, nil
		case "none":
			return nil, fmt.Errorf("Anthropic does not support tool_choice=none when tools are supplied")
		default:
			return nil, fmt.Errorf("unsupported tool_choice %q", text)
		}
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("invalid tool_choice")
	}
	var object struct {
		Type     string `json:"type"`
		Function *struct {
			Name string `json:"name"`
		} `json:"function"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &object); err != nil {
		return nil, fmt.Errorf("invalid tool_choice")
	}
	name := strings.TrimSpace(object.Name)
	if object.Function != nil {
		name = strings.TrimSpace(object.Function.Name)
	}
	if name == "" {
		return nil, fmt.Errorf("named tool_choice requires function.name")
	}
	return &anthropicToolChoice{Type: "tool", Name: name, DisableParallelToolUse: disableParallel}, nil
}

func anthropicMetadataFromOpenAI(metadata map[string]interface{}, user string) *anthropicMetadata {
	userID := strings.TrimSpace(user)
	if userID == "" && metadata != nil {
		if value, ok := metadata["user_id"].(string); ok {
			userID = strings.TrimSpace(value)
		}
	}
	if userID == "" {
		return nil
	}
	return &anthropicMetadata{UserID: userID}
}

func anthropicOutputConfigFromOpenAI(
	runtimeProfile platform.ModelRuntimeProfile,
	responseFormat interface{},
	output map[string]interface{},
	effort string,
) (*anthropicOutputConfig, error) {
	config := &anthropicOutputConfig{}
	if value, ok := output["effort"].(string); ok {
		config.Effort = strings.TrimSpace(value)
	} else if normalized := normalizeAnthropicEffort(effort); normalized != "" &&
		runtimeProfile.AllowsEffort(normalized) {
		config.Effort = normalized
	}
	if responseFormat != nil {
		raw, err := json.Marshal(responseFormat)
		if err != nil {
			return nil, fmt.Errorf("invalid response_format")
		}
		var format struct {
			Type       string          `json:"type"`
			JSONSchema json.RawMessage `json:"json_schema"`
		}
		if err := json.Unmarshal(raw, &format); err != nil {
			return nil, fmt.Errorf("invalid response_format")
		}
		switch format.Type {
		case "", "text":
		case "json_schema":
			if len(format.JSONSchema) == 0 || string(format.JSONSchema) == "null" {
				return nil, fmt.Errorf("response_format.json_schema is required")
			}
			var envelope struct {
				Schema json.RawMessage `json:"schema"`
			}
			if err := json.Unmarshal(format.JSONSchema, &envelope); err != nil {
				return nil, fmt.Errorf("invalid response_format.json_schema")
			}
			schema := format.JSONSchema
			if len(envelope.Schema) > 0 && string(envelope.Schema) != "null" {
				schema = envelope.Schema
			}
			var preserved interface{}
			if err := json.Unmarshal(schema, &preserved); err != nil {
				return nil, fmt.Errorf("invalid response_format.json_schema.schema")
			}
			config.Format = map[string]interface{}{"type": "json_schema", "schema": preserved}
		default:
			return nil, fmt.Errorf("unsupported response_format type %q", format.Type)
		}
	}
	if config.Format == nil {
		if value, ok := output["format"]; ok {
			config.Format = value
		}
	}
	if config.Effort == "" && config.Format == nil {
		return nil, nil
	}
	return config, nil
}

func validateAnthropicRequestControls(
	runtimeProfile platform.ModelRuntimeProfile,
	serviceTier, speed, inferenceGeo string,
) error {
	serviceTier = strings.ToLower(strings.TrimSpace(serviceTier))
	speed = strings.ToLower(strings.TrimSpace(speed))
	inferenceGeo = strings.ToLower(strings.TrimSpace(inferenceGeo))
	if serviceTier != "" &&
		!runtimeProfile.AllowsServiceTier(serviceTier) {
		return fmt.Errorf("unsupported service_tier %q", serviceTier)
	}
	if inferenceGeo != "" &&
		!runtimeProfile.AllowsInferenceGeo(inferenceGeo) {
		return fmt.Errorf("unsupported inference_geo %q", inferenceGeo)
	}
	if speed != "" && !runtimeProfile.AllowsSpeed(speed) {
		return fmt.Errorf("unsupported speed %q", speed)
	}
	return nil
}

func normalizeAnthropicEffort(effort string) string {
	switch strings.ToLower(strings.TrimSpace(effort)) {
	case "minimal", "low":
		return "low"
	case "medium":
		return "medium"
	case "high":
		return "high"
	case "xhigh", "max":
		return "max"
	default:
		return ""
	}
}

func anthropicBlocksFromToolCalls(calls []contractsopenai.ToolCall) ([]anthropicContentBlock, error) {
	out := make([]anthropicContentBlock, 0, len(calls))
	for _, call := range calls {
		if call.Function == nil {
			return nil, fmt.Errorf("anthropic tool call requires a function")
		}
		var input interface{} = map[string]interface{}{}
		if strings.TrimSpace(call.Function.Arguments) != "" {
			if err := json.Unmarshal([]byte(call.Function.Arguments), &input); err != nil {
				return nil, fmt.Errorf("tool call %q arguments must be valid JSON", call.ID)
			}
		}
		out = append(out, anthropicContentBlock{
			Type:  "tool_use",
			ID:    call.ID,
			Name:  call.Function.Name,
			Input: input,
		})
	}
	return out, nil
}

func anthropicToolResultContent(content interface{}) interface{} {
	if text := contractsopenai.ChatContentText(content); text != "" {
		return text
	}
	return content
}

func cloneStringAnyMap(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	out := make(map[string]interface{}, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func anthropicBlocksFromThinkingBlocks(blocks []contractsopenai.ThinkingBlock) []anthropicContentBlock {
	out := make([]anthropicContentBlock, 0, len(blocks))
	for _, block := range blocks {
		blockType, _ := block["type"].(string)
		blockType = strings.TrimSpace(blockType)
		switch blockType {
		case "thinking":
			item := anthropicContentBlock{Type: "thinking"}
			if thinking, ok := block["thinking"].(string); ok {
				item.Thinking = thinking
			}
			if signature, ok := block["signature"].(string); ok {
				item.Signature = signature
			}
			out = append(out, item)
		case "redacted_thinking":
			item := anthropicContentBlock{Type: "redacted_thinking"}
			if data, ok := block["data"].(string); ok {
				item.Data = data
			}
			out = append(out, item)
		}
	}
	return out
}

func anthropicBlocksFromContent(content interface{}) ([]anthropicContentBlock, error) {
	parts, ok, err := contractsopenai.ChatContentParts(content)
	if err != nil {
		return nil, err
	}
	if !ok {
		text := contractsopenai.ChatContentText(content)
		if text == "" {
			text = " "
		}
		return []anthropicContentBlock{{Type: "text", Text: text}}, nil
	}
	blocks := make([]anthropicContentBlock, 0, len(parts))
	for _, part := range parts {
		switch part.Type {
		case "text":
			blocks = append(blocks, anthropicContentBlock{Type: "text", Text: part.Text})
		case "image_url":
			source := &anthropicSource{}
			if strings.HasPrefix(strings.ToLower(strings.TrimSpace(part.ImageURL.URL)), "data:") {
				mediaType, data, err := contractsopenai.ParseImageDataURL(part.ImageURL.URL)
				if err != nil {
					return nil, err
				}
				source.Type = "base64"
				source.MediaType = mediaType
				source.Data = data
			} else {
				if !isHTTPURL(part.ImageURL.URL) {
					return nil, fmt.Errorf("image_url must use https or a base64 data URL")
				}
				source.Type = "url"
				source.URL = strings.TrimSpace(part.ImageURL.URL)
			}
			blocks = append(blocks, anthropicContentBlock{
				Type:   "image",
				Source: source,
			})
		case "file":
			if strings.TrimSpace(part.File.FileID) != "" {
				blocks = append(blocks, anthropicContentBlock{
					Type: "document",
					Source: &anthropicSource{
						Type:   "file",
						FileID: strings.TrimSpace(part.File.FileID),
					},
				})
				continue
			}
			mediaType, data, err := contractsopenai.ParseFileDataURL(part.File.FileData)
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, anthropicContentBlock{
				Type: "document",
				Source: &anthropicSource{
					Type:      "base64",
					MediaType: mediaType,
					Data:      data,
				},
			})
		}
	}
	if len(blocks) == 0 {
		blocks = append(blocks, anthropicContentBlock{Type: "text", Text: " "})
	}
	return blocks, nil
}

func isHTTPURL(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(value, "https://")
}

func canonicalAnthropicModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	if strings.Count(model, ":") == 2 {
		parts := strings.SplitN(model, ":", 3)
		model = strings.TrimSpace(parts[2])
	}
	const providerPrefix = "anthropic/"
	if strings.HasPrefix(strings.ToLower(model), providerPrefix) {
		return strings.TrimSpace(model[len(providerPrefix):])
	}
	return model
}

func anthropicThinkingConfig(
	req contractsopenai.ChatCompletionRequest,
	maxTokens int,
	runtimeProfile platform.ModelRuntimeProfile,
) (*contractsopenai.ThinkingConfig, error) {
	if req.Thinking != nil {
		if strings.EqualFold(strings.TrimSpace(req.Thinking.Type), "disabled") {
			return nil, nil
		}
		cfg := *req.Thinking
		if strings.TrimSpace(cfg.Type) == "" {
			cfg.Type = "enabled"
		}
		if strings.EqualFold(cfg.Type, "adaptive") {
			cfg.BudgetTokens = 0
			return &cfg, nil
		}
		if !strings.EqualFold(cfg.Type, "enabled") {
			return nil, fmt.Errorf("unsupported thinking type %q", cfg.Type)
		}
		if cfg.BudgetTokens <= 0 {
			cfg.BudgetTokens = defaultAnthropicThinkingBudget(req.ReasoningEffort, maxTokens)
		}
		return &cfg, nil
	}

	effort := strings.ToLower(strings.TrimSpace(req.ReasoningEffort))
	switch strings.ToLower(
		strings.TrimSpace(runtimeProfile.Thinking.DefaultMode),
	) {
	case "adaptive":
		if effort == "none" {
			return nil, nil
		}
		return &contractsopenai.ThinkingConfig{Type: "adaptive"}, nil
	case "disabled":
		return nil, nil
	}
	if effort == "" || effort == "none" {
		return nil, nil
	}

	return &contractsopenai.ThinkingConfig{
		Type:         "enabled",
		BudgetTokens: defaultAnthropicThinkingBudget(effort, maxTokens),
	}, nil
}

func defaultAnthropicThinkingBudget(effort string, maxTokens int) int {
	budget := 2048
	switch strings.ToLower(strings.TrimSpace(effort)) {
	case "minimal", "low":
		budget = 1024
	case "medium", "":
		budget = 2048
	case "high":
		budget = 3200
	case "xhigh", "max":
		budget = 8192
	}
	if maxTokens > 1024 && budget >= maxTokens {
		return maxTokens - 1
	}
	return budget
}

func convertToOpenAIResponse(resp anthropicResponse, model string) *contractsopenai.ChatCompletionResponse {
	content := ""
	reasoningContent := ""
	thinkingBlocks := make([]contractsopenai.ThinkingBlock, 0)
	toolCalls := make([]contractsopenai.ToolCall, 0)
	citations := make([]interface{}, 0)
	providerContent := make([]map[string]interface{}, 0)
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			content += block.Text
		case "thinking":
			thinkingBlock := contractsopenai.ThinkingBlock{"type": block.Type}
			if block.Thinking != "" {
				thinkingBlock["thinking"] = block.Thinking
				reasoningContent += block.Thinking
			}
			if block.Signature != "" {
				thinkingBlock["signature"] = block.Signature
			}
			thinkingBlocks = append(thinkingBlocks, thinkingBlock)
		case "redacted_thinking":
			thinkingBlock := contractsopenai.ThinkingBlock{"type": block.Type}
			if block.Data != "" {
				thinkingBlock["data"] = block.Data
			}
			thinkingBlocks = append(thinkingBlocks, thinkingBlock)
		case "tool_use":
			arguments, _ := json.Marshal(block.Input)
			toolCalls = append(toolCalls, contractsopenai.ToolCall{
				ID:   block.ID,
				Type: "function",
				Function: &contractsopenai.FunctionCall{
					Name:      block.Name,
					Arguments: string(arguments),
				},
			})
		case "server_tool_use", "web_search_tool_result",
			"web_fetch_tool_result", "code_execution_tool_result":
			if len(block.ProviderData) > 0 {
				providerContent = append(providerContent, block.ProviderData)
			}
		default:
			content += block.Text
		}
		citations = append(citations, block.Citations...)
	}

	message := &contractsopenai.ChatMessage{
		Role:    "assistant",
		Content: content,
	}
	if reasoningContent != "" {
		message.ReasoningContent = reasoningContent
	}
	if len(thinkingBlocks) > 0 {
		message.ThinkingBlocks = thinkingBlocks
	}
	if len(toolCalls) > 0 {
		message.ToolCalls = toolCalls
	}
	if len(citations) > 0 {
		message.Citations = citations
	}
	if len(providerContent) > 0 {
		message.ProviderContent = providerContent
	}

	return &contractsopenai.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []contractsopenai.Choice{{
			Index:        0,
			Message:      message,
			FinishReason: mapAnthropicStopReason(resp.StopReason),
		}},
		Usage: &contractsopenai.Usage{
			PromptTokens:               resp.Usage.InputTokens,
			CompletionTokens:           resp.Usage.OutputTokens,
			TotalTokens:                resp.Usage.InputTokens + resp.Usage.OutputTokens,
			CacheCreationInputTokens:   resp.Usage.CacheCreationInputTokens,
			CacheCreation5mInputTokens: resp.Usage.CacheCreation.Ephemeral5mInputTokens,
			CacheCreation1hInputTokens: resp.Usage.CacheCreation.Ephemeral1hInputTokens,
			CacheReadInputTokens:       resp.Usage.CacheReadInputTokens,
			ServerToolUse:              resp.Usage.ServerToolUse,
			ServiceTier:                resp.Usage.ServiceTier,
			Speed:                      resp.Usage.Speed,
		},
	}
}

func mapAnthropicStopReason(reason string) string {
	switch strings.TrimSpace(reason) {
	case "end_turn", "stop_sequence", "pause_turn", "refusal":
		return "stop"
	case "max_tokens", "model_context_window_exceeded":
		return "length"
	case "tool_use":
		return "tool_calls"
	default:
		return strings.TrimSpace(reason)
	}
}

func isAnthropicServerContentBlock(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "server_tool_use", "web_search_tool_result",
		"web_fetch_tool_result", "code_execution_tool_result":
		return true
	default:
		return false
	}
}

func openAIUsageFromAnthropic(usage anthropicUsage) *contractsopenai.Usage {
	return &contractsopenai.Usage{
		PromptTokens:               usage.InputTokens,
		CompletionTokens:           usage.OutputTokens,
		TotalTokens:                usage.InputTokens + usage.OutputTokens,
		CacheCreationInputTokens:   usage.CacheCreationInputTokens,
		CacheCreation5mInputTokens: usage.CacheCreation.Ephemeral5mInputTokens,
		CacheCreation1hInputTokens: usage.CacheCreation.Ephemeral1hInputTokens,
		CacheReadInputTokens:       usage.CacheReadInputTokens,
		ServerToolUse:              usage.ServerToolUse,
		ServiceTier:                usage.ServiceTier,
		Speed:                      usage.Speed,
	}
}
