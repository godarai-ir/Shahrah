-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS provider_native_resource_mappings (
    public_id text NOT NULL,
    tenant_id uuid NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    provider text NOT NULL,
    provider_account_id uuid NOT NULL REFERENCES tenant_provider_accounts(account_id) ON DELETE CASCADE,
    resource_kind text NOT NULL,
    upstream_id text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    metadata_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    pricing_revision text,
    reservation_id text,
    reservation_json jsonb,
    estimated_usage_json jsonb,
    expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT provider_native_resource_provider_check
        CHECK (provider = lower(provider) AND provider <> ''),
    CONSTRAINT provider_native_resource_kind_check
        CHECK (resource_kind = lower(resource_kind) AND resource_kind <> ''),
    PRIMARY KEY (tenant_id, provider, resource_kind, public_id),
    UNIQUE (tenant_id, provider, provider_account_id, resource_kind, upstream_id)
);

CREATE INDEX IF NOT EXISTS idx_provider_native_resource_tenant_lookup
    ON provider_native_resource_mappings
    (tenant_id, provider, resource_kind, public_id);

CREATE INDEX IF NOT EXISTS idx_provider_native_resource_reconcile
    ON provider_native_resource_mappings
    (provider, resource_kind, status, updated_at)
    WHERE status IN ('queued', 'in_progress', 'canceling');

CREATE INDEX IF NOT EXISTS idx_provider_native_resource_expiry
    ON provider_native_resource_mappings (expires_at)
    WHERE expires_at IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS provider_native_resource_mappings;
-- +goose StatementEnd
