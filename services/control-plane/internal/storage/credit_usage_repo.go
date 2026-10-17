package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
)

const creditUsageGroupLimit = 10

// creditUsageGroupExpressions are trusted, constant ClickHouse group expressions.
var creditUsageGroupExpressions = map[string]string{
	"model":   "model_name",
	"team":    "arrayJoin(if(empty(team_ids), [''], team_ids))",
	"user":    "ifNull(user_email, '')",
	"product": "if(credit_product = '', 'gateway', credit_product)",
}

// GetCreditUsageSummary aggregates credits from model events. It reports
// Available=false until ClickHouse metrics and the credit columns are present.
func (s *MetricsRepository) GetCreditUsageSummary(
	ctx context.Context,
	tenantID string,
	start time.Time,
	end time.Time,
) (*platform.CreditUsageSummary, error) {
	out := &platform.CreditUsageSummary{
		PeriodStart:       start,
		PeriodEnd:         end,
		TotalCredits:      credits.Format(0),
		CacheSavedCredits: credits.Format(0),
		ByModel:           []platform.CreditUsageItem{},
		ByTeam:            []platform.CreditUsageItem{},
		ByUser:            []platform.CreditUsageItem{},
		ByProduct:         []platform.CreditUsageItem{},
	}

	if !s.clickHouseMetricsReady() || !s.clickhouseMetricCredits {
		return out, nil
	}

	out.Available = true

	var totalMC, avoidedMC int64

	err := s.clickhouse.QueryRowContext(ctx, `
		SELECT toInt64(sum(credits_mc)), toInt64(sum(credits_avoided_mc)), toInt64(count())
		FROM gateway_model_events
		WHERE tenant_id = ? AND ts >= ? AND ts < ?
	`, tenantID, start, end).Scan(&totalMC, &avoidedMC, &out.RequestCount)
	if err != nil {
		return nil, fmt.Errorf("query credit usage totals: %w", err)
	}

	out.TotalCredits = credits.Format(credits.Millicredits(totalMC))
	out.CacheSavedCredits = credits.Format(credits.Millicredits(avoidedMC))

	if out.ByModel, err = s.queryCreditUsageGroup(ctx, "model", tenantID, start, end); err != nil {
		return nil, err
	}

	if out.ByTeam, err = s.queryCreditUsageGroup(ctx, "team", tenantID, start, end); err != nil {
		return nil, err
	}

	if out.ByUser, err = s.queryCreditUsageGroup(ctx, "user", tenantID, start, end); err != nil {
		return nil, err
	}

	if out.ByProduct, err = s.queryCreditUsageGroup(ctx, "product", tenantID, start, end); err != nil {
		return nil, err
	}

	if err := s.labelCreditUsageTeams(ctx, tenantID, out.ByTeam); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *MetricsRepository) queryCreditUsageGroup(
	ctx context.Context,
	group string,
	tenantID string,
	start time.Time,
	end time.Time,
) ([]platform.CreditUsageItem, error) {
	expression, ok := creditUsageGroupExpressions[group]
	if !ok {
		return nil, fmt.Errorf("unknown credit usage group %q", group)
	}

	rows, err := s.clickhouse.QueryContext(ctx, fmt.Sprintf(`
		SELECT %s AS group_key, toInt64(sum(credits_mc)) AS credits_total, toInt64(count()) AS request_total
		FROM gateway_model_events
		WHERE tenant_id = ? AND ts >= ? AND ts < ?
		GROUP BY group_key
		ORDER BY credits_total DESC, group_key ASC
		LIMIT %d
	`, expression, creditUsageGroupLimit), tenantID, start, end)
	if err != nil {
		return nil, fmt.Errorf("query credit usage by %s: %w", group, err)
	}
	defer closeBestEffort(rows)

	items := make([]platform.CreditUsageItem, 0, creditUsageGroupLimit)

	for rows.Next() {
		var (
			item      platform.CreditUsageItem
			creditsMC int64
		)

		if err := rows.Scan(&item.Key, &creditsMC, &item.RequestCount); err != nil {
			return nil, fmt.Errorf("scan credit usage by %s: %w", group, err)
		}

		item.Credits = credits.Format(credits.Millicredits(creditsMC))
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credit usage by %s: %w", group, err)
	}

	return items, nil
}

// labelCreditUsageTeams resolves team display names for the team breakdown.
func (s *MetricsRepository) labelCreditUsageTeams(ctx context.Context, tenantID string, items []platform.CreditUsageItem) error {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if item.Key != "" {
			ids = append(ids, item.Key)
		}
	}

	if len(ids) == 0 || s.pool == nil {
		return nil
	}

	rows, err := s.pool.Query(ctx, `
		SELECT team_id::text, COALESCE(team_alias, '')
		FROM teams
		WHERE tenant_id::text = $1 AND team_id::text = ANY($2)
	`, tenantID, ids)
	if err != nil {
		return fmt.Errorf("load credit usage team names: %w", err)
	}
	defer rows.Close()

	names := make(map[string]string, len(ids))

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return fmt.Errorf("scan credit usage team name: %w", err)
		}

		names[id] = name
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate credit usage team names: %w", err)
	}

	for i := range items {
		items[i].Label = names[items[i].Key]
	}

	return nil
}
