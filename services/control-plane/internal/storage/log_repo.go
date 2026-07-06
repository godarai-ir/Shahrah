package storage

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// LogSpend writes the spend log entry.
func (s *Store) LogSpend(ctx context.Context, log platform.SpendLog) error {
	metaRaw := marshalJSONBestEffort(log.RequestMetadata)

	if _, err := s.pool.Exec(ctx, `
		INSERT INTO spend_logs (
			request_id, token_hash, tenant_id, organization_id, team_id, model, input_tokens, output_tokens,
			total_tokens, cost_usd, cost_irr, response_time_ms, status, request_metadata,
			completion_start_time, completion_end_time
		)
		VALUES ($1, $2, $3, NULLIF($4, '')::uuid, NULLIF($5, '')::uuid, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, log.RequestID, emptyToNil(log.TokenHash), nullUUID(log.TenantID), log.OrganizationID, log.TeamID, log.Model, log.InputTokens,
		log.OutputTokens, log.TotalTokens, log.CostUSD, log.CostIRR, log.ResponseTimeMS, log.Status, metaRaw, log.CompletionStart, log.CompletionEnd); err != nil {
		return err
	}

	return s.queueTokenSpendDelta(ctx, log.TokenHash, log.CostUSD)
}
