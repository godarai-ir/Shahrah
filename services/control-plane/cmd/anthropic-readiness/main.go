// Command anthropic-readiness runs the live, credential-backed release matrix
// against GodarAI's Anthropic-native API. It never writes the virtual key,
// Anthropic Admin key, prompts, or upstream opaque identifiers to its report.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

var defaultModels = []string{
	"anthropic/claude-haiku-4-5",
	"anthropic/claude-sonnet-5",
	"anthropic/claude-opus-5",
	"anthropic/claude-fable-5",
}

type modelCapabilitySpec struct {
	ProfileVersion int
	Thinking       string
	WebSearchType  string
	Batch300K      bool
	ExpectedIDs    []string
}

var modelCapabilities = map[string]modelCapabilitySpec{
	"anthropic/claude-haiku-4-5": {
		ProfileVersion: 2,
		Thinking:       "manual",
		WebSearchType:  "web_search_20250305",
		ExpectedIDs: []string{
			"claude-haiku-4-5",
			"claude-haiku-4-5-20251001",
		},
	},
	"anthropic/claude-sonnet-5": {
		ProfileVersion: 2,
		Thinking:       "adaptive",
		WebSearchType:  "web_search_20260318",
		Batch300K:      true,
		ExpectedIDs:    []string{"claude-sonnet-5"},
	},
	"anthropic/claude-opus-5": {
		ProfileVersion: 2,
		Thinking:       "adaptive",
		WebSearchType:  "web_search_20260318",
		Batch300K:      true,
		ExpectedIDs:    []string{"claude-opus-5"},
	},
	"anthropic/claude-fable-5": {
		ProfileVersion: 2,
		Thinking:       "adaptive",
		WebSearchType:  "web_search_20260318",
		ExpectedIDs:    []string{"claude-fable-5"},
	},
}

type config struct {
	baseURL         string
	virtualKey      string
	authLabel       string
	models          []string
	providerAccount string
	reportPath      string
	adminKey        string
	workspaceID     string
	runResources    bool
	test300K        bool
	timeout         time.Duration
}

type report struct {
	GeneratedAt       time.Time          `json:"generated_at"`
	BaseURL           string             `json:"base_url"`
	AuthLabel         string             `json:"auth_label"`
	Models            []modelReport      `json:"models"`
	Resources         []checkResult      `json:"resources,omitempty"`
	RateLimitSnapshot *rateLimitSnapshot `json:"rate_limit_snapshot,omitempty"`
	Summary           reportSummary      `json:"summary"`
}

type modelReport struct {
	Model          string        `json:"model"`
	ProfileVersion int           `json:"profile_version"`
	Checks         []checkResult `json:"checks"`
}

type checkResult struct {
	Name          string `json:"name"`
	Passed        bool   `json:"passed"`
	StatusCode    int    `json:"status_code,omitempty"`
	RequestIDHash string `json:"request_id_sha256,omitempty"`
	Detail        string `json:"detail,omitempty"`
}

type reportSummary struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Blocked int `json:"blocked"`
}

type rateLimitSnapshot struct {
	ObservedAt time.Time         `json:"observed_at"`
	Org        rateLimitSummary  `json:"organization"`
	Workspace  *rateLimitSummary `json:"workspace,omitempty"`
}

