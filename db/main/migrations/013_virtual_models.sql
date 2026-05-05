-- Migration: 013_virtual_models.sql
-- Description: Tenant-scoped virtual models for gateway routing/fallback

CREATE TABLE IF NOT EXISTS virtual_models (
    virtual_model_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    alias VARCHAR(255) NOT NULL,
    routing_strategy VARCHAR(64) NOT NULL DEFAULT 'priority',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    version INT NOT NULL DEFAULT 1,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_virtual_models_tenant_alias UNIQUE (tenant_id, alias)
);

CREATE TABLE IF NOT EXISTS virtual_model_targets (
    target_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    virtual_model_id UUID NOT NULL REFERENCES virtual_models(virtual_model_id) ON DELETE CASCADE,
    target_order INT NOT NULL,
    provider VARCHAR(64) NOT NULL,
    model VARCHAR(255) NOT NULL,
    timeout_sec FLOAT,
    retry_attempts INT NOT NULL DEFAULT 2,
    retry_delay_ms INT NOT NULL DEFAULT 100,
    retry_status_codes INT[] NOT NULL DEFAULT ARRAY[429, 500, 502, 503],
    fallback_status_codes INT[] NOT NULL DEFAULT ARRAY[429, 500, 502, 503, 504],
    fallback_candidate BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_virtual_model_targets_order UNIQUE (virtual_model_id, target_order)
);

CREATE INDEX IF NOT EXISTS idx_virtual_models_tenant_id ON virtual_models(tenant_id);
CREATE INDEX IF NOT EXISTS idx_virtual_models_alias ON virtual_models(alias);
CREATE INDEX IF NOT EXISTS idx_virtual_model_targets_model ON virtual_model_targets(model);

