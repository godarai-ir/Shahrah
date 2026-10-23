package billing

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	jalaali "github.com/jalaali/go-jalaali"
)

// FinancialCalendar is the calendar used by financial reports.
const FinancialCalendar = "jalali"

var (
	// ErrBillingReportPresetInvalid indicates an unsupported report window.
	ErrBillingReportPresetInvalid = errors.New("billing report preset is invalid")
	// ErrBillingJalaliMonthInvalid indicates an invalid Jalali month.
	ErrBillingJalaliMonthInvalid = errors.New("billing jalali month is invalid")
	// ErrBillingJalaliDateInvalid indicates an invalid Jalali date.
	ErrBillingJalaliDateInvalid = errors.New("billing jalali date is invalid")
	financialTehranLocation     = loadFinancialTehranLocation()
)

// BillingReportWindow resolves a named financial report window.
func BillingReportWindow(now time.Time, preset string) (time.Time, time.Time, string, error) {
	localNow := now.In(financialTehranLocation)

	jy, jm, _, err := jalaali.ToJalaali(localNow.Year(), localNow.Month(), localNow.Day())
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}

	switch strings.TrimSpace(strings.ToLower(preset)) {
	case "current_month":
		start, startErr := jalaliStartOfDayUTC(jy, int(jm), 1)

		return start, now.UTC(), "current_month", startErr
	case "current_quarter":
		quarterMonth := ((int(jm)-1)/3)*3 + 1
		start, startErr := jalaliStartOfDayUTC(jy, quarterMonth, 1)

		return start, now.UTC(), "current_quarter", startErr
	case "current_year":
		start, startErr := jalaliStartOfDayUTC(jy, 1, 1)

		return start, now.UTC(), "current_year", startErr
	case "last_3_months":
		startYear, startMonth := shiftJalaliMonth(jy, int(jm), -2)
		start, startErr := jalaliStartOfDayUTC(startYear, startMonth, 1)

		return start, now.UTC(), "last_3_months", startErr
	case "last_6_months":
		startYear, startMonth := shiftJalaliMonth(jy, int(jm), -5)
		start, startErr := jalaliStartOfDayUTC(startYear, startMonth, 1)

		return start, now.UTC(), "last_6_months", startErr
	case "last_12_months":
		startYear, startMonth := shiftJalaliMonth(jy, int(jm), -11)
		start, startErr := jalaliStartOfDayUTC(startYear, startMonth, 1)

		return start, now.UTC(), "last_12_months", startErr
	default:
		return time.Time{}, time.Time{}, "", ErrBillingReportPresetInvalid
	}
}

// JalaliMonthWindow resolves the UTC bounds of a Jalali month.
func JalaliMonthWindow(jalaliYear, jalaliMonth int) (time.Time, time.Time, error) {
	if !jalaali.IsValidDate(jalaliYear, jalaliMonth, 1) {
		return time.Time{}, time.Time{}, ErrBillingJalaliMonthInvalid
	}

	start, err := jalaliStartOfDayUTC(jalaliYear, jalaliMonth, 1)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	nextYear, nextMonth := shiftJalaliMonth(jalaliYear, jalaliMonth, 1)

	end, err := jalaliStartOfDayUTC(nextYear, nextMonth, 1)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return start, end, nil
}

// ParseJalaliDateAtTehranMidday parses a Jalali date at Tehran midday.
func ParseJalaliDateAtTehranMidday(value string) (time.Time, error) {
	year, month, day, err := parseJalaliDateParts(value)
	if err != nil || !jalaali.IsValidDate(year, month, day) {
		return time.Time{}, ErrBillingJalaliDateInvalid
	}

	gYear, gMonth, gDay, err := jalaali.ToGregorian(year, jalaali.Month(month), day)
	if err != nil {
		return time.Time{}, ErrBillingJalaliDateInvalid
	}

	return time.Date(gYear, gMonth, gDay, 12, 0, 0, 0, financialTehranLocation).UTC(), nil
}

// FormatJalaliMonthKey formats a stable Jalali year-month key.
func FormatJalaliMonthKey(year, month int) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}

// AddJalaliMonths moves t by months on the Jalali calendar in Tehran, keeping
// its Tehran time of day. A day the target month lacks becomes that month's
// last day, so 31 Shahrivar plus one month is 30 Mehr and 30 Esfand of a leap
// year plus twelve months is 29 Esfand.
func AddJalaliMonths(t time.Time, months int) (time.Time, error) {
	local := t.In(financialTehranLocation)

	jy, jm, jd, err := jalaali.ToJalaali(local.Year(), local.Month(), local.Day())
	if err != nil {
		return time.Time{}, err
	}

	year, month := shiftJalaliMonth(jy, int(jm), months)

	length, err := jalaali.MonthLength(year, month)
	if err != nil {
		return time.Time{}, err
	}

	gYear, gMonth, gDay, err := jalaali.ToGregorian(year, jalaali.Month(month), min(jd, length))
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(gYear, gMonth, gDay, local.Hour(), local.Minute(), local.Second(), local.Nanosecond(), financialTehranLocation).UTC(), nil
}

// FinancialTehranLocation returns the timezone used by financial date calculations.
func FinancialTehranLocation() *time.Location {
	return financialTehranLocation
}

func shiftJalaliMonth(year, month, delta int) (int, int) {
	totalMonths := year*12 + (month - 1) + delta
	if totalMonths < 0 {
		return 0, 1
	}

	return totalMonths / 12, totalMonths%12 + 1
}

func jalaliStartOfDayUTC(year, month, day int) (time.Time, error) {
	if !jalaali.IsValidDate(year, month, day) {
		return time.Time{}, ErrBillingJalaliDateInvalid
	}

	gYear, gMonth, gDay, err := jalaali.ToGregorian(year, jalaali.Month(month), day)
	if err != nil {
		return time.Time{}, ErrBillingJalaliDateInvalid
	}

	return time.Date(gYear, gMonth, gDay, 0, 0, 0, 0, financialTehranLocation).UTC(), nil
}

func parseJalaliDateParts(value string) (int, int, int, error) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 3 {
		return 0, 0, 0, ErrBillingJalaliDateInvalid
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, ErrBillingJalaliDateInvalid
	}

	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, ErrBillingJalaliDateInvalid
	}

	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, ErrBillingJalaliDateInvalid
	}

	return year, month, day, nil
}

func loadFinancialTehranLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Tehran")
	if err == nil {
		return location
	}

	return time.FixedZone("Asia/Tehran", (3*60+30)*60)
}
