package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/jackc/pgx/v5"
)

// CreateFileDocument registers an already-scanned private object for asynchronous parsing.
func (s *Store) CreateFileDocument(
	ctx context.Context,
	tenantID, knowledgeBaseID, actor, title, mediaType, objectKey, contentSHA256 string,
	sizeBytes int64,
) (*platform.KnowledgeBaseDocument, error) {
	return s.createDocumentVersion(ctx, createDocumentInput{
		TenantID:        tenantID,
		KnowledgeBaseID: knowledgeBaseID,
		Actor:           actor,
		Title:           title,
		SourceType:      "file",
		MediaType:       mediaType,
		ObjectKey:       &objectKey,
		ContentSHA256:   contentSHA256,
		SizeBytes:       sizeBytes,
	})
}

// ListDocuments lists active logical documents and their processing or active version state.
func (s *Store) ListDocuments(ctx context.Context, tenantID, knowledgeBaseID string, limit, offset int) (*platform.DocumentList, error) {
	rows, err := s.metadata.Query(ctx, `
		SELECT d.document_id::text, d.knowledge_base_id::text, d.title, d.source_type, d.media_type,
		       COALESCE(pv.status, av.status, 'queued'), d.active_version_id::text, d.processing_version_id::text,
		       COALESCE(pv.size_bytes, av.size_bytes, 0), COALESCE(pv.page_count, av.page_count),
		       COALESCE(pv.chunk_count, av.chunk_count, 0), COALESCE(pv.detected_language, av.detected_language, ''),
		       COALESCE(pv.failure_code, av.failure_code, ''), COALESCE(pv.failure_message, av.failure_message, ''),
		       d.created_at, d.updated_at, count(*) OVER()::int
		FROM knowledge_base_documents d
		LEFT JOIN document_versions pv ON pv.tenant_id = d.tenant_id AND pv.version_id = d.processing_version_id
		LEFT JOIN document_versions av ON av.tenant_id = d.tenant_id AND av.version_id = d.active_version_id
		WHERE d.tenant_id = $1 AND d.knowledge_base_id = $2 AND d.deleted_at IS NULL
		ORDER BY d.created_at DESC
		LIMIT $3 OFFSET $4`, tenantID, knowledgeBaseID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	result := &platform.DocumentList{Items: []platform.KnowledgeBaseDocument{}, Limit: limit, Offset: offset}

	for rows.Next() {
		var item platform.KnowledgeBaseDocument
		if err := rows.Scan(
			&item.ID, &item.KnowledgeBaseID, &item.Title, &item.SourceType, &item.MediaType,
			&item.Status, &item.ActiveVersionID, &item.ProcessingVersionID, &item.SizeBytes,
			&item.PageCount, &item.ChunkCount, &item.DetectedLanguage, &item.FailureCode, &item.FailureMessage,
			&item.CreatedAt, &item.UpdatedAt, &result.Total,
		); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}

		result.Items = append(result.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate documents: %w", err)
	}

	return result, nil
}

// GetDocument returns one tenant-scoped document.
func (s *Store) GetDocument(ctx context.Context, tenantID, knowledgeBaseID, documentID string) (*platform.KnowledgeBaseDocument, error) {
	var item platform.KnowledgeBaseDocument

	err := s.metadata.QueryRow(ctx, `
		SELECT d.document_id::text, d.knowledge_base_id::text, d.title, d.source_type, d.media_type,
		       COALESCE(pv.status, av.status, 'queued'), d.active_version_id::text, d.processing_version_id::text,
		       COALESCE(pv.size_bytes, av.size_bytes, 0), COALESCE(pv.page_count, av.page_count),
		       COALESCE(pv.chunk_count, av.chunk_count, 0), COALESCE(pv.detected_language, av.detected_language, ''),
		       COALESCE(pv.failure_code, av.failure_code, ''), COALESCE(pv.failure_message, av.failure_message, ''),
		       d.created_at, d.updated_at
		FROM knowledge_base_documents d
		LEFT JOIN document_versions pv ON pv.tenant_id = d.tenant_id AND pv.version_id = d.processing_version_id
		LEFT JOIN document_versions av ON av.tenant_id = d.tenant_id AND av.version_id = d.active_version_id
		WHERE d.tenant_id = $1 AND d.knowledge_base_id = $2 AND d.document_id = $3 AND d.deleted_at IS NULL`,
		tenantID, knowledgeBaseID, documentID).Scan(
		&item.ID, &item.KnowledgeBaseID, &item.Title, &item.SourceType, &item.MediaType,
		&item.Status, &item.ActiveVersionID, &item.ProcessingVersionID, &item.SizeBytes,
		&item.PageCount, &item.ChunkCount, &item.DetectedLanguage, &item.FailureCode, &item.FailureMessage,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}

	return &item, nil
}

// DeleteDocument immediately removes a document from retrieval and marks metadata for purge.
func (s *Store) DeleteDocument(ctx context.Context, tenantID, knowledgeBaseID, documentID string) error {
	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete document: %w", err)
	}
	defer rollback(ctx, tx)

	command, err := tx.Exec(ctx, `
		UPDATE knowledge_base_documents SET deleted_at = now(), active_version_id = NULL,
		       processing_version_id = NULL, updated_at = now()
		WHERE tenant_id = $1 AND knowledge_base_id = $2 AND document_id = $3 AND deleted_at IS NULL`,
		tenantID, knowledgeBaseID, documentID)
	if err != nil {
		return fmt.Errorf("hide document: %w", err)
	}

	if command.RowsAffected() == 0 {
		return ErrNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE rag_ingestion_jobs SET status='canceled', lease_owner=NULL, lease_expires_at=NULL, updated_at=now()
		WHERE tenant_id=$1 AND document_id=$2 AND status NOT IN ('completed','failed','canceled')`, tenantID, documentID); err != nil {
		return fmt.Errorf("cancel document ingestion: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO rag_purge_jobs (tenant_id, object_key)
		SELECT DISTINCT tenant_id, object_key FROM document_versions
		WHERE tenant_id=$1 AND document_id=$2 AND object_key IS NOT NULL
		ON CONFLICT (tenant_id, object_key) DO UPDATE
		SET status='queued', attempt=0, lease_owner=NULL, lease_expires_at=NULL,
		    next_attempt_at=now(), error_code=NULL, updated_at=now()`, tenantID, documentID); err != nil {
		return fmt.Errorf("queue document object purge: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO rag_purge_jobs (tenant_id, object_key)
		VALUES ($1, $2)
		ON CONFLICT (tenant_id, object_key) DO UPDATE
		SET status='queued', attempt=0, lease_owner=NULL, lease_expires_at=NULL,
		    next_attempt_at=now(), error_code=NULL, updated_at=now()`,
		tenantID, DocumentVectorPurgeObjectKey(documentID)); err != nil {
		return fmt.Errorf("queue document vector purge: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE document_versions SET status = 'deleting', updated_at = now()
		WHERE tenant_id = $1 AND knowledge_base_id = $2 AND document_id = $3`,
		tenantID, knowledgeBaseID, documentID); err != nil {
		return fmt.Errorf("mark document versions deleting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete document: %w", err)
	}

	return nil
}

// RetryDocument creates a fresh durable job for the current failed processing version.
func (s *Store) RetryDocument(ctx context.Context, tenantID, knowledgeBaseID, documentID string) error {
	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin retry document: %w", err)
	}
	defer rollback(ctx, tx)

	var versionID string

	err = tx.QueryRow(ctx, `
		SELECT v.version_id::text
		FROM knowledge_base_documents d
		JOIN document_versions v ON v.tenant_id = d.tenant_id AND v.version_id = d.processing_version_id
		WHERE d.tenant_id = $1 AND d.knowledge_base_id = $2 AND d.document_id = $3
		  AND d.deleted_at IS NULL AND v.status = 'failed'
		FOR UPDATE`, tenantID, knowledgeBaseID, documentID).Scan(&versionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	if err != nil {
		return fmt.Errorf("load failed document version: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE document_versions
		SET status = 'queued', failure_code = NULL, failure_message = NULL, updated_at = now()
		WHERE tenant_id = $1 AND version_id = $2`, tenantID, versionID); err != nil {
		return fmt.Errorf("reset document version: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO rag_ingestion_jobs (
			tenant_id, knowledge_base_id, document_id, version_id, idempotency_key
		) VALUES ($1, $2, $3, $4, $5)`,
		tenantID, knowledgeBaseID, documentID, versionID, "retry:"+versionID+":"+time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("queue document retry: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit document retry: %w", err)
	}

	return nil
}
