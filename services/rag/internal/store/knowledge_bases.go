package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned without revealing whether another tenant owns the identifier.
var ErrNotFound = errors.New("resource not found")

// CreateKnowledgeBase persists a tenant-scoped managed knowledge base.
func (s *Store) CreateKnowledgeBase(ctx context.Context, tenantID, actor string, request platform.CreateKnowledgeBaseRequest) (*platform.KnowledgeBase, error) {
	var result platform.KnowledgeBase

	err := s.metadata.QueryRow(ctx, `
		INSERT INTO knowledge_bases (tenant_id, name, description, rag_profile_revision, created_by)
		SELECT $1, $2, $3, revision, $4 FROM rag_profiles WHERE active = true
		RETURNING knowledge_base_id::text, tenant_id::text, name, description, rag_profile_revision,
		          0, 0, created_at, updated_at, deleted_at`,
		tenantID, strings.TrimSpace(request.Name), strings.TrimSpace(request.Description), actor,
	).Scan(
		&result.ID, &result.TenantID, &result.Name, &result.Description, &result.RAGProfileRevision,
		&result.DocumentCount, &result.ReadyDocumentCount, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create knowledge base: %w", err)
	}

	return &result, nil
}

// ListKnowledgeBases lists only the caller tenant's active knowledge bases.
func (s *Store) ListKnowledgeBases(ctx context.Context, tenantID string, limit, offset int) (*platform.KnowledgeBaseList, error) {
	rows, err := s.metadata.Query(ctx, `
		SELECT kb.knowledge_base_id::text, kb.tenant_id::text, kb.name, kb.description, kb.rag_profile_revision,
		       count(d.document_id)::int,
		       count(d.document_id) FILTER (WHERE v.status = 'ready')::int,
		       kb.created_at, kb.updated_at, kb.deleted_at,
		       count(*) OVER()::int
		FROM knowledge_bases kb
		LEFT JOIN knowledge_base_documents d
		  ON d.tenant_id = kb.tenant_id AND d.knowledge_base_id = kb.knowledge_base_id AND d.deleted_at IS NULL
		LEFT JOIN document_versions v
		  ON v.tenant_id = kb.tenant_id AND v.version_id = d.active_version_id
		WHERE kb.tenant_id = $1 AND kb.deleted_at IS NULL
		GROUP BY kb.knowledge_base_id
		ORDER BY kb.created_at DESC
		LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list knowledge bases: %w", err)
	}
	defer rows.Close()

	result := &platform.KnowledgeBaseList{Items: []platform.KnowledgeBase{}, Limit: limit, Offset: offset}

	for rows.Next() {
		var item platform.KnowledgeBase
		if err := rows.Scan(
			&item.ID, &item.TenantID, &item.Name, &item.Description, &item.RAGProfileRevision,
			&item.DocumentCount, &item.ReadyDocumentCount, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt,
			&result.Total,
		); err != nil {
			return nil, fmt.Errorf("scan knowledge base: %w", err)
		}

		result.Items = append(result.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate knowledge bases: %w", err)
	}

	return result, nil
}

// GetKnowledgeBase returns a single tenant-scoped knowledge base.
func (s *Store) GetKnowledgeBase(ctx context.Context, tenantID, knowledgeBaseID string) (*platform.KnowledgeBase, error) {
	var result platform.KnowledgeBase

	err := s.metadata.QueryRow(ctx, `
		SELECT kb.knowledge_base_id::text, kb.tenant_id::text, kb.name, kb.description, kb.rag_profile_revision,
		       count(d.document_id)::int,
		       count(d.document_id) FILTER (WHERE v.status = 'ready')::int,
		       kb.created_at, kb.updated_at, kb.deleted_at
		FROM knowledge_bases kb
		LEFT JOIN knowledge_base_documents d
		  ON d.tenant_id = kb.tenant_id AND d.knowledge_base_id = kb.knowledge_base_id AND d.deleted_at IS NULL
		LEFT JOIN document_versions v ON v.tenant_id = kb.tenant_id AND v.version_id = d.active_version_id
		WHERE kb.tenant_id = $1 AND kb.knowledge_base_id = $2 AND kb.deleted_at IS NULL
		GROUP BY kb.knowledge_base_id`, tenantID, knowledgeBaseID).Scan(
		&result.ID, &result.TenantID, &result.Name, &result.Description, &result.RAGProfileRevision,
		&result.DocumentCount, &result.ReadyDocumentCount, &result.CreatedAt, &result.UpdatedAt, &result.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get knowledge base: %w", err)
	}

	return &result, nil
}

// UpdateKnowledgeBase updates metadata without changing its pinned profile.
func (s *Store) UpdateKnowledgeBase(ctx context.Context, tenantID, knowledgeBaseID string, request platform.UpdateKnowledgeBaseRequest) (*platform.KnowledgeBase, error) {
	command, err := s.metadata.Exec(ctx, `
		UPDATE knowledge_bases
		SET name = COALESCE($3, name), description = COALESCE($4, description), updated_at = now()
		WHERE tenant_id = $1 AND knowledge_base_id = $2 AND deleted_at IS NULL`,
		tenantID, knowledgeBaseID, request.Name, request.Description)
	if err != nil {
		return nil, fmt.Errorf("update knowledge base: %w", err)
	}

	if command.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return s.GetKnowledgeBase(ctx, tenantID, knowledgeBaseID)
}

// DeleteKnowledgeBase immediately hides the collection and queues its documents for purge.
func (s *Store) DeleteKnowledgeBase(ctx context.Context, tenantID, knowledgeBaseID string) error {
	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete knowledge base: %w", err)
	}
	defer rollback(ctx, tx)

	command, err := tx.Exec(ctx, `
		UPDATE knowledge_bases SET deleted_at = now(), updated_at = now()
		WHERE tenant_id = $1 AND knowledge_base_id = $2 AND deleted_at IS NULL`, tenantID, knowledgeBaseID)
	if err != nil {
		return fmt.Errorf("hide knowledge base: %w", err)
	}

	if command.RowsAffected() == 0 {
		return ErrNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE knowledge_base_documents SET deleted_at = now(), updated_at = now()
		WHERE tenant_id = $1 AND knowledge_base_id = $2 AND deleted_at IS NULL`, tenantID, knowledgeBaseID); err != nil {
		return fmt.Errorf("hide knowledge base documents: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE rag_ingestion_jobs SET status='canceled', lease_owner=NULL, lease_expires_at=NULL, updated_at=now()
		WHERE tenant_id=$1 AND knowledge_base_id=$2 AND status NOT IN ('completed','failed','canceled')`, tenantID, knowledgeBaseID); err != nil {
		return fmt.Errorf("cancel knowledge base ingestion: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO rag_purge_jobs (tenant_id, object_key)
		SELECT DISTINCT tenant_id, object_key FROM document_versions
		WHERE tenant_id=$1 AND knowledge_base_id=$2 AND object_key IS NOT NULL
		ON CONFLICT (tenant_id, object_key) DO UPDATE
		SET status='queued', attempt=0, lease_owner=NULL, lease_expires_at=NULL,
		    next_attempt_at=now(), error_code=NULL, updated_at=now()`, tenantID, knowledgeBaseID); err != nil {
		return fmt.Errorf("queue knowledge base object purge: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO rag_purge_jobs (tenant_id, object_key)
		VALUES ($1, $2)
		ON CONFLICT (tenant_id, object_key) DO UPDATE
		SET status='queued', attempt=0, lease_owner=NULL, lease_expires_at=NULL,
		    next_attempt_at=now(), error_code=NULL, updated_at=now()`,
		tenantID, KnowledgeBaseVectorPurgeObjectKey(knowledgeBaseID)); err != nil {
		return fmt.Errorf("queue knowledge base vector purge: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE document_versions SET status = 'deleting', updated_at = now()
		WHERE tenant_id = $1 AND knowledge_base_id = $2 AND status <> 'deleting'`, tenantID, knowledgeBaseID); err != nil {
		return fmt.Errorf("mark knowledge base versions deleting: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete knowledge base: %w", err)
	}

	return nil
}

// CreateTextDocument creates a logical document, version, and durable ingestion job atomically.
func (s *Store) CreateTextDocument(ctx context.Context, tenantID, knowledgeBaseID, actor string, request platform.CreateTextDocumentRequest) (*platform.KnowledgeBaseDocument, error) {
	text := domain.NormalizePersianText(request.Text).Display
	hash := sha256.Sum256([]byte(text))

	return s.createDocumentVersion(ctx, createDocumentInput{
		TenantID:        tenantID,
		KnowledgeBaseID: knowledgeBaseID,
		Actor:           actor,
		Title:           strings.TrimSpace(request.Title),
		SourceType:      "text",
		MediaType:       "text/plain",
		SourceText:      &text,
		ContentSHA256:   hex.EncodeToString(hash[:]),
		SizeBytes:       int64(len([]byte(text))),
	})
}

type createDocumentInput struct {
	TenantID        string
	KnowledgeBaseID string
	Actor           string
	Title           string
	SourceType      string
	MediaType       string
	ObjectKey       *string
	SourceText      *string
	ContentSHA256   string
	SizeBytes       int64
}

func (s *Store) createDocumentVersion(ctx context.Context, input createDocumentInput) (*platform.KnowledgeBaseDocument, error) {
	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create document: %w", err)
	}
	defer rollback(ctx, tx)

	var document platform.KnowledgeBaseDocument

	err = tx.QueryRow(ctx, `
		INSERT INTO knowledge_base_documents (tenant_id, knowledge_base_id, title, source_type, media_type, created_by)
		SELECT $1, kb.knowledge_base_id, $3, $4, $5, $6
		FROM knowledge_bases kb
		WHERE kb.tenant_id = $1 AND kb.knowledge_base_id = $2 AND kb.deleted_at IS NULL
		RETURNING document_id::text, knowledge_base_id::text, title, source_type, media_type,
		          'queued', active_version_id::text, processing_version_id::text, 0, NULL, 0, '', '', created_at, updated_at`,
		input.TenantID, input.KnowledgeBaseID, input.Title, input.SourceType, input.MediaType, input.Actor,
	).Scan(
		&document.ID, &document.KnowledgeBaseID, &document.Title, &document.SourceType, &document.MediaType,
		&document.Status, &document.ActiveVersionID, &document.ProcessingVersionID, &document.SizeBytes,
		&document.PageCount, &document.ChunkCount, &document.FailureCode, &document.FailureMessage,
		&document.CreatedAt, &document.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("insert document: %w", err)
	}

	versionID := uuid.NewString()

	_, err = tx.Exec(ctx, `
		INSERT INTO document_versions (
			version_id, tenant_id, knowledge_base_id, document_id, version, status,
			object_key, source_text, content_sha256, size_bytes, normalization_revision
		)
		SELECT $1, $2, $3, $4, 1, 'queued', $5, $6, $7, $8,
		       profile.configuration->>'normalization_revision'
		FROM knowledge_bases kb
		JOIN rag_profiles profile ON profile.revision = kb.rag_profile_revision
		WHERE kb.tenant_id = $2 AND kb.knowledge_base_id = $3 AND kb.deleted_at IS NULL`,
		versionID, input.TenantID, input.KnowledgeBaseID, document.ID,
		input.ObjectKey, input.SourceText, input.ContentSHA256, input.SizeBytes)
	if err != nil {
		return nil, fmt.Errorf("insert document version: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE knowledge_base_documents SET processing_version_id = $3, updated_at = now()
		WHERE tenant_id = $1 AND document_id = $2`, input.TenantID, document.ID, versionID); err != nil {
		return nil, fmt.Errorf("attach processing version: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO rag_ingestion_jobs (
			tenant_id, knowledge_base_id, document_id, version_id, idempotency_key
		) VALUES ($1, $2, $3, $4, $5)`,
		input.TenantID, input.KnowledgeBaseID, document.ID, versionID, "ingest:"+versionID); err != nil {
		return nil, fmt.Errorf("insert ingestion job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit create document: %w", err)
	}

	document.ProcessingVersionID = &versionID
	document.SizeBytes = input.SizeBytes

	return &document, nil
}
