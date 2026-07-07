-- +goose Up
CREATE TABLE IF NOT EXISTS public.tenant_log_saved_views (
    saved_view_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    filters_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    visible_columns_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    shared_scope TEXT NOT NULL DEFAULT 'tenant' CHECK (shared_scope IN ('tenant', 'private')),
    created_by_user_id UUID REFERENCES public.users(user_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS tenant_log_saved_views_tenant_name_ci_idx
    ON public.tenant_log_saved_views (tenant_id, lower(name));

CREATE INDEX IF NOT EXISTS tenant_log_saved_views_tenant_updated_idx
    ON public.tenant_log_saved_views (tenant_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS public.tenant_log_export_jobs (
    export_id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    filters_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    format TEXT NOT NULL DEFAULT 'csv' CHECK (format IN ('csv', 'jsonl')),
    include_content BOOLEAN NOT NULL DEFAULT false,
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('queued', 'running', 'completed', 'failed', 'canceled')),
    row_count INTEGER NOT NULL DEFAULT 0,
    created_by_user_id UUID REFERENCES public.users(user_id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS tenant_log_export_jobs_tenant_created_idx
    ON public.tenant_log_export_jobs (tenant_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS public.tenant_log_export_jobs;
DROP TABLE IF EXISTS public.tenant_log_saved_views;
