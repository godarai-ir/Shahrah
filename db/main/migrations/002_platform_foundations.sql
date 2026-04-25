-- Migration: 002_platform_foundations.sql
-- Description: Tenant, wallet, exchange-rate, and tenant-scoped model/key foundations.

CREATE TABLE IF NOT EXISTS tenants (
    tenant_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(255) NOT NULL UNIQUE,
    domain VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE verification_tokens
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(tenant_id);

ALTER TABLE model_deployments
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(tenant_id);

ALTER TABLE request_logs
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(tenant_id);

ALTER TABLE request_logs
    ADD COLUMN IF NOT EXISTS model VARCHAR(255);

ALTER TABLE spend_logs
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(tenant_id);

ALTER TABLE spend_logs
    ADD COLUMN IF NOT EXISTS request_id VARCHAR(255);

ALTER TABLE spend_logs
    ADD COLUMN IF NOT EXISTS cost_usd DECIMAL(10,6);

ALTER TABLE spend_logs
    ADD COLUMN IF NOT EXISTS cost_irr DECIMAL(18,2);

CREATE TABLE IF NOT EXISTS wallets (
    wallet_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL UNIQUE REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    balance_irr DOUBLE PRECISION NOT NULL DEFAULT 0,
    pending_irr DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(wallet_id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    tx_type VARCHAR(50) NOT NULL,
    amount_irr DOUBLE PRECISION NOT NULL,
    amount_usd DOUBLE PRECISION,
    exchange_rate DOUBLE PRECISION,
    status VARCHAR(50) NOT NULL,
    description TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS exchange_rates (
    rate_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    usd_to_irr DOUBLE PRECISION NOT NULL,
    source VARCHAR(255) NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tenants (slug, domain, status)
VALUES ('default', 'default', 'active')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
SELECT tenant_id, 0, 0
FROM tenants
WHERE slug = 'default'
ON CONFLICT (tenant_id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_verification_tokens_tenant_id ON verification_tokens(tenant_id);
CREATE INDEX IF NOT EXISTS idx_model_deployments_tenant_id ON model_deployments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_tenant_id ON wallet_transactions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_exchange_rates_fetched_at ON exchange_rates(fetched_at DESC);
