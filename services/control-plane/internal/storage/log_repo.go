package storage

import (
	"context"
	"encoding/json"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// LogRequest writes the normalized request log entry.
func (s *Store) LogRequest(ctx context.Context, entry platform.RequestLogEntry) error {
	reqRaw, _ := json.Marshal(entry.RequestBody)
	respRaw, _ := json.Marshal(entry.ResponseBody)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO request_logs (request_id, token_hash, tenant_id, model, request_body, response_body, status, cost, response_time_ms)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, entry.RequestID, emptyToNil(entry.TokenHash), nullUUID(entry.TenantID), entry.Model, reqRaw, respRaw, entry.Status, entry.CostUSD, entry.ResponseTimeMS)
	return err
}

// LogSpend writes the spend log entry.
func (s *Store) LogSpend(ctx context.Context, log platform.SpendLog) error {
	metaRaw, _ := json.Marshal(log.RequestMetadata)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
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

	if log.TokenHash != "" && log.CostUSD > 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE verification_tokens
			SET spend = spend + $2, updated_at = NOW()
			WHERE token_hash = $1
		`, log.TokenHash, log.CostUSD); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
