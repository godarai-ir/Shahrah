-- +goose Up
-- Credit plans (Phase 2b). This migration only seeds DRAFT plan versions:
-- publishing changes what new purchases pay, so a support-admin publishes them.
-- Prices are IRR (10 × the PRD's toman prices) and exclude VAT.
-- See plans/credit-pricing/phase-2-design.md.

-- +goose StatementBegin
INSERT INTO public.feature_catalog (feature_key, value_type, default_value, allow_override, category, description)
VALUES
    ('credits.monthly_grant', 'integer', '0'::jsonb, true, 'credits', 'Credits granted each month of a subscription period.'),
    ('credits.bonus_percent', 'decimal', '0'::jsonb, true, 'credits', 'Bonus credits over the list price, for display.'),
    ('credits.overage_allowed', 'boolean', 'false'::jsonb, true, 'credits', 'Whether the tenant admin may enable overage billing.'),
    ('credits.rollover_allowed', 'boolean', 'false'::jsonb, true, 'credits', 'Whether unused plan credits roll over on renewal.'),
    ('credits.byok_fee_percent', 'decimal', '5'::jsonb, true, 'credits', 'Platform fee on externally billed (BYOK) usage, in percent of the credit equivalent.'),
    ('credits.trial_grant', 'integer', '500'::jsonb, true, 'credits', 'Credits granted once for a trial.'),
    ('credits.trial_days', 'integer', '14'::jsonb, true, 'credits', 'Days before trial credits expire.')
ON CONFLICT (feature_key) DO NOTHING;

INSERT INTO public.commercial_plans (slug, display_name, description, sort_order, active)
VALUES ('business', 'Business', 'Plan for larger organizations with up to 100 members.', 35, true)
ON CONFLICT (slug) DO NOTHING;

ALTER TABLE public.commercial_signup_settings
    ADD COLUMN IF NOT EXISTS trial_credit_mc bigint NOT NULL DEFAULT 500000,
    ADD COLUMN IF NOT EXISTS trial_credit_days integer NOT NULL DEFAULT 14,
    ADD COLUMN IF NOT EXISTS trial_max_users integer NOT NULL DEFAULT 3;

ALTER TABLE public.commercial_signup_settings
    ADD CONSTRAINT chk_commercial_signup_trial_credit CHECK (
        trial_credit_mc >= 0 AND trial_credit_days > 0 AND trial_credit_days <= 365 AND trial_max_users > 0
    );

DO $seed$
DECLARE
    offer record;
    target_plan uuid;
    source_version uuid;
    draft uuid;
    next_version integer;
BEGIN
    FOR offer IN
        SELECT *
        FROM (VALUES
            ('developer', 'developer', '3', '5000', '0', 'true', 'false', 50000000::numeric, 540000000::numeric),
            ('team', 'team', '10', '16500', '10', 'true', 'false', 150000000::numeric, 1620000000::numeric),
            ('growth', 'growth', '30', '54000', '20', 'true', 'false', 450000000::numeric, 4860000000::numeric),
            ('business', 'growth', '100', '190000', '26.67', 'true', 'false', 1500000000::numeric, 16200000000::numeric),
            ('enterprise', 'enterprise', 'null', 'null', '0', 'true', 'true', NULL::numeric, NULL::numeric)
        ) AS o(slug, copy_from, max_users, monthly_credits, bonus_percent, overage_allowed, rollover_allowed, monthly_irr, yearly_irr)
    LOOP
        SELECT plan_id INTO target_plan FROM public.commercial_plans WHERE slug = offer.slug;

        SELECT pv.version_id INTO source_version
        FROM public.plan_versions pv
        JOIN public.commercial_plans p ON p.plan_id = pv.plan_id
        WHERE p.slug = offer.copy_from AND pv.status = 'published';

        SELECT COALESCE(MAX(version), 0) + 1 INTO next_version
        FROM public.plan_versions WHERE plan_id = target_plan;

        INSERT INTO public.plan_versions (plan_id, version, status)
        VALUES (target_plan, next_version, 'draft')
        RETURNING version_id INTO draft;

        -- Keep every non-credit entitlement of the plan (Business starts from Growth).
        IF source_version IS NOT NULL THEN
            INSERT INTO public.plan_entitlements (version_id, feature_key, value_json)
            SELECT draft, feature_key, value_json
            FROM public.plan_entitlements
            WHERE version_id = source_version;
        END IF;

        INSERT INTO public.plan_entitlements (version_id, feature_key, value_json)
        VALUES
            (draft, 'users.max', offer.max_users::jsonb),
            (draft, 'credits.monthly_grant', offer.monthly_credits::jsonb),
            (draft, 'credits.bonus_percent', offer.bonus_percent::jsonb),
            (draft, 'credits.overage_allowed', offer.overage_allowed::jsonb),
            (draft, 'credits.rollover_allowed', offer.rollover_allowed::jsonb),
            (draft, 'credits.byok_fee_percent', '5'::jsonb)
        ON CONFLICT (version_id, feature_key) DO UPDATE SET value_json = EXCLUDED.value_json, updated_at = NOW();

        INSERT INTO public.plan_prices (version_id, billing_interval, currency, amount_minor, discount_percent)
        VALUES
            (draft, 'monthly', 'IRR', offer.monthly_irr, 0),
            (draft, 'yearly', 'IRR', offer.yearly_irr, CASE WHEN offer.yearly_irr IS NULL THEN 0 ELSE 10 END);
    END LOOP;
END
$seed$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove the seeded drafts that were never published or used.
DELETE FROM public.plan_versions pv
WHERE pv.status = 'draft'
  AND EXISTS (
      SELECT 1 FROM public.plan_entitlements pe
      WHERE pe.version_id = pv.version_id AND pe.feature_key = 'credits.monthly_grant'
  )
  AND NOT EXISTS (SELECT 1 FROM public.tenant_subscriptions ts WHERE ts.version_id = pv.version_id OR ts.pending_version_id = pv.version_id)
  AND NOT EXISTS (SELECT 1 FROM public.subscription_payment_intents spi WHERE spi.version_id = pv.version_id);

DELETE FROM public.commercial_plans p
WHERE p.slug = 'business'
  AND NOT EXISTS (SELECT 1 FROM public.plan_versions pv WHERE pv.plan_id = p.plan_id);

DELETE FROM public.feature_catalog
WHERE feature_key IN (
    'credits.monthly_grant', 'credits.bonus_percent', 'credits.overage_allowed', 'credits.rollover_allowed',
    'credits.byok_fee_percent', 'credits.trial_grant', 'credits.trial_days'
);

ALTER TABLE public.commercial_signup_settings DROP CONSTRAINT IF EXISTS chk_commercial_signup_trial_credit;
ALTER TABLE public.commercial_signup_settings
    DROP COLUMN IF EXISTS trial_max_users,
    DROP COLUMN IF EXISTS trial_credit_days,
    DROP COLUMN IF EXISTS trial_credit_mc;
-- +goose StatementEnd
