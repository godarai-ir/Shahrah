-- +goose Up
ALTER TABLE public.wallets
    ALTER COLUMN balance_irr TYPE numeric(18,0) USING ROUND(balance_irr)::numeric(18,0),
    ALTER COLUMN pending_irr TYPE numeric(18,0) USING ROUND(pending_irr)::numeric(18,0);

ALTER TABLE public.wallet_transactions
    ALTER COLUMN amount_irr TYPE numeric(18,0) USING ROUND(amount_irr)::numeric(18,0),
    ADD COLUMN IF NOT EXISTS balance_after_irr numeric(18,0),
    ADD COLUMN IF NOT EXISTS idempotency_key character varying(128),
    ADD COLUMN IF NOT EXISTS source_type character varying(64),
    ADD COLUMN IF NOT EXISTS source_id character varying(255),
    ADD COLUMN IF NOT EXISTS effective_at timestamp with time zone NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS completed_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS created_by character varying(255);

CREATE UNIQUE INDEX IF NOT EXISTS uq_wallet_transactions_tenant_idempotency
    ON public.wallet_transactions (tenant_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_wallet_transactions_tenant_source
    ON public.wallet_transactions (tenant_id, source_type, source_id);

ALTER TABLE public.wallet_topup_intents
    ALTER COLUMN amount_irr TYPE numeric(18,0) USING ROUND(amount_irr)::numeric(18,0),
    ALTER COLUMN tax_amount_irr TYPE numeric(18,0) USING ROUND(tax_amount_irr)::numeric(18,0),
    ALTER COLUMN fee_amount_irr TYPE numeric(18,0) USING ROUND(fee_amount_irr)::numeric(18,0),
    ALTER COLUMN payable_irr TYPE numeric(18,0) USING ROUND(payable_irr)::numeric(18,0);

CREATE TABLE IF NOT EXISTS public.bank_transfer_requests (
    bank_transfer_request_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    public_id uuid DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    amount_irr numeric(18,0) NOT NULL CHECK (amount_irr > 0),
    tracking_number character varying(255) NOT NULL,
    bank_transaction_at timestamp with time zone NOT NULL,
    receipt_file_url text,
    status character varying(32) NOT NULL DEFAULT 'pending_review',
    rejection_reason text,
    wallet_tx_id uuid REFERENCES public.wallet_transactions(transaction_id) ON DELETE SET NULL,
    reviewed_by character varying(255),
    reviewed_at timestamp with time zone,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_by character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_bank_transfer_status CHECK (status IN ('pending_review', 'approved', 'rejected', 'cancelled'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_bank_transfer_tenant_tracking
    ON public.bank_transfer_requests (tenant_id, tracking_number);

CREATE INDEX IF NOT EXISTS idx_bank_transfer_tenant_created
    ON public.bank_transfer_requests (tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS public.gift_codes (
    gift_code_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    code character varying(128) NOT NULL,
    tenant_id uuid REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    amount_irr numeric(18,0) NOT NULL CHECK (amount_irr > 0),
    starts_at timestamp with time zone,
    expires_at timestamp with time zone,
    max_redemptions integer CHECK (max_redemptions IS NULL OR max_redemptions > 0),
    redeemed_count integer NOT NULL DEFAULT 0 CHECK (redeemed_count >= 0),
    active boolean NOT NULL DEFAULT true,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_gift_codes_code_ci
    ON public.gift_codes (LOWER(code));

CREATE TABLE IF NOT EXISTS public.gift_code_redemptions (
    redemption_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    gift_code_id uuid NOT NULL REFERENCES public.gift_codes(gift_code_id) ON DELETE CASCADE,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    wallet_tx_id uuid REFERENCES public.wallet_transactions(transaction_id) ON DELETE SET NULL,
    amount_irr numeric(18,0) NOT NULL CHECK (amount_irr > 0),
    status character varying(32) NOT NULL DEFAULT 'used',
    redeemed_by character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_gift_redemption_status CHECK (status IN ('active', 'used', 'expired', 'cancelled'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_gift_redemptions_code_tenant
    ON public.gift_code_redemptions (gift_code_id, tenant_id);

CREATE TABLE IF NOT EXISTS public.discount_codes (
    discount_code_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    code character varying(128) NOT NULL,
    tenant_id uuid REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    discount_type character varying(32) NOT NULL,
    value_irr numeric(18,0),
    percent numeric(5,2),
    max_discount_irr numeric(18,0),
    plan_slug character varying(64),
    first_payment_only boolean NOT NULL DEFAULT false,
    first_invoice_only boolean NOT NULL DEFAULT false,
    starts_at timestamp with time zone,
    expires_at timestamp with time zone,
    max_redemptions integer CHECK (max_redemptions IS NULL OR max_redemptions > 0),
    redeemed_count integer NOT NULL DEFAULT 0 CHECK (redeemed_count >= 0),
    active boolean NOT NULL DEFAULT true,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_discount_code_type CHECK (discount_type IN ('percent', 'fixed')),
    CONSTRAINT chk_discount_code_value CHECK (
        (discount_type = 'percent' AND percent IS NOT NULL AND percent > 0 AND percent <= 100)
        OR (discount_type = 'fixed' AND value_irr IS NOT NULL AND value_irr > 0)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_discount_codes_code_ci
    ON public.discount_codes (LOWER(code));

CREATE TABLE IF NOT EXISTS public.discount_redemptions (
    redemption_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    discount_code_id uuid NOT NULL REFERENCES public.discount_codes(discount_code_id) ON DELETE CASCADE,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    code character varying(128) NOT NULL,
    discount_type character varying(32) NOT NULL,
    value_irr numeric(18,0),
    percent numeric(5,2),
    starts_at timestamp with time zone,
    expires_at timestamp with time zone,
    status character varying(32) NOT NULL DEFAULT 'active',
    used_amount_irr numeric(18,0) NOT NULL DEFAULT 0,
    redeemed_by character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_discount_redemption_status CHECK (status IN ('active', 'used', 'expired', 'cancelled'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_discount_redemptions_code_tenant
    ON public.discount_redemptions (discount_code_id, tenant_id);

CREATE TABLE IF NOT EXISTS public.billing_profiles (
    billing_profile_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL UNIQUE REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    profile_type character varying(16) NOT NULL,
    data jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_by character varying(255),
    updated_by character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_billing_profile_type CHECK (profile_type IN ('individual', 'company'))
);

CREATE TABLE IF NOT EXISTS public.invoices (
    invoice_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    invoice_number character varying(64) NOT NULL UNIQUE,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    subscription_id uuid REFERENCES public.tenant_subscriptions(subscription_id) ON DELETE SET NULL,
    status character varying(32) NOT NULL DEFAULT 'draft',
    issued_at timestamp with time zone,
    period_start timestamp with time zone NOT NULL,
    period_end timestamp with time zone NOT NULL,
    tenant_name text,
    billing_profile_snapshot jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_usage_irr numeric(18,0) NOT NULL DEFAULT 0,
    discount_irr numeric(18,0) NOT NULL DEFAULT 0,
    gift_credit_irr numeric(18,0) NOT NULL DEFAULT 0,
    cache_savings_irr numeric(18,0) NOT NULL DEFAULT 0,
    final_amount_irr numeric(18,0) NOT NULL DEFAULT 0,
    net_adjustment_irr numeric(18,0) NOT NULL DEFAULT 0,
    wallet_tx_id uuid REFERENCES public.wallet_transactions(transaction_id) ON DELETE SET NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_invoice_status CHECK (status IN ('draft', 'issued', 'paid', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_invoices_tenant_period
    ON public.invoices (tenant_id, period_start DESC, period_end DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_subscription_period
    ON public.invoices (subscription_id, period_start, period_end)
    WHERE subscription_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS public.invoice_line_items (
    line_item_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    invoice_id uuid NOT NULL REFERENCES public.invoices(invoice_id) ON DELETE CASCADE,
    item_type character varying(64) NOT NULL,
    provider_name character varying(255),
    description text NOT NULL,
    quantity numeric(18,6) NOT NULL DEFAULT 1,
    amount_irr numeric(18,0) NOT NULL DEFAULT 0,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_invoice_line_items_invoice
    ON public.invoice_line_items (invoice_id);

CREATE TABLE IF NOT EXISTS public.financial_audit_events (
    audit_event_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    event_type character varying(128) NOT NULL,
    actor_id character varying(255),
    actor_role character varying(255),
    source_type character varying(64),
    source_id character varying(255),
    idempotency_key character varying(128),
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_financial_audit_tenant_created
    ON public.financial_audit_events (tenant_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS public.financial_audit_events;
DROP TABLE IF EXISTS public.invoice_line_items;
DROP TABLE IF EXISTS public.invoices;
DROP TABLE IF EXISTS public.billing_profiles;
DROP TABLE IF EXISTS public.discount_redemptions;
DROP TABLE IF EXISTS public.discount_codes;
DROP TABLE IF EXISTS public.gift_code_redemptions;
DROP TABLE IF EXISTS public.gift_codes;
DROP TABLE IF EXISTS public.bank_transfer_requests;

DROP INDEX IF EXISTS public.idx_wallet_transactions_tenant_source;
DROP INDEX IF EXISTS public.uq_wallet_transactions_tenant_idempotency;

ALTER TABLE public.wallet_transactions
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS completed_at,
    DROP COLUMN IF EXISTS effective_at,
    DROP COLUMN IF EXISTS source_id,
    DROP COLUMN IF EXISTS source_type,
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS balance_after_irr;
