package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// LogSpend writes the spend log entry.
func (s *Store) LogSpend(ctx context.Context, log platform.SpendLog) error {
	return s.LogSpends(ctx, []platform.SpendLog{log})
}

// LogSpends writes spend log entries in one PostgreSQL batch.
func (s *Store) LogSpends(ctx context.Context, logs []platform.SpendLog) error {
	if len(logs) == 0 {
		return nil
	}

	values := make([]string, 0, len(logs))

	args := make([]interface{}, 0, len(logs)*spendLogInsertColumnCount)
	for i, log := range logs {
		values = append(values, spendLogInsertPlaceholders(i, spendLogInsertColumnCount))
		args = append(args,
			log.RequestID,
			emptyToNil(log.TokenHash),
			nullUUID(log.TenantID),
			log.OrganizationID,
			log.TeamID,
			log.Model,
			log.InputTokens,
			log.OutputTokens,
			log.TotalTokens,
			log.CostUSD,
			log.CostIRR,
			log.ResponseTimeMS,
			log.Status,
			marshalJSONBestEffort(log.RequestMetadata),
			log.CompletionStart,
			log.CompletionEnd,
			marshalJSONBestEffort(log.UsageMeasures),
			nullableJSON(log.RatingSnapshot),
			marshalJSONBestEffort(log.CostComponents),
			emptyToNil(log.PricingSource),
			emptyToNil(log.RatingStatus),
			emptyToNil(log.VideoSeconds),
			emptyToNil(log.VideoResolution),
			nullPositiveInt(log.OutputCount),
		)
	}

	rows, err := s.pool.Query(ctx, `
		WITH inserted_spend AS (
			INSERT INTO spend_logs (
				request_id, token_hash, tenant_id, organization_id, team_id, model, input_tokens, output_tokens,
				total_tokens, cost_usd, cost_irr, response_time_ms, status, request_metadata,
				completion_start_time, completion_end_time, usage_measures, rating_snapshot,
				cost_components, pricing_source, rating_status, video_seconds, video_resolution, output_count
			)
			VALUES `+strings.Join(values, ", ")+`
			ON CONFLICT (tenant_id, request_id)
			WHERE tenant_id IS NOT NULL AND request_id IS NOT NULL
			DO NOTHING
			RETURNING COALESCE(token_hash, '') AS token_hash,
			          COALESCE(cost_usd, 0)::double precision AS cost_usd,
			          rating_snapshot
		)
		SELECT token_hash, cost_usd
		FROM inserted_spend
		WHERE rating_snapshot IS NULL
	`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Conflict replays are intentionally absent from RETURNING, so their
	// token-spend aggregation side effect cannot be applied a second time.
	insertedLegacySpends := make([]tokenSpendUpdate, 0, len(logs))
	for rows.Next() {
		var inserted tokenSpendUpdate
		if err := rows.Scan(&inserted.TokenHash, &inserted.DeltaUSD); err != nil {
			return err
		}
		insertedLegacySpends = append(insertedLegacySpends, inserted)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	for _, inserted := range insertedLegacySpends {
		if err := s.queueTokenSpendDelta(ctx, inserted.TokenHash, inserted.DeltaUSD); err != nil {
			return err
		}
	}

	return nil
}

const spendLogInsertColumnCount = 24

func nullPositiveInt(value int) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func spendLogInsertPlaceholders(rowIndex, columns int) string {
	start := rowIndex*columns + 1

	placeholders := make([]string, 0, columns)
	for i := 0; i < columns; i++ {
		placeholder := fmt.Sprintf("$%d", start+i)
		switch i {
		case 3, 4:
			placeholder = fmt.Sprintf("NULLIF(%s, '')::uuid", placeholder)
		}

		placeholders = append(placeholders, placeholder)
	}

	return "(" + strings.Join(placeholders, ", ") + ")"
}
