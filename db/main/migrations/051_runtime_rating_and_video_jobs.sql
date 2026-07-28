-- +goose Up
ALTER TABLE spend_logs
    ADD COLUMN IF NOT EXISTS usage_measures JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS rating_snapshot JSONB,
    ADD COLUMN IF NOT EXISTS cost_components JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS pricing_source VARCHAR(20),
    ADD COLUMN IF NOT EXISTS rating_status VARCHAR(32),
    ADD COLUMN IF NOT EXISTS video_seconds NUMERIC(12,3),
    ADD COLUMN IF NOT EXISTS video_resolution VARCHAR(32),
    ADD COLUMN IF NOT EXISTS output_count INTEGER;

ALTER TABLE model_deployments
    ADD COLUMN IF NOT EXISTS credential_source VARCHAR(32),
    ADD COLUMN IF NOT EXISTS billing_mode VARCHAR(32),
    ADD COLUMN IF NOT EXISTS effective_pricing JSONB,
    ADD COLUMN IF NOT EXISTS pricing_source VARCHAR(32);

UPDATE model_deployments
SET credential_source = CASE
        WHEN lower(provider) = 'mock' THEN 'godar_managed_key'
        WHEN btrim(COALESCE(provider_config->>'api_key', '')) <> '' THEN 'user_byok'
        ELSE credential_source
    END,
    billing_mode = CASE
        WHEN lower(provider) = 'mock' THEN 'external_billed'
        WHEN btrim(COALESCE(provider_config->>'api_key', '')) <> '' THEN 'external_billed'
        ELSE billing_mode
    END
WHERE credential_source IS NULL OR billing_mode IS NULL;

ALTER TABLE model_deployments
    DROP CONSTRAINT IF EXISTS ck_model_deployments_credential_source,
    ADD CONSTRAINT ck_model_deployments_credential_source CHECK (
        credential_source IS NULL OR credential_source IN ('godar_managed_key', 'user_byok', 'platform_managed', 'platform_env_break_glass')
    ),
    DROP CONSTRAINT IF EXISTS ck_model_deployments_billing_mode,
    ADD CONSTRAINT ck_model_deployments_billing_mode CHECK (
        billing_mode IS NULL OR billing_mode IN ('wallet_billed', 'external_billed')
    );

CREATE UNIQUE INDEX IF NOT EXISTS uq_spend_logs_tenant_request
    ON spend_logs (tenant_id, request_id)
    WHERE tenant_id IS NOT NULL AND request_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_spend_logs_video_usage
    ON spend_logs (tenant_id, video_resolution, created_at DESC)
    WHERE video_seconds IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_spend_logs_rating_status
    ON spend_logs (tenant_id, rating_status, created_at DESC);

CREATE TABLE IF NOT EXISTS video_generation_jobs (
    job_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    request_id VARCHAR(255) NOT NULL,
    idempotency_key VARCHAR(128),
    token_hash VARCHAR(64),
    actor_id VARCHAR(255),
    owner_principal VARCHAR(320),
    model VARCHAR(255) NOT NULL,
    provider VARCHAR(100) NOT NULL,
    provider_model VARCHAR(255) NOT NULL,
    deployment_id VARCHAR(512) NOT NULL,
    provider_account_id UUID,
    region VARCHAR(100),
    billing_mode VARCHAR(32) NOT NULL,
    pricing_source VARCHAR(20),
    provider_job_id VARCHAR(512),
    generation_status VARCHAR(32) NOT NULL DEFAULT 'submitting',
    delivery_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    settlement_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    request_snapshot JSONB NOT NULL,
    requested_usage JSONB NOT NULL DEFAULT '[]'::jsonb,
    actual_usage JSONB,
    pricing_snapshot JSONB,
    estimated_rating_snapshot JSONB NOT NULL,
    rating_snapshot JSONB,
    exchange_rate NUMERIC(24,8) NOT NULL,
    estimated_cost_usd NUMERIC(24,12) NOT NULL DEFAULT 0,
    estimated_cost_irr NUMERIC(24,2) NOT NULL DEFAULT 0,
    final_cost_usd NUMERIC(24,12),
    final_cost_irr NUMERIC(24,2),
    budget_reservation JSONB,
    provider_response JSONB,
    output_artifacts JSONB NOT NULL DEFAULT '[]'::jsonb,
    last_error TEXT,
    next_poll_at TIMESTAMPTZ,
    lease_owner VARCHAR(255),
    lease_expires_at TIMESTAMPTZ,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_video_generation_status CHECK (
        generation_status IN ('submitting', 'queued', 'running', 'succeeded', 'failed', 'cancelled', 'billing_review')
    ),
    CONSTRAINT ck_video_delivery_status CHECK (
        delivery_status IN ('pending', 'copying', 'ready', 'failed', 'deleted')
    ),
    CONSTRAINT ck_video_settlement_status CHECK (
        settlement_status IN ('pending', 'committed', 'released', 'review')
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_video_jobs_tenant_request
    ON video_generation_jobs (tenant_id, request_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_video_jobs_tenant_owner_idempotency
    ON video_generation_jobs (tenant_id, owner_principal, idempotency_key)
    WHERE owner_principal IS NOT NULL AND idempotency_key IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uq_video_jobs_provider_job
    ON video_generation_jobs (provider, provider_job_id)
    WHERE provider_job_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_video_jobs_poll
    ON video_generation_jobs (next_poll_at, lease_expires_at)
    WHERE generation_status IN ('submitting', 'queued', 'running', 'succeeded', 'failed', 'cancelled');
CREATE INDEX IF NOT EXISTS idx_video_jobs_tenant_created
    ON video_generation_jobs (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_video_jobs_tenant_owner_created
    ON video_generation_jobs (tenant_id, owner_principal, created_at DESC);

CREATE TABLE IF NOT EXISTS video_job_submissions (
    submission_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES video_generation_jobs(job_id) ON DELETE CASCADE,
    attempt INTEGER NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    request_payload JSONB NOT NULL,
    response_payload JSONB,
    error TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_video_job_submission_attempt UNIQUE (job_id, attempt),
    CONSTRAINT ck_video_submission_status CHECK (
        status IN ('pending', 'processing', 'submitted', 'failed')
    )
);

CREATE INDEX IF NOT EXISTS idx_video_submissions_pending
    ON video_job_submissions (created_at)
    WHERE status IN ('pending', 'processing');

-- +goose Down
DROP TABLE IF EXISTS video_job_submissions;
DROP TABLE IF EXISTS video_generation_jobs;
DROP INDEX IF EXISTS idx_spend_logs_rating_status;
DROP INDEX IF EXISTS idx_spend_logs_video_usage;
DROP INDEX IF EXISTS uq_spend_logs_tenant_request;
ALTER TABLE model_deployments
    DROP CONSTRAINT IF EXISTS ck_model_deployments_billing_mode,
    DROP CONSTRAINT IF EXISTS ck_model_deployments_credential_source,
    DROP COLUMN IF EXISTS pricing_source,
    DROP COLUMN IF EXISTS effective_pricing,
    DROP COLUMN IF EXISTS billing_mode,
    DROP COLUMN IF EXISTS credential_source;
ALTER TABLE spend_logs
    DROP COLUMN IF EXISTS output_count,
    DROP COLUMN IF EXISTS video_resolution,
    DROP COLUMN IF EXISTS video_seconds,
    DROP COLUMN IF EXISTS rating_status,
    DROP COLUMN IF EXISTS pricing_source,
    DROP COLUMN IF EXISTS cost_components,
    DROP COLUMN IF EXISTS rating_snapshot,
    DROP COLUMN IF EXISTS usage_measures;
