package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/jackc/pgx/v5"
)

// IngestionJob is a leased unit of ingestion work.
type IngestionJob struct {
	ID              string
	TenantID        string
	TenantSlug      string
	KnowledgeBaseID string
	DocumentID      string
	VersionID       string
	Title           string
	SourceType      string
	MediaType       string
	ObjectKey       *string
	SourceText      *string
	Attempt         int
	MaxAttempts     int
	Profile         domain.RAGProfile
}

// ClaimIngestionJob uses SKIP LOCKED so multiple workers can safely share the queue.
func (s *Store) ClaimIngestionJob(ctx context.Context, owner string, lease time.Duration) (*IngestionJob, error) {
	if lease <= 0 {
		lease = 2 * time.Minute
	}

	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin claim ingestion job: %w", err)
	}
	defer rollback(ctx, tx)

	var (
		job                  IngestionJob
		profileConfiguration []byte
	)

	err = tx.QueryRow(ctx, `
		SELECT j.job_id::text, j.tenant_id::text, t.slug, j.knowledge_base_id::text,
		       j.document_id::text, j.version_id::text, d.title, d.source_type, d.media_type,
		       v.object_key, v.source_text, j.attempt, j.max_attempts,
		       profile.revision, profile.embedding_model, profile.embedding_dimensions,
		       profile.rerank_model, profile.chunk_tokens, profile.chunk_overlap_tokens,
		       profile.max_chunks, profile.max_context_tokens, profile.configuration
		FROM rag_ingestion_jobs j
		JOIN tenants t ON t.tenant_id = j.tenant_id
		JOIN knowledge_base_documents d ON d.tenant_id = j.tenant_id AND d.document_id = j.document_id
		JOIN document_versions v ON v.tenant_id = j.tenant_id AND v.version_id = j.version_id
		JOIN knowledge_bases kb ON kb.tenant_id = j.tenant_id AND kb.knowledge_base_id = j.knowledge_base_id
		JOIN rag_profiles profile ON profile.revision = kb.rag_profile_revision
		WHERE (
			(j.status IN ('queued','retry') AND j.next_attempt_at <= now())
			OR (j.status = 'running' AND j.lease_expires_at < now())
		) AND d.deleted_at IS NULL
		ORDER BY j.created_at
		FOR UPDATE OF j SKIP LOCKED
		LIMIT 1`).Scan(
		&job.ID, &job.TenantID, &job.TenantSlug, &job.KnowledgeBaseID,
		&job.DocumentID, &job.VersionID, &job.Title, &job.SourceType, &job.MediaType,
		&job.ObjectKey, &job.SourceText, &job.Attempt, &job.MaxAttempts,
		&job.Profile.Revision, &job.Profile.EmbeddingModel, &job.Profile.EmbeddingDimensions,
		&job.Profile.RerankModel, &job.Profile.ChunkTokens, &job.Profile.ChunkOverlapTokens,
		&job.Profile.MaxChunks, &job.Profile.MaxContextTokens, &profileConfiguration,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("select ingestion job: %w", err)
	}

	if err := json.Unmarshal(profileConfiguration, &job.Profile.Configuration); err != nil {
		return nil, fmt.Errorf("decode ingestion RAG profile %q: %w", job.Profile.Revision, err)
	}

	if err := job.Profile.Validate(); err != nil {
		return nil, fmt.Errorf("validate ingestion RAG profile %q: %w", job.Profile.Revision, err)
	}

	job.Attempt++
	if _, err := tx.Exec(ctx, `
		UPDATE rag_ingestion_jobs
		SET status = 'running', attempt = $2, lease_owner = $3, lease_expires_at = now() + $4::interval,
		    updated_at = now()
		WHERE job_id = $1`, job.ID, job.Attempt, owner, intervalLiteral(lease)); err != nil {
		return nil, fmt.Errorf("lease ingestion job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit ingestion lease: %w", err)
	}

	return &job, nil
}