type rateLimitSummary struct {
	ResponseBytes  int    `json:"response_bytes"`
	ResponseSHA256 string `json:"response_sha256"`
	ResourceCount  int    `json:"resource_count"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)

	client := &http.Client{Timeout: 90 * time.Second}
	result := run(ctx, client, cfg)

	cancel()

	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := os.WriteFile(cfg.reportPath, encoded, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	fmt.Printf(
		"Anthropic readiness report: %s (passed=%d failed=%d blocked=%d)\n",
		cfg.reportPath,
		result.Summary.Passed,
		result.Summary.Failed,
		result.Summary.Blocked,
	)

	if result.Summary.Failed > 0 || result.Summary.Blocked > 0 {
		os.Exit(1)
	}
}

func loadConfig() (config, error) {
	var (
		cfg      config
		modelCSV string
	)

	flag.StringVar(&cfg.baseURL, "base-url", envOr("GODARAI_BASE_URL", "http://localhost:3002"), "GodarAI control-plane base URL")
	flag.StringVar(&cfg.authLabel, "auth-label", envOr("GODARAI_ANTHROPIC_AUTH_LABEL", "unspecified"), "sanitized label such as api_key or wif")
	flag.StringVar(&modelCSV, "models", strings.Join(defaultModels, ","), "comma-separated GodarAI Anthropic model FQNs")
	flag.StringVar(&cfg.providerAccount, "provider-account", os.Getenv("GODARAI_ANTHROPIC_PROVIDER_ACCOUNT"), "provider account selector required by Files API")
	flag.StringVar(&cfg.reportPath, "report", "anthropic-readiness-report.json", "sanitized JSON report path")
	flag.StringVar(&cfg.workspaceID, "workspace-id", os.Getenv("ANTHROPIC_ADMIN_WORKSPACE_ID"), "optional Anthropic workspace ID for Rate Limits API")
	flag.BoolVar(&cfg.runResources, "run-resources", false, "run live Files and Message Batch lifecycle checks (may incur cost)")
	flag.BoolVar(&cfg.test300K, "test-300k", false, "create a 300k max-output Batch check on eligible models (may incur cost)")
	flag.DurationVar(&cfg.timeout, "timeout", 20*time.Minute, "overall matrix timeout")
	flag.Parse()

	cfg.virtualKey = strings.TrimSpace(os.Getenv("GODARAI_VIRTUAL_KEY"))
	cfg.adminKey = strings.TrimSpace(os.Getenv("ANTHROPIC_ADMIN_KEY"))
	cfg.baseURL = strings.TrimRight(strings.TrimSpace(cfg.baseURL), "/")

	cfg.models = splitCSV(modelCSV)
	if cfg.virtualKey == "" {
		return cfg, errors.New("GODARAI_VIRTUAL_KEY is required")
	}

	if len(cfg.models) == 0 {
		return cfg, errors.New("at least one model is required")
	}

	if cfg.runResources && strings.TrimSpace(cfg.providerAccount) == "" {
		return cfg, errors.New("-provider-account is required with -run-resources")
	}

	return cfg, nil
}

func run(ctx context.Context, client *http.Client, cfg config) report {
	out := report{
		GeneratedAt: time.Now().UTC(),
		BaseURL:     cfg.baseURL,
		AuthLabel:   cfg.authLabel,
	}
	for _, model := range cfg.models {
		spec := modelCapabilities[model]
		item := modelReport{
			Model:          model,
			ProfileVersion: spec.ProfileVersion,
		}
		item.Checks = append(item.Checks,
			checkModel(ctx, client, cfg, model, spec),
			checkCountTokens(ctx, client, cfg, model),
			checkMessage(ctx, client, cfg, model, spec, false),
			checkMessage(ctx, client, cfg, model, spec, true),
			checkStructuredOutput(ctx, client, cfg, model),
			checkThinking(ctx, client, cfg, model, spec),
			checkOpenAIServerToolContract(ctx, client, cfg, model, spec),
		)
		out.Models = append(out.Models, item)
	}

	if cfg.runResources {
		out.Resources = runResourceChecks(ctx, client, cfg)
	}

	if cfg.adminKey != "" {
		snapshot, checks := fetchRateLimitSnapshot(ctx, client, cfg)
		out.RateLimitSnapshot = snapshot
		out.Resources = append(out.Resources, checks...)
	} else {
		out.Resources = append(out.Resources, checkResult{
			Name:   "rate_limits.admin_api",
			Detail: "BLOCKED: ANTHROPIC_ADMIN_KEY was not provided",
		})
	}

	for _, item := range out.Models {
		for _, check := range item.Checks {
			accumulate(&out.Summary, check)
		}
	}

	for _, check := range out.Resources {
		accumulate(&out.Summary, check)
	}

	return out
}

func checkModel(
	ctx context.Context,
	client *http.Client,
	cfg config,
	model string,
	spec modelCapabilitySpec,
) checkResult {
	status, headers, body, err := request(
		ctx,
		client,
		cfg,
		http.MethodGet,
		"/v1/models/"+url.PathEscape(strings.TrimPrefix(model, "anthropic/")),
		nil,
		nil,
	)

	result := resultFor("models.get", status, headers, body, err)
	if result.Passed {
		var payload struct {
			ID             string `json:"id"`
			MaxInputTokens int    `json:"max_input_tokens"`
			MaxTokens      int    `json:"max_tokens"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			result.Passed = false
			result.Detail = "Models API response was invalid JSON"
		} else if !containsExact(spec.ExpectedIDs, payload.ID) {
			result.Passed = false
			result.Detail = fmt.Sprintf(
				"Models API returned %q for requested model",
				payload.ID,
			)
		} else if payload.MaxInputTokens <= 0 || payload.MaxTokens <= 0 {
			result.Passed = false
			result.Detail = "Models API response omitted token limits"
		}
	}

	return result
}

