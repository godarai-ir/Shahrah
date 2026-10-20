-- +goose Up
-- Member seats: a personal credit quota per tenant member for GodarChat, on top
-- of the tenant's shared credit pool. See docs/PRD صندلی‌ها و اعتبار شخصی گدارAI.md.
--
-- Personal credits are ordinary credit lots with an owner_user_id. The tenant
-- balance projection still sums every lot; the pool is that balance less the
-- personal lots, so ledger reconciliation is unchanged.

-- +goose StatementBegin
CREATE TABLE public.seat_types (
    seat_type character varying(16) PRIMARY KEY,
    monthly_grant_mc bigint NOT NULL,
    lifetime boolean NOT NULL DEFAULT false,
    discount_percent numeric(5,2) NOT NULL DEFAULT 0,
    pool_access boolean NOT NULL DEFAULT false,
    rank smallint NOT NULL,
    active boolean NOT NULL DEFAULT true,
    updated_by character varying(255),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_seat_types_type CHECK (seat_type IN ('free', 'pro', 'max')),
    CONSTRAINT chk_seat_types_grant CHECK (monthly_grant_mc > 0),
    CONSTRAINT chk_seat_types_discount CHECK (discount_percent >= 0 AND discount_percent < 100)
);

INSERT INTO public.seat_types (seat_type, monthly_grant_mc, lifetime, discount_percent, pool_access, rank, updated_by)
VALUES
    ('free', 500000, true, 0, false, 1, 'migration'),
    ('pro', 2000000, false, 10, true, 2, 'migration'),
    ('max', 8000000, false, 20, true, 3, 'migration');

-- Paid seats a tenant bought, as blocks paid through a date. A mid-period
-- purchase is one block paid through the subscription's period end; a renewal
-- moves paid_through forward. Expired blocks no longer count.
CREATE TABLE public.tenant_seat_blocks (
    block_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    seat_type character varying(16) NOT NULL REFERENCES public.seat_types(seat_type),
    quantity integer NOT NULL,
    -- renew_quantity, when set, is how many seats of this block the next
    -- renewal keeps; the rest lapse at paid_through.
    renew_quantity integer,
    paid_through timestamp with time zone NOT NULL,
    source_order_id uuid,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_tenant_seat_blocks_type CHECK (seat_type IN ('pro', 'max')),
    CONSTRAINT chk_tenant_seat_blocks_quantity CHECK (quantity > 0),
    CONSTRAINT chk_tenant_seat_blocks_renew CHECK (renew_quantity IS NULL OR (renew_quantity >= 0 AND renew_quantity <= quantity))
);

CREATE INDEX idx_tenant_seat_blocks_tenant ON public.tenant_seat_blocks (tenant_id, seat_type, paid_through);

-- One row per member with a seat; no row means no seat. pending_seat_type is
-- a downgrade that applies at pending_effective_at ('unassigned' removes it).
CREATE TABLE public.tenant_seat_assignments (
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    seat_type character varying(16) NOT NULL REFERENCES public.seat_types(seat_type),
    pending_seat_type character varying(16),
    pending_effective_at timestamp with time zone,
    assigned_by character varying(255),
    assigned_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT tenant_seat_assignments_pkey PRIMARY KEY (tenant_id, user_id),
    CONSTRAINT chk_tenant_seat_assignments_pending CHECK (
        (pending_seat_type IS NULL AND pending_effective_at IS NULL)
        OR (pending_seat_type IN ('pro', 'unassigned') AND pending_effective_at IS NOT NULL)
    )
);

CREATE INDEX idx_tenant_seat_assignments_type ON public.tenant_seat_assignments (tenant_id, seat_type);
CREATE INDEX idx_tenant_seat_assignments_pending
    ON public.tenant_seat_assignments (pending_effective_at) WHERE pending_effective_at IS NOT NULL;

-- A member gets the free seat once per tenant, even after leaving and rejoining.
CREATE TABLE public.tenant_free_seat_grants (
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    granted_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT tenant_free_seat_grants_pkey PRIMARY KEY (tenant_id, user_id)
);

CREATE TABLE public.seat_upgrade_requests (
    request_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    user_id uuid NOT NULL,
    current_seat_type character varying(16) NOT NULL,
    requested_seat_type character varying(16) NOT NULL,
    reason text,
    block_reason character varying(64),
    status character varying(16) NOT NULL DEFAULT 'pending',
    decided_by character varying(255),
    decision_note text,
    decided_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_seat_upgrade_requests_type CHECK (requested_seat_type IN ('pro', 'max')),
    CONSTRAINT chk_seat_upgrade_requests_status CHECK (status IN ('pending', 'approved', 'denied', 'canceled'))
);

