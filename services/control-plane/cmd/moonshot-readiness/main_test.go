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
	t.Setenv("GODARAI_MOONSHOT_PROVIDER_ACCOUNT", "")

	report, err := runFromEnvironment(context.Background(), http.DefaultClient, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if report.Blocked != len(readinessProbes()) || report.ProductionReady {
		t.Fatalf("unexpected blocked report: %+v", report)
	}
}

func TestReadinessUsesOnlyGatewayCredentialAndRedactsReport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer gateway-secret" {
			t.Error("unexpected authorization header")

			return
		}

		if request.Header.Get("x-godarai-provider") != "moonshot" ||
			request.Header.Get("x-godarai-provider-account") != "acct_moonshot" {
			t.Error("missing Moonshot account affinity")

			return
		}

		w.Header().Set("Content-Type", "application/json")

		if _, writeErr := w.Write([]byte(`{"ok":true,"secret":"must-not-enter-report"}`)); writeErr != nil {
			t.Errorf("write response: %v", writeErr)
		}
	}))
	defer server.Close()

	t.Setenv("GODARAI_BASE_URL", server.URL)
	t.Setenv("GODARAI_VIRTUAL_KEY", "gateway-secret")
	t.Setenv("GODARAI_MOONSHOT_PROVIDER_ACCOUNT", "acct_moonshot")
	t.Setenv("GODARAI_MOONSHOT_MAX_SPEND_USD", "0.10")
	t.Setenv("MOONSHOT_API_KEY", "provider-secret")

	report, err := runFromEnvironment(context.Background(), server.Client(), time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if !report.SmokeReady || report.ProductionReady || report.Passed != len(readinessProbes()) {
		t.Fatalf("unexpected report: %+v", report)
	}

	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}

	for _, secret := range []string{"gateway-secret", "provider-secret", "must-not-enter-report"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("report leaked %q: %s", secret, encoded)
		}
	}
}

func TestReadinessHonorsApprovedSpendCeiling(t *testing.T) {
	requests := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++

		if _, writeErr := w.Write([]byte(`{"ok":true}`)); writeErr != nil {
			t.Errorf("write response: %v", writeErr)
		}
	}))
	defer server.Close()

	t.Setenv("GODARAI_BASE_URL", server.URL)
	t.Setenv("GODARAI_VIRTUAL_KEY", "gateway-secret")
	t.Setenv("GODARAI_MOONSHOT_PROVIDER_ACCOUNT", "acct_moonshot")
	t.Setenv("GODARAI_MOONSHOT_MAX_SPEND_USD", "0.025")

	report, err := runFromEnvironment(context.Background(), server.Client(), time.Second)
	if err != nil {
		t.Fatal(err)
	}

	if report.Blocked == 0 || report.SmokeReady || report.ProductionReady ||
		report.PlannedSpendUSD > report.SpendCeilingUSD {
		t.Fatalf("spend ceiling was not enforced: %+v", report)
	}

	if requests >= len(readinessProbes()) {
		t.Fatalf("expected cost-bearing probes to be blocked, got %d requests", requests)
	}
}
