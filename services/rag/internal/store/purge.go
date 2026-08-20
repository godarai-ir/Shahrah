package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	// TenantVectorPurgeObjectKey identifies a tenant-wide vector purge task.
	TenantVectorPurgeObjectKey     = "__tenant_vectors__"
	knowledgeBaseVectorPurgePrefix = "__knowledge_base_vectors__:"
	documentVectorPurgePrefix      = "__document_vectors__:"
)

// KnowledgeBaseVectorPurgeObjectKey creates a durable vector-purge marker.
func KnowledgeBaseVectorPurgeObjectKey(knowledgeBaseID string) string {
	return knowledgeBaseVectorPurgePrefix + strings.TrimSpace(knowledgeBaseID)
}

// DocumentVectorPurgeObjectKey creates a durable vector-purge marker.
func DocumentVectorPurgeObjectKey(documentID string) string {
	return documentVectorPurgePrefix + strings.TrimSpace(documentID)
}

// ParseVectorPurgeObjectKey resolves an internal purge marker into its scope and identifier.
func ParseVectorPurgeObjectKey(objectKey string) (string, string, bool) {
	switch {
	case objectKey == TenantVectorPurgeObjectKey:
		return "tenant", "", true
	case strings.HasPrefix(objectKey, knowledgeBaseVectorPurgePrefix):
		identifier := strings.TrimPrefix(objectKey, knowledgeBaseVectorPurgePrefix)
		if identifier == "" {
			return "", "", false
		}

		return "knowledge_base", identifier, true
	case strings.HasPrefix(objectKey, documentVectorPurgePrefix):
		identifier := strings.TrimPrefix(objectKey, documentVectorPurgePrefix)
		if identifier == "" {
			return "", "", false
		}

		return "document", identifier, true
	default:
		return "", "", false
	}
}

// PurgeJob is a leased private-object deletion task.
type PurgeJob struct {
	ID          string
	TenantID    string
	ObjectKey   string
	Attempt     int
	MaxAttempts int
}

// ClaimPurgeJob leases one deletion task with crash recovery.
func (s *Store) ClaimPurgeJob(ctx context.Context, owner string, lease time.Duration) (*PurgeJob, error) {
	tx, err := s.metadata.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin claim purge job: %w", err)
	}
	defer rollback(ctx, tx)

	var job PurgeJob

	err = tx.QueryRow(ctx, `
		SELECT purge_job_id::text, tenant_id::text, object_key, attempt, max_attempts
		FROM rag_purge_jobs
		WHERE (status IN ('queued','retry') AND next_attempt_at <= now())
		   OR (status='running' AND lease_expires_at < now())
		ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT 1`).Scan(
		&job.ID, &job.TenantID, &job.ObjectKey, &job.Attempt, &job.MaxAttempts,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("select purge job: %w", err)
	}

	job.Attempt++
	if _, err := tx.Exec(ctx, `
		UPDATE rag_purge_jobs SET status='running', attempt=$2, lease_owner=$3,
		lease_expires_at=now()+$4::interval, updated_at=now() WHERE purge_job_id=$1`,
		job.ID, job.Attempt, owner, intervalLiteral(lease)); err != nil {
		return nil, fmt.Errorf("lease purge job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit purge lease: %w", err)
	}

	return &job, nil
}

// CompletePurge marks the private object deletion durable.
func (s *Store) CompletePurge(ctx context.Context, jobID, tenantID, owner string) error {
	command, err := s.metadata.Exec(ctx, `
		UPDATE rag_purge_jobs SET status='completed', lease_owner=NULL, lease_expires_at=NULL, updated_at=now()
		WHERE purge_job_id=$1 AND tenant_id=$2 AND status='running' AND lease_owner=$3`, jobID, tenantID, owner)
	if err != nil {
		return fmt.Errorf("complete purge job: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrLeaseLost
	}

	return nil
}

// FailPurge records a content-free failure and retries for up to 24 hours under normal backoff.
func (s *Store) FailPurge(ctx context.Context, job PurgeJob, owner, code string) error {
	status := "retry"
	if job.Attempt >= job.MaxAttempts {
		status = "failed"
	}

	command, err := s.metadata.Exec(ctx, `
		UPDATE rag_purge_jobs SET status=$3, error_code=$4, lease_owner=NULL, lease_expires_at=NULL,
		next_attempt_at=now()+make_interval(secs => LEAST(3600, 30*(2 ^ attempt))), updated_at=now()
		WHERE purge_job_id=$1 AND tenant_id=$2 AND status='running' AND lease_owner=$5`,
		job.ID, job.TenantID, status, code, owner)
	if err != nil {
		return fmt.Errorf("fail purge job: %w", err)
	}

	if command.RowsAffected() != 1 {
		return ErrLeaseLost
	}

	return nil
}