CREATE UNIQUE INDEX uq_seat_upgrade_requests_open
    ON public.seat_upgrade_requests (tenant_id, user_id) WHERE status = 'pending';
CREATE INDEX idx_seat_upgrade_requests_tenant ON public.seat_upgrade_requests (tenant_id, status, created_at DESC);

-- Seat purchases. A purchase order adds a block for the rest of the period;
-- a renewal order moves blocks to the next period.
CREATE TABLE public.seat_orders (
    order_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    public_id uuid DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id),
    kind character varying(16) NOT NULL DEFAULT 'purchase',
    seat_type character varying(16) NOT NULL REFERENCES public.seat_types(seat_type),
    quantity integer NOT NULL,
    billing_interval character varying(16) NOT NULL,
    period_start timestamp with time zone NOT NULL,
    period_end timestamp with time zone NOT NULL,
    remaining_days integer NOT NULL,
    period_days integer NOT NULL,
    list_price_irr_per_credit numeric(18,0) NOT NULL,
    unit_period_price_irr numeric(18,0) NOT NULL,
    price_irr numeric(18,0) NOT NULL,
    tax_percent numeric(5,2) NOT NULL DEFAULT 0,
    tax_irr numeric(18,0) NOT NULL DEFAULT 0,
    payable_irr numeric(18,0) NOT NULL,
    -- assign_user_id is the member who gets one of the seats once paid, for an
    -- upgrade that had no free seat in stock.
    assign_user_id uuid,
    upgrade_request_id uuid REFERENCES public.seat_upgrade_requests(request_id),
    renew_block_ids uuid[] NOT NULL DEFAULT '{}',
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
    block_id uuid,
    created_by character varying(255),
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_seat_orders_kind CHECK (kind IN ('purchase', 'renewal')),
    CONSTRAINT chk_seat_orders_type CHECK (seat_type IN ('pro', 'max')),
    CONSTRAINT chk_seat_orders_quantity CHECK (quantity > 0 AND quantity <= 1000),
    CONSTRAINT chk_seat_orders_status CHECK (status IN ('PENDING', 'AWAITING_TRANSFER', 'COMPLETED', 'FAILED', 'CANCELED')),
    CONSTRAINT chk_seat_orders_amounts CHECK (price_irr > 0 AND tax_irr >= 0 AND payable_irr = price_irr + tax_irr)
);

CREATE UNIQUE INDEX uq_seat_orders_authority ON public.seat_orders (authority) WHERE authority IS NOT NULL;
CREATE INDEX idx_seat_orders_tenant_created ON public.seat_orders (tenant_id, created_at DESC);

ALTER TABLE public.bank_transfer_requests
    ADD COLUMN IF NOT EXISTS seat_order_id uuid REFERENCES public.seat_orders(order_id);

CREATE UNIQUE INDEX uq_bank_transfer_seat_order_approved
    ON public.bank_transfer_requests (seat_order_id)
    WHERE seat_order_id IS NOT NULL AND status = 'approved';

-- Invitations carry the seat the member gets on acceptance.
ALTER TABLE public.tenant_admin_invitations
    ADD COLUMN IF NOT EXISTS seat_type character varying(16);

-- Personal credit lots.
ALTER TABLE public.credit_lots ADD COLUMN IF NOT EXISTS owner_user_id uuid;
ALTER TABLE public.credit_lots DROP CONSTRAINT chk_credit_lots_source;
ALTER TABLE public.credit_lots ADD CONSTRAINT chk_credit_lots_source CHECK (source IN (
    'plan', 'trial', 'promo', 'purchase_addon', 'migration', 'compensation', 'adjust', 'overage', 'seat', 'seat_free'
));
ALTER TABLE public.credit_lots ADD CONSTRAINT chk_credit_lots_owner CHECK (
    (source IN ('seat', 'seat_free')) = (owner_user_id IS NOT NULL)
);

ALTER TABLE public.credit_lot_balances ADD COLUMN IF NOT EXISTS owner_user_id uuid;

CREATE INDEX idx_credit_lot_balances_owner_spendable
    ON public.credit_lot_balances (tenant_id, owner_user_id, expires_at ASC NULLS LAST, source_rank, lot_created_at)
    WHERE remaining_mc > 0 AND owner_user_id IS NOT NULL;

-- Consumption settled from a member's personal lots per calendar-month window,
-- the counterpart of credit_budget_window_settlements for seat windows. The
-- pool window's settled_mc no longer includes personal-lot consumption.
CREATE TABLE public.credit_user_window_settlements (
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    window_id character varying(16) NOT NULL,
    settled_mc bigint NOT NULL DEFAULT 0,
    pool_settled_mc bigint NOT NULL DEFAULT 0,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT credit_user_window_settlements_pkey PRIMARY KEY (tenant_id, user_id, window_id)
);

