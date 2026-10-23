package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
)

// LogSpend writes the spend log entry.
func (s *SpendRepository) LogSpend(ctx context.Context, log platform.SpendLog) error {
	return s.LogSpends(ctx, []platform.SpendLog{log})
}

// LogSpends writes spend log entries in one PostgreSQL batch.
func (s *SpendRepository) LogSpends(ctx context.Context, logs []platform.SpendLog) error {
	if len(logs) == 0 {
		return nil
	}

	values := make([]string, 0, len(logs))
	args := make([]interface{}, 0, len(logs)*spendLogInsertColumnCount)
	spendDeltas := make(map[string]float64, len(logs))

	for i, log := range logs {
		values = append(values, spendLogInsertPlaceholders(i, spendLogInsertColumnCount))
		args = append(args,
			log.RequestID,
			emptyToNil(log.TokenHash),
			nullUUID(log.TenantID),
			log.TeamID,
			attributionUUID(log.UserID),
			log.Model,
			log.InputTokens,
			log.OutputTokens,
			log.TotalTokens,
			log.CostUSD,
			log.CostIRR,
			log.ResponseTimeMS,
			log.Status,
			marshalJSONBestEffort(log.RequestMetadata),
			marshalJSONBestEffort(log.UsageBreakdown),
			marshalJSONBestEffort(log.PricingSnapshot),
			log.CompletionStart,
			log.CompletionEnd,
		)
		args = append(args, spendLogCreditArgs(log.Credits)...)

		tokenHash := strings.TrimSpace(log.TokenHash)
		if tokenHash != "" && log.CostUSD > 0 {
			spendDeltas[tokenHash] += log.CostUSD
		}
	}

	if _, err := s.pool.Exec(ctx, `
		INSERT INTO spend_logs (
			request_id, token_hash, tenant_id, team_id, user_id, model, input_tokens, output_tokens,
			total_tokens, cost_usd, cost_irr, response_time_ms, status, request_metadata,
			usage_breakdown_json, pricing_snapshot_json,
			completion_start_time, completion_end_time,
			credits_mc, credits_avoided_mc, rate_card_version, product, billing_mode, credit_source
		)
		VALUES `+strings.Join(values, ", "), args...); err != nil {
		return err
	}

	return s.tokenUsage.queueTokenSpendDeltas(ctx, spendDeltas)
}

const spendLogInsertColumnCount = 24

// spendLogCreditArgs returns the credit column values, all NULL when credits were not computed.
func spendLogCreditArgs(credits *platform.SpendLogCredits) []interface{} {
	if credits == nil {
		return []interface{}{nil, nil, nil, nil, nil, nil}
	}

	return []interface{}{
		credits.Millicredits,
		credits.AvoidedMillicredits,
		credits.RateCardVersion,
		emptyToNil(credits.Product),
		emptyToNil(credits.BillingMode),
		emptyToNil(credits.Source),
	}
}

// attributionUUID keeps an attribution id only when it is a UUID. The ids are
// optional context, and a malformed one must not fail the whole spend-log batch.
func attributionUUID(value string) interface{} {
	parsed, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}

	return parsed.String()
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
