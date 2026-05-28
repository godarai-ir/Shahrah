-- +goose Up
CREATE TABLE IF NOT EXISTS public.rate_limit_rules (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    rule_id character varying(255) NOT NULL,
    name character varying(255),
    description text,
    enabled boolean NOT NULL DEFAULT true,
    priority integer NOT NULL DEFAULT 100,
    conditions_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    limit_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    conflict_behavior character varying(64) NOT NULL DEFAULT 'most_restrictive',
    response_behavior character varying(64) NOT NULL DEFAULT 'return_429',
    soft_limit boolean NOT NULL DEFAULT false,
    retry_after_header boolean NOT NULL DEFAULT true,
    effective_from timestamp with time zone,
    expires_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_rate_limit_rules_tenant_rule_id
    ON public.rate_limit_rules (tenant_id, rule_id);

CREATE INDEX IF NOT EXISTS idx_rate_limit_rules_enabled
    ON public.rate_limit_rules (enabled);

CREATE INDEX IF NOT EXISTS idx_rate_limit_rules_priority
    ON public.rate_limit_rules (priority);

CREATE INDEX IF NOT EXISTS idx_rate_limit_rules_updated_at
    ON public.rate_limit_rules (updated_at DESC);

-- +goose Down
DROP TABLE IF EXISTS public.rate_limit_rules;
