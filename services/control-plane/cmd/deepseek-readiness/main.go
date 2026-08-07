// Command deepseek-readiness runs a bounded gateway-only DeepSeek evidence check.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	controlconfig "github.com/Delpak/shahrah/services/control-plane/internal/config/environment"
)

const (
	deepSeekReadinessEstimatedMaximumUSD = 0.001
	deepSeekReadinessMaximumCeilingUSD   = 10.0
)

var deepSeekReadinessModels = []string{"deepseek-v4-flash", "deepseek-v4-pro"}

type readinessConfig struct {
	BaseURL         string
	VirtualKey      string
	ProviderAccount string
	SpendCeilingUSD float64
	ReportPath      string
	Timeout         time.Duration
}

type readinessCheck struct {
	Name        string `json:"name"`
	Passed      bool   `json:"passed"`
	StatusCode  int    `json:"status_code,omitempty"`
	ServedModel string `json:"served_model,omitempty"`
	CacheHit    int    `json:"cache_hit_tokens,omitempty"`
	CacheMiss   int    `json:"cache_miss_tokens,omitempty"`
	Output      int    `json:"output_tokens,omitempty"`
	Error       string `json:"error,omitempty"`
}

type readinessReport struct {
	Provider        string           `json:"provider"`
	ProviderAccount string           `json:"provider_account"`
	GeneratedAt     time.Time        `json:"generated_at"`
	SpendCeilingUSD float64          `json:"spend_ceiling_usd"`
	EstimatedMaxUSD float64          `json:"estimated_maximum_usd"`
	Checks          []readinessCheck `json:"checks"`
	ProductionReady bool             `json:"production_ready"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	config, err := parseConfig(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	report := readinessReport{
		Provider: "deepseek", ProviderAccount: config.ProviderAccount,
		GeneratedAt: time.Now().UTC(), SpendCeilingUSD: config.SpendCeilingUSD,
		EstimatedMaxUSD: deepSeekReadinessEstimatedMaximumUSD,
		ProductionReady: false,
	}
	client := &http.Client{Timeout: min(config.Timeout, 30*time.Second)}

	modelsCheck := checkModels(ctx, client, config)

	report.Checks = append(report.Checks, modelsCheck)
	if modelsCheck.Passed {
		for _, model := range deepSeekReadinessModels {
			report.Checks = append(report.Checks, checkChat(ctx, client, config, model))
		}

		report.Checks = append(report.Checks, checkResponses(ctx, client, config))
	}

	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("encode readiness report: %w", err)
	}

	if containsSecret(reportBytes, config.VirtualKey) {
		return errors.New("readiness report contains a secret")
	}

	if config.ReportPath != "" {
		if err := os.WriteFile(config.ReportPath, append(reportBytes, '\n'), 0o600); err != nil {
			return fmt.Errorf("write readiness report: %w", err)
		}
	}

	if _, err := fmt.Fprintln(output, string(reportBytes)); err != nil {
		return fmt.Errorf("write readiness output: %w", err)
	}

	for _, check := range report.Checks {
		if !check.Passed {
			return errors.New("one or more DeepSeek readiness checks failed")
		}
	}

	return nil
}

func parseConfig(args []string) (readinessConfig, error) {
	set := flag.NewFlagSet("deepseek-readiness", flag.ContinueOnError)
	set.SetOutput(io.Discard)

	config := readinessConfig{}
	set.StringVar(&config.BaseURL, "base-url", controlconfig.Environment("GODARAI_BASE_URL"), "GodarAI base URL")
	set.StringVar(&config.ProviderAccount, "provider-account", controlconfig.Environment("GODARAI_DEEPSEEK_PROVIDER_ACCOUNT"), "DeepSeek provider account")
	set.StringVar(&config.ReportPath, "report", "deepseek-readiness-report.json", "sanitized JSON report path")
	set.DurationVar(&config.Timeout, "timeout", 90*time.Second, "total readiness timeout")

	if err := set.Parse(args); err != nil {
		return readinessConfig{}, err
	}

	config.VirtualKey = strings.TrimSpace(controlconfig.Environment("GODARAI_VIRTUAL_KEY"))
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.ProviderAccount = strings.TrimSpace(config.ProviderAccount)

	ceiling, err := strconv.ParseFloat(strings.TrimSpace(controlconfig.Environment("GODARAI_DEEPSEEK_MAX_SPEND_USD")), 64)
	if err != nil || ceiling < deepSeekReadinessEstimatedMaximumUSD || ceiling > deepSeekReadinessMaximumCeilingUSD {
		return readinessConfig{}, fmt.Errorf(
			"GODARAI_DEEPSEEK_MAX_SPEND_USD must be at least %.6f and no more than %.0f",
			deepSeekReadinessEstimatedMaximumUSD,
			deepSeekReadinessMaximumCeilingUSD,
		)
	}

	config.SpendCeilingUSD = ceiling

	parsed, err := url.Parse(config.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return readinessConfig{}, errors.New("GODARAI_BASE_URL is invalid")
	}

	if config.VirtualKey == "" {
		return readinessConfig{}, errors.New("GODARAI_VIRTUAL_KEY is required")
	}

	if config.ProviderAccount == "" {
		return readinessConfig{}, errors.New("GODARAI_DEEPSEEK_PROVIDER_ACCOUNT is required")
	}

	if config.Timeout <= 0 || config.Timeout > 5*time.Minute {
		return readinessConfig{}, errors.New("timeout must be greater than zero and no more than five minutes")
	}

	return config, nil
}

func checkModels(ctx context.Context, client *http.Client, config readinessConfig) readinessCheck {
	check := readinessCheck{Name: "catalog_models"}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, config.BaseURL+"/v1/models", http.NoBody)
	if err != nil {
		check.Error = "request_build_failed"

		return check
	}

	applyHeaders(req, config)

	resp, err := client.Do(req)
	if err != nil {
		check.Error = "request_failed"

		return check
	}
	defer closeReadinessBody(resp.Body, &check)

	check.StatusCode = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		check.Error = "unexpected_status"

		return check
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 4*1024*1024)).Decode(&payload) != nil {
		check.Error = "invalid_model_list"

		return check
	}

	found := make(map[string]bool, len(payload.Data))
	for _, model := range payload.Data {
		found[model.ID] = true
	}

	for _, model := range deepSeekReadinessModels {
		if !found[model] {
			check.Error = "canonical_model_missing"

			return check
		}
	}

	check.Passed = true

	return check
}

func checkChat(
	ctx context.Context,
	client *http.Client,
	config readinessConfig,
	model string,
) readinessCheck {
	check := readinessCheck{Name: "chat_" + model}
	body := map[string]interface{}{
		"model":      model,
		"messages":   []map[string]string{{"role": "user", "content": "Reply with OK."}},
		"thinking":   map[string]string{"type": "disabled"},
		"max_tokens": 16,
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		check.Error = "request_encode_failed"

		return check
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.BaseURL+"/v1/chat/completions", bytes.NewReader(encoded))
	if err != nil {
		check.Error = "request_build_failed"

		return check
	}

	applyHeaders(req, config)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		check.Error = "request_failed"

		return check
	}
	defer closeReadinessBody(resp.Body, &check)

	check.StatusCode = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		check.Error = "unexpected_status"

		return check
	}

	var payload struct {
		Model string `json:"model"`
		Usage struct {
			PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
			PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
			CompletionTokens      int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 4*1024*1024)).Decode(&payload) != nil {
		check.Error = "invalid_chat_response"

		return check
	}

	check.ServedModel = payload.Model
	check.CacheHit = payload.Usage.PromptCacheHitTokens
	check.CacheMiss = payload.Usage.PromptCacheMissTokens

	check.Output = payload.Usage.CompletionTokens
	if payload.Model != model {
		check.Error = "served_model_mismatch"

		return check
	}

	if check.CacheHit+check.CacheMiss <= 0 || check.Output <= 0 {
		check.Error = "usage_incomplete"

		return check
	}

	check.Passed = true

	return check
}

func checkResponses(
	ctx context.Context,
	client *http.Client,
	config readinessConfig,
) readinessCheck {
	const model = "deepseek-v4-flash"

	check := readinessCheck{Name: "responses_" + model}
	body := map[string]interface{}{
		"model":             model,
		"input":             "Reply with OK.",
		"reasoning":         map[string]string{"effort": "none"},
		"max_output_tokens": 16,
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		check.Error = "request_encode_failed"

		return check
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.BaseURL+"/v1/responses", bytes.NewReader(encoded))
	if err != nil {
		check.Error = "request_build_failed"

		return check
	}

	applyHeaders(req, config)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		check.Error = "request_failed"

		return check
	}
	defer closeReadinessBody(resp.Body, &check)

	check.StatusCode = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		check.Error = "unexpected_status"

		return check
	}

	var payload struct {
		Model string `json:"model"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
			InputDetails struct {
				CachedTokens int `json:"cached_tokens"`
			} `json:"input_tokens_details"`
		} `json:"usage"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 4*1024*1024)).Decode(&payload) != nil {
		check.Error = "invalid_responses_response"

		return check
	}

	check.ServedModel = payload.Model
	check.CacheHit = payload.Usage.InputDetails.CachedTokens
	check.CacheMiss = payload.Usage.InputTokens - check.CacheHit
	check.Output = payload.Usage.OutputTokens

	if payload.Model != model {
		check.Error = "served_model_mismatch"

		return check
	}

	if check.CacheHit < 0 || check.CacheMiss < 0 || check.CacheHit+check.CacheMiss <= 0 || check.Output <= 0 {
		check.Error = "usage_incomplete"

		return check
	}

	check.Passed = true

	return check
}

func applyHeaders(request *http.Request, config readinessConfig) {
	request.Header.Set("Authorization", "Bearer "+config.VirtualKey)
	request.Header.Set("x-godarai-provider", "deepseek")
	request.Header.Set("x-godarai-provider-account", config.ProviderAccount)
}

func containsSecret(report []byte, secret string) bool {
	secret = strings.TrimSpace(secret)

	return secret != "" && bytes.Contains(report, []byte(secret))
}

func closeReadinessBody(body io.Closer, check *readinessCheck) {
	if err := body.Close(); err != nil && check.Error == "" {
		check.Passed = false
		check.Error = "response_close_failed"
	}
}
