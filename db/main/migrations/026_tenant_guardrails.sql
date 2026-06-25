-- +goose Up
CREATE TABLE IF NOT EXISTS public.tenant_guardrails (
    guardrail_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    slug varchar(96) NOT NULL,
    name varchar(255) NOT NULL,
    description text,
    guardrail_type varchar(64) NOT NULL,
    operation varchar(32) NOT NULL DEFAULT 'validate',
    enforcement varchar(32) NOT NULL DEFAULT 'enforce',
    enabled boolean NOT NULL DEFAULT true,
    config_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_by varchar(255),
    updated_by varchar(255),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT tenant_guardrails_pkey PRIMARY KEY (guardrail_id),
    CONSTRAINT tenant_guardrails_tenant_fk FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT tenant_guardrails_slug_chk CHECK (slug ~ '^[a-zA-Z0-9][a-zA-Z0-9_-]{1,95}$'),
    CONSTRAINT tenant_guardrails_type_chk CHECK (guardrail_type IN ('keyword_blocklist', 'regex', 'secrets_detection', 'basic_pii')),
    CONSTRAINT tenant_guardrails_operation_chk CHECK (operation IN ('validate', 'mutate')),
    CONSTRAINT tenant_guardrails_enforcement_chk CHECK (enforcement IN ('audit', 'enforce', 'enforce_ignore_error'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_tenant_guardrails_tenant_slug
    ON public.tenant_guardrails (tenant_id, slug);

CREATE INDEX IF NOT EXISTS idx_tenant_guardrails_tenant
    ON public.tenant_guardrails (tenant_id);

CREATE INDEX IF NOT EXISTS idx_tenant_guardrails_tenant_enabled
    ON public.tenant_guardrails (tenant_id, enabled);

-- +goose Down
DROP TABLE IF EXISTS public.tenant_guardrails;
