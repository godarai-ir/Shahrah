BEGIN;

CREATE TABLE IF NOT EXISTS anthropic_file_mappings (
    public_id text PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_account_id uuid NOT NULL REFERENCES tenant_provider_accounts(account_id) ON DELETE CASCADE,
    upstream_id text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, provider_account_id, upstream_id)
);

CREATE INDEX IF NOT EXISTS idx_anthropic_file_mappings_tenant_account
    ON anthropic_file_mappings (tenant_id, provider_account_id, created_at DESC);

CREATE TABLE IF NOT EXISTS anthropic_batch_mappings (
    public_id text PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_account_id uuid NOT NULL REFERENCES tenant_provider_accounts(account_id) ON DELETE CASCADE,
    upstream_id text NOT NULL,
    status text NOT NULL DEFAULT 'in_progress',
    reservation_id text,
    reservation_json jsonb,
    pricing_revision text,
    estimated_usage_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    ended_at timestamptz,
    UNIQUE (tenant_id, provider_account_id, upstream_id)
);

CREATE INDEX IF NOT EXISTS idx_anthropic_batch_mappings_tenant_account_status
    ON anthropic_batch_mappings (tenant_id, provider_account_id, status, created_at DESC);

COMMIT;
