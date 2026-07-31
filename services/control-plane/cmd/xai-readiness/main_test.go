package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestReadinessBlocksWithoutGatewayCredentials(t *testing.T) {
	t.Setenv("GODARAI_BASE_URL", "")
	t.Setenv("GODARAI_VIRTUAL_KEY", "")
	t.Setenv("GODARAI_XAI_PROVIDER_ACCOUNT", "")
	report, err := runFromEnvironment(context.Background(), http.DefaultClient, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if report.Blocked != len(readinessProbes()) || report.ProductionReady {
		t.Fatalf("unexpected blocked report: %+v", report)
	}
}

func TestReadinessUsesOnlyGatewayCredentialAndRedactsReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer gateway-secret" {
			t.Fatalf("unexpected authorization header")
		}
		if r.Header.Get("x-godarai-provider-account") != "acct_xai" {
			t.Fatalf("missing account affinity")
		}
		if r.URL.Path == "/v1/batches/readiness-invalid/requests" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"secret":"must-not-enter-report"}`))
	}))
	defer server.Close()
	t.Setenv("GODARAI_BASE_URL", server.URL)
	t.Setenv("GODARAI_VIRTUAL_KEY", "gateway-secret")
	t.Setenv("GODARAI_XAI_PROVIDER_ACCOUNT", "acct_xai")
	t.Setenv("GODARAI_XAI_MAX_SPEND_USD", "0.25")
	t.Setenv("XAI_API_KEY", "provider-secret")
	report, err := runFromEnvironment(context.Background(), server.Client(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !report.SmokeReady || report.ProductionReady || report.Passed != len(readinessProbes()) {
		t.Fatalf("unexpected report: %+v", report)
	}
	encoded := string(mustJSON(t, report))
	for _, secret := range []string{"gateway-secret", "provider-secret", "must-not-enter-report"} {
		if contains(encoded, secret) {
			t.Fatalf("report leaked %q: %s", secret, encoded)
		}
	}
}

func TestReadinessHonorsApprovedSpendCeiling(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path == "/v1/batches/readiness-invalid/requests" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	t.Setenv("GODARAI_BASE_URL", server.URL)
	t.Setenv("GODARAI_VIRTUAL_KEY", "gateway-secret")
	t.Setenv("GODARAI_XAI_PROVIDER_ACCOUNT", "acct_xai")
	t.Setenv("GODARAI_XAI_MAX_SPEND_USD", "0.03")
	report, err := runFromEnvironment(context.Background(), server.Client(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if report.Blocked == 0 || report.SmokeReady || report.ProductionReady || report.PlannedSpendUSD > report.SpendCeilingUSD {
		t.Fatalf("spend ceiling was not enforced: %+v", report)
	}
	if requests >= len(readinessProbes()) {
		t.Fatalf("expected cost-bearing probes to be blocked, got %d requests", requests)
	}
}

func mustJSON(t *testing.T, value interface{}) []byte {
	t.Helper()
	encoded, err := jsonMarshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

var jsonMarshal = func(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func contains(value, part string) bool { return strings.Contains(value, part) }