func checkCountTokens(ctx context.Context, client *http.Client, cfg config, model string) checkResult {
	return jsonCheck(ctx, client, cfg, "messages.count_tokens", "/v1/messages/count_tokens", map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{{
			"role": "user", "content": "Reply with OK.",
		}},
	})
}

func checkMessage(
	ctx context.Context,
	client *http.Client,
	cfg config,
	model string,
	spec modelCapabilitySpec,
	stream bool,
) checkResult {
	name := "messages.sync"
	if stream {
		name = "messages.stream"
	}

	body := map[string]interface{}{
		"model": model, "max_tokens": 32, "stream": stream,
		"messages": []map[string]interface{}{{
			"role": "user", "content": "Reply with exactly OK.",
		}},
	}
	encoded, _ := json.Marshal(body) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	status, headers, responseBody, err := request(ctx, client, cfg, http.MethodPost, "/v1/messages", encoded, nil)

	result := resultFor(name, status, headers, responseBody, err)
	if result.Passed && stream &&
		(!bytes.Contains(responseBody, []byte("message_stop")) ||
			bytes.Contains(responseBody, []byte("anthropic_stream_error"))) {
		result.Passed = false
		result.Detail = "stream did not end with message_stop or contained an error"
	}

	if result.Passed && !responseUsesExpectedModel(responseBody, spec.ExpectedIDs) {
		result.Passed = false
		result.Detail = "response omitted the served model or returned an unexpected model"
	}

	return result
}

func responseUsesExpectedModel(body []byte, expected []string) bool {
	compact := bytes.Map(func(value rune) rune {
		switch value {
		case ' ', '\t', '\r', '\n':
			return -1
		default:
			return value
		}
	}, body)

	for _, model := range expected {
		encoded, _ := json.Marshal(model) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
		if bytes.Contains(compact, append([]byte(`"model":`), encoded...)) {
			return true
		}
	}

	return false
}

func containsExact(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}

	return false
}

func checkStructuredOutput(ctx context.Context, client *http.Client, cfg config, model string) checkResult {
	return jsonCheck(ctx, client, cfg, "messages.structured_output", "/v1/messages", map[string]interface{}{
		"model": model, "max_tokens": 64,
		"messages": []map[string]interface{}{{"role": "user", "content": "Return status OK."}},
		"output_config": map[string]interface{}{
			"format": map[string]interface{}{
				"type": "json_schema",
				"schema": map[string]interface{}{
					"type":                 "object",
					"properties":           map[string]interface{}{"status": map[string]string{"type": "string"}},
					"required":             []string{"status"},
					"additionalProperties": false,
				},
			},
		},
	})
}

func checkThinking(
	ctx context.Context,
	client *http.Client,
	cfg config,
	model string,
	spec modelCapabilitySpec,
) checkResult {
	thinking := map[string]interface{}{"type": "enabled", "budget_tokens": 1024}
	output := map[string]interface{}{}

	if spec.Thinking == "adaptive" {
		thinking = map[string]interface{}{"type": "adaptive"}
		output["effort"] = "low"
	} else if spec.Thinking != "manual" {
		return checkResult{
			Name:   "messages.thinking",
			Detail: "BLOCKED: runtime capability spec is unavailable",
		}
	}

	payload := map[string]interface{}{
		"model": model, "max_tokens": 2048, "thinking": thinking,
		"messages": []map[string]interface{}{{"role": "user", "content": "Reply with OK."}},
	}
	if len(output) > 0 {
		payload["output_config"] = output
	}

	return jsonCheck(ctx, client, cfg, "messages.thinking", "/v1/messages", payload)
}

