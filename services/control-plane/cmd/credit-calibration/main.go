// Package main produces the credit calibration report from ClickHouse model
// events: credits per user per month (median, p75, p95) and the share of the
// matching PRD plan quota each tenant would consume. It is read-only.
package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

const queryTimeout = 5 * time.Minute

// calibrationQuery sums credits per (tenant, user, Tehran month), rounding each
// request up to 0.01 credit (500 credits per USD, so 50,000 centi-credits per USD).
const calibrationQuery = `
	SELECT
		tenant_id,
		ifNull(user_id, '') AS user,
		toDateTime(toStartOfMonth(ts, 'Asia/Tehran'), 'Asia/Tehran') AS month,
		toInt64(sum(ceil(toFloat64(cost_usd) * 50000))) AS centi_credits
	FROM gateway_model_events
	WHERE ts >= ? AND ts < ? AND status_code < 400
	GROUP BY tenant_id, user, month
`

func main() {
	dsn := flag.String("dsn", os.Getenv("CLICKHOUSE_DSN"), "ClickHouse DSN (defaults to CLICKHOUSE_DSN)")
	days := flag.Int("days", 90, "look-back window in days")
	outDir := flag.String("out", "output/credit-calibration", "output directory")

	flag.Parse()

	if err := run(*dsn, *days, *outDir); err != nil {
		log.Fatal(err)
	}
}

func run(dsn string, days int, outDir string) error {
	if strings.TrimSpace(dsn) == "" {
		return errors.New("a ClickHouse DSN is required (--dsn or CLICKHOUSE_DSN)")
	}

	if days < 31 {
		return errors.New("--days must be at least 31 to contain a complete month")
	}

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return fmt.Errorf("open clickhouse: %w", err)
	}
	defer closeQuietly(db)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	end := time.Now().UTC()
	start := end.AddDate(0, 0, -days)

	rows, err := loadUsageRows(ctx, db, start, end)
	if err != nil {
		return err
	}

	report := buildCalibrationReport(rows, start, end)
	stamp := end.Format("20060102-150405")

	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	markdownPath := filepath.Join(outDir, "credit-calibration-"+stamp+".md")
	if err := os.WriteFile(markdownPath, []byte(renderMarkdown(report)), 0o600); err != nil {
		return fmt.Errorf("write markdown report: %w", err)
	}

	csvPath := filepath.Join(outDir, "credit-calibration-"+stamp+".csv")
	if err := writeCSV(csvPath, report); err != nil {
		return err
	}

	log.Printf("wrote %s and %s (%d tenants)", markdownPath, csvPath, len(report.Tenants))

	return nil
}

func loadUsageRows(ctx context.Context, db *sql.DB, start, end time.Time) ([]usageRow, error) {
	result, err := db.QueryContext(ctx, calibrationQuery, start, end)
	if err != nil {
		return nil, fmt.Errorf("query usage: %w", err)
	}
	defer closeQuietly(result)

	var rows []usageRow

	for result.Next() {
		var row usageRow
		if err := result.Scan(&row.TenantID, &row.UserID, &row.Month, &row.CentiCredits); err != nil {
			return nil, fmt.Errorf("scan usage: %w", err)
		}

		row.Month = row.Month.UTC()
		rows = append(rows, row)
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate usage: %w", err)
	}

	return rows, nil
}

func renderMarkdown(report calibrationReport) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Credit Calibration Report\n\n")
	fmt.Fprintf(&b, "Window: %s to %s (UTC). Complete months used: %d.\n\n",
		report.WindowStart.Format(time.DateOnly), report.WindowEnd.Format(time.DateOnly), len(report.Months))
	fmt.Fprintf(&b, "Credits = 500 × provider USD cost, rounded up to 0.01 credit per request. ")
	fmt.Fprintf(&b, "Months are calendar months in Asia/Tehran; partial months are excluded.\n\n")
	fmt.Fprintf(&b, "## Credits per user per month (all tenants)\n\n")
	fmt.Fprintf(&b, "| Samples | Median | p75 | p95 |\n| --- | --- | --- | --- |\n")
	fmt.Fprintf(&b, "| %d | %.2f | %.2f | %.2f |\n\n", report.Overall.Samples, report.Overall.P50, report.Overall.P75, report.Overall.P95)
	fmt.Fprintf(&b, "Target: the median tenant should use 60 to 70%% of its plan quota.\n\n")
	fmt.Fprintf(&b, "## Tenants\n\n")
	fmt.Fprintf(&b, "| Tenant | Active users | Months | Median / user | p75 / user | p95 / user | Credits / month | Plan | Quota share |\n")
	fmt.Fprintf(&b, "| --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")

	for _, tenant := range report.Tenants {
		share := "n/a"
		if tenant.PlanCredits > 0 {
			share = fmt.Sprintf("%.0f%%", tenant.QuotaShare*100)
		}

		fmt.Fprintf(&b, "| `%s` | %d | %d | %.2f | %.2f | %.2f | %.2f | %s | %s |\n",
			tenant.TenantID, tenant.ActiveUsers, tenant.CompleteMonths,
			tenant.PerUser.P50, tenant.PerUser.P75, tenant.PerUser.P95,
			tenant.MonthlyCredits, tenant.Plan, share)
	}

	return b.String()
}

func writeCSV(path string, report calibrationReport) (err error) {
	file, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create csv report: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close csv report: %w", closeErr)
		}
	}()

	writer := csv.NewWriter(file)
	records := [][]string{{
		"tenant_id", "active_users", "complete_months", "p50_per_user", "p75_per_user", "p95_per_user",
		"monthly_credits", "plan", "plan_credits", "quota_share",
	}}

	for _, tenant := range report.Tenants {
		records = append(records, []string{
			tenant.TenantID,
			strconv.Itoa(tenant.ActiveUsers),
			strconv.Itoa(tenant.CompleteMonths),
			formatCredits(tenant.PerUser.P50),
			formatCredits(tenant.PerUser.P75),
			formatCredits(tenant.PerUser.P95),
			formatCredits(tenant.MonthlyCredits),
			tenant.Plan,
			strconv.FormatInt(tenant.PlanCredits, 10),
			strconv.FormatFloat(tenant.QuotaShare, 'f', 4, 64),
		})
	}

	if err := writer.WriteAll(records); err != nil {
		return fmt.Errorf("write csv report: %w", err)
	}

	return nil
}

func formatCredits(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func closeQuietly(closer interface{ Close() error }) {
	if err := closer.Close(); err != nil {
		log.Printf("close: %v", err)
	}
}
