package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrEntitlementDenied is returned when the managed RAG add-on or a quota is unavailable.
var ErrEntitlementDenied = errors.New("managed RAG entitlement denied")

// RequireRAGEnabled fails closed when the add-on is not active for the tenant.
func (s *Store) RequireRAGEnabled(ctx context.Context, tenantID string) error {
	enabled, err := s.featureBool(ctx, tenantID, "rag.enabled")
	if err != nil {
		return err
	}

	if !enabled {
		return ErrEntitlementDenied
	}

	return nil
}

// RequireKnowledgeBaseCapacity checks the configured tenant KB maximum.
func (s *Store) RequireKnowledgeBaseCapacity(ctx context.Context, tenantID string) error {
	return s.requireCountBelow(ctx, tenantID, "rag.knowledge_bases.max", `
		SELECT count(*)::bigint FROM knowledge_bases WHERE tenant_id=$1 AND deleted_at IS NULL`)
}

// RequireDocumentCapacity checks both document count and total private/vector storage allowance.
func (s *Store) RequireDocumentCapacity(ctx context.Context, tenantID string, incomingBytes int64) error {
	if err := s.requireCountBelow(ctx, tenantID, "rag.documents.max", `
		SELECT count(*)::bigint FROM knowledge_base_documents WHERE tenant_id=$1 AND deleted_at IS NULL`); err != nil {
		return err
	}

	storageGiB, unlimited, err := s.featureNumber(ctx, tenantID, "rag.storage_gib")
	if err != nil {
		return err
	}

	if unlimited {
		return nil
	}

	var usedBytes int64
	if err := s.metadata.QueryRow(ctx, `
		SELECT COALESCE(sum(size_bytes),0)::bigint FROM document_versions
		WHERE tenant_id=$1 AND status <> 'deleting'`, tenantID).Scan(&usedBytes); err != nil {
		return fmt.Errorf("read RAG storage usage: %w", err)
	}

	limitBytes := int64(storageGiB * 1024 * 1024 * 1024)
	if limitBytes <= 0 || usedBytes+incomingBytes > limitBytes {
		return ErrEntitlementDenied
	}

	return nil
}

// RequireChatbotCapacity checks the configured chatbot maximum.
func (s *Store) RequireChatbotCapacity(ctx context.Context, tenantID string) error {
	return s.requireCountBelow(ctx, tenantID, "rag.chatbots.max", `
		SELECT count(*)::bigint FROM chatbots WHERE tenant_id=$1 AND deleted_at IS NULL`)
}

// RequirePublicBots verifies that public publication is included in the add-on.
func (s *Store) RequirePublicBots(ctx context.Context, tenantID string) error {
	enabled, err := s.featureBool(ctx, tenantID, "rag.public_bots.enabled")
	if err != nil {
		return err
	}

	if !enabled {
		return ErrEntitlementDenied
	}

	return nil
}

// RequireKnowledgeBasesPerBot enforces the tenant-specific limit capped by the product maximum.
func (s *Store) RequireKnowledgeBasesPerBot(ctx context.Context, tenantID string, count int) error {
	limit, unlimited, err := s.featureNumber(ctx, tenantID, "rag.max_kbs_per_bot")
	if err != nil {
		return err
	}

	if count < 1 || count > 5 || !unlimited && float64(count) > limit {
		return ErrEntitlementDenied
	}

	return nil
}

// RequireQueryCapacity enforces the current calendar-month completed query allowance.
func (s *Store) RequireQueryCapacity(ctx context.Context, tenantID string) error {
	limit, unlimited, err := s.featureNumber(ctx, tenantID, "rag.queries.monthly")
	if err != nil {
		return err
	}

	if unlimited {
		return nil
	}

	now := time.Now().UTC()
	windowStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	var used int64
	if err := s.metadata.QueryRow(ctx, `
		SELECT count(*)::bigint FROM rag_usage_events
		WHERE tenant_id=$1 AND operation='retrieval' AND status='completed' AND created_at >= $2`,
		tenantID, windowStart).Scan(&used); err != nil {
		return fmt.Errorf("read RAG query usage: %w", err)
	}

	if limit <= 0 || float64(used) >= limit {
		return ErrEntitlementDenied
	}

	return nil
}

func (s *Store) requireCountBelow(ctx context.Context, tenantID, featureKey, query string) error {
	limit, unlimited, err := s.featureNumber(ctx, tenantID, featureKey)
	if err != nil {
		return err
	}

	if unlimited {
		return nil
	}

	var used int64
	if err := s.metadata.QueryRow(ctx, query, tenantID).Scan(&used); err != nil {
		return fmt.Errorf("read %s usage: %w", featureKey, err)
	}

	if limit <= 0 || float64(used) >= limit {
		return ErrEntitlementDenied
	}

	return nil
}

func (s *Store) featureBool(ctx context.Context, tenantID, featureKey string) (bool, error) {
	value, err := s.featureValue(ctx, tenantID, featureKey)
	if err != nil {
		return false, err
	}

	var result bool
	if err := json.Unmarshal(value, &result); err != nil {
		return false, fmt.Errorf("decode boolean entitlement %s: %w", featureKey, err)
	}

	return result, nil
}

func (s *Store) featureNumber(ctx context.Context, tenantID, featureKey string) (float64, bool, error) {
	value, err := s.featureValue(ctx, tenantID, featureKey)
	if err != nil {
		return 0, false, err
	}

	if string(value) == "null" {
		return 0, true, nil
	}

	var result float64
	if err := json.Unmarshal(value, &result); err != nil {
		return 0, false, fmt.Errorf("decode numeric entitlement %s: %w", featureKey, err)
	}

	return result, false, nil
}

func (s *Store) featureValue(ctx context.Context, tenantID, featureKey string) ([]byte, error) {
	var value []byte

	err := s.metadata.QueryRow(ctx, `
		WITH candidates AS (
			SELECT fc.default_value AS value_json, 0 AS priority
			FROM feature_catalog fc WHERE fc.feature_key=$2
			UNION ALL
			SELECT pe.value_json, 1
			FROM tenant_subscriptions ts
			JOIN plan_entitlements pe ON pe.version_id=ts.version_id AND pe.feature_key=$2
			WHERE ts.tenant_id=$1 AND ts.status IN ('active','past_due')
			UNION ALL
			SELECT o.value_json, 2
			FROM tenant_entitlement_overrides o
			WHERE o.tenant_id=$1 AND o.feature_key=$2 AND o.active=true
			  AND (o.expires_at IS NULL OR o.expires_at > now())
		)
		SELECT value_json FROM candidates ORDER BY priority DESC LIMIT 1`, tenantID, featureKey).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEntitlementDenied
	}

	if err != nil {
		return nil, fmt.Errorf("resolve entitlement %s: %w", featureKey, err)
	}

	return value, nil
}