func checkOpenAIServerToolContract(
	ctx context.Context,
	client *http.Client,
	cfg config,
	model string,
	spec modelCapabilitySpec,
) checkResult {
	if spec.WebSearchType == "" {
		return checkResult{
			Name:   "openai.server_tool_contract",
			Detail: "BLOCKED: runtime capability spec is unavailable",
		}
	}

	payload := map[string]interface{}{
		"model": model, "max_tokens": 32,
		"messages": []map[string]interface{}{{"role": "user", "content": "Reply OK without using tools."}},
		"tools": []map[string]interface{}{{
			"type": spec.WebSearchType, "name": "web_search", "max_uses": 1,
		}},
		"tool_choice": "auto",
	}

	return jsonCheck(ctx, client, cfg, "openai.server_tool_contract", "/v1/chat/completions", payload)
}

func runResourceChecks(ctx context.Context, client *http.Client, cfg config) []checkResult {
	checks := make([]checkResult, 0, 12)

	var upload bytes.Buffer

	writer := multipart.NewWriter(&upload)
	part, _ := writer.CreateFormFile("file", "readiness.txt")  //nolint:errcheck // A downstream disconnect is handled by terminating the stream.
	_, _ = io.WriteString(part, "GodarAI Anthropic readiness") //nolint:errcheck // A downstream disconnect is handled by terminating the stream.
	_ = writer.Close()                                         //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.
	extra := map[string]string{"Content-Type": writer.FormDataContentType()}
	status, headers, body, err := request(ctx, client, cfg, http.MethodPost, "/v1/files", upload.Bytes(), extra)
	uploadResult := resultFor("files.create", status, headers, body, err)
	checks = append(checks,
		uploadResult,
		methodCheck(ctx, client, cfg, "files.list", http.MethodGet, "/v1/files", nil),
	)

	if fileID := responseID(body); uploadResult.Passed && fileID != "" {
		filePath := "/v1/files/" + url.PathEscape(fileID)
		checks = append(checks,
			methodCheck(ctx, client, cfg, "files.get", http.MethodGet, filePath, nil),
			methodCheck(ctx, client, cfg, "files.content", http.MethodGet, filePath+"/content", nil),
			methodCheck(ctx, client, cfg, "files.delete", http.MethodDelete, filePath, nil),
		)
	} else if uploadResult.Passed {
		checks = append(checks, checkResult{
			Name: "files.lifecycle", Detail: "create response omitted an opaque file id",
		})
	}

	batchModel := cfg.models[0]
	maxTokens := 32

	if cfg.test300K {
		for _, model := range cfg.models {
			if modelCapabilities[model].Batch300K {
				batchModel = model
				maxTokens = 300000

				break
			}
		}
	}

	batch := map[string]interface{}{
		"requests": []map[string]interface{}{{
			"custom_id": "readiness",
			"params": map[string]interface{}{
				"model": batchModel, "max_tokens": maxTokens,
				"messages": []map[string]interface{}{{"role": "user", "content": "Reply exactly OK."}},
			},
		}},
	}
	batchBody, _ := json.Marshal(batch) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	status, headers, body, err = request(
		ctx, client, cfg, http.MethodPost, "/v1/messages/batches", batchBody, nil,
	)
	createResult := resultFor("batches.create", status, headers, body, err)

	checks = append(checks,
		createResult,
		methodCheck(ctx, client, cfg, "batches.list", http.MethodGet, "/v1/messages/batches", nil),
	)
	if batchID := responseID(body); createResult.Passed && batchID != "" {
		checks = append(checks, waitForBatch(ctx, client, cfg, batchID)...)
	} else if createResult.Passed {
		checks = append(checks, checkResult{
			Name: "batches.lifecycle", Detail: "create response omitted an opaque batch id",
		})
	}

	cancelBody, _ := json.Marshal(map[string]interface{}{ //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
		"requests": []map[string]interface{}{{
			"custom_id": "readiness-cancel",
			"params": map[string]interface{}{
				"model": cfg.models[0], "max_tokens": 32,
				"messages": []map[string]interface{}{{"role": "user", "content": "Reply exactly OK."}},
			},
		}},
	})
	status, headers, body, err = request(
		ctx, client, cfg, http.MethodPost, "/v1/messages/batches", cancelBody, nil,
	)
	cancelCreate := resultFor("batches.cancel.create", status, headers, body, err)

	checks = append(checks, cancelCreate)
	if batchID := responseID(body); cancelCreate.Passed && batchID != "" {
		checks = append(checks, methodCheck(
			ctx,
			client,
			cfg,
			"batches.cancel",
			http.MethodPost,
			"/v1/messages/batches/"+url.PathEscape(batchID)+"/cancel",
			nil,
		))
	}

	return checks
}

