-- +goose Up
CREATE TABLE IF NOT EXISTS public.tenant_runtime_profiles (
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    version bigint NOT NULL,
    profile_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    checksum varchar(64) NOT NULL,
    compiled_at timestamptz NOT NULL DEFAULT now(),
    source_updated_at timestamptz NOT NULL DEFAULT now(),
    status varchar(32) NOT NULL DEFAULT 'active',
    error text,
    CONSTRAINT tenant_runtime_profiles_pkey PRIMARY KEY (tenant_id, version),
    CONSTRAINT tenant_runtime_profiles_status_chk CHECK (status IN ('active', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_tenant_runtime_profiles_latest
    ON public.tenant_runtime_profiles (tenant_id, version DESC);

CREATE INDEX IF NOT EXISTS idx_tenant_runtime_profiles_compiled_at
    ON public.tenant_runtime_profiles (compiled_at DESC);

-- +goose Down
DROP TABLE IF EXISTS public.tenant_runtime_profiles;
