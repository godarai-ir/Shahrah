-- +goose Up
CREATE TABLE IF NOT EXISTS public.commercial_signup_settings (
    settings_id boolean PRIMARY KEY DEFAULT true CHECK (settings_id),
    trial_days integer NOT NULL DEFAULT 15 CHECK (trial_days > 0 AND trial_days <= 365),
    signup_credit_irr numeric(18,0) NOT NULL DEFAULT 2000000 CHECK (signup_credit_irr >= 0),
    updated_by character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

INSERT INTO public.commercial_signup_settings (settings_id, trial_days, signup_credit_irr)
VALUES (true, 15, 2000000)
ON CONFLICT (settings_id) DO NOTHING;

ALTER TABLE public.tenant_subscriptions
    ADD COLUMN IF NOT EXISTS trial_started_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS trial_ends_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS converted_at timestamp with time zone;

ALTER TABLE public.tenant_subscriptions DROP CONSTRAINT IF EXISTS chk_tenant_subscriptions_status;
ALTER TABLE public.tenant_subscriptions
    ADD CONSTRAINT chk_tenant_subscriptions_status
    CHECK (status IN ('trialing', 'active', 'past_due', 'canceled', 'expired'));

DROP INDEX IF EXISTS public.uq_tenant_subscriptions_active;
CREATE UNIQUE INDEX uq_tenant_subscriptions_active
    ON public.tenant_subscriptions (tenant_id)
    WHERE status IN ('trialing', 'active', 'past_due');

-- +goose Down
UPDATE public.tenant_subscriptions SET status = 'expired' WHERE status = 'trialing';
DROP INDEX IF EXISTS public.uq_tenant_subscriptions_active;
CREATE UNIQUE INDEX uq_tenant_subscriptions_active
    ON public.tenant_subscriptions (tenant_id)
    WHERE status IN ('active', 'past_due');

ALTER TABLE public.tenant_subscriptions DROP CONSTRAINT IF EXISTS chk_tenant_subscriptions_status;
ALTER TABLE public.tenant_subscriptions
    ADD CONSTRAINT chk_tenant_subscriptions_status
    CHECK (status IN ('active', 'past_due', 'canceled', 'expired'));
ALTER TABLE public.tenant_subscriptions
    DROP COLUMN IF EXISTS converted_at,
    DROP COLUMN IF EXISTS trial_ends_at,
    DROP COLUMN IF EXISTS trial_started_at;
DROP TABLE IF EXISTS public.commercial_signup_settings;
