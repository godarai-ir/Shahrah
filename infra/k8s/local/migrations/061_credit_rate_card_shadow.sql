-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.credit_rate_card_versions (
    version integer PRIMARY KEY CHECK (version > 0),
    status character varying(16) NOT NULL DEFAULT 'draft',
    effective_from timestamp with time zone NOT NULL DEFAULT now(),
    notice_until timestamp with time zone,
    changelog_fa text NOT NULL DEFAULT '',
    changelog_en text NOT NULL DEFAULT '',
    published_by character varying(255),
    published_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_credit_rate_card_versions_status CHECK (status IN ('draft', 'published', 'superseded'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_credit_rate_card_versions_one_published
    ON public.credit_rate_card_versions ((true))
    WHERE status = 'published';

CREATE TABLE IF NOT EXISTS public.credit_model_multipliers (
    version integer NOT NULL REFERENCES public.credit_rate_card_versions(version) ON DELETE CASCADE,
    model character varying(255) NOT NULL,
    multiplier numeric(10,4) NOT NULL,
    reason text NOT NULL,
    created_by character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (version, model),
    CONSTRAINT chk_credit_model_multipliers_positive CHECK (multiplier > 0),
    CONSTRAINT chk_credit_model_multipliers_reason CHECK (length(btrim(reason)) > 0),
    CONSTRAINT chk_credit_model_multipliers_model CHECK (model = lower(btrim(model)) AND model <> '')
);

INSERT INTO public.credit_rate_card_versions (
    version, status, effective_from, changelog_fa, changelog_en, published_by, published_at
)
VALUES (
    1, 'published', now(),
    'نسخه‌ی اول نرخ‌نامه‌ی کردیتی: نرخ هر مدل برابر قیمت رسمی provider است.',
    'Initial credit rate card: every model is priced at the official provider rate.',
    'system', now()
)
ON CONFLICT (version) DO NOTHING;

ALTER TABLE public.spend_logs
    ADD COLUMN IF NOT EXISTS credits_mc bigint,
    ADD COLUMN IF NOT EXISTS credits_avoided_mc bigint,
    ADD COLUMN IF NOT EXISTS rate_card_version integer,
    ADD COLUMN IF NOT EXISTS product character varying(32),
    ADD COLUMN IF NOT EXISTS billing_mode character varying(32),
    ADD COLUMN IF NOT EXISTS credit_source character varying(32);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.spend_logs
    DROP COLUMN IF EXISTS credits_mc,
    DROP COLUMN IF EXISTS credits_avoided_mc,
    DROP COLUMN IF EXISTS rate_card_version,
    DROP COLUMN IF EXISTS product,
    DROP COLUMN IF EXISTS billing_mode,
    DROP COLUMN IF EXISTS credit_source;

DROP TABLE IF EXISTS public.credit_model_multipliers;
DROP TABLE IF EXISTS public.credit_rate_card_versions;
-- +goose StatementEnd
