-- +goose Up
CREATE TABLE IF NOT EXISTS public.commercial_plans (
    plan_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    slug character varying(64) NOT NULL,
    display_name character varying(255) NOT NULL,
    description text,
    sort_order integer NOT NULL DEFAULT 100,
    active boolean NOT NULL DEFAULT true,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_commercial_plans_slug
    ON public.commercial_plans (slug);

CREATE TABLE IF NOT EXISTS public.plan_versions (
    version_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    plan_id uuid NOT NULL REFERENCES public.commercial_plans(plan_id) ON DELETE CASCADE,
    version integer NOT NULL,
    status character varying(32) NOT NULL DEFAULT 'draft',
    effective_from timestamp with time zone,
    published_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_plan_versions_status CHECK (status IN ('draft', 'published', 'archived'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_plan_versions_plan_version
    ON public.plan_versions (plan_id, version);

CREATE UNIQUE INDEX IF NOT EXISTS uq_plan_versions_one_published
    ON public.plan_versions (plan_id)
    WHERE status = 'published';

CREATE TABLE IF NOT EXISTS public.feature_catalog (
    feature_key character varying(128) PRIMARY KEY,
    value_type character varying(32) NOT NULL,
    default_value jsonb NOT NULL DEFAULT 'null'::jsonb,
    allow_override boolean NOT NULL DEFAULT true,
    category character varying(64),
    description text,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_feature_catalog_value_type CHECK (value_type IN ('boolean', 'integer', 'decimal', 'string', 'json'))
);

CREATE TABLE IF NOT EXISTS public.plan_entitlements (
    entitlement_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES public.plan_versions(version_id) ON DELETE CASCADE,
    feature_key character varying(128) NOT NULL REFERENCES public.feature_catalog(feature_key) ON DELETE CASCADE,
    value_json jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_plan_entitlements_version_feature
    ON public.plan_entitlements (version_id, feature_key);

CREATE TABLE IF NOT EXISTS public.plan_prices (
    price_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    version_id uuid NOT NULL REFERENCES public.plan_versions(version_id) ON DELETE CASCADE,
    billing_interval character varying(32) NOT NULL,
    currency character varying(16) NOT NULL DEFAULT 'IRR',
    amount_minor numeric(18,0),
    discount_percent numeric(5,2) NOT NULL DEFAULT 0,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_plan_prices_interval CHECK (billing_interval IN ('monthly', 'quarterly', 'yearly')),
    CONSTRAINT chk_plan_prices_discount CHECK (discount_percent >= 0 AND discount_percent <= 100)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_plan_prices_version_interval
    ON public.plan_prices (version_id, billing_interval);

CREATE TABLE IF NOT EXISTS public.tenant_subscriptions (
    subscription_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    plan_id uuid NOT NULL REFERENCES public.commercial_plans(plan_id),
    version_id uuid NOT NULL REFERENCES public.plan_versions(version_id),
    status character varying(32) NOT NULL DEFAULT 'active',
    billing_interval character varying(32) NOT NULL DEFAULT 'monthly',
    current_period_start timestamp with time zone NOT NULL,
    current_period_end timestamp with time zone NOT NULL,
    cancel_at_period_end boolean NOT NULL DEFAULT false,
    pending_version_id uuid REFERENCES public.plan_versions(version_id),
    price_snapshot_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_tenant_subscriptions_status CHECK (status IN ('active', 'past_due', 'canceled', 'expired')),
    CONSTRAINT chk_tenant_subscriptions_interval CHECK (billing_interval IN ('monthly', 'quarterly', 'yearly'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tenant_subscriptions_active
    ON public.tenant_subscriptions (tenant_id)
    WHERE status IN ('active', 'past_due');

CREATE INDEX IF NOT EXISTS idx_tenant_subscriptions_tenant
    ON public.tenant_subscriptions (tenant_id);

CREATE TABLE IF NOT EXISTS public.tenant_entitlement_overrides (
    override_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    feature_key character varying(128) NOT NULL REFERENCES public.feature_catalog(feature_key) ON DELETE CASCADE,
    value_json jsonb NOT NULL,
    reason text,
    active boolean NOT NULL DEFAULT true,
    expires_at timestamp with time zone,
    created_by character varying(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tenant_entitlement_overrides_active
    ON public.tenant_entitlement_overrides (tenant_id, feature_key)
    WHERE active = true;

CREATE INDEX IF NOT EXISTS idx_tenant_entitlement_overrides_tenant
    ON public.tenant_entitlement_overrides (tenant_id);

CREATE TABLE IF NOT EXISTS public.subscription_events (
    event_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    subscription_id uuid REFERENCES public.tenant_subscriptions(subscription_id) ON DELETE SET NULL,
    event_type character varying(64) NOT NULL,
    actor character varying(255),
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_subscription_events_tenant_created
    ON public.subscription_events (tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS public.subscription_payment_intents (
    payment_intent_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    public_id uuid DEFAULT gen_random_uuid() NOT NULL UNIQUE,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    plan_id uuid NOT NULL REFERENCES public.commercial_plans(plan_id),
    version_id uuid NOT NULL REFERENCES public.plan_versions(version_id),
    billing_interval character varying(32) NOT NULL,
    provider character varying(32) NOT NULL,
    status character varying(32) NOT NULL DEFAULT 'PENDING',
    currency character varying(16) NOT NULL DEFAULT 'IRR',
    amount_minor numeric(18,0) NOT NULL,
    discount_percent numeric(5,2) NOT NULL DEFAULT 0,
    price_snapshot_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    authority character varying(255),
    payment_url text,
    callback_url text NOT NULL,
    return_url text NOT NULL,
    ref_id character varying(255),
    gateway_txn_id character varying(255),
    card_mask character varying(64),
    payer_phone character varying(64),
    failure_reason text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    raw_payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    completed_subscription_id uuid REFERENCES public.tenant_subscriptions(subscription_id) ON DELETE SET NULL,
    completed_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT chk_subscription_payment_intents_interval CHECK (billing_interval IN ('monthly', 'quarterly', 'yearly')),
    CONSTRAINT chk_subscription_payment_intents_status CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'CANCELED')),
    CONSTRAINT chk_subscription_payment_intents_amount CHECK (amount_minor >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_subscription_payment_intents_authority
    ON public.subscription_payment_intents (authority)
    WHERE authority IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_subscription_payment_intents_tenant_created
    ON public.subscription_payment_intents (tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS public.tenant_usage_counters (
    counter_id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    feature_key character varying(128) NOT NULL REFERENCES public.feature_catalog(feature_key) ON DELETE CASCADE,
    window_start timestamp with time zone NOT NULL,
    window_end timestamp with time zone NOT NULL,
    used numeric(18,6) NOT NULL DEFAULT 0,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tenant_usage_counters_window
    ON public.tenant_usage_counters (tenant_id, feature_key, window_start);

ALTER TABLE IF EXISTS public.tenant_onboarding_requests
    ALTER COLUMN plan_slug SET DEFAULT 'developer';

INSERT INTO public.commercial_plans (slug, display_name, description, sort_order, active)
VALUES
    ('developer', 'Developer', 'Starter plan for individual developers and small experiments.', 10, true),
    ('team', 'Team', 'Operational plan for small teams.', 20, true),
    ('growth', 'Growth', 'Scale plan for growing production workloads.', 30, true),
    ('enterprise', 'Enterprise', 'Custom plan for enterprise workloads and deployments.', 40, true)
ON CONFLICT (slug) DO UPDATE
SET display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    sort_order = EXCLUDED.sort_order,
    active = EXCLUDED.active,
    updated_at = NOW();

INSERT INTO public.plan_versions (plan_id, version, status, effective_from, published_at)
SELECT plan_id, 1, 'published', NOW(), NOW()
FROM public.commercial_plans
WHERE slug IN ('developer', 'team', 'growth', 'enterprise')
ON CONFLICT (plan_id, version) DO UPDATE
SET status = EXCLUDED.status,
    effective_from = COALESCE(public.plan_versions.effective_from, EXCLUDED.effective_from),
    published_at = COALESCE(public.plan_versions.published_at, EXCLUDED.published_at),
    updated_at = NOW();

INSERT INTO public.feature_catalog (feature_key, value_type, default_value, allow_override, category, description)
VALUES
    ('gateway.requests.monthly', 'integer', '0'::jsonb, true, 'limits', 'Monthly gateway request quota.'),
    ('users.max', 'integer', '0'::jsonb, true, 'limits', 'Maximum active tenant users.'),
    ('teams.enabled', 'boolean', 'false'::jsonb, true, 'security', 'Team management access.'),
    ('cache.simple.enabled', 'boolean', 'true'::jsonb, true, 'gateway', 'Simple request caching access.'),
    ('cache.simple.ttl_days', 'integer', '1'::jsonb, true, 'gateway', 'Simple cache TTL in days.'),
    ('cache.semantic.enabled', 'boolean', 'false'::jsonb, true, 'gateway', 'Semantic caching access.'),
    ('routing.weighted.enabled', 'boolean', 'false'::jsonb, true, 'gateway', 'Weighted routing strategy access.'),
    ('routing.latency.enabled', 'boolean', 'false'::jsonb, true, 'gateway', 'Latency-based routing access.'),
    ('rate_limit_rules.enabled', 'boolean', 'false'::jsonb, true, 'policies', 'Tenant rate-limit rule management access.'),
    ('budget_limit_rules.enabled', 'boolean', 'false'::jsonb, true, 'policies', 'Tenant budget-limit rule management access.'),
    ('logs.retention_days', 'integer', '3'::jsonb, true, 'observability', 'Log retention in days.'),
    ('traces.retention_days', 'integer', '30'::jsonb, true, 'observability', 'Trace retention in days.'),
    ('metadata.custom.enabled', 'boolean', 'false'::jsonb, true, 'observability', 'Custom metadata capture access.'),
    ('llm_pricing.custom.enabled', 'boolean', 'false'::jsonb, true, 'observability', 'Custom LLM pricing access.'),
    ('analytics.enabled', 'boolean', 'true'::jsonb, true, 'observability', 'Operational and cost analytics access.'),
    ('alerts.enabled', 'boolean', 'false'::jsonb, true, 'observability', 'Operational alerting access.'),
    ('otel_export.enabled', 'boolean', 'false'::jsonb, true, 'observability', 'OpenTelemetry export access.'),
    ('prompt_templates.max', 'integer', '3'::jsonb, true, 'prompt', 'Maximum prompt templates. Use -1 for unlimited.'),
    ('prompt_versions.enabled', 'boolean', 'true'::jsonb, true, 'prompt', 'Prompt version and variable management access.'),
    ('guardrails.third_party.enabled', 'boolean', 'false'::jsonb, true, 'guardrails', 'Third-party guardrail integration access.'),
    ('guardrails.webhook.enabled', 'boolean', 'false'::jsonb, true, 'guardrails', 'Custom guardrail webhook access.'),
    ('auth.pat.enabled', 'boolean', 'true'::jsonb, true, 'security', 'Personal access token access.'),
    ('auth.vat.enabled', 'boolean', 'false'::jsonb, true, 'security', 'Virtual access token access.'),
    ('rbac.enabled', 'boolean', 'false'::jsonb, true, 'security', 'Role-based access control access.'),
    ('sso.enabled', 'boolean', 'false'::jsonb, true, 'security', 'SSO access.'),
    ('org_units.enabled', 'boolean', 'false'::jsonb, true, 'security', 'Organization unit management access.'),
    ('admin_audit.enabled', 'boolean', 'false'::jsonb, true, 'security', 'Administrative audit log access.'),
    ('deployment.saas.enabled', 'boolean', 'true'::jsonb, false, 'deployment', 'SaaS deployment access.'),
    ('deployment.vpc.enabled', 'boolean', 'false'::jsonb, true, 'deployment', 'VPC or on-prem deployment access.'),
    ('support.tier', 'string', '"chatbot"'::jsonb, true, 'support', 'Support tier.'),
    ('support.sla', 'string', 'null'::jsonb, true, 'support', 'Support SLA tier.')
ON CONFLICT (feature_key) DO UPDATE
SET value_type = EXCLUDED.value_type,
    default_value = EXCLUDED.default_value,
    allow_override = EXCLUDED.allow_override,
    category = EXCLUDED.category,
    description = EXCLUDED.description,
    updated_at = NOW();

WITH entitlement_seed(plan_slug, feature_key, value_json) AS (
    VALUES
        ('developer', 'gateway.requests.monthly', '10000'),
        ('team', 'gateway.requests.monthly', '100000'),
        ('growth', 'gateway.requests.monthly', '1000000'),
        ('enterprise', 'gateway.requests.monthly', 'null'),
        ('developer', 'users.max', '3'),
        ('team', 'users.max', '10'),
        ('growth', 'users.max', '30'),
        ('enterprise', 'users.max', 'null'),
        ('developer', 'teams.enabled', 'false'),
        ('team', 'teams.enabled', 'true'),
        ('growth', 'teams.enabled', 'true'),
        ('enterprise', 'teams.enabled', 'true'),
        ('developer', 'cache.simple.ttl_days', '1'),
        ('team', 'cache.simple.ttl_days', 'null'),
        ('growth', 'cache.simple.ttl_days', 'null'),
        ('enterprise', 'cache.simple.ttl_days', 'null'),
        ('developer', 'cache.semantic.enabled', 'false'),
        ('team', 'cache.semantic.enabled', 'true'),
        ('growth', 'cache.semantic.enabled', 'true'),
        ('enterprise', 'cache.semantic.enabled', 'true'),
        ('developer', 'routing.weighted.enabled', 'false'),
        ('team', 'routing.weighted.enabled', 'true'),
        ('growth', 'routing.weighted.enabled', 'true'),
        ('enterprise', 'routing.weighted.enabled', 'true'),
        ('developer', 'routing.latency.enabled', 'false'),
        ('team', 'routing.latency.enabled', 'true'),
        ('growth', 'routing.latency.enabled', 'true'),
        ('enterprise', 'routing.latency.enabled', 'true'),
        ('developer', 'rate_limit_rules.enabled', 'false'),
        ('team', 'rate_limit_rules.enabled', 'true'),
        ('growth', 'rate_limit_rules.enabled', 'true'),
        ('enterprise', 'rate_limit_rules.enabled', 'true'),
        ('developer', 'budget_limit_rules.enabled', 'false'),
        ('team', 'budget_limit_rules.enabled', 'true'),
        ('growth', 'budget_limit_rules.enabled', 'true'),
        ('enterprise', 'budget_limit_rules.enabled', 'true'),
        ('developer', 'logs.retention_days', '3'),
        ('team', 'logs.retention_days', '30'),
        ('growth', 'logs.retention_days', '30'),
        ('enterprise', 'logs.retention_days', 'null'),
        ('developer', 'traces.retention_days', '30'),
        ('team', 'traces.retention_days', '90'),
        ('growth', 'traces.retention_days', '90'),
        ('enterprise', 'traces.retention_days', 'null'),
        ('developer', 'metadata.custom.enabled', 'false'),
        ('team', 'metadata.custom.enabled', 'true'),
        ('growth', 'metadata.custom.enabled', 'true'),
        ('enterprise', 'metadata.custom.enabled', 'true'),
        ('developer', 'llm_pricing.custom.enabled', 'false'),
        ('team', 'llm_pricing.custom.enabled', 'true'),
        ('growth', 'llm_pricing.custom.enabled', 'true'),
        ('enterprise', 'llm_pricing.custom.enabled', 'true'),
        ('developer', 'alerts.enabled', 'false'),
        ('team', 'alerts.enabled', 'false'),
        ('growth', 'alerts.enabled', 'true'),
        ('enterprise', 'alerts.enabled', 'true'),
        ('developer', 'otel_export.enabled', 'false'),
        ('team', 'otel_export.enabled', 'false'),
        ('growth', 'otel_export.enabled', 'true'),
        ('enterprise', 'otel_export.enabled', 'true'),
        ('developer', 'prompt_templates.max', '3'),
        ('team', 'prompt_templates.max', '-1'),
        ('growth', 'prompt_templates.max', '-1'),
        ('enterprise', 'prompt_templates.max', '-1'),
        ('developer', 'guardrails.third_party.enabled', 'false'),
        ('team', 'guardrails.third_party.enabled', 'true'),
        ('growth', 'guardrails.third_party.enabled', 'true'),
        ('enterprise', 'guardrails.third_party.enabled', 'true'),
        ('developer', 'guardrails.webhook.enabled', 'false'),
        ('team', 'guardrails.webhook.enabled', 'false'),
        ('growth', 'guardrails.webhook.enabled', 'true'),
        ('enterprise', 'guardrails.webhook.enabled', 'true'),
        ('developer', 'auth.vat.enabled', 'false'),
        ('team', 'auth.vat.enabled', 'true'),
        ('growth', 'auth.vat.enabled', 'true'),
        ('enterprise', 'auth.vat.enabled', 'true'),
        ('developer', 'rbac.enabled', 'false'),
        ('team', 'rbac.enabled', 'true'),
        ('growth', 'rbac.enabled', 'true'),
        ('enterprise', 'rbac.enabled', 'true'),
        ('developer', 'sso.enabled', 'false'),
        ('team', 'sso.enabled', 'false'),
        ('growth', 'sso.enabled', 'true'),
        ('enterprise', 'sso.enabled', 'true'),
        ('developer', 'org_units.enabled', 'false'),
        ('team', 'org_units.enabled', 'false'),
        ('growth', 'org_units.enabled', 'false'),
        ('enterprise', 'org_units.enabled', 'true'),
        ('developer', 'admin_audit.enabled', 'false'),
        ('team', 'admin_audit.enabled', 'false'),
        ('growth', 'admin_audit.enabled', 'true'),
        ('enterprise', 'admin_audit.enabled', 'true'),
        ('developer', 'deployment.vpc.enabled', 'false'),
        ('team', 'deployment.vpc.enabled', 'false'),
        ('growth', 'deployment.vpc.enabled', 'false'),
        ('enterprise', 'deployment.vpc.enabled', 'true'),
        ('developer', 'support.tier', '"chatbot"'),
        ('team', 'support.tier', '"product"'),
        ('growth', 'support.tier', '"product"'),
        ('enterprise', 'support.tier', '"priority"'),
        ('developer', 'support.sla', 'null'),
        ('team', 'support.sla', '"standard"'),
        ('growth', 'support.sla', '"enterprise"'),
        ('enterprise', 'support.sla', '"enterprise"')
)
INSERT INTO public.plan_entitlements (version_id, feature_key, value_json)
SELECT pv.version_id, s.feature_key, s.value_json::jsonb
FROM entitlement_seed s
JOIN public.commercial_plans p ON p.slug = s.plan_slug
JOIN public.plan_versions pv ON pv.plan_id = p.plan_id AND pv.version = 1
ON CONFLICT (version_id, feature_key) DO UPDATE
SET value_json = EXCLUDED.value_json,
    updated_at = NOW();

WITH price_seed(plan_slug, billing_interval, discount_percent) AS (
    VALUES
        ('developer', 'monthly', 0::numeric),
        ('developer', 'quarterly', 10::numeric),
        ('developer', 'yearly', 20::numeric),
        ('team', 'monthly', 0::numeric),
        ('team', 'quarterly', 10::numeric),
        ('team', 'yearly', 20::numeric),
        ('growth', 'monthly', 0::numeric),
        ('growth', 'quarterly', 10::numeric),
        ('growth', 'yearly', 20::numeric),
        ('enterprise', 'monthly', 0::numeric),
        ('enterprise', 'quarterly', 0::numeric),
        ('enterprise', 'yearly', 0::numeric)
)
INSERT INTO public.plan_prices (version_id, billing_interval, currency, amount_minor, discount_percent)
SELECT pv.version_id, s.billing_interval, 'IRR', NULL, s.discount_percent
FROM price_seed s
JOIN public.commercial_plans p ON p.slug = s.plan_slug
JOIN public.plan_versions pv ON pv.plan_id = p.plan_id AND pv.version = 1
ON CONFLICT (version_id, billing_interval) DO UPDATE
SET currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    discount_percent = EXCLUDED.discount_percent,
    updated_at = NOW();

INSERT INTO public.tenant_subscriptions (
    tenant_id, plan_id, version_id, status, billing_interval,
    current_period_start, current_period_end, price_snapshot_json, metadata
)
SELECT t.tenant_id, p.plan_id, pv.version_id, 'active', 'monthly', NOW(), NOW() + INTERVAL '1 month',
       jsonb_build_object('currency', 'IRR', 'amount_minor', 0, 'billing_interval', 'monthly'),
       jsonb_build_object('source', 'migration_default')
FROM public.tenants t
JOIN public.commercial_plans p ON p.slug = 'developer'
JOIN public.plan_versions pv ON pv.plan_id = p.plan_id AND pv.version = 1
WHERE NOT EXISTS (
    SELECT 1
    FROM public.tenant_subscriptions ts
    WHERE ts.tenant_id = t.tenant_id
      AND ts.status IN ('active', 'past_due')
);

-- +goose Down
ALTER TABLE IF EXISTS public.tenant_onboarding_requests
    ALTER COLUMN plan_slug SET DEFAULT 'free';

DROP TABLE IF EXISTS public.tenant_usage_counters;
DROP TABLE IF EXISTS public.subscription_payment_intents;
DROP TABLE IF EXISTS public.subscription_events;
DROP TABLE IF EXISTS public.tenant_entitlement_overrides;
DROP TABLE IF EXISTS public.tenant_subscriptions;
DROP TABLE IF EXISTS public.plan_prices;
DROP TABLE IF EXISTS public.plan_entitlements;
DROP TABLE IF EXISTS public.feature_catalog;
DROP TABLE IF EXISTS public.plan_versions;
DROP TABLE IF EXISTS public.commercial_plans;
