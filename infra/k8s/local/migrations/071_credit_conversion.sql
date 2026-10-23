-- +goose Up
-- Credit cutover (Phase 3e): converting a tenant's IRR wallet into credits and
-- back. A conversion is a ledger entry of kind 'convert' plus one
-- credit_conversions row that records the rate it used.
-- See plans/credit-pricing/phase-3-design.md.

-- +goose StatementBegin
ALTER TABLE public.credit_ledger DROP CONSTRAINT chk_credit_ledger_kind;
ALTER TABLE public.credit_ledger DROP CONSTRAINT chk_credit_ledger_sign;
ALTER TABLE public.credit_ledger DROP CONSTRAINT chk_credit_ledger_manual_reason;

ALTER TABLE public.credit_ledger
    ADD CONSTRAINT chk_credit_ledger_kind CHECK (kind IN (
        'purchase', 'plan_grant', 'promo_grant', 'consume', 'refund', 'expire', 'adjust', 'convert'
    )),
    ADD CONSTRAINT chk_credit_ledger_sign CHECK (
        (kind IN ('purchase', 'plan_grant', 'promo_grant', 'refund') AND amount_mc > 0)
        OR (kind IN ('consume', 'expire') AND amount_mc <= 0)
        OR (kind IN ('adjust', 'convert') AND amount_mc <> 0)
    ),
    ADD CONSTRAINT chk_credit_ledger_manual_reason CHECK (
        kind NOT IN ('adjust', 'refund', 'convert') OR (reason IS NOT NULL AND btrim(reason) <> '')
    );

CREATE TABLE public.credit_conversions (
    conversion_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id),
    sequence integer NOT NULL,
    direction character varying(16) NOT NULL,
    price_toman_per_credit bigint NOT NULL,
    price_revision_id uuid,
    wallet_irr bigint NOT NULL,
    converted_irr bigint NOT NULL,
    residue_irr bigint NOT NULL,
    credits_mc bigint NOT NULL,
    debt_mc bigint NOT NULL DEFAULT 0,
    uncollected_irr bigint NOT NULL DEFAULT 0,
    rules_converted integer NOT NULL DEFAULT 0,
    previous_subscription_id uuid,
    subscription_id uuid,
    plan_slug character varying(64),
    actor character varying(255) NOT NULL,
    reason text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT uq_credit_conversions_sequence UNIQUE (tenant_id, sequence),
    CONSTRAINT chk_credit_conversions_direction CHECK (direction IN ('to_credit', 'to_irr')),
    CONSTRAINT chk_credit_conversions_amounts CHECK (
        price_toman_per_credit > 0 AND wallet_irr >= 0 AND converted_irr >= 0 AND residue_irr >= 0
        AND credits_mc >= 0 AND debt_mc >= 0 AND uncollected_irr >= 0
    ),
    CONSTRAINT chk_credit_conversions_reason CHECK (btrim(reason) <> '')
);

CREATE INDEX idx_credit_conversions_created ON public.credit_conversions (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.credit_conversions;

-- Down is refused while convert entries exist: the ledger is append-only, so
-- they cannot be removed to satisfy the narrower constraints.
ALTER TABLE public.credit_ledger DROP CONSTRAINT chk_credit_ledger_manual_reason;
ALTER TABLE public.credit_ledger DROP CONSTRAINT chk_credit_ledger_sign;
ALTER TABLE public.credit_ledger DROP CONSTRAINT chk_credit_ledger_kind;

ALTER TABLE public.credit_ledger
    ADD CONSTRAINT chk_credit_ledger_kind CHECK (kind IN (
        'purchase', 'plan_grant', 'promo_grant', 'consume', 'refund', 'expire', 'adjust'
    )),
    ADD CONSTRAINT chk_credit_ledger_sign CHECK (
        (kind IN ('purchase', 'plan_grant', 'promo_grant', 'refund') AND amount_mc > 0)
        OR (kind IN ('consume', 'expire') AND amount_mc <= 0)
        OR (kind = 'adjust' AND amount_mc <> 0)
    ),
    ADD CONSTRAINT chk_credit_ledger_manual_reason CHECK (
        kind NOT IN ('adjust', 'refund') OR (reason IS NOT NULL AND btrim(reason) <> '')
    );
-- +goose StatementEnd