// UpdateIngestionStage persists a resumable job and document lifecycle stage.
func (s *Store) UpdateIngestionStage(
	ctx context.Context,
	jobID, tenantID, versionID, stage, owner string,
	lease time.Duration,
) error {
	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update ingestion stage: %w", err)
	}
	defer rollback(ctx, tx)

	command, err := tx.Exec(ctx, `
		UPDATE rag_ingestion_jobs
		SET stage = $4, lease_expires_at = now() + $6::interval, updated_at = now()
		WHERE job_id = $1 AND tenant_id = $2 AND version_id = $3
		  AND status = 'running' AND lease_owner = $5`,
		jobID, tenantID, versionID, stage, owner, intervalLiteral(lease))
	if err != nil {
		return fmt.Errorf("update job stage: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrLeaseLost
	}

	command, err = tx.Exec(ctx, `
		UPDATE document_versions SET status = $3, updated_at = now()
		WHERE tenant_id = $1 AND version_id = $2 AND status <> 'deleting'`, tenantID, versionID, stage)
	if err != nil {
		return fmt.Errorf("update document stage: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrLeaseLost
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit ingestion stage: %w", err)
	}

	return nil
}

// RenewIngestionLease keeps ownership during long parser and provider calls.
func (s *Store) RenewIngestionLease(
	ctx context.Context,
	jobID, tenantID, owner string,
	lease time.Duration,
) error {
	command, err := s.metadata.Exec(ctx, `
		UPDATE rag_ingestion_jobs
		SET lease_expires_at = now() + $4::interval, updated_at = now()
		WHERE job_id = $1 AND tenant_id = $2 AND status = 'running' AND lease_owner = $3`,
		jobID, tenantID, owner, intervalLiteral(lease))
	if err != nil {
		return fmt.Errorf("renew ingestion lease: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrLeaseLost
	}

	return nil
}

// CompleteIngestion atomically activates metadata after vector activation succeeds.
func (s *Store) CompleteIngestion(
	ctx context.Context,
	job IngestionJob,
	owner string,
	pageCount *int,
	chunkCount int,
	detectedLanguage string,
) error {
	metadataTx, err := s.metadata.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin metadata activation: %w", err)
	}
	defer rollback(ctx, metadataTx)

	var (
		ownedJobID      string
		activeVersionID *string
	)

	lockErr := metadataTx.QueryRow(ctx, `
		SELECT j.job_id::text, d.active_version_id::text
		FROM knowledge_base_documents d
		JOIN rag_ingestion_jobs j
		  ON j.tenant_id = d.tenant_id AND j.document_id = d.document_id
		WHERE j.job_id = $1 AND j.tenant_id = $2 AND j.version_id = $3
		  AND j.status = 'running' AND j.lease_owner = $4
		  AND d.deleted_at IS NULL AND d.processing_version_id = j.version_id
		FOR UPDATE OF d, j`, job.ID, job.TenantID, job.VersionID, owner).Scan(&ownedJobID, &activeVersionID)
	if errors.Is(lockErr, pgx.ErrNoRows) {
		return ErrLeaseLost
	}

	if lockErr != nil {
		return fmt.Errorf("lock ingestion activation: %w", lockErr)
	}

	vectorTx, beginErr := s.vectors.Begin(ctx)
	if beginErr != nil {
		return fmt.Errorf("begin vector activation: %w", beginErr)
	}
	defer rollback(ctx, vectorTx)

	if _, err = vectorTx.Exec(ctx, `
		UPDATE rag_chunks SET active = false
		WHERE tenant_id = $1 AND document_id = $2 AND active = true`, job.TenantID, job.DocumentID); err != nil {
		return fmt.Errorf("deactivate old chunks: %w", err)
	}

	if _, err = vectorTx.Exec(ctx, `
		UPDATE rag_chunks SET active = true
		WHERE tenant_id = $1 AND document_version_id = $2`, job.TenantID, job.VersionID); err != nil {
		return fmt.Errorf("activate new chunks: %w", err)
	}

	err = vectorTx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit vector activation: %w", err)
	}

	compensateVectorActivation := func(cause error) error {
		compensationContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()

		compensationErr := s.restoreVectorActivation(
			compensationContext, job.TenantID, job.DocumentID, job.VersionID, activeVersionID,
		)
		if compensationErr != nil {
			return errors.Join(cause, fmt.Errorf("compensate vector activation: %w", compensationErr))
		}

		return cause
	}

	command, err := metadataTx.Exec(ctx, `
		UPDATE document_versions
		SET status = 'ready', page_count = $3, chunk_count = $4, detected_language = $5,
		    activated_at = now(), updated_at = now()
		WHERE tenant_id = $1 AND version_id = $2 AND status = 'indexing'`,
		job.TenantID, job.VersionID, pageCount, chunkCount, strings.TrimSpace(detectedLanguage))
	if err != nil {
		return compensateVectorActivation(fmt.Errorf("activate document version: %w", err))
	}

	if command.RowsAffected() != 1 {
		return compensateVectorActivation(ErrLeaseLost)
	}

	command, err = metadataTx.Exec(ctx, `
		UPDATE knowledge_base_documents
		SET active_version_id = $3, processing_version_id = NULL, updated_at = now()
		WHERE tenant_id = $1 AND document_id = $2 AND deleted_at IS NULL
		  AND processing_version_id = $3`, job.TenantID, job.DocumentID, job.VersionID)
	if err != nil {
		return compensateVectorActivation(fmt.Errorf("activate logical document: %w", err))
	}

	if command.RowsAffected() != 1 {
		return compensateVectorActivation(ErrLeaseLost)
	}

	command, err = metadataTx.Exec(ctx, `
		UPDATE rag_ingestion_jobs
		SET status = 'completed', stage = 'ready', lease_owner = NULL, lease_expires_at = NULL, updated_at = now()
		WHERE job_id = $1 AND tenant_id = $2 AND status = 'running' AND lease_owner = $3`,
		job.ID, job.TenantID, owner)
	if err != nil {
		return compensateVectorActivation(fmt.Errorf("complete ingestion job: %w", err))
	}

	if command.RowsAffected() != 1 {
		return compensateVectorActivation(ErrLeaseLost)
	}

	if err := metadataTx.Commit(ctx); err != nil {
		return compensateVectorActivation(fmt.Errorf("commit metadata activation: %w", err))
	}

	return nil
}

