-- +goose Up
-- Credit add-on packs and overage settings (Phase 2c).
-- Pack prices are a discount on the active credit list price, computed per order
-- and snapshotted on it. Packs can only be bought by tenants billed in credits.
-- See plans/credit-pricing/phase-2-design.md.

-- +goose StatementBegin
CREATE TABLE public.credit_packs (
    pack_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    slug character varying(64) NOT NULL,
    version integer NOT NULL,
    credits bigint NOT NULL,
    discount_percent numeric(5,2) NOT NULL DEFAULT 0,
    validity_months integer NOT NULL DEFAULT 12,
    status character varying(16) NOT NULL DEFAULT 'draft',
    sort_order integer NOT NULL DEFAULT 100,
    created_by character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_credit_packs_credits CHECK (credits > 0),
    CONSTRAINT chk_credit_packs_discount CHECK (discount_percent >= 0 AND discount_percent < 100),
    CONSTRAINT chk_credit_packs_validity CHECK (validity_months > 0 AND validity_months <= 36),
    CONSTRAINT chk_credit_packs_status CHECK (status IN ('draft', 'published', 'archived'))
);

CREATE UNIQUE INDEX uq_credit_packs_slug_version ON public.credit_packs (slug, version);
CREATE UNIQUE INDEX uq_credit_packs_one_published ON public.credit_packs (slug) WHERE status = 'published';

INSERT INTO public.credit_packs (slug, version, credits, discount_percent, validity_months, status, sort_order, created_by)
VALUES
    ('credits-1000', 1, 1000, 0, 12, 'published', 10, 'migration'),
    ('credits-5000', 1, 5000, 10, 12, 'published', 20, 'migration'),
    ('credits-20000', 1, 20000, 20, 12, 'published', 30, 'migration');

CREATE TABLE public.credit_pack_orders (
    order_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    public_id uuid DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id),
    pack_id uuid NOT NULL REFERENCES public.credit_packs(pack_id),
    pack_slug character varying(64) NOT NULL,
    pack_version integer NOT NULL,
    credits bigint NOT NULL,
    validity_months integer NOT NULL,
    list_price_irr_per_credit numeric(18,0) NOT NULL,
    discount_percent numeric(5,2) NOT NULL,
    price_irr numeric(18,0) NOT NULL,
    tax_percent numeric(5,2) NOT NULL DEFAULT 0,
    tax_irr numeric(18,0) NOT NULL DEFAULT 0,
    payable_irr numeric(18,0) NOT NULL,
    provider character varying(32) NOT NULL,
    status character varying(32) NOT NULL DEFAULT 'PENDING',
    authority character varying(255),
    payment_url text,
    callback_url text,
    return_url text,
    ref_id character varying(255),
    gateway_txn_id character varying(255),
    card_mask character varying(64),
    payer_phone character varying(64),
    failure_reason text,
    raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    lot_id uuid,
    created_by character varying(255),
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_credit_pack_orders_status CHECK (status IN ('PENDING', 'AWAITING_TRANSFER', 'COMPLETED', 'FAILED', 'CANCELED')),
    CONSTRAINT chk_credit_pack_orders_amounts CHECK (price_irr > 0 AND tax_irr >= 0 AND payable_irr = price_irr + tax_irr)
);

CREATE UNIQUE INDEX uq_credit_pack_orders_authority ON public.credit_pack_orders (authority) WHERE authority IS NOT NULL;
CREATE INDEX idx_credit_pack_orders_tenant_created ON public.credit_pack_orders (tenant_id, created_at DESC);

ALTER TABLE public.bank_transfer_requests
    ADD COLUMN IF NOT EXISTS credit_pack_order_id uuid REFERENCES public.credit_pack_orders(order_id);

CREATE UNIQUE INDEX uq_bank_transfer_credit_pack_order_approved
    ON public.bank_transfer_requests (credit_pack_order_id)
    WHERE credit_pack_order_id IS NOT NULL AND status = 'approved';

CREATE TABLE public.credit_tenant_settings (
    tenant_id uuid PRIMARY KEY REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    overage_enabled boolean NOT NULL DEFAULT false,
    overage_cap_mc bigint,
    updated_by character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_credit_tenant_settings_cap CHECK (overage_cap_mc IS NULL OR overage_cap_mc > 0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.credit_tenant_settings;
DROP INDEX IF EXISTS public.uq_bank_transfer_credit_pack_order_approved;
ALTER TABLE public.bank_transfer_requests DROP COLUMN IF EXISTS credit_pack_order_id;
DROP TABLE IF EXISTS public.credit_pack_orders;
DROP TABLE IF EXISTS public.credit_packs;
-- +goose StatementEnd
