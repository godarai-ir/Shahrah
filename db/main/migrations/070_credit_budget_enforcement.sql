-- +goose Up
-- Credit enforcement (Phase 3a): budget rules in credits with an explicit
-- exceed behavior, and the per-window record of ledger-settled consumption that
-- keeps the credit pool window's allowance exact.
-- See plans/credit-pricing/phase-3-design.md.

-- +goose StatementBegin
ALTER TABLE public.budget_limit_rules
    ADD COLUMN limit_mc bigint,
    ADD COLUMN on_exceed character varying(16) NOT NULL DEFAULT 'block',
    ADD COLUMN downgrade_model character varying(255),
    ADD COLUMN downgrade_daily_limit_mc bigint,
    ADD COLUMN alert_thresholds jsonb NOT NULL DEFAULT '[80, 100]'::jsonb;

-- Rules that did not block before still do not block.
UPDATE public.budget_limit_rules
SET on_exceed = CASE WHEN block_on_budget_exceed THEN 'block' ELSE 'alert' END;

ALTER TABLE public.budget_limit_rules
    ADD CONSTRAINT chk_budget_limit_rules_on_exceed CHECK (on_exceed IN ('block', 'downgrade', 'alert')),
    ADD CONSTRAINT chk_budget_limit_rules_limit_mc CHECK (limit_mc IS NULL OR limit_mc > 0),
    ADD CONSTRAINT chk_budget_limit_rules_downgrade CHECK (
        on_exceed <> 'downgrade' OR (downgrade_model IS NOT NULL AND btrim(downgrade_model) <> '')
    ),
    ADD CONSTRAINT chk_budget_limit_rules_downgrade_limit CHECK (
        downgrade_daily_limit_mc IS NULL OR downgrade_daily_limit_mc > 0
    ),
    ADD CONSTRAINT chk_budget_limit_rules_alert_thresholds CHECK (jsonb_typeof(alert_thresholds) = 'array');

CREATE TABLE public.credit_budget_window_settlements (
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    window_id character varying(16) NOT NULL,
    settled_mc bigint NOT NULL DEFAULT 0,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, window_id),
    CONSTRAINT chk_credit_budget_window_settlements_settled CHECK (settled_mc >= 0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.credit_budget_window_settlements;

ALTER TABLE public.budget_limit_rules
    DROP CONSTRAINT IF EXISTS chk_budget_limit_rules_alert_thresholds,
    DROP CONSTRAINT IF EXISTS chk_budget_limit_rules_downgrade_limit,
    DROP CONSTRAINT IF EXISTS chk_budget_limit_rules_downgrade,
    DROP CONSTRAINT IF EXISTS chk_budget_limit_rules_limit_mc,
    DROP CONSTRAINT IF EXISTS chk_budget_limit_rules_on_exceed;

ALTER TABLE public.budget_limit_rules
    DROP COLUMN IF EXISTS alert_thresholds,
    DROP COLUMN IF EXISTS downgrade_daily_limit_mc,
    DROP COLUMN IF EXISTS downgrade_model,
    DROP COLUMN IF EXISTS on_exceed,
    DROP COLUMN IF EXISTS limit_mc;
-- +goose StatementEnd
