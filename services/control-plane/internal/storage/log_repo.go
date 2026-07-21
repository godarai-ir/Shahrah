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
		)
	}

	if _, err := s.pool.Exec(ctx, `
		INSERT INTO spend_logs (
			request_id, token_hash, tenant_id, organization_id, team_id, model, input_tokens, output_tokens,
			total_tokens, cost_usd, cost_irr, response_time_ms, status, request_metadata,
			completion_start_time, completion_end_time
		)
		VALUES `+strings.Join(values, ", "), args...); err != nil {
		return err
	}

	for _, log := range logs {
		if err := s.queueTokenSpendDelta(ctx, log.TokenHash, log.CostUSD); err != nil {
			return err
		}
	}

	return nil
}

const spendLogInsertColumnCount = 16

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
