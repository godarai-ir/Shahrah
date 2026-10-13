-- +goose Up
-- +goose StatementBegin
-- Plans no longer cap the number of gateway requests per subscription period.
-- Remove the quota feature (catalog row, plan values and tenant overrides) and
-- the per-tenant usage counters that only existed to enforce it.
DROP TABLE IF EXISTS public.tenant_usage_counters;

DELETE FROM public.tenant_entitlement_overrides WHERE feature_key = 'gateway.requests.monthly';
DELETE FROM public.plan_entitlements WHERE feature_key = 'gateway.requests.monthly';
DELETE FROM public.feature_catalog WHERE feature_key = 'gateway.requests.monthly';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
INSERT INTO public.feature_catalog (feature_key, value_type, default_value, allow_override, category, description)
VALUES ('gateway.requests.monthly', 'integer', '0'::jsonb, true, 'limits', 'Monthly gateway request quota.')
ON CONFLICT (feature_key) DO NOTHING;

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
-- +goose StatementEnd
