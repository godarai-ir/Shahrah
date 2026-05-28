-- +goose Up
CREATE TABLE IF NOT EXISTS public.budget_limit_rules (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    rule_id character varying(255) NOT NULL,
    name character varying(255),
    description text,
    enabled boolean NOT NULL DEFAULT true,
    priority integer NOT NULL DEFAULT 100,
    conditions_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    limit_to numeric(18,6) NOT NULL,
    unit character varying(64) NOT NULL,
    budget_applies_per character varying(255),
    block_on_budget_exceed boolean NOT NULL DEFAULT true,
    alerts_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_budget_limit_rules_tenant_rule_id
    ON public.budget_limit_rules (tenant_id, rule_id);

CREATE INDEX IF NOT EXISTS idx_budget_limit_rules_tenant_id
    ON public.budget_limit_rules (tenant_id);

CREATE INDEX IF NOT EXISTS idx_budget_limit_rules_enabled
    ON public.budget_limit_rules (enabled);

CREATE INDEX IF NOT EXISTS idx_budget_limit_rules_priority
    ON public.budget_limit_rules (priority);

CREATE INDEX IF NOT EXISTS idx_budget_limit_rules_updated_at
    ON public.budget_limit_rules (updated_at DESC);

-- +goose Down
DROP TABLE IF EXISTS public.budget_limit_rules;
