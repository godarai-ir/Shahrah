// Command xai-readiness runs a bounded, credentialed conformance smoke test
// through GodarAI. It never sends an xAI provider key to the gateway and never
// writes credentials or response bodies to its report.
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
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultSpendCeilingUSD = 0.25

type readinessCase struct {
	Name     string `json:"name"`
	Model    string `json:"model,omitempty"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"`
	HTTPCode int    `json:"http_status,omitempty"`
	Error    string `json:"error,omitempty"`
}

type readinessReport struct {
	GeneratedAt     time.Time       `json:"generated_at"`
	BaseURL         string          `json:"base_url"`
	Provider        string          `json:"provider"`
	ProviderAccount string          `json:"provider_account,omitempty"`
	SpendCeilingUSD float64         `json:"spend_ceiling_usd"`
	PlannedSpendUSD float64         `json:"planned_spend_usd"`
	Passed          int             `json:"passed"`
	Failed          int             `json:"failed"`
	Blocked         int             `json:"blocked"`
	SmokeReady      bool            `json:"smoke_ready"`
	ProductionReady bool            `json:"production_ready"`
	Cases           []readinessCase `json:"cases"`
}

type probe struct {
	name, model, method, path, body string
	wantStatus                      int
	maxCostUSD                      float64
}

func main() {
	var reportPath string
	var timeout time.Duration
	flag.StringVar(&reportPath, "report", "xai-readiness-report.json", "sanitized JSON report output")
	flag.DurationVar(&timeout, "timeout", 45*time.Second, "timeout for each GodarAI request")
	flag.Parse()

	report, err := runFromEnvironment(context.Background(), http.DefaultClient, timeout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := os.WriteFile(reportPath, encoded, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Printf("xAI readiness: passed=%d failed=%d blocked=%d report=%s\n", report.Passed, report.Failed, report.Blocked, reportPath)
	if !report.SmokeReady {
		os.Exit(1)
	}
}

func runFromEnvironment(ctx context.Context, client *http.Client, timeout time.Duration) (readinessReport, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("GODARAI_BASE_URL")), "/")
	virtualKey := strings.TrimSpace(os.Getenv("GODARAI_VIRTUAL_KEY"))
	account := strings.TrimSpace(os.Getenv("GODARAI_XAI_PROVIDER_ACCOUNT"))
	ceiling, err := spendCeiling(os.Getenv("GODARAI_XAI_MAX_SPEND_USD"))
	if err != nil {
		return readinessReport{}, err
	}
	report := readinessReport{
		GeneratedAt: time.Now().UTC(), BaseURL: baseURL, Provider: "xai",
		ProviderAccount: account, SpendCeilingUSD: ceiling,
	}
	probes := readinessProbes()
	if baseURL == "" || virtualKey == "" || account == "" {
		for _, item := range probes {
			report.Cases = append(report.Cases, readinessCase{
				Name: item.name, Model: item.model, Endpoint: item.method + " " + item.path,
				Status: "blocked", Error: "required GodarAI readiness environment is incomplete",
			})
		}
		report.Blocked = len(report.Cases)
		return report, nil
	}
	if ceiling <= 0 {
		return readinessReport{}, errors.New("GODARAI_XAI_MAX_SPEND_USD must be greater than zero")
	}
	plannedSpend := 0.0
	for _, item := range probes {
		if plannedSpend+item.maxCostUSD > ceiling {
			report.Cases = append(report.Cases, readinessCase{
				Name: item.name, Model: item.model, Endpoint: item.method + " " + item.path,
				Status: "blocked", Error: "approved readiness spend ceiling would be exceeded",
			})
			report.Blocked++
			continue
		}
		result := executeProbe(ctx, client, timeout, baseURL, virtualKey, account, item)
		plannedSpend += item.maxCostUSD
		report.PlannedSpendUSD = plannedSpend
		report.Cases = append(report.Cases, result)
		switch result.Status {
		case "passed":
			report.Passed++
		case "blocked":
			report.Blocked++
		default:
			report.Failed++
		}
	}
	report.SmokeReady = report.Failed == 0 && report.Blocked == 0
	// This bounded runner cannot certify load, regional, reconciliation,
	// privacy, dashboard, approval, or rollback gates on its own.
	report.ProductionReady = false
	return report, nil
}

