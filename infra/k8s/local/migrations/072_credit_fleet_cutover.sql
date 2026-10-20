-- +goose Up
-- One-shot fleet cutover to credit billing (product decision, 2026-10-17).
-- See plans/credit-pricing/phase-3-design.md, "Fleet cutover migration".
--
-- * The credit plan versions seeded as drafts by 067 are published.
-- * Every tenant becomes a credit tenant on a 14-day trial of the credit
--   version of its plan (developer when it has none). The credit ledger worker
--   then grants each trial 500 credits that expire with it.
-- * IRR budget limits get credit limits at the active list price, rounded up.
-- * All IRR financial history is deleted: wallets, wallet transactions,
--   top-ups, bank transfers, subscription payments, invoices, discount and gift
--   redemptions, window settlements, and spend logs. The owner accepted this
--   loss; the IRR wallet balances are not converted.
--
-- This migration cannot be reversed.

-- +goose StatementBegin
DO $cutover$
DECLARE
    credit_plan record;
    price_toman bigint;
BEGIN
    -- 1. Publish the credit plan versions.
    FOR credit_plan IN
        SELECT DISTINCT ON (pv.plan_id) pv.plan_id, pv.version_id
        FROM public.plan_versions pv
        WHERE pv.status = 'draft'
          AND EXISTS (
              SELECT 1 FROM public.plan_entitlements pe
              WHERE pe.version_id = pv.version_id AND pe.feature_key = 'credits.monthly_grant'
          )
          AND NOT EXISTS (
              SELECT 1 FROM public.plan_versions published
              JOIN public.plan_entitlements pe ON pe.version_id = published.version_id AND pe.feature_key = 'credits.monthly_grant'
              WHERE published.plan_id = pv.plan_id AND published.status = 'published'
          )
        ORDER BY pv.plan_id, pv.version DESC
    LOOP
        UPDATE public.plan_versions SET status = 'archived', updated_at = NOW()
        WHERE plan_id = credit_plan.plan_id AND status = 'published';

        UPDATE public.plan_versions SET status = 'published', published_at = NOW(), updated_at = NOW()
        WHERE version_id = credit_plan.version_id;
    END LOOP;

    -- 2. Delete the IRR financial history. Children go first; nothing here is
    -- referenced by the credit ledger.
    DELETE FROM public.invoice_line_items;
    DELETE FROM public.invoices;
    DELETE FROM public.wallet_manual_adjustments;
    DELETE FROM public.gift_code_redemptions;
    DELETE FROM public.discount_redemptions;
    DELETE FROM public.bank_transfer_requests;
    DELETE FROM public.credit_pack_orders;
    DELETE FROM public.wallet_topup_intents;
    DELETE FROM public.subscription_payment_intents;
    DELETE FROM public.wallet_budget_window_settlements;
    DELETE FROM public.wallet_transactions;
    DELETE FROM public.wallets;
    TRUNCATE public.spend_logs;

    -- New tenants get no IRR signup credit; credit trials replace it.
    UPDATE public.commercial_signup_settings
    SET signup_credit_irr = 0, trial_credit_mc = 500000, trial_credit_days = 14
    WHERE settings_id = true;

    -- 3. Give IRR budget limits a credit limit, rounded up to a whole credit.
    SELECT COALESCE(
        (SELECT price_toman_per_credit FROM public.credit_price_revisions WHERE status = 'active'), 1000
    ) INTO price_toman;

    UPDATE public.budget_limit_rules
    SET limit_mc = CEIL(limit_to / (price_toman * 10)::numeric)::bigint * 1000, updated_at = NOW()
    WHERE limit_mc IS NULL AND limit_to > 0;

    -- 4. Move every tenant to a 14-day trial of its credit plan.
    CREATE TEMPORARY TABLE cutover_targets ON COMMIT DROP AS
    SELECT t.tenant_id,
           current_sub.subscription_id AS previous_subscription_id,
           COALESCE(own_plan.plan_id, developer.plan_id) AS plan_id,
           COALESCE(own_plan.version_id, developer.version_id) AS version_id
    FROM public.tenants t
    LEFT JOIN LATERAL (
        SELECT ts.subscription_id, ts.plan_id
        FROM public.tenant_subscriptions ts
        WHERE ts.tenant_id = t.tenant_id AND ts.status IN ('trialing', 'active', 'past_due')
        ORDER BY ts.updated_at DESC
        LIMIT 1
    ) current_sub ON true
    LEFT JOIN LATERAL (
        SELECT pv.plan_id, pv.version_id
        FROM public.plan_versions pv
        WHERE pv.plan_id = current_sub.plan_id AND pv.status = 'published'
          AND EXISTS (SELECT 1 FROM public.plan_entitlements pe
                      WHERE pe.version_id = pv.version_id AND pe.feature_key = 'credits.monthly_grant')
    ) own_plan ON true
    LEFT JOIN LATERAL (
        SELECT pv.plan_id, pv.version_id
        FROM public.plan_versions pv
        JOIN public.commercial_plans p ON p.plan_id = pv.plan_id AND p.slug = 'developer'
        WHERE pv.status = 'published'
          AND EXISTS (SELECT 1 FROM public.plan_entitlements pe
                      WHERE pe.version_id = pv.version_id AND pe.feature_key = 'credits.monthly_grant')
    ) developer ON true;

    IF EXISTS (SELECT 1 FROM cutover_targets WHERE version_id IS NULL) THEN
        RAISE EXCEPTION 'credit cutover: no published credit plan version for some tenants';
    END IF;

    UPDATE public.tenant_subscriptions
    SET status = 'canceled', updated_at = NOW()
    WHERE status IN ('trialing', 'active', 'past_due');

    INSERT INTO public.tenant_subscriptions (
        tenant_id, plan_id, version_id, status, billing_interval,
        current_period_start, current_period_end, trial_started_at, trial_ends_at,
        cancel_at_period_end, price_snapshot_json, metadata
    )
    SELECT target.tenant_id, target.plan_id, target.version_id, 'trialing', 'monthly',
           NOW(), NOW() + INTERVAL '14 days', NOW(), NOW() + INTERVAL '14 days',
           false,
           jsonb_build_object('currency', 'IRR', 'amount_minor', 0, 'billing_interval', 'monthly',
                              'discount_percent', 100, 'regular_amount_minor', COALESCE(pp.amount_minor, 0)),
           jsonb_build_object('source', 'credit_cutover', 'previous_subscription_id', target.previous_subscription_id)
    FROM cutover_targets target
    LEFT JOIN public.plan_prices pp ON pp.version_id = target.version_id AND pp.billing_interval = 'monthly';

    INSERT INTO public.subscription_events (tenant_id, subscription_id, event_type, actor, payload)
    SELECT ts.tenant_id, ts.subscription_id, 'credit_cutover', 'migration',
           jsonb_build_object('trial_ends_at', ts.trial_ends_at, 'trial_credits', 500)
    FROM public.tenant_subscriptions ts
    WHERE ts.metadata->>'source' = 'credit_cutover' AND ts.status = 'trialing';

    -- 5. Bill every tenant in credits.
    UPDATE public.tenants SET billing_unit = 'credit' WHERE billing_unit <> 'credit';
END
$cutover$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'the credit fleet cutover deleted the IRR financial history and cannot be reversed';
END
$$;
-- +goose StatementEnd
