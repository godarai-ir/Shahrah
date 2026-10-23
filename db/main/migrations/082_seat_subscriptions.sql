-- +goose Up
-- Seat subscriptions replace commercial plans (product decision, 2026-10-22).
--
-- * Two internal plans are left: `free` (every tenant without paid seats; always
--   active, never expires) and `seats` (the subscription that the first paid
--   seat purchase starts). Both enable every feature, grant no pool credits and
--   have no users.max: free seats are capped at five in code, and every paid
--   seat adds one member.
-- * Every other plan is retired: its published version is archived and the
--   plan can no longer be bought.
-- * Trials end now. A tenant that holds paid seats moves to a `seats`
--   subscription that runs until its seats are paid through; every other
--   tenant moves to `free`. Plan credits already granted expire with their
--   slice; no new plan credits are granted. Support settles any unused paid
--   plan by hand (see the `seat_cutover` subscription events).
--
-- This migration cannot be reversed.

-- +goose StatementBegin
DO $seat_cutover$
DECLARE
    offer record;
    target_plan uuid;
    target_version uuid;
    enterprise_version uuid;
    developer_version uuid;
BEGIN
    SELECT pv.version_id INTO enterprise_version
    FROM public.plan_versions pv
    JOIN public.commercial_plans p ON p.plan_id = pv.plan_id
    WHERE p.slug = 'enterprise' AND pv.status = 'published';

    SELECT pv.version_id INTO developer_version
    FROM public.plan_versions pv
    JOIN public.commercial_plans p ON p.plan_id = pv.plan_id
    WHERE p.slug = 'developer' AND pv.status = 'published';

    -- 1. Retire every sellable plan.
    UPDATE public.plan_versions SET status = 'archived', updated_at = NOW()
    WHERE status = 'published';

    UPDATE public.commercial_plans SET active = false, updated_at = NOW();

    -- 2. Add the two internal plans with every feature enabled.
    FOR offer IN
        SELECT * FROM (VALUES
            ('free', 'Free', 'Free seats only. Paid seats start a seat subscription.', 1),
            ('seats', 'Seats', 'Seat subscription: prepaid Pro and Max seats.', 2)
        ) AS o(slug, display_name, description, sort_order)
    LOOP
        INSERT INTO public.commercial_plans (slug, display_name, description, sort_order, active)
        VALUES (offer.slug, offer.display_name, offer.description, offer.sort_order, true)
        ON CONFLICT (slug) DO UPDATE
        SET display_name = EXCLUDED.display_name, description = EXCLUDED.description,
            sort_order = EXCLUDED.sort_order, active = true, updated_at = NOW()
        RETURNING plan_id INTO target_plan;

        INSERT INTO public.plan_versions (plan_id, version, status, effective_from, published_at)
        SELECT target_plan, COALESCE(MAX(version), 0) + 1, 'published', NOW(), NOW()
        FROM public.plan_versions WHERE plan_id = target_plan
        RETURNING version_id INTO target_version;

        -- Start from the catalog defaults, take the richest plan's values, and
        -- turn every boolean feature on.
        INSERT INTO public.plan_entitlements (version_id, feature_key, value_json)
        SELECT target_version, fc.feature_key,
               CASE WHEN fc.value_type = 'boolean' THEN 'true'::jsonb
                    ELSE COALESCE(pe.value_json, fc.default_value)
               END
        FROM public.feature_catalog fc
        LEFT JOIN public.plan_entitlements pe
            ON pe.version_id = enterprise_version AND pe.feature_key = fc.feature_key;

        -- Pool credits come from credit packs; members come from seats. Support
        -- tiers and on-prem deployment are not part of the self-serve offer.
        UPDATE public.plan_entitlements pe
        SET value_json = CASE pe.feature_key
                WHEN 'users.max' THEN 'null'::jsonb
                WHEN 'credits.monthly_grant' THEN '0'::jsonb
                WHEN 'credits.bonus_percent' THEN '0'::jsonb
                WHEN 'credits.rollover_allowed' THEN 'false'::jsonb
                WHEN 'credits.trial_grant' THEN '0'::jsonb
                WHEN 'deployment.vpc.enabled' THEN 'false'::jsonb
                ELSE COALESCE(
                    (SELECT dev.value_json FROM public.plan_entitlements dev
                     WHERE dev.version_id = developer_version AND dev.feature_key = pe.feature_key),
                    pe.value_json)
            END,
            updated_at = NOW()
        WHERE pe.version_id = target_version
          AND pe.feature_key IN ('users.max', 'credits.monthly_grant', 'credits.bonus_percent',
                                 'credits.rollover_allowed', 'credits.trial_grant',
                                 'deployment.vpc.enabled', 'support.tier', 'support.sla');
    END LOOP;

    -- 3. Move every tenant to `free` or `seats`.
    CREATE TEMPORARY TABLE seat_cutover_targets ON COMMIT DROP AS
    SELECT t.tenant_id,
           current_sub.subscription_id AS previous_subscription_id,
           current_sub.status AS previous_status,
           current_sub.plan_slug AS previous_plan,
           current_sub.billing_interval,
           current_sub.current_period_start,
           current_sub.current_period_end,
           blocks.paid_through
    FROM public.tenants t
    LEFT JOIN LATERAL (
        SELECT ts.subscription_id, ts.status, p.slug AS plan_slug,
               CASE WHEN ts.billing_interval = 'yearly' THEN 'yearly' ELSE 'monthly' END AS billing_interval,
               ts.current_period_start, ts.current_period_end
        FROM public.tenant_subscriptions ts
        JOIN public.commercial_plans p ON p.plan_id = ts.plan_id
        WHERE ts.tenant_id = t.tenant_id AND ts.status IN ('trialing', 'active', 'past_due')
        ORDER BY ts.updated_at DESC
        LIMIT 1
    ) current_sub ON true
    LEFT JOIN LATERAL (
        SELECT MAX(b.paid_through) AS paid_through
        FROM public.tenant_seat_blocks b
        WHERE b.tenant_id = t.tenant_id AND b.paid_through > NOW()
    ) blocks ON true;

    UPDATE public.tenant_subscriptions
    SET status = 'canceled', updated_at = NOW()
    WHERE status IN ('trialing', 'active', 'past_due');

    INSERT INTO public.tenant_subscriptions (
        tenant_id, plan_id, version_id, status, billing_interval,
        current_period_start, current_period_end, cancel_at_period_end,
        price_snapshot_json, metadata
    )
    SELECT target.tenant_id, p.plan_id, pv.version_id, 'active',
           CASE WHEN target.paid_through IS NULL THEN 'monthly' ELSE target.billing_interval END,
           CASE WHEN target.paid_through IS NULL THEN NOW()
                ELSE LEAST(COALESCE(target.current_period_start, NOW()), NOW()) END,
           COALESCE(target.paid_through, TIMESTAMPTZ '9999-12-31 00:00:00+00'),
           false,
           jsonb_build_object('currency', 'IRR', 'amount_minor', 0),
           jsonb_build_object('source', 'seat_cutover',
                              'previous_subscription_id', target.previous_subscription_id)
    FROM seat_cutover_targets target
    JOIN public.commercial_plans p
        ON p.slug = CASE WHEN target.paid_through IS NULL THEN 'free' ELSE 'seats' END
    JOIN public.plan_versions pv ON pv.plan_id = p.plan_id AND pv.status = 'published';

    INSERT INTO public.subscription_events (tenant_id, subscription_id, event_type, actor, payload)
    SELECT ts.tenant_id, ts.subscription_id, 'seat_cutover', 'migration',
           jsonb_build_object(
               'previous_subscription_id', target.previous_subscription_id,
               'previous_plan', target.previous_plan,
               'previous_status', target.previous_status,
               'previous_period_end', target.current_period_end,
               'unused_paid_plan', target.previous_status = 'active'
                   AND target.current_period_end > NOW())
    FROM public.tenant_subscriptions ts
    JOIN seat_cutover_targets target ON target.tenant_id = ts.tenant_id
    WHERE ts.metadata->>'source' = 'seat_cutover' AND ts.status = 'active';
END
$seat_cutover$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'the seat subscription cutover retired every plan and cannot be reversed';
END
$$;
-- +goose StatementEnd
