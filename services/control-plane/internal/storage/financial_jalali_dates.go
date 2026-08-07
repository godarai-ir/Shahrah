package storage

import (
	"time"

	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
)

const FinancialCalendar = billingdomain.FinancialCalendar

var (
	ErrBillingReportPresetInvalid = billingdomain.ErrBillingReportPresetInvalid
	ErrBillingJalaliMonthInvalid  = billingdomain.ErrBillingJalaliMonthInvalid
	ErrBillingJalaliDateInvalid   = billingdomain.ErrBillingJalaliDateInvalid
	financialTehranLocation       = billingdomain.FinancialTehranLocation()
)

func BillingReportWindow(now time.Time, preset string) (time.Time, time.Time, string, error) {
	return billingdomain.BillingReportWindow(now, preset)
}

func JalaliMonthWindow(jalaliYear, jalaliMonth int) (time.Time, time.Time, error) {
	return billingdomain.JalaliMonthWindow(jalaliYear, jalaliMonth)
}

func ParseJalaliDateAtTehranMidday(value string) (time.Time, error) {
	return billingdomain.ParseJalaliDateAtTehranMidday(value)
}

func FormatJalaliMonthKey(year, month int) string {
	return billingdomain.FormatJalaliMonthKey(year, month)
}