func waitForBatch(
	ctx context.Context,
	client *http.Client,
	cfg config,
	batchID string,
) []checkResult {
	target := "/v1/messages/batches/" + url.PathEscape(batchID)
	for attempt := 0; attempt < 90; attempt++ {
		status, headers, body, err := request(
			ctx, client, cfg, http.MethodGet, target, nil, nil,
		)

		result := resultFor("batches.get", status, headers, body, err)
		if !result.Passed {
			return []checkResult{result}
		}

		var payload struct {
			ProcessingStatus string `json:"processing_status"`
		}
		if json.Unmarshal(body, &payload) != nil {
			result.Passed = false
			result.Detail = "batch status response was invalid"

			return []checkResult{result}
		}

		if payload.ProcessingStatus == "ended" {
			results := methodCheck(
				ctx,
				client,
				cfg,
				"batches.results",
				http.MethodGet,
				target+"/results",
				nil,
			)

			return []checkResult{result, results}
		}

		select {
		case <-ctx.Done():
			return []checkResult{{
				Name: "batches.get", Detail: ctx.Err().Error(),
			}}
		case <-time.After(2 * time.Second):
		}
	}

	return []checkResult{{
		Name: "batches.get", Detail: "batch did not end within three minutes",
	}}
}

func fetchRateLimitSnapshot(
	ctx context.Context,
	client *http.Client,
	cfg config,
) (*rateLimitSnapshot, []checkResult) {
	snapshot := &rateLimitSnapshot{ObservedAt: time.Now().UTC()}
	org, orgCheck := adminGet(ctx, client, cfg, "/v1/organizations/rate_limits")
	snapshot.Org = org
	checks := []checkResult{orgCheck}

	if cfg.workspaceID != "" {
		workspacePath := "/v1/organizations/workspaces/" +
			url.PathEscape(cfg.workspaceID) + "/rate_limits"
		workspace, workspaceCheck := adminGet(ctx, client, cfg, workspacePath)
		snapshot.Workspace = &workspace

		checks = append(checks, workspaceCheck)
	}

	return snapshot, checks
}

func adminGet(
	ctx context.Context,
	client *http.Client,
	cfg config,
	targetPath string,
) (rateLimitSummary, checkResult) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.anthropic.com"+targetPath, http.NoBody)
	if err != nil {
		return rateLimitSummary{}, checkResult{Name: "rate_limits.snapshot", Detail: err.Error()}
	}

	req.Header.Set("x-api-key", cfg.adminKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return rateLimitSummary{}, checkResult{Name: "rate_limits.snapshot", Detail: err.Error()}
	}
	defer resp.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))

	name := "rate_limits.organization"
	if strings.Contains(targetPath, "/workspaces/") {
		name = "rate_limits.workspace"
	}

	result := resultFor(name, resp.StatusCode, resp.Header, body, readErr)
	if !json.Valid(body) {
		result.Passed = false
		result.Detail = "Rate Limits API returned invalid JSON"

		return rateLimitSummary{}, result
	}

	var envelope struct {
		Data []json.RawMessage `json:"data"`
	}

	_ = json.Unmarshal(body, &envelope) //nolint:errcheck // Invalid optional response metadata is treated as absent.
	digest := sha256.Sum256(body)

	return rateLimitSummary{
		ResponseBytes:  len(body),
		ResponseSHA256: fmt.Sprintf("%x", digest[:]),
		ResourceCount:  len(envelope.Data),
	}, result
}

