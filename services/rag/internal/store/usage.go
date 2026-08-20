package store

import (
	"context"
	"fmt"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// UsageEvent is a content-free managed RAG billing and quality event.
type UsageEvent struct {
	TenantID               string
	RequestID              string
	KnowledgeBaseIDs       []string
	ChatbotID              *string
	Operation              string
	Status                 string
	EmbeddingTokens        int
	RerankUnits            int
	OCRPages               int
	GenerationInputTokens  int
	GenerationOutputTokens int
	ProviderCostUSD        float64
	LatencyMS              int
	Refusal                bool
	CitationValid          *bool
}

// RecordUsage stores metering without document, question, chunk, or answer content.
func (s *Store) RecordUsage(ctx context.Context, event UsageEvent) error {
	_, err := s.metadata.Exec(ctx, `
		INSERT INTO rag_usage_events (
			tenant_id, request_id, knowledge_base_ids, chatbot_id, operation, status,
			embedding_tokens, rerank_units, ocr_pages, generation_input_tokens, generation_output_tokens,
			provider_cost_usd, latency_ms, refusal, citation_valid, platform_query_fee_irr
		) VALUES ($1,$2,$3::uuid[],$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,
			CASE WHEN $5='retrieval' AND $6='completed' THEN COALESCE((
				SELECT unit_price_irr FROM rag_rate_catalog
				WHERE dimension='platform_query_completed' AND active=true
				  AND effective_from <= now() AND (effective_until IS NULL OR effective_until > now())
				ORDER BY effective_from DESC LIMIT 1
			),0) ELSE 0 END)`,
		event.TenantID, event.RequestID, event.KnowledgeBaseIDs, event.ChatbotID,
		event.Operation, event.Status, event.EmbeddingTokens, event.RerankUnits,
		event.OCRPages, event.GenerationInputTokens, event.GenerationOutputTokens, event.ProviderCostUSD,
		event.LatencyMS, event.Refusal, event.CitationValid)
	if err != nil {
		return fmt.Errorf("record RAG usage: %w", err)
	}

	return nil
}

// GetChatbotAnalytics returns content-free aggregate metrics for one tenant chatbot.
func (s *Store) GetChatbotAnalytics(ctx context.Context, tenantID, chatbotID string) (*platform.ChatbotAnalytics, error) {
	var result platform.ChatbotAnalytics

	err := s.metadata.QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE operation='answer')::bigint,
		       count(*) FILTER (WHERE operation='answer' AND refusal)::bigint,
		       count(*) FILTER (WHERE operation='answer' AND citation_valid=false)::bigint,
		       COALESCE(avg(latency_ms) FILTER (WHERE operation='answer'),0)::float8,
		       COALESCE(sum(provider_cost_usd),0)::float8,
		       COALESCE(sum(platform_query_fee_irr),0)::float8
		FROM rag_usage_events
		WHERE tenant_id=$1 AND chatbot_id=$2`, tenantID, chatbotID).Scan(
		&result.QueryCount, &result.RefusalCount, &result.CitationFailureCount,
		&result.AverageLatencyMS, &result.ProviderCostUSD, &result.PlatformCostIRR,
	)
	if err != nil {
		return nil, fmt.Errorf("get chatbot analytics: %w", err)
	}

	return &result, nil
}
