package storage

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrVideoJobConflict = errors.New("video job version conflict")

func (s *Store) CreateVideoJob(ctx context.Context, job *platform.VideoGenerationJob) (*platform.VideoGenerationJob, error) {
	if job.ID == "" {
		job.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if job.CreatedAt.IsZero() {
		job.CreatedAt = now
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = now
	}
	if job.ExpiresAt.IsZero() {
		job.ExpiresAt = now.Add(7 * 24 * time.Hour)
	}
	if job.GenerationStatus == "" {
		job.GenerationStatus = platform.VideoGenerationSubmitting
	}
	if job.DeliveryStatus == "" {
		job.DeliveryStatus = platform.VideoDeliveryPending
	}
	if job.SettlementStatus == "" {
		job.SettlementStatus = platform.VideoSettlementPending
	}
	if job.Version == 0 {
		job.Version = 1
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	_, err = tx.Exec(ctx, `
		INSERT INTO video_generation_jobs (
			job_id, tenant_id, request_id, idempotency_key, token_hash, actor_id,
			model, provider, provider_model, deployment_id, provider_account_id,
			region, billing_mode, pricing_source, provider_job_id,
			generation_status, delivery_status, settlement_status,
			request_snapshot, requested_usage, actual_usage, pricing_snapshot,
			estimated_rating_snapshot, rating_snapshot, exchange_rate, estimated_cost_usd, estimated_cost_irr,
			final_cost_usd, final_cost_irr, budget_reservation, provider_response,
			output_artifacts, last_error, next_poll_at, attempt_count, version,
			completed_at, cancelled_at, expires_at, created_at, updated_at
		) VALUES (
			$1::uuid, $2::uuid, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''),
			$7, $8, $9, $10, NULLIF($11, '')::uuid,
			NULLIF($12, ''), $13, NULLIF($14, ''), NULLIF($15, ''),
			$16, $17, $18,
			$19, $20, $21, $22,
			$23, $24, $25, $26, $27,
			$28, $29, $30, $31,
			$32, NULLIF($33, ''), $34, $35, $36,
			$37, $38, $39, $40, $41
		)
	`, job.ID, job.TenantID, job.RequestID, job.IdempotencyKey, job.TokenHash, job.ActorID,
		job.Model, job.Provider, job.ProviderModel, job.DeploymentID, job.ProviderAccountID,
		job.Region, job.BillingMode, job.PricingSource, job.ProviderJobID,
		job.GenerationStatus, job.DeliveryStatus, job.SettlementStatus,
		marshalJSONBestEffort(job.Request), marshalJSONBestEffort(job.RequestedUsage), nullableJSON(job.ActualUsage),
		nullableJSON(job.PricingSnapshot), marshalJSONBestEffort(job.EstimatedRating), nullableJSON(job.FinalRating), job.ExchangeRate,
		ratedCost(job.EstimatedRating, false), ratedCost(job.EstimatedRating, true),
		ratedCost(job.FinalRating, false), ratedCost(job.FinalRating, true),
		nullableJSON(job.BudgetReservation), nullableJSON(job.ProviderResponse),
		marshalJSONBestEffort(job.Artifacts), job.LastError, job.NextPollAt, job.AttemptCount, job.Version,
		job.CompletedAt, job.CancelledAt, job.ExpiresAt, job.CreatedAt, job.UpdatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO video_job_submissions (job_id, attempt, status, request_payload)
		VALUES ($1::uuid, 1, 'pending', $2)
	`, job.ID, marshalJSONBestEffort(job.Request))
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *Store) GetVideoJob(ctx context.Context, tenantID, jobID string) (*platform.VideoGenerationJob, error) {
	return scanVideoJob(s.pool.QueryRow(ctx, videoJobSelect+` WHERE tenant_id = $1::uuid AND job_id = $2::uuid`, tenantID, jobID))
}

func (s *Store) GetVideoJobByIdempotency(ctx context.Context, tenantID, key string) (*platform.VideoGenerationJob, error) {
	return scanVideoJob(s.pool.QueryRow(ctx, videoJobSelect+` WHERE tenant_id = $1::uuid AND idempotency_key = $2`, tenantID, key))
}

func (s *Store) GetVideoJobByProviderID(ctx context.Context, provider, providerJobID string) (*platform.VideoGenerationJob, error) {
	return scanVideoJob(s.pool.QueryRow(ctx, videoJobSelect+`
		WHERE LOWER(provider) = LOWER($1) AND provider_job_id = $2
	`, strings.TrimSpace(provider), strings.TrimSpace(providerJobID)))
}

func (s *Store) ListVideoJobs(ctx context.Context, tenantID string, limit int, before time.Time) ([]platform.VideoGenerationJob, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if before.IsZero() {
		before = time.Now().UTC().Add(time.Second)
	}
	rows, err := s.pool.Query(ctx, videoJobSelect+`
		WHERE tenant_id = $1::uuid AND created_at < $2
		ORDER BY created_at DESC
		LIMIT $3
	`, tenantID, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.VideoGenerationJob, 0, limit)
	for rows.Next() {
		item, scanErr := scanVideoJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (s *Store) SaveVideoJob(ctx context.Context, job *platform.VideoGenerationJob, expectedVersion int64) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE video_generation_jobs SET
			provider_job_id = NULLIF($3, ''),
			generation_status = $4,
			delivery_status = $5,
			settlement_status = $6,
			actual_usage = $7,
			rating_snapshot = $8,
			final_cost_usd = $9,
			final_cost_irr = $10,
			budget_reservation = $11,
			provider_response = $12,
			output_artifacts = $13,
			last_error = NULLIF($14, ''),
			next_poll_at = $15,
			lease_owner = NULL,
			lease_expires_at = NULL,
			attempt_count = $16,
			completed_at = $17,
			cancelled_at = $18,
			version = version + 1,
			updated_at = NOW()
		WHERE job_id = $1::uuid AND tenant_id = $2::uuid AND version = $19
	`, job.ID, job.TenantID, job.ProviderJobID, job.GenerationStatus, job.DeliveryStatus,
		job.SettlementStatus, nullableJSON(job.ActualUsage), nullableJSON(job.FinalRating),
		ratedCost(job.FinalRating, false), ratedCost(job.FinalRating, true),
		nullableJSON(job.BudgetReservation), nullableJSON(job.ProviderResponse),
		marshalJSONBestEffort(job.Artifacts), job.LastError, job.NextPollAt, job.AttemptCount,
		job.CompletedAt, job.CancelledAt, expectedVersion)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrVideoJobConflict
	}
	job.Version = expectedVersion + 1
	job.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Store) MarkVideoSubmission(ctx context.Context, jobID string, attempt int, status string, response map[string]interface{}, submissionErr string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE video_job_submissions SET
			status = $3,
			response_payload = $4,
			error = NULLIF($5, ''),
			started_at = CASE WHEN $3 = 'processing' THEN COALESCE(started_at, NOW()) ELSE started_at END,
			completed_at = CASE WHEN $3 IN ('submitted', 'failed') THEN NOW() ELSE completed_at END,
			updated_at = NOW()
		WHERE job_id = $1::uuid AND attempt = $2
	`, jobID, attempt, status, nullableJSON(response), submissionErr)
	return err
}

func (s *Store) ClaimVideoJobs(ctx context.Context, worker string, limit int, lease time.Duration) ([]platform.VideoGenerationJob, error) {
	if limit <= 0 {
		limit = 10
	}
	if lease <= 0 {
		lease = 30 * time.Second
	}
	rows, err := s.pool.Query(ctx, `
		WITH candidates AS (
			SELECT job_id
			FROM video_generation_jobs
			WHERE (
				generation_status IN ('submitting', 'queued', 'running')
				OR (
					generation_status = 'succeeded'
					AND (
						settlement_status = 'pending'
						OR delivery_status IN ('pending', 'copying', 'failed')
					)
				)
				OR (
					generation_status IN ('failed', 'cancelled')
					AND settlement_status = 'pending'
				)
			)
			  AND (next_poll_at IS NULL OR next_poll_at <= NOW())
			  AND (lease_expires_at IS NULL OR lease_expires_at < NOW())
			ORDER BY COALESCE(next_poll_at, created_at)
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE video_generation_jobs j
		SET lease_owner = $2, lease_expires_at = NOW() + $3::interval, updated_at = NOW()
		FROM candidates c
		WHERE j.job_id = c.job_id
		RETURNING `+videoJobColumns("j"), limit, worker, lease.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.VideoGenerationJob, 0, limit)
	for rows.Next() {
		item, scanErr := scanVideoJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

// ClaimExpiredVideoJobs leases terminal jobs whose managed artifacts have
// reached their retention deadline. Deleting an S3 object is idempotent, and
// SaveVideoJob's version check prevents two janitors from finalizing the same
// job concurrently.
func (s *Store) ClaimExpiredVideoJobs(ctx context.Context, worker string, limit int, lease time.Duration) ([]platform.VideoGenerationJob, error) {
	if limit <= 0 {
		limit = 10
	}
	if lease <= 0 {
		lease = 30 * time.Second
	}
	rows, err := s.pool.Query(ctx, `
		WITH candidates AS (
			SELECT job_id
			FROM video_generation_jobs
			WHERE expires_at <= NOW()
			  AND delivery_status IN ('ready', 'failed')
			  AND generation_status IN ('succeeded', 'failed', 'cancelled', 'billing_review')
			  AND (lease_expires_at IS NULL OR lease_expires_at < NOW())
			ORDER BY expires_at
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE video_generation_jobs j
		SET lease_owner = $2, lease_expires_at = NOW() + $3::interval, updated_at = NOW()
		FROM candidates c
		WHERE j.job_id = c.job_id
		RETURNING `+videoJobColumns("j"), limit, worker, lease.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.VideoGenerationJob, 0, limit)
	for rows.Next() {
		item, scanErr := scanVideoJob(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

const videoJobSelect = `SELECT ` + `
	job_id::text, tenant_id::text, request_id, COALESCE(idempotency_key, ''),
	COALESCE(token_hash, ''), COALESCE(actor_id, ''), model, provider, provider_model,
	deployment_id, COALESCE(provider_account_id::text, ''), COALESCE(region, ''),
	billing_mode, COALESCE(pricing_source, ''), COALESCE(provider_job_id, ''),
	generation_status, delivery_status, settlement_status, request_snapshot,
	requested_usage, actual_usage, pricing_snapshot, estimated_rating_snapshot, rating_snapshot,
	exchange_rate::text, budget_reservation, provider_response, output_artifacts,
	COALESCE(last_error, ''), next_poll_at, lease_owner, lease_expires_at,
	attempt_count, version, completed_at, cancelled_at, expires_at, created_at, updated_at
FROM video_generation_jobs`

func videoJobColumns(alias string) string {
	p := alias + "."
	return p + `job_id::text, ` + p + `tenant_id::text, ` + p + `request_id, COALESCE(` + p + `idempotency_key, ''),
		COALESCE(` + p + `token_hash, ''), COALESCE(` + p + `actor_id, ''), ` + p + `model, ` + p + `provider, ` + p + `provider_model,
		` + p + `deployment_id, COALESCE(` + p + `provider_account_id::text, ''), COALESCE(` + p + `region, ''),
		` + p + `billing_mode, COALESCE(` + p + `pricing_source, ''), COALESCE(` + p + `provider_job_id, ''),
		` + p + `generation_status, ` + p + `delivery_status, ` + p + `settlement_status, ` + p + `request_snapshot,
		` + p + `requested_usage, ` + p + `actual_usage, ` + p + `pricing_snapshot, ` + p + `estimated_rating_snapshot, ` + p + `rating_snapshot,
		` + p + `exchange_rate::text, ` + p + `budget_reservation, ` + p + `provider_response, ` + p + `output_artifacts,
		COALESCE(` + p + `last_error, ''), ` + p + `next_poll_at, ` + p + `lease_owner, ` + p + `lease_expires_at,
		` + p + `attempt_count, ` + p + `version, ` + p + `completed_at, ` + p + `cancelled_at, ` + p + `expires_at, ` + p + `created_at, ` + p + `updated_at`
}

type videoRow interface {
	Scan(dest ...interface{}) error
}

func scanVideoJob(row videoRow) (*platform.VideoGenerationJob, error) {
	var (
		job                                    platform.VideoGenerationJob
		requestRaw, requestedRaw, artifactsRaw []byte
		actualRaw, pricingRaw                  []byte
		estimatedRatingRaw, finalRatingRaw     []byte
		reservationRaw, providerResponseRaw    []byte
	)
	err := row.Scan(
		&job.ID, &job.TenantID, &job.RequestID, &job.IdempotencyKey,
		&job.TokenHash, &job.ActorID, &job.Model, &job.Provider, &job.ProviderModel,
		&job.DeploymentID, &job.ProviderAccountID, &job.Region,
		&job.BillingMode, &job.PricingSource, &job.ProviderJobID,
		&job.GenerationStatus, &job.DeliveryStatus, &job.SettlementStatus,
		&requestRaw, &requestedRaw, &actualRaw, &pricingRaw, &estimatedRatingRaw, &finalRatingRaw,
		&job.ExchangeRate, &reservationRaw, &providerResponseRaw, &artifactsRaw,
		&job.LastError, &job.NextPollAt, &job.LeaseOwner, &job.LeaseExpiresAt,
		&job.AttemptCount, &job.Version, &job.CompletedAt, &job.CancelledAt,
		&job.ExpiresAt, &job.CreatedAt, &job.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(requestRaw, &job.Request); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(requestedRaw, &job.RequestedUsage); err != nil {
		return nil, err
	}
	if len(actualRaw) > 0 {
		_ = json.Unmarshal(actualRaw, &job.ActualUsage)
	}
	if len(pricingRaw) > 0 {
		_ = json.Unmarshal(pricingRaw, &job.PricingSnapshot)
	}
	if len(estimatedRatingRaw) > 0 {
		var rating platform.RatedUsage
		if json.Unmarshal(estimatedRatingRaw, &rating) == nil {
			job.EstimatedRating = &rating
		}
	}
	if len(finalRatingRaw) > 0 {
		var rating platform.RatedUsage
		if json.Unmarshal(finalRatingRaw, &rating) == nil {
			job.FinalRating = &rating
		}
	}
	if len(reservationRaw) > 0 {
		_ = json.Unmarshal(reservationRaw, &job.BudgetReservation)
	}
	if len(providerResponseRaw) > 0 {
		_ = json.Unmarshal(providerResponseRaw, &job.ProviderResponse)
	}
	if len(artifactsRaw) > 0 {
		_ = json.Unmarshal(artifactsRaw, &job.Artifacts)
	}
	return &job, nil
}

func nullableJSON(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil || string(raw) == "null" {
		return nil
	}
	return raw
}

func ratedCost(rating *platform.RatedUsage, irr bool) interface{} {
	if rating == nil {
		return nil
	}
	if irr {
		if strings.TrimSpace(rating.TotalCostIRR) == "" {
			return "0"
		}
		return rating.TotalCostIRR
	}
	if strings.TrimSpace(rating.TotalCostUSD) == "" {
		return "0"
	}
	return rating.TotalCostUSD
}

var _ videoRow = pgx.Row(nil)
