// Command anthropic-readiness runs the live, credential-backed release matrix
// against GodarAI's Anthropic-native API. It never writes the virtual key,
// Anthropic Admin key, prompts, or upstream opaque identifiers to its report.
package main

import (
	"bytes"
	"context"
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
	"anthropic/claude-haiku-4-5-20251001",
	"anthropic/claude-opus-4-1",
	"anthropic/claude-opus-4-1-20250805",
	"anthropic/claude-opus-4-5",
	"anthropic/claude-opus-4-5-20251101",
	"anthropic/claude-opus-4-6",
	"anthropic/claude-opus-4-7",
	"anthropic/claude-sonnet-4-5",
	"anthropic/claude-sonnet-4-5-20250929",
	"anthropic/claude-sonnet-4-6",
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
	Model  string        `json:"model"`
	Checks []checkResult `json:"checks"`
}

type checkResult struct {
	Name       string `json:"name"`
	Passed     bool   `json:"passed"`
	StatusCode int    `json:"status_code,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

type reportSummary struct {
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

type rateLimitSnapshot struct {
	ObservedAt time.Time       `json:"observed_at"`
	Org        json.RawMessage `json:"organization"`
	Workspace  json.RawMessage `json:"workspace,omitempty"`
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()
	client := &http.Client{Timeout: 90 * time.Second}
	result := run(ctx, client, cfg)
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
		"Anthropic readiness report: %s (passed=%d failed=%d)\n",
		cfg.reportPath,
		result.Summary.Passed,
		result.Summary.Failed,
	)
	if result.Summary.Failed > 0 {
		os.Exit(1)
	}
}

func loadConfig() (config, error) {
	var cfg config
	var modelCSV string
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
		item := modelReport{Model: model}
		item.Checks = append(item.Checks,
			checkModel(ctx, client, cfg, model),
			checkCountTokens(ctx, client, cfg, model),
			checkMessage(ctx, client, cfg, model, false),
			checkMessage(ctx, client, cfg, model, true),
			checkStructuredOutput(ctx, client, cfg, model),
			checkThinking(ctx, client, cfg, model),
			checkOpenAIServerToolContract(ctx, client, cfg, model),
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

func checkModel(ctx context.Context, client *http.Client, cfg config, model string) checkResult {
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
	if result.Passed && !bytes.Contains(body, []byte(`"max_tokens"`)) {
		result.Passed = false
		result.Detail = "Models API response omitted max_tokens"
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

func checkMessage(ctx context.Context, client *http.Client, cfg config, model string, stream bool) checkResult {
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
	encoded, _ := json.Marshal(body)
	status, headers, responseBody, err := request(ctx, client, cfg, http.MethodPost, "/v1/messages", encoded, nil)
	result := resultFor(name, status, headers, responseBody, err)
	if result.Passed && stream &&
		(!bytes.Contains(responseBody, []byte("message_stop")) ||
			bytes.Contains(responseBody, []byte("anthropic_stream_error"))) {
		result.Passed = false
		result.Detail = "stream did not end with message_stop or contained an error"
	}
	return result
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

func checkThinking(ctx context.Context, client *http.Client, cfg config, model string) checkResult {
	thinking := map[string]interface{}{"type": "enabled", "budget_tokens": 1024}
	output := map[string]interface{}{}
	if strings.Contains(model, "-4-6") || strings.Contains(model, "-4-7") {
		thinking = map[string]interface{}{"type": "adaptive"}
		output["effort"] = "low"
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
) checkResult {
	payload := map[string]interface{}{
		"model": model, "max_tokens": 32,
		"messages": []map[string]interface{}{{"role": "user", "content": "Reply OK without using tools."}},
		"tools": []map[string]interface{}{{
			// The baseline version works across the legacy and current Claude 4
			// models in this readiness matrix. Newer versions default to
			// programmatic calling, which is not supported by every legacy model.
			"type": "web_search_20250305", "name": "web_search", "max_uses": 1,
		}},
		"tool_choice": "auto",
	}
	return jsonCheck(ctx, client, cfg, "openai.server_tool_contract", "/v1/chat/completions", payload)
}

func runResourceChecks(ctx context.Context, client *http.Client, cfg config) []checkResult {
	checks := make([]checkResult, 0, 12)
	var upload bytes.Buffer
	writer := multipart.NewWriter(&upload)
	part, _ := writer.CreateFormFile("file", "readiness.txt")
	_, _ = io.WriteString(part, "GodarAI Anthropic readiness")
	_ = writer.Close()
	extra := map[string]string{"Content-Type": writer.FormDataContentType()}
	status, headers, body, err := request(ctx, client, cfg, http.MethodPost, "/v1/files", upload.Bytes(), extra)
	uploadResult := resultFor("files.create", status, headers, body, err)
	checks = append(checks, uploadResult)
	checks = append(checks, methodCheck(ctx, client, cfg, "files.list", http.MethodGet, "/v1/files", nil))
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
			if strings.Contains(model, "opus-4-6") ||
				strings.Contains(model, "opus-4-7") ||
				strings.Contains(model, "sonnet-4-6") {
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
	batchBody, _ := json.Marshal(batch)
	status, headers, body, err = request(
		ctx, client, cfg, http.MethodPost, "/v1/messages/batches", batchBody, nil,
	)
	createResult := resultFor("batches.create", status, headers, body, err)
	checks = append(checks, createResult)
	checks = append(checks, methodCheck(
		ctx, client, cfg, "batches.list", http.MethodGet, "/v1/messages/batches", nil,
	))
	if batchID := responseID(body); createResult.Passed && batchID != "" {
		checks = append(checks, waitForBatch(ctx, client, cfg, batchID)...)
	} else if createResult.Passed {
		checks = append(checks, checkResult{
			Name: "batches.lifecycle", Detail: "create response omitted an opaque batch id",
		})
	}

	cancelBody, _ := json.Marshal(map[string]interface{}{
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
		snapshot.Workspace = workspace
		checks = append(checks, workspaceCheck)
	}
	return snapshot, checks
}

func adminGet(
	ctx context.Context,
	client *http.Client,
	cfg config,
	targetPath string,
) (json.RawMessage, checkResult) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.anthropic.com"+targetPath,
		nil,
	)
	if err != nil {
		return nil, checkResult{Name: "rate_limits.snapshot", Detail: err.Error()}
	}
	req.Header.Set("x-api-key", cfg.adminKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := client.Do(req)
	if err != nil {
		return nil, checkResult{Name: "rate_limits.snapshot", Detail: err.Error()}
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	name := "rate_limits.organization"
	if strings.Contains(targetPath, "/workspaces/") {
		name = "rate_limits.workspace"
	}
	result := resultFor(name, resp.StatusCode, resp.Header, body, readErr)
	if !json.Valid(body) {
		result.Passed = false
		result.Detail = "Rate Limits API returned invalid JSON"
		return nil, result
	}
	return append(json.RawMessage(nil), body...), result
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
		encoded, _ = json.Marshal(payload)
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
		body, _ = json.Marshal(payload)
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
	defer resp.Body.Close()
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
		RequestID:  firstHeader(headers, "request-id", "x-request-id"),
	}
	if err != nil {
		result.Detail = err.Error()
	} else if !result.Passed {
		result.Detail = sanitizedError(body)
	}
	return result
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
	_ = json.Unmarshal(body, &response)
	return strings.TrimSpace(response.ID)
}

func accumulate(summary *reportSummary, result checkResult) {
	if result.Passed {
		summary.Passed++
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
