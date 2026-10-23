-- +goose Up
-- Credit price policy (Phase 2d): hourly dollar-rate samples from the existing
-- exchange-rate pipeline, banded list-price revisions, scheduled plan versions,
-- and scheduled rate-card versions. See plans/credit-pricing/phase-2-design.md.

-- +goose StatementBegin
CREATE TABLE public.exchange_rate_samples (
    sample_hour timestamp with time zone PRIMARY KEY,
    source character varying(255) NOT NULL,
    usd_to_irr numeric(18,2) NOT NULL,
    rate_fetched_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_exchange_rate_samples_positive CHECK (usd_to_irr > 0)
);

CREATE TABLE public.credit_price_revisions (
    revision_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    price_toman_per_credit integer NOT NULL,
    previous_price_toman integer,
    base_usd_toman numeric(18,2) NOT NULL,
    observed_avg_usd_toman numeric(18,2),
    band_pct numeric(5,2) NOT NULL DEFAULT 10,
    status character varying(16) NOT NULL,
    effective_from timestamp with time zone,
    plan_effective_from timestamp with time zone,
    notice_sent_at timestamp with time zone,
    proposed_by character varying(255),
    decided_by character varying(255),
    decided_at timestamp with time zone,
    decision_reason text,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_credit_price_revisions_price CHECK (price_toman_per_credit > 0 AND base_usd_toman > 0),
    CONSTRAINT chk_credit_price_revisions_status CHECK (status IN ('proposed', 'active', 'rejected', 'superseded'))
);

CREATE UNIQUE INDEX uq_credit_price_revisions_one_active ON public.credit_price_revisions ((true)) WHERE status = 'active';
CREATE UNIQUE INDEX uq_credit_price_revisions_one_proposed ON public.credit_price_revisions ((true)) WHERE status = 'proposed';
CREATE INDEX idx_credit_price_revisions_effective ON public.credit_price_revisions (effective_from DESC);

INSERT INTO public.credit_price_revisions (
    price_toman_per_credit, base_usd_toman, band_pct, status, effective_from, plan_effective_from,
    proposed_by, decided_by, decided_at, decision_reason
)
VALUES (1000, 270000, 10, 'active', now(), now(), 'migration', 'migration', now(), 'Launch list price from the credit pricing PRD.');

ALTER TABLE public.plan_versions
    ADD COLUMN IF NOT EXISTS credit_price_revision_id uuid REFERENCES public.credit_price_revisions(revision_id);

ALTER TABLE public.credit_rate_card_versions DROP CONSTRAINT IF EXISTS chk_credit_rate_card_versions_status;
ALTER TABLE public.credit_rate_card_versions
    ADD CONSTRAINT chk_credit_rate_card_versions_status CHECK (status IN ('draft', 'scheduled', 'published', 'superseded'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE public.credit_rate_card_versions SET status = 'draft' WHERE status = 'scheduled';
ALTER TABLE public.credit_rate_card_versions DROP CONSTRAINT IF EXISTS chk_credit_rate_card_versions_status;
ALTER TABLE public.credit_rate_card_versions
    ADD CONSTRAINT chk_credit_rate_card_versions_status CHECK (status IN ('draft', 'published', 'superseded'));

ALTER TABLE public.plan_versions DROP COLUMN IF EXISTS credit_price_revision_id;
DROP TABLE IF EXISTS public.credit_price_revisions;
DROP TABLE IF EXISTS public.exchange_rate_samples;
-- +goose StatementEnd
