-- 003_tenant_provider_accounts

-- +goose Up

-- Migration: 003_tenant_provider_accounts
-- Description: Tenant provider accounts, selected models, and access control.

CREATE TABLE IF NOT EXISTS tenant_provider_accounts (
    account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    provider VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    auth_mode VARCHAR(50) NOT NULL,
    provider_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    credential_data JSONB,
    has_credential BOOLEAN NOT NULL DEFAULT FALSE,
    allow_virtual_keys_only BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_provider_accounts_tenant_name UNIQUE (tenant_id, name),
    CONSTRAINT chk_tenant_provider_accounts_auth_mode CHECK (
        auth_mode IN (
            'api_key',
            'azure_entra_cert',
            'aws_sigv4',
            'google_service_account',
            'no_credential'
        )
    )
);

CREATE TABLE IF NOT EXISTS tenant_provider_account_models (
    account_model_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES tenant_provider_accounts(account_id) ON DELETE CASCADE,
    model VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    types TEXT[] NOT NULL DEFAULT '{}',
    pricing_mode VARCHAR(20) NOT NULL DEFAULT 'public',
    private_pricing JSONB,
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_provider_account_models_account_model UNIQUE (account_id, model),
    CONSTRAINT chk_tenant_provider_account_models_pricing_mode CHECK (
        pricing_mode IN ('public', 'private')
    )
);

CREATE TABLE IF NOT EXISTS tenant_provider_account_access (
    account_access_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES tenant_provider_accounts(account_id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    principal_type VARCHAR(20) NOT NULL,
    principal_id VARCHAR(255),
    display_name VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_tenant_provider_account_access_role CHECK (
        role IN ('manager', 'user')
    ),
    CONSTRAINT chk_tenant_provider_account_access_principal_type CHECK (
        principal_type IN ('user', 'team', 'everyone')
    ),
    CONSTRAINT chk_tenant_provider_account_access_everyone CHECK (
        (principal_type = 'everyone' AND principal_id IS NULL)
        OR
        (principal_type <> 'everyone' AND principal_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_tenant_provider_accounts_tenant_id
    ON tenant_provider_accounts(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenant_provider_accounts_tenant_provider
    ON tenant_provider_accounts(tenant_id, provider);
CREATE INDEX IF NOT EXISTS idx_tenant_provider_account_models_account_id
    ON tenant_provider_account_models(account_id);
CREATE INDEX IF NOT EXISTS idx_tenant_provider_account_access_account_id
    ON tenant_provider_account_access(account_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_tenant_provider_account_access_principal
    ON tenant_provider_account_access(account_id, role, principal_type, COALESCE(principal_id, ''));

-- +goose Down

-- Migration: 003_tenant_provider_accounts (DOWN)
-- Description: Rollback tenant provider accounts, selected models, and access control.

DROP INDEX IF EXISTS uq_tenant_provider_account_access_principal;
DROP INDEX IF EXISTS idx_tenant_provider_account_access_account_id;
DROP INDEX IF EXISTS idx_tenant_provider_account_models_account_id;
DROP INDEX IF EXISTS idx_tenant_provider_accounts_tenant_provider;
DROP INDEX IF EXISTS idx_tenant_provider_accounts_tenant_id;

DROP TABLE IF EXISTS tenant_provider_account_access CASCADE;
DROP TABLE IF EXISTS tenant_provider_account_models CASCADE;
DROP TABLE IF EXISTS tenant_provider_accounts CASCADE;
