-- Migration: 005_wallet_topup_intents.sql
-- Description: External payment intents for wallet top-up with idempotent callback processing.

CREATE TABLE IF NOT EXISTS wallet_topup_settings (
    settings_id BOOLEAN PRIMARY KEY DEFAULT TRUE,
    min_amount_usd DOUBLE PRECISION NOT NULL DEFAULT 1,
    max_amount_usd DOUBLE PRECISION NOT NULL DEFAULT 1000,
    tax_percent DOUBLE PRECISION NOT NULL DEFAULT 0,
    fee_percent DOUBLE PRECISION NOT NULL DEFAULT 2,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (min_amount_usd > 0),
    CHECK (max_amount_usd >= min_amount_usd),
    CHECK (tax_percent >= 0),
    CHECK (fee_percent >= 0)
);

INSERT INTO wallet_topup_settings (settings_id)
VALUES (TRUE)
ON CONFLICT (settings_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS wallet_topup_intents (
    topup_intent_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    public_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    amount_usd DOUBLE PRECISION NOT NULL,
    amount_irr DOUBLE PRECISION NOT NULL,
    exchange_rate DOUBLE PRECISION NOT NULL,
    tax_percent DOUBLE PRECISION NOT NULL,
    fee_percent DOUBLE PRECISION NOT NULL,
    tax_amount_irr DOUBLE PRECISION NOT NULL,
    fee_amount_irr DOUBLE PRECISION NOT NULL,
    payable_irr DOUBLE PRECISION NOT NULL,
    authority VARCHAR(255),
    payment_url TEXT,
    ref_id VARCHAR(255),
    gateway_txn_id VARCHAR(255),
    card_mask VARCHAR(64),
    payer_phone VARCHAR(64),
    callback_url TEXT,
    return_url TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    failure_reason TEXT,
    wallet_tx_id UUID REFERENCES wallet_transactions(transaction_id) ON DELETE SET NULL,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'CANCELED'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallet_topup_intents_authority_unique
ON wallet_topup_intents(authority)
WHERE authority IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_wallet_topup_intents_tenant_created
ON wallet_topup_intents(tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_wallet_topup_intents_status
ON wallet_topup_intents(status);
