package main

import (
	"math"
	"sort"
	"time"
)

// usageRow is the credit total of one (tenant, user, month) bucket in
// centi-credits (0.01 credit), already rounded up per request.
type usageRow struct {
	TenantID     string
	UserID       string
	Month        time.Time
	CentiCredits int64
}

// plan is a commercial plan from the credit pricing PRD used for quota fit.
type plan struct {
	Name          string
	MaxUsers      int
	MonthlyCredit int64
}

// calibrationPlans lists the PRD plans ordered by user cap.
var calibrationPlans = []plan{
	{Name: "Developer", MaxUsers: 3, MonthlyCredit: 5000},
	{Name: "Team", MaxUsers: 10, MonthlyCredit: 16500},
	{Name: "Growth", MaxUsers: 30, MonthlyCredit: 54000},
	{Name: "Business", MaxUsers: 100, MonthlyCredit: 190000},
}

// distribution summarizes credits per user per month.
type distribution struct {
	Samples int
	P50     float64
	P75     float64
	P95     float64
}

// tenantCalibration is the per-tenant calibration result.
type tenantCalibration struct {
	TenantID       string
	ActiveUsers    int
	CompleteMonths int
	PerUser        distribution
	MonthlyCredits float64
	Plan           string
	PlanCredits    int64
	QuotaShare     float64
}

// calibrationReport is the full calibration result.
type calibrationReport struct {
	WindowStart time.Time
	WindowEnd   time.Time
	Months      []time.Time
	Overall     distribution
	Tenants     []tenantCalibration
}

// buildCalibrationReport keeps only complete months inside the window, then
// computes per-user percentiles and plan quota fit per tenant. Usage without a
// user (service keys) counts toward tenant totals but not per-user percentiles.
func buildCalibrationReport(rows []usageRow, windowStart, windowEnd time.Time) calibrationReport {
	report := calibrationReport{WindowStart: windowStart, WindowEnd: windowEnd}

	months := map[time.Time]struct{}{}
	byTenant := map[string][]usageRow{}

	var overall []float64

	for _, row := range rows {
		if !completeMonth(row.Month, windowStart, windowEnd) {
			continue
		}

		months[row.Month] = struct{}{}
		byTenant[row.TenantID] = append(byTenant[row.TenantID], row)

		if row.UserID != "" {
			overall = append(overall, centiToCredits(row.CentiCredits))
		}
	}

	for month := range months {
		report.Months = append(report.Months, month)
	}

	sort.Slice(report.Months, func(i, j int) bool { return report.Months[i].Before(report.Months[j]) })

	report.Overall = summarize(overall)

	for tenantID, tenantRows := range byTenant {
		report.Tenants = append(report.Tenants, calibrateTenant(tenantID, tenantRows))
	}

	sort.Slice(report.Tenants, func(i, j int) bool {
		if report.Tenants[i].QuotaShare != report.Tenants[j].QuotaShare {
			return report.Tenants[i].QuotaShare > report.Tenants[j].QuotaShare
		}

		return report.Tenants[i].TenantID < report.Tenants[j].TenantID
	})

	return report
}

func calibrateTenant(tenantID string, rows []usageRow) tenantCalibration {
	users := map[string]struct{}{}
	months := map[time.Time]struct{}{}

	var (
		perUser []float64
		total   int64
	)

	for _, row := range rows {
		months[row.Month] = struct{}{}
		total += row.CentiCredits

		if row.UserID != "" {
			users[row.UserID] = struct{}{}
			perUser = append(perUser, centiToCredits(row.CentiCredits))
		}
	}

	out := tenantCalibration{
		TenantID:       tenantID,
		ActiveUsers:    len(users),
		CompleteMonths: len(months),
		PerUser:        summarize(perUser),
	}

	if out.CompleteMonths > 0 {
		out.MonthlyCredits = centiToCredits(total) / float64(out.CompleteMonths)
	}

	if fit, ok := planForUsers(out.ActiveUsers); ok {
		out.Plan = fit.Name
		out.PlanCredits = fit.MonthlyCredit
		out.QuotaShare = out.MonthlyCredits / float64(fit.MonthlyCredit)
	} else {
		out.Plan = "Enterprise"
	}

	return out
}

// planForUsers returns the smallest PRD plan whose user cap fits the count.
func planForUsers(users int) (plan, bool) {
	for _, candidate := range calibrationPlans {
		if users <= candidate.MaxUsers {
			return candidate, true
		}
	}

	return plan{}, false
}

// completeMonth reports whether the whole month lies inside [windowStart, windowEnd).
func completeMonth(month, windowStart, windowEnd time.Time) bool {
	return !month.Before(windowStart) && !month.AddDate(0, 1, 0).After(windowEnd)
}

func summarize(values []float64) distribution {
	if len(values) == 0 {
		return distribution{}
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	return distribution{
		Samples: len(sorted),
		P50:     percentile(sorted, 50),
		P75:     percentile(sorted, 75),
		P95:     percentile(sorted, 95),
	}
}

// percentile uses the nearest-rank method on sorted values.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	rank := int(math.Ceil(p / 100 * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}

	return sorted[rank-1]
}

func centiToCredits(value int64) float64 {
	return float64(value) / 100
}
