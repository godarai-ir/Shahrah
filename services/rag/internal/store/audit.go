package store

import (
	"context"
	"encoding/json"
	"fmt"
)

// RecordAudit stores a content-free security event for managed RAG administration.
func (s *Store) RecordAudit(
	ctx context.Context,
	tenantID, actorID, action, resourceType, resourceID string,
	metadata map[string]any,
) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode RAG audit metadata: %w", err)
	}

	_, err = s.metadata.Exec(ctx, `
		INSERT INTO rag_audit_events (tenant_id, actor_id, action, resource_type, resource_id, metadata)
		VALUES ($1,$2,$3,$4,$5,$6)`, tenantID, actorID, action, resourceType, resourceID, payload)
	if err != nil {
		return fmt.Errorf("record RAG audit event: %w", err)
	}

	return nil
}
