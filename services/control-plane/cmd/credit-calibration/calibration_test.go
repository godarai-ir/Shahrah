package main

import (
	"strings"
	"testing"
	"time"
)

func month(year int, m time.Month) time.Time {
	return time.Date(year, m, 1, 0, 0, 0, 0, time.UTC)
}

func TestBuildCalibrationReportUsesCompleteMonthsOnly(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	windowEnd := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)

	rows := []usageRow{
		// July is partial (window starts on the 14th) and October is the current month.
		{TenantID: "t1", UserID: "u1", Month: month(2026, time.July), CentiCredits: 999_999_00},
		{TenantID: "t1", UserID: "u1", Month: month(2026, time.October), CentiCredits: 999_999_00},
		{TenantID: "t1", UserID: "u1", Month: month(2026, time.August), CentiCredits: 1000_00},
		{TenantID: "t1", UserID: "u2", Month: month(2026, time.August), CentiCredits: 3000_00},
		{TenantID: "t1", UserID: "u1", Month: month(2026, time.September), CentiCredits: 2000_00},
		{TenantID: "t1", UserID: "", Month: month(2026, time.September), CentiCredits: 500_00},
	}

	report := buildCalibrationReport(rows, windowStart, windowEnd)

	if len(report.Months) != 2 {
		t.Fatalf("months = %v, want August and September only", report.Months)
	}

	if report.Overall.Samples != 3 || report.Overall.P50 != 2000 || report.Overall.P95 != 3000 {
		t.Fatalf("overall = %+v, want 3 samples, p50 2000, p95 3000", report.Overall)
	}

	if len(report.Tenants) != 1 {
		t.Fatalf("tenants = %+v", report.Tenants)
	}

	tenant := report.Tenants[0]
	if tenant.ActiveUsers != 2 || tenant.CompleteMonths != 2 {
		t.Fatalf("tenant = %+v, want 2 users over 2 months", tenant)
	}

	// (1000 + 3000 + 2000 + 500 service) / 2 months = 3250 credits per month on Developer (5000).
	if tenant.MonthlyCredits != 3250 || tenant.Plan != "Developer" || tenant.QuotaShare != 0.65 {
		t.Fatalf("tenant fit = %+v, want 3250 credits/month = 65%% of Developer", tenant)
	}
}

func TestPlanForUsers(t *testing.T) {
	t.Parallel()

	tests := map[int]string{0: "Developer", 3: "Developer", 4: "Team", 30: "Growth", 100: "Business"}
	for users, want := range tests {
		got, ok := planForUsers(users)
		if !ok || got.Name != want {
			t.Fatalf("planForUsers(%d) = %v %v, want %s", users, got, ok, want)
		}
	}

	if _, ok := planForUsers(101); ok {
		t.Fatal("planForUsers(101) should fall through to Enterprise")
	}
}

func TestPercentileNearestRank(t *testing.T) {
	t.Parallel()

	sorted := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	if got := percentile(sorted, 50); got != 5 {
		t.Fatalf("p50 = %v, want 5", got)
	}

	if got := percentile(sorted, 95); got != 10 {
		t.Fatalf("p95 = %v, want 10", got)
	}

	if got := percentile(nil, 50); got != 0 {
		t.Fatalf("empty p50 = %v, want 0", got)
	}
}

func TestRenderMarkdownMarksEnterpriseShare(t *testing.T) {
	t.Parallel()

	report := calibrationReport{Tenants: []tenantCalibration{{TenantID: "big", ActiveUsers: 150, Plan: "Enterprise"}}}

	out := renderMarkdown(report)
	if !strings.Contains(out, "| `big` | 150 |") || !strings.Contains(out, "| Enterprise | n/a |") {
		t.Fatalf("markdown missing enterprise row:\n%s", out)
	}
}
