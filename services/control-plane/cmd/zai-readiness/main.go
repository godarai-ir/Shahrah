// Command zai-readiness runs a bounded, gateway-only Z.ai evidence matrix.
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
	zaiEstimatedMaximumUSD = 0.02
	zaiMaximumCeilingUSD   = 10.0
)

type readinessConfig struct {
	BaseURL, VirtualKey, ProviderAccount        string
	ImageURL, VideoURL, DocumentURL, OCRFileURL string
	ReportPath                                  string
	SpendCeilingUSD                             float64
	Timeout                                     time.Duration
}

type readinessCheck struct {
	Name       string `json:"name"`
	Model      string `json:"model"`
	Passed     bool   `json:"passed"`
	StatusCode int    `json:"status_code,omitempty"`
	Input      int    `json:"input_tokens,omitempty"`
	Cached     int    `json:"cached_input_tokens,omitempty"`
	Output     int    `json:"output_tokens,omitempty"`
	Error      string `json:"error,omitempty"`
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

	client := &http.Client{Timeout: min(config.Timeout, 90*time.Second)}
	report := readinessReport{
		Provider: "zai", ProviderAccount: config.ProviderAccount,
		GeneratedAt: time.Now().UTC(), SpendCeilingUSD: config.SpendCeilingUSD,
		EstimatedMaxUSD: zaiEstimatedMaximumUSD,
	}

	report.Checks = append(report.Checks,
		postCheck(ctx, client, config, "chat_glm_4_7_flashx", "glm-4.7-flashx", "/v1/chat/completions", map[string]interface{}{
			"messages": []map[string]interface{}{{"role": "user", "content": "Reply with OK."}},
			"thinking": map[string]string{"type": "disabled"}, "do_sample": false, "max_tokens": 8,
		}),
		postCheck(ctx, client, config, "chat_glm_5_3_flash", "glm-5.3-flash", "/v1/chat/completions", map[string]interface{}{
			"messages": []map[string]interface{}{{"role": "user", "content": "Reply with OK."}},
			"thinking": map[string]string{"type": "enabled"}, "max_tokens": 8,
		}),
		postCheck(ctx, client, config, "tools_glm_5_2", "glm-5.2", "/v1/chat/completions", map[string]interface{}{
			"messages": []map[string]interface{}{{"role": "user", "content": "Call the ping function."}},
			"thinking": map[string]string{"type": "disabled"}, "max_tokens": 32,
			"tools": []map[string]interface{}{{"type": "function", "function": map[string]interface{}{
				"name": "ping", "description": "Return pong", "parameters": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
			}}}, "tool_choice": "auto",
		}),
		postCheck(ctx, client, config, "vision_image", "glm-5v-turbo", "/v1/chat/completions", visionBody("image_url", config.ImageURL)),
		postCheck(ctx, client, config, "vision_video", "glm-5v-turbo", "/v1/chat/completions", visionBody("video_url", config.VideoURL)),
		postCheck(ctx, client, config, "vision_document", "glm-5v-turbo", "/v1/chat/completions", visionBody("file_url", config.DocumentURL)),
		postCheck(ctx, client, config, "layout_parsing", "glm-ocr", "/v1/layout_parsing", map[string]interface{}{
			"file": config.OCRFileURL, "start_page_id": 1, "end_page_id": 1,
			"return_crop_images": false, "need_layout_visualization": false,
		}),
	)

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
			return errors.New("one or more Z.ai readiness checks failed")
		}
	}

	return nil
}

