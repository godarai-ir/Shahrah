-- +goose Up
-- Credit ledger (Phase 2): lots, an append-only ledger with allocations, and
-- rebuildable balance projections. Amounts are millicredits (1 credit = 1,000 mc).
-- See plans/credit-pricing/phase-2-design.md.

-- +goose StatementBegin
ALTER TABLE public.tenants
    ADD COLUMN IF NOT EXISTS billing_unit character varying(16) NOT NULL DEFAULT 'irr';

ALTER TABLE public.tenants
    ADD CONSTRAINT chk_tenants_billing_unit CHECK (billing_unit IN ('irr', 'credit'));

ALTER TABLE public.spend_logs
    ADD COLUMN IF NOT EXISTS org_unit_id uuid;

CREATE TABLE public.credit_lots (
    lot_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id),
    source character varying(32) NOT NULL,
    amount_mc bigint NOT NULL,
    expires_at timestamp with time zone,
    purchase_price_irr numeric(18,0),
    price_per_credit_irr numeric(18,4),
    source_ref jsonb NOT NULL DEFAULT '{}'::jsonb,
    grant_entry_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_credit_lots_source CHECK (source IN (
        'plan', 'trial', 'promo', 'purchase_addon', 'migration', 'compensation', 'adjust', 'overage'
    )),
    CONSTRAINT chk_credit_lots_amount CHECK (amount_mc > 0)
);

CREATE INDEX idx_credit_lots_tenant_created ON public.credit_lots (tenant_id, created_at DESC);

CREATE TABLE public.credit_ledger (
    entry_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    kind character varying(32) NOT NULL,
    amount_mc bigint NOT NULL,
    idempotency_key character varying(200) NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    lot_id uuid,
    request_id character varying(255),
    trace_id character varying(255),
    model character varying(255),
    user_id uuid,
    team_id uuid,
    org_unit_id uuid,
    product character varying(32),
    billing_mode character varying(32),
    pricing_revision character varying(128),
    rate_card_version integer,
    internal_cost_usd character varying(64),
    credit_equivalent_mc bigint,
    spend_log_id bigint,
    reason text,
    actor character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT credit_ledger_pkey PRIMARY KEY (entry_id, occurred_at),
    CONSTRAINT chk_credit_ledger_kind CHECK (kind IN (
        'purchase', 'plan_grant', 'promo_grant', 'consume', 'refund', 'expire', 'adjust'
    )),
    CONSTRAINT chk_credit_ledger_sign CHECK (
        (kind IN ('purchase', 'plan_grant', 'promo_grant', 'refund') AND amount_mc > 0)
        OR (kind IN ('consume', 'expire') AND amount_mc <= 0)
        OR (kind = 'adjust' AND amount_mc <> 0)
    ),
    CONSTRAINT chk_credit_ledger_manual_reason CHECK (
        kind NOT IN ('adjust', 'refund') OR (reason IS NOT NULL AND btrim(reason) <> '')
    )
) PARTITION BY RANGE (occurred_at);

CREATE INDEX idx_credit_ledger_tenant_occurred ON public.credit_ledger (tenant_id, occurred_at DESC);

-- Unique indexes on a partitioned table must include the partition key, so the
-- ledger's idempotency guarantee lives here and is written in the same transaction.
CREATE TABLE public.credit_ledger_keys (
    idempotency_key character varying(200) PRIMARY KEY,
    entry_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    occurred_at timestamp with time zone NOT NULL
);

CREATE TABLE public.credit_ledger_allocations (
    allocation_id uuid NOT NULL,
    entry_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    lot_id uuid REFERENCES public.credit_lots(lot_id),
    amount_mc bigint NOT NULL,
    occurred_at timestamp with time zone NOT NULL,
    CONSTRAINT credit_ledger_allocations_pkey PRIMARY KEY (allocation_id, occurred_at),
    CONSTRAINT chk_credit_ledger_allocations_amount CHECK (amount_mc <> 0)
) PARTITION BY RANGE (occurred_at);

CREATE INDEX idx_credit_allocations_entry ON public.credit_ledger_allocations (entry_id);
CREATE INDEX idx_credit_allocations_tenant_lot ON public.credit_ledger_allocations (tenant_id, lot_id);

CREATE TABLE public.credit_lot_balances (
    lot_id uuid PRIMARY KEY REFERENCES public.credit_lots(lot_id),
    tenant_id uuid NOT NULL,
    remaining_mc bigint NOT NULL,
    expires_at timestamp with time zone,
    source_rank smallint NOT NULL,
    lot_created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_credit_lot_balances_remaining CHECK (remaining_mc >= 0)
);

CREATE INDEX idx_credit_lot_balances_spendable
    ON public.credit_lot_balances (tenant_id, expires_at ASC NULLS LAST, source_rank, lot_created_at)
    WHERE remaining_mc > 0;

CREATE INDEX idx_credit_lot_balances_expiring
    ON public.credit_lot_balances (expires_at)
    WHERE remaining_mc > 0 AND expires_at IS NOT NULL;

CREATE TABLE public.credit_tenant_balances (
    tenant_id uuid PRIMARY KEY REFERENCES public.tenants(tenant_id),
    available_mc bigint NOT NULL DEFAULT 0,
    debt_mc bigint NOT NULL DEFAULT 0,
    ledger_sum_mc bigint NOT NULL DEFAULT 0,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_credit_tenant_balances_non_negative CHECK (available_mc >= 0 AND debt_mc >= 0)
);

