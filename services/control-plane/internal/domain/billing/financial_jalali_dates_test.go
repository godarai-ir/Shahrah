package billing

import (
	"testing"
	"time"

	jalaali "github.com/jalaali/go-jalaali"
	"github.com/stretchr/testify/require"
)

// jalaliAt is a Tehran wall-clock time on a Jalali date, in UTC.
func jalaliAt(t *testing.T, year, month, day, hour, minute int) time.Time {
	t.Helper()

	gYear, gMonth, gDay, err := jalaali.ToGregorian(year, jalaali.Month(month), day)
	require.NoError(t, err)

	return time.Date(gYear, gMonth, gDay, hour, minute, 0, 0, FinancialTehranLocation()).UTC()
}

func TestAddJalaliMonths(t *testing.T) {
	tests := []struct {
		name   string
		from   time.Time
		months int
		want   time.Time
	}{
		{"same day of the next month", jalaliAt(t, 1405, 7, 8, 14, 30), 1, jalaliAt(t, 1405, 8, 8, 14, 30)},
		{"31 Shahrivar ends on the last day of Mehr", jalaliAt(t, 1405, 6, 31, 10, 0), 1, jalaliAt(t, 1405, 7, 30, 10, 0)},
		{"30 Bahman ends on 29 Esfand of a common year", jalaliAt(t, 1404, 11, 30, 9, 0), 1, jalaliAt(t, 1404, 12, 29, 9, 0)},
		{"30 Esfand of a leap year plus a year is 29 Esfand", jalaliAt(t, 1403, 12, 30, 9, 0), 12, jalaliAt(t, 1404, 12, 29, 9, 0)},
		{"months roll over the year", jalaliAt(t, 1405, 12, 15, 9, 0), 1, jalaliAt(t, 1406, 1, 15, 9, 0)},
		{"a year keeps the date", jalaliAt(t, 1405, 7, 30, 12, 30), 12, jalaliAt(t, 1406, 7, 30, 12, 30)},
		// 01:30 in Tehran is the previous evening in UTC: the Tehran date decides.
		{"the Tehran day counts, not the UTC day", jalaliAt(t, 1405, 8, 1, 1, 30), 1, jalaliAt(t, 1405, 9, 1, 1, 30)},
		{"negative months go back", jalaliAt(t, 1405, 1, 31, 9, 0), -1, jalaliAt(t, 1404, 12, 29, 9, 0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AddJalaliMonths(tt.from, tt.months)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
			require.Equal(t, time.UTC, got.Location())
		})
	}
}