func parseConfig(args []string) (readinessConfig, error) {
	set := flag.NewFlagSet("zai-readiness", flag.ContinueOnError)
	set.SetOutput(io.Discard)

	config := readinessConfig{}
	set.StringVar(&config.BaseURL, "base-url", controlconfig.Environment("GODARAI_BASE_URL"), "GodarAI base URL")
	set.StringVar(&config.ProviderAccount, "provider-account", controlconfig.Environment("GODARAI_ZAI_PROVIDER_ACCOUNT"), "Z.ai provider account alias")
	set.StringVar(&config.ReportPath, "report", "zai-readiness-report.json", "sanitized JSON report path")
	set.DurationVar(&config.Timeout, "timeout", 4*time.Minute, "total readiness timeout")

	if err := set.Parse(args); err != nil {
		return readinessConfig{}, err
	}

	if !strings.EqualFold(strings.TrimSpace(controlconfig.Environment("GODARAI_ZAI_ALLOW_LIVE")), "true") {
		return readinessConfig{}, errors.New("GODARAI_ZAI_ALLOW_LIVE=true is required for billable live probes")
	}

	config.VirtualKey = strings.TrimSpace(controlconfig.Environment("GODARAI_VIRTUAL_KEY"))
	config.BaseURL = strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")
	config.ProviderAccount = strings.TrimSpace(config.ProviderAccount)
	config.ImageURL = strings.TrimSpace(controlconfig.Environment("GODARAI_ZAI_IMAGE_URL"))
	config.VideoURL = strings.TrimSpace(controlconfig.Environment("GODARAI_ZAI_VIDEO_URL"))
	config.DocumentURL = strings.TrimSpace(controlconfig.Environment("GODARAI_ZAI_DOCUMENT_URL"))
	config.OCRFileURL = strings.TrimSpace(controlconfig.Environment("GODARAI_ZAI_OCR_FILE_URL"))

	ceiling, err := strconv.ParseFloat(strings.TrimSpace(controlconfig.Environment("GODARAI_ZAI_MAX_SPEND_USD")), 64)
	if err != nil || ceiling < zaiEstimatedMaximumUSD || ceiling > zaiMaximumCeilingUSD {
		return readinessConfig{}, fmt.Errorf("GODARAI_ZAI_MAX_SPEND_USD must be at least %.2f and no more than %.0f", zaiEstimatedMaximumUSD, zaiMaximumCeilingUSD)
	}

	config.SpendCeilingUSD = ceiling

	parsed, err := url.Parse(config.BaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return readinessConfig{}, errors.New("GODARAI_BASE_URL is invalid")
	}

	if config.VirtualKey == "" || config.ProviderAccount == "" {
		return readinessConfig{}, errors.New("GODARAI_VIRTUAL_KEY and GODARAI_ZAI_PROVIDER_ACCOUNT are required")
	}

	for name, value := range map[string]string{
		"GODARAI_ZAI_IMAGE_URL": config.ImageURL, "GODARAI_ZAI_VIDEO_URL": config.VideoURL,
		"GODARAI_ZAI_DOCUMENT_URL": config.DocumentURL, "GODARAI_ZAI_OCR_FILE_URL": config.OCRFileURL,
	} {
		fixture, fixtureErr := url.Parse(value)
		if fixtureErr != nil || fixture.Scheme != "https" || fixture.Host == "" {
			return readinessConfig{}, fmt.Errorf("%s must be an explicit HTTPS fixture URL", name)
		}
	}

	if config.Timeout <= 0 || config.Timeout > 5*time.Minute {
		return readinessConfig{}, errors.New("timeout must be greater than zero and no more than five minutes")
	}

	return config, nil
}

func visionBody(kind, fixture string) map[string]interface{} {
	return map[string]interface{}{
		"messages": []map[string]interface{}{{"role": "user", "content": []map[string]interface{}{
			{"type": "text", "text": "Describe this input briefly."},
			{"type": kind, kind: map[string]string{"url": fixture}},
		}}}, "thinking": map[string]string{"type": "disabled"}, "max_tokens": 32,
	}
}

func postCheck(ctx context.Context, client *http.Client, config readinessConfig, name, model, path string, body map[string]interface{}) readinessCheck {
	check := readinessCheck{Name: name, Model: model}
	body["model"] = "zai:" + config.ProviderAccount + ":" + model

	encoded, err := json.Marshal(body)
	if err != nil {
		check.Error = "request_encode_failed"
		return check
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.BaseURL+path, bytes.NewReader(encoded))
	if err != nil {
		check.Error = "request_build_failed"
		return check
	}

	req.Header.Set("Authorization", "Bearer "+config.VirtualKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-godarai-provider", "zai")
	req.Header.Set("x-godarai-provider-account", config.ProviderAccount)

	resp, err := client.Do(req)
	if err != nil {
		check.Error = "request_failed"
		return check
	}
	defer closeBody(resp.Body, &check)

	check.StatusCode = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		check.Error = "unexpected_status"
		return check
	}

	var payload struct {
		Usage struct {
			PromptTokens        int `json:"prompt_tokens"`
			CompletionTokens    int `json:"completion_tokens"`
			PromptTokensDetails struct {
				CachedTokens int `json:"cached_tokens"`
			} `json:"prompt_tokens_details"`
		} `json:"usage"`
	}
	if json.NewDecoder(io.LimitReader(resp.Body, 4*1024*1024)).Decode(&payload) != nil {
		check.Error = "invalid_response"
		return check
	}

	check.Input, check.Cached, check.Output = payload.Usage.PromptTokens, payload.Usage.PromptTokensDetails.CachedTokens, payload.Usage.CompletionTokens
	if check.Input <= 0 || check.Output <= 0 || check.Cached < 0 || check.Cached > check.Input {
		check.Error = "usage_incomplete"
		return check
	}

	check.Passed = true

	return check
}

func containsSecret(report []byte, secret string) bool {
	return secret != "" && bytes.Contains(report, []byte(secret))
}

func closeBody(body io.Closer, check *readinessCheck) {
	if err := body.Close(); err != nil && check.Error == "" {
		check.Passed, check.Error = false, "response_close_failed"
	}
}