CREATE TABLE public.credit_settlement_cursors (
    cursor_id character varying(64) PRIMARY KEY,
    last_spend_log_id bigint NOT NULL DEFAULT 0,
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

-- Row UPDATE and DELETE are always refused. TRUNCATE is refused unless the
-- session explicitly opts in with SET LOCAL shahrah.credit_ledger_maintenance = 'on',
-- which only disposable test databases do; it stops an accidental
-- "TRUNCATE tenants CASCADE" from wiping lots and allocations.
CREATE FUNCTION public.credit_ledger_forbid_mutation() RETURNS trigger
    LANGUAGE plpgsql AS $fn$
BEGIN
    IF TG_OP = 'TRUNCATE' AND current_setting('shahrah.credit_ledger_maintenance', true) = 'on' THEN
        RETURN NULL;
    END IF;

    RAISE EXCEPTION 'credit ledger table % is append-only (% refused)', TG_TABLE_NAME, TG_OP
        USING ERRCODE = 'restrict_violation';
END;
$fn$;

CREATE TRIGGER trg_credit_ledger_append_only
    BEFORE UPDATE OR DELETE ON public.credit_ledger
    FOR EACH ROW EXECUTE FUNCTION public.credit_ledger_forbid_mutation();
CREATE TRIGGER trg_credit_ledger_no_truncate
    BEFORE TRUNCATE ON public.credit_ledger
    FOR EACH STATEMENT EXECUTE FUNCTION public.credit_ledger_forbid_mutation();

CREATE TRIGGER trg_credit_allocations_append_only
    BEFORE UPDATE OR DELETE ON public.credit_ledger_allocations
    FOR EACH ROW EXECUTE FUNCTION public.credit_ledger_forbid_mutation();
CREATE TRIGGER trg_credit_allocations_no_truncate
    BEFORE TRUNCATE ON public.credit_ledger_allocations
    FOR EACH STATEMENT EXECUTE FUNCTION public.credit_ledger_forbid_mutation();

CREATE TRIGGER trg_credit_ledger_keys_append_only
    BEFORE UPDATE OR DELETE ON public.credit_ledger_keys
    FOR EACH ROW EXECUTE FUNCTION public.credit_ledger_forbid_mutation();
CREATE TRIGGER trg_credit_ledger_keys_no_truncate
    BEFORE TRUNCATE ON public.credit_ledger_keys
    FOR EACH STATEMENT EXECUTE FUNCTION public.credit_ledger_forbid_mutation();

CREATE TRIGGER trg_credit_lots_append_only
    BEFORE UPDATE OR DELETE ON public.credit_lots
    FOR EACH ROW EXECUTE FUNCTION public.credit_ledger_forbid_mutation();
CREATE TRIGGER trg_credit_lots_no_truncate
    BEFORE TRUNCATE ON public.credit_lots
    FOR EACH STATEMENT EXECUTE FUNCTION public.credit_ledger_forbid_mutation();

-- Creates the monthly partitions of the ledger and allocation tables from the
-- month before `from_ts` through `months_ahead` months after it. Idempotent.
CREATE FUNCTION public.credit_ledger_ensure_partitions(from_ts timestamp with time zone, months_ahead integer)
    RETURNS integer
    LANGUAGE plpgsql AS $fn$
DECLARE
    month_start date;
    month_end date;
    parent text;
    partition_name text;
    created integer := 0;
BEGIN
    FOR i IN -1..months_ahead LOOP
        month_start := (date_trunc('month', from_ts AT TIME ZONE 'UTC') + make_interval(months => i))::date;
        month_end := (month_start + interval '1 month')::date;

        FOREACH parent IN ARRAY ARRAY['credit_ledger', 'credit_ledger_allocations'] LOOP
            partition_name := format('%s_y%sm%s', parent, to_char(month_start, 'YYYY'), to_char(month_start, 'MM'));

            IF to_regclass('public.' || partition_name) IS NULL THEN
                EXECUTE format(
                    'CREATE TABLE public.%I PARTITION OF public.%I FOR VALUES FROM (%L) TO (%L)',
                    partition_name, parent,
                    (month_start::timestamp AT TIME ZONE 'UTC'), (month_end::timestamp AT TIME ZONE 'UTC')
                );
                created := created + 1;
            END IF;
        END LOOP;
    END LOOP;

    RETURN created;
END;
$fn$;

CREATE TABLE public.credit_ledger_default PARTITION OF public.credit_ledger DEFAULT;
CREATE TABLE public.credit_ledger_allocations_default PARTITION OF public.credit_ledger_allocations DEFAULT;

SELECT public.credit_ledger_ensure_partitions(now(), 3);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.credit_settlement_cursors;
DROP TABLE IF EXISTS public.credit_tenant_balances;
DROP TABLE IF EXISTS public.credit_lot_balances;
DROP TABLE IF EXISTS public.credit_ledger_allocations;
DROP TABLE IF EXISTS public.credit_ledger_keys;
DROP TABLE IF EXISTS public.credit_ledger;
DROP TABLE IF EXISTS public.credit_lots;
DROP FUNCTION IF EXISTS public.credit_ledger_ensure_partitions(timestamp with time zone, integer);
DROP FUNCTION IF EXISTS public.credit_ledger_forbid_mutation();

ALTER TABLE public.spend_logs DROP COLUMN IF EXISTS org_unit_id;

ALTER TABLE public.tenants DROP CONSTRAINT IF EXISTS chk_tenants_billing_unit;
ALTER TABLE public.tenants DROP COLUMN IF EXISTS billing_unit;
-- +goose StatementEnd