func readinessProbes() []probe {
	return []probe{
		{name: "model discovery", method: http.MethodGet, path: "/v1/models", wantStatus: http.StatusOK},
		{name: "grok-4.3 chat", model: "grok-4.3", method: http.MethodPost, path: "/v1/chat/completions", body: `{"model":"grok-4.3","messages":[{"role":"user","content":"Reply with OK."}],"max_tokens":8}`, wantStatus: http.StatusOK, maxCostUSD: 0.01},
		{name: "grok-4.5 chat", model: "grok-4.5", method: http.MethodPost, path: "/v1/chat/completions", body: `{"model":"grok-4.5","messages":[{"role":"user","content":"Reply with OK."}],"max_tokens":8,"reasoning_effort":"low"}`, wantStatus: http.StatusOK, maxCostUSD: 0.01},
		{name: "grok-4.3 responses", model: "grok-4.3", method: http.MethodPost, path: "/v1/responses", body: `{"model":"grok-4.3","input":"Reply with OK.","max_output_tokens":8}`, wantStatus: http.StatusOK, maxCostUSD: 0.01},
		{name: "grok-4.5 responses", model: "grok-4.5", method: http.MethodPost, path: "/v1/responses", body: `{"model":"grok-4.5","input":"Reply with OK.","max_output_tokens":8,"reasoning":{"effort":"low"}}`, wantStatus: http.StatusOK, maxCostUSD: 0.01},
		{name: "imagine generation", model: "grok-imagine-image", method: http.MethodPost, path: "/v1/images/generations", body: `{"model":"grok-imagine-image","prompt":"A small blue circle on white","n":1,"resolution":"1k","response_format":"b64_json"}`, wantStatus: http.StatusOK, maxCostUSD: 0.02},
		{name: "imagine quality generation", model: "grok-imagine-image-quality", method: http.MethodPost, path: "/v1/images/generations", body: `{"model":"grok-imagine-image-quality","prompt":"A small blue circle on white","n":1,"resolution":"1k","response_format":"b64_json"}`, wantStatus: http.StatusOK, maxCostUSD: 0.05},
		{name: "unknown batch ownership rejection", method: http.MethodPost, path: "/v1/batches/readiness-invalid/requests", body: `{"batch_requests":[{"batch_request_id":"negative","batch_request":{"chat_get_completion":{"model":"grok-4.5","messages":[]}}}]}`, wantStatus: http.StatusNotFound},
	}
}

func executeProbe(parent context.Context, client *http.Client, timeout time.Duration, baseURL, key, account string, item probe) readinessCase {
	result := readinessCase{Name: item.name, Model: item.model, Endpoint: item.method + " " + item.path, Status: "failed"}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, item.method, baseURL+item.path, bytes.NewBufferString(item.body))
	if err != nil {
		result.Error = "request construction failed"
		return result
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-godarai-provider", "xai")
	req.Header.Set("x-godarai-provider-account", account)
	if item.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		result.Error = "request failed: " + safeError(err)
		return result
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	result.HTTPCode = resp.StatusCode
	if resp.StatusCode == item.wantStatus {
		result.Status = "passed"
		return result
	}
	result.Error = fmt.Sprintf("unexpected HTTP status: got %d want %d", resp.StatusCode, item.wantStatus)
	return result
}

func spendCeiling(raw string) (float64, error) {
	if strings.TrimSpace(raw) == "" {
		return defaultSpendCeilingUSD, nil
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || value <= 0 || value > 10 {
		return 0, errors.New("GODARAI_XAI_MAX_SPEND_USD must be a number greater than zero and no more than 10")
	}
	return value, nil
}

func safeError(err error) string {
	message := err.Error()
	for _, secret := range []string{os.Getenv("GODARAI_VIRTUAL_KEY"), os.Getenv("XAI_API_KEY")} {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[REDACTED]")
		}
	}
	return message
}