func jsonCheck(
	ctx context.Context,
	client *http.Client,
	cfg config,
	name string,
	targetPath string,
	payload interface{},
) checkResult {
	var encoded []byte

	method := http.MethodGet
	if payload != nil {
		method = http.MethodPost
		encoded, _ = json.Marshal(payload) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	}

	status, headers, body, err := request(ctx, client, cfg, method, targetPath, encoded, nil)

	return resultFor(name, status, headers, body, err)
}

func methodCheck(
	ctx context.Context,
	client *http.Client,
	cfg config,
	name string,
	method string,
	targetPath string,
	payload interface{},
) checkResult {
	var body []byte
	if payload != nil {
		body, _ = json.Marshal(payload) //nolint:errcheck // The value is composed exclusively of JSON-safe fields.
	}

	status, headers, responseBody, err := request(
		ctx, client, cfg, method, targetPath, body, nil,
	)

	return resultFor(name, status, headers, responseBody, err)
}

func request(
	ctx context.Context,
	client *http.Client,
	cfg config,
	method string,
	targetPath string,
	body []byte,
	extraHeaders map[string]string,
) (int, http.Header, []byte, error) {
	target, err := url.Parse(cfg.baseURL)
	if err != nil {
		return 0, nil, nil, err
	}

	target.Path = path.Join(target.Path, targetPath)
	if strings.HasSuffix(targetPath, "/") {
		target.Path += "/"
	}

	req, err := http.NewRequestWithContext(ctx, method, target.String(), bytes.NewReader(body))
	if err != nil {
		return 0, nil, nil, err
	}

	req.Header.Set("x-api-key", cfg.virtualKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	if cfg.providerAccount != "" {
		req.Header.Set("x-godarai-provider-account", cfg.providerAccount)
	}

	for name, value := range extraHeaders {
		req.Header.Set(name, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // Cleanup is best-effort and cannot change the completed operation.

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 32*1024*1024))

	return resp.StatusCode, resp.Header.Clone(), responseBody, err
}

func resultFor(
	name string,
	status int,
	headers http.Header,
	body []byte,
	err error,
) checkResult {
	result := checkResult{
		Name:       name,
		Passed:     err == nil && status >= 200 && status < 300,
		StatusCode: status,
		RequestIDHash: opaqueIDHash(
			firstHeader(headers, "request-id", "x-request-id"),
		),
	}
	if err != nil {
		result.Detail = err.Error()
	} else if !result.Passed {
		result.Detail = sanitizedError(body)
	}

	return result
}

func opaqueIDHash(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	digest := sha256.Sum256([]byte(value))

	return fmt.Sprintf("%x", digest[:])
}

func sanitizedError(body []byte) string {
	var envelope struct {
		Error struct {
			Type    string `json:"type"`
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil {
		values := []string{envelope.Error.Type, envelope.Error.Code, envelope.Error.Message}
		return strings.TrimSpace(strings.Join(nonEmpty(values), ": "))
	}

	return fmt.Sprintf("HTTP response body omitted from sanitized report (%d bytes)", len(body))
}

func firstHeader(headers http.Header, names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(headers.Get(name)); value != "" {
			return value
		}
	}

	return ""
}

func responseID(body []byte) string {
	var response struct {
		ID string `json:"id"`
	}

	_ = json.Unmarshal(body, &response) //nolint:errcheck // Invalid optional response metadata is treated as absent.

	return strings.TrimSpace(response.ID)
}

func accumulate(summary *reportSummary, result checkResult) {
	if result.Passed {
		summary.Passed++
	} else if strings.HasPrefix(result.Detail, "BLOCKED:") {
		summary.Blocked++
	} else {
		summary.Failed++
	}
}

func splitCSV(value string) []string {
	seen := map[string]struct{}{}

	var result []string

	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		result = append(result, item)
	}

	sort.Strings(result)

	return result
}

func nonEmpty(values []string) []string {
	out := values[:0]
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, strings.TrimSpace(value))
		}
	}

	return out
}

func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}

	return fallback
}
