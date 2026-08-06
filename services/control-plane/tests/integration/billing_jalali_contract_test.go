package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"go.uber.org/zap"
)

func TestIntegration_BillingReportEndpointsUseJalaliContract(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	summaryResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/billing/reports/summary?preset=current_month", "Bearer sk-master", "")
	if summaryResp.StatusCode != http.StatusOK {
		t.Fatalf("expected summary status 200, got %d", summaryResp.StatusCode)
	}

	var summary platform.BillingUsageSummary
	decodeJSONBody(t, summaryResp, &summary)

	if summary.Calendar != "jalali" {
		t.Fatalf("expected calendar jalali, got %q", summary.Calendar)
	}

	if summary.Preset != "current_month" {
		t.Fatalf("expected preset current_month, got %q", summary.Preset)
	}

	legacySummaryResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/billing/reports/summary?range=current_month", "Bearer sk-master", "")
	if legacySummaryResp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected legacy summary param to return 422, got %d", legacySummaryResp.StatusCode)
	}

	monthlyResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/billing/reports/monthly?jalali_year=1405&jalali_month=4", "Bearer sk-master", "")
	if monthlyResp.StatusCode != http.StatusOK {
		t.Fatalf("expected monthly status 200, got %d", monthlyResp.StatusCode)
	}

	var monthly platform.MonthlyUsageDetail
	decodeJSONBody(t, monthlyResp, &monthly)

	if monthly.Calendar != "jalali" {
		t.Fatalf("expected monthly calendar jalali, got %q", monthly.Calendar)
	}

	if monthly.JalaliYear != 1405 || monthly.JalaliMonth != 4 {
		t.Fatalf("expected monthly Jalali selection 1405/4, got %d/%d", monthly.JalaliYear, monthly.JalaliMonth)
	}

	legacyMonthlyResp := performJSONRequest(t, app, http.MethodGet, "http://localhost/api/v1/billing/reports/monthly?month=2026-07", "Bearer sk-master", "")
	if legacyMonthlyResp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected legacy monthly param to return 422, got %d", legacyMonthlyResp.StatusCode)
	}
}

func TestIntegration_CreateBankTransferAcceptsJalaliDate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	memCache, _ := testcache.New(t, logger)

	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer provider.Close()

	app, _, _, _ := newIntegrationApp(t, logger, memCache, provider.URL)

	resp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/wallet/bank-transfers", "Bearer sk-master", `{
		"amount_irr": 250000,
		"tracking_number": "ABC-123",
		"bank_transaction_date_jalali": "1405-04-17"
	}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create bank transfer status 201, got %d", resp.StatusCode)
	}

	var item platform.BankTransferRequest
	decodeJSONBody(t, resp, &item)

	expected := time.Date(2026, time.July, 8, 8, 30, 0, 0, time.UTC)
	if !item.BankTransactionAt.Equal(expected) {
		t.Fatalf("expected normalized bank transaction time %s, got %s", expected.Format(time.RFC3339), item.BankTransactionAt.Format(time.RFC3339))
	}

	invalidResp := performJSONRequest(t, app, http.MethodPost, "http://localhost/api/v1/wallet/bank-transfers", "Bearer sk-master", `{
		"amount_irr": 250000,
		"tracking_number": "ABC-123",
		"bank_transaction_date_jalali": "1405-13-17"
	}`)
	if invalidResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected invalid Jalali date to return 400, got %d", invalidResp.StatusCode)
	}
}
