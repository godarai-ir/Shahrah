package storage

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	jalaali "github.com/jalaali/go-jalaali"
)

const FinancialCalendar = "jalali"

var (
	ErrBillingReportPresetInvalid = errors.New("billing report preset is invalid")
	ErrBillingJalaliMonthInvalid  = errors.New("billing jalali month is invalid")
	ErrBillingJalaliDateInvalid   = errors.New("billing jalali date is invalid")

	financialTehranLocation = loadFinancialTehranLocation()
)

func BillingReportWindow(now time.Time, preset string) (time.Time, time.Time, string, error) {
	localNow := now.In(financialTehranLocation)
	jy, jm, _, err := jalaali.ToJalaali(localNow.Year(), localNow.Month(), localNow.Day())
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}

	preset = strings.TrimSpace(strings.ToLower(preset))
	switch preset {
	case "current_month":
		start, err := jalaliStartOfDayUTC(jy, int(jm), 1)
		return start, now.UTC(), "current_month", err
	case "current_quarter":
		quarterMonth := ((int(jm)-1)/3)*3 + 1
		start, err := jalaliStartOfDayUTC(jy, quarterMonth, 1)
		return start, now.UTC(), "current_quarter", err
	case "current_year":
		start, err := jalaliStartOfDayUTC(jy, 1, 1)
		return start, now.UTC(), "current_year", err
	case "last_3_months":
		startYear, startMonth := shiftJalaliMonth(jy, int(jm), -2)
		start, err := jalaliStartOfDayUTC(startYear, startMonth, 1)
		return start, now.UTC(), "last_3_months", err
	case "last_6_months":
		startYear, startMonth := shiftJalaliMonth(jy, int(jm), -5)
		start, err := jalaliStartOfDayUTC(startYear, startMonth, 1)
		return start, now.UTC(), "last_6_months", err
	case "last_12_months":
		startYear, startMonth := shiftJalaliMonth(jy, int(jm), -11)
		start, err := jalaliStartOfDayUTC(startYear, startMonth, 1)
		return start, now.UTC(), "last_12_months", err
	default:
		return time.Time{}, time.Time{}, "", ErrBillingReportPresetInvalid
	}
}

func JalaliMonthWindow(jalaliYear int, jalaliMonth int) (time.Time, time.Time, error) {
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

func ParseJalaliDateAtTehranMidday(value string) (time.Time, error) {
	year, month, day, err := parseJalaliDateParts(value)
	if err != nil {
		return time.Time{}, err
	}
	if !jalaali.IsValidDate(year, month, day) {
		return time.Time{}, ErrBillingJalaliDateInvalid
	}

	gYear, gMonth, gDay, err := jalaali.ToGregorian(year, jalaali.Month(month), day)
	if err != nil {
		return time.Time{}, ErrBillingJalaliDateInvalid
	}

	return time.Date(gYear, gMonth, gDay, 12, 0, 0, 0, financialTehranLocation).UTC(), nil
}

func shiftJalaliMonth(year int, month int, delta int) (int, int) {
	totalMonths := year*12 + (month - 1) + delta
	if totalMonths < 0 {
		return 0, 1
	}

	return totalMonths / 12, totalMonths%12 + 1
}

func jalaliStartOfDayUTC(year int, month int, day int) (time.Time, error) {
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

func FormatJalaliMonthKey(year int, month int) string {
	return fmt.Sprintf("%04d-%02d", year, month)
}