ALTER TABLE public.credit_tenant_settings
    ADD COLUMN IF NOT EXISTS seat_upgrade_requests_enabled boolean NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS seat_pool_limit_pro_mc bigint,
    ADD COLUMN IF NOT EXISTS seat_pool_limit_max_mc bigint,
    ADD COLUMN IF NOT EXISTS seats_enforced_from timestamp with time zone;

ALTER TABLE public.credit_tenant_settings
    ADD CONSTRAINT chk_credit_tenant_settings_seat_limits CHECK (
        (seat_pool_limit_pro_mc IS NULL OR seat_pool_limit_pro_mc >= 0)
        AND (seat_pool_limit_max_mc IS NULL OR seat_pool_limit_max_mc >= 0)
    );

-- Existing members of credit tenants get their free seat now. Seat limits
-- start with the tenant's next billing period, so nobody's chat stops mid-period.
INSERT INTO public.tenant_seat_assignments (tenant_id, user_id, seat_type, assigned_by, assigned_at)
SELECT u.tenant_id, u.user_id, 'free', 'migration', now()
FROM public.users u
JOIN public.tenants t ON t.tenant_id = u.tenant_id AND t.billing_unit = 'credit'
WHERE u.tenant_id IS NOT NULL AND COALESCE(u.active, true)
ON CONFLICT DO NOTHING;

INSERT INTO public.tenant_free_seat_grants (tenant_id, user_id, granted_at)
SELECT tenant_id, user_id, assigned_at FROM public.tenant_seat_assignments WHERE seat_type = 'free'
ON CONFLICT DO NOTHING;

INSERT INTO public.credit_tenant_settings (tenant_id, seats_enforced_from, updated_by)
SELECT ts.tenant_id, ts.current_period_end, 'migration'
FROM public.tenant_subscriptions ts
JOIN public.tenants t ON t.tenant_id = ts.tenant_id AND t.billing_unit = 'credit'
WHERE ts.status = 'active' AND ts.current_period_end > now()
ON CONFLICT (tenant_id) DO UPDATE
SET seats_enforced_from = EXCLUDED.seats_enforced_from, updated_at = now();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.credit_tenant_settings
    DROP CONSTRAINT IF EXISTS chk_credit_tenant_settings_seat_limits,
    DROP COLUMN IF EXISTS seats_enforced_from,
    DROP COLUMN IF EXISTS seat_pool_limit_max_mc,
    DROP COLUMN IF EXISTS seat_pool_limit_pro_mc,
    DROP COLUMN IF EXISTS seat_upgrade_requests_enabled;
DROP TABLE IF EXISTS public.credit_user_window_settlements;
DROP INDEX IF EXISTS public.idx_credit_lot_balances_owner_spendable;
ALTER TABLE public.credit_lot_balances DROP COLUMN IF EXISTS owner_user_id;
ALTER TABLE public.credit_lots DROP CONSTRAINT IF EXISTS chk_credit_lots_owner;
-- Personal lots cannot be rolled back without rewriting the append-only ledger.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.credit_lots WHERE owner_user_id IS NOT NULL) THEN
        RAISE EXCEPTION 'migration 077_seats cannot be rolled back after seat credits were granted';
    END IF;
END
$$;
ALTER TABLE public.credit_lots DROP CONSTRAINT IF EXISTS chk_credit_lots_source;
ALTER TABLE public.credit_lots ADD CONSTRAINT chk_credit_lots_source CHECK (source IN (
    'plan', 'trial', 'promo', 'purchase_addon', 'migration', 'compensation', 'adjust', 'overage'
));
ALTER TABLE public.credit_lots DROP COLUMN IF EXISTS owner_user_id;
ALTER TABLE public.tenant_admin_invitations DROP COLUMN IF EXISTS seat_type;
DROP INDEX IF EXISTS public.uq_bank_transfer_seat_order_approved;
ALTER TABLE public.bank_transfer_requests DROP COLUMN IF EXISTS seat_order_id;
DROP TABLE IF EXISTS public.seat_orders;
DROP TABLE IF EXISTS public.seat_upgrade_requests;
DROP TABLE IF EXISTS public.tenant_free_seat_grants;
DROP TABLE IF EXISTS public.tenant_seat_assignments;
DROP TABLE IF EXISTS public.tenant_seat_blocks;
DROP TABLE IF EXISTS public.seat_types;
-- +goose StatementEnd
