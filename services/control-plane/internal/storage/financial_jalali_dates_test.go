package storage

import (
	"testing"
	"time"

	jalaali "github.com/jalaali/go-jalaali"
)

func TestBillingReportWindowUsesJalaliBoundaries(t *testing.T) {
	now := time.Date(2026, time.July, 8, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		name          string
		preset        string
		expectedYear  int
		expectedMonth int
		expectedDay   int
	}{
		{name: "current month", preset: "current_month", expectedYear: 1405, expectedMonth: 4, expectedDay: 1},
		{name: "current quarter", preset: "current_quarter", expectedYear: 1405, expectedMonth: 4, expectedDay: 1},
		{name: "current year", preset: "current_year", expectedYear: 1405, expectedMonth: 1, expectedDay: 1},
		{name: "last 3 months", preset: "last_3_months", expectedYear: 1405, expectedMonth: 2, expectedDay: 1},
		{name: "last 6 months", preset: "last_6_months", expectedYear: 1404, expectedMonth: 11, expectedDay: 1},
		{name: "last 12 months", preset: "last_12_months", expectedYear: 1404, expectedMonth: 5, expectedDay: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			start, end, normalized, err := BillingReportWindow(now, tc.preset)
			if err != nil {
				t.Fatalf("BillingReportWindow returned error: %v", err)
			}
			if normalized != tc.preset {
				t.Fatalf("expected normalized preset %q, got %q", tc.preset, normalized)
			}
			if !start.Equal(mustStartOfJalaliDay(t, tc.expectedYear, tc.expectedMonth, tc.expectedDay)) {
				t.Fatalf("unexpected start: got %s", start.Format(time.RFC3339))
			}
			if !end.Equal(now.UTC()) {
				t.Fatalf("expected end to match now UTC, got %s", end.Format(time.RFC3339))
			}
		})
	}
}

func TestBillingReportWindowUsesTehranDateNearMidnight(t *testing.T) {
	now := time.Date(2026, time.June, 21, 20, 35, 0, 0, time.UTC)

	start, _, _, err := BillingReportWindow(now, "current_month")
	if err != nil {
		t.Fatalf("BillingReportWindow returned error: %v", err)
	}

	expected := mustStartOfJalaliDay(t, 1405, 4, 1)
	if !start.Equal(expected) {
		t.Fatalf("expected Tehran-local month boundary %s, got %s", expected.Format(time.RFC3339), start.Format(time.RFC3339))
	}
}

func TestBillingReportWindowRejectsInvalidPreset(t *testing.T) {
	_, _, _, err := BillingReportWindow(time.Now().UTC(), "3m")
	if err != ErrBillingReportPresetInvalid {
		t.Fatalf("expected ErrBillingReportPresetInvalid, got %v", err)
	}
}

func TestJalaliMonthWindowSupportsLeapMonthAndValidation(t *testing.T) {
	start, end, err := JalaliMonthWindow(1403, 12)
	if err != nil {
		t.Fatalf("JalaliMonthWindow returned error: %v", err)
	}

	if duration := end.Sub(start); duration != 30*24*time.Hour {
		t.Fatalf("expected leap Esfand window to span 30 days, got %s", duration)
	}

	if _, _, err := JalaliMonthWindow(1405, 13); err != ErrBillingJalaliMonthInvalid {
		t.Fatalf("expected ErrBillingJalaliMonthInvalid, got %v", err)
	}
}

func TestParseJalaliDateAtTehranMidday(t *testing.T) {
	parsed, err := ParseJalaliDateAtTehranMidday("1405-04-17")
	if err != nil {
		t.Fatalf("ParseJalaliDateAtTehranMidday returned error: %v", err)
	}

	expected := time.Date(2026, time.July, 8, 8, 30, 0, 0, time.UTC)
	if !parsed.Equal(expected) {
		t.Fatalf("expected midday UTC %s, got %s", expected.Format(time.RFC3339), parsed.Format(time.RFC3339))
	}

	if _, err := ParseJalaliDateAtTehranMidday("1405-13-01"); err != ErrBillingJalaliDateInvalid {
		t.Fatalf("expected ErrBillingJalaliDateInvalid, got %v", err)
	}
}

func mustStartOfJalaliDay(t *testing.T, year int, month int, day int) time.Time {
	t.Helper()

	gYear, gMonth, gDay, err := jalaali.ToGregorian(year, jalaali.Month(month), day)
	if err != nil {
		t.Fatalf("failed to convert Jalali date: %v", err)
	}

	return time.Date(gYear, gMonth, gDay, 0, 0, 0, 0, financialTehranLocation).UTC()
}