// FailIngestion records a structured, content-free error and applies bounded retry.
func (s *Store) FailIngestion(ctx context.Context, job IngestionJob, owner, code string, cause error) error {
	status := "retry"
	documentStatus := "queued"

	if job.Attempt >= job.MaxAttempts {
		status = "failed"
		documentStatus = "failed"
	}

	details, err := json.Marshal(map[string]any{"code": code, "retryable": status == "retry"})
	if err != nil {
		return fmt.Errorf("encode ingestion failure: %w", err)
	}

	message := code
	if cause != nil {
		message = fmt.Sprintf("%s: %s", code, safeErrorClass(cause))
	}

	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin fail ingestion: %w", err)
	}
	defer rollback(ctx, tx)

	command, err := tx.Exec(ctx, `
		UPDATE rag_ingestion_jobs
		SET status = $3, error_code = $4, error_details = $5, lease_owner = NULL,
		    lease_expires_at = NULL, next_attempt_at = now() + make_interval(secs => LEAST(300, 5 * (2 ^ attempt))),
		    updated_at = now()
		WHERE job_id = $1 AND tenant_id = $2 AND status = 'running' AND lease_owner = $6`,
		job.ID, job.TenantID, status, code, details, owner)
	if err != nil {
		return fmt.Errorf("fail ingestion job: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrLeaseLost
	}

	command, err = tx.Exec(ctx, `
		UPDATE document_versions
		SET status = $3, failure_code = $4, failure_message = $5, updated_at = now()
		WHERE tenant_id = $1 AND version_id = $2 AND status <> 'deleting'`,
		job.TenantID, job.VersionID, documentStatus, code, message)
	if err != nil {
		return fmt.Errorf("fail document version: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrLeaseLost
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit ingestion failure: %w", err)
	}

	return nil
}

func intervalLiteral(duration time.Duration) string {
	return fmt.Sprintf("%f seconds", duration.Seconds())
}

func safeErrorClass(err error) string {
	if err == nil {
		return "unknown"
	}

	return fmt.Sprintf("%T", err)
}
