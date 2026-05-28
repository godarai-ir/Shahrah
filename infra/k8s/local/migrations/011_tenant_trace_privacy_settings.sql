-- +goose Up
CREATE TABLE IF NOT EXISTS public.tenant_trace_privacy_settings (
    tenant_id uuid PRIMARY KEY REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    capture_policy text NOT NULL DEFAULT 'metadata-only',
    updated_by_user_id uuid NULL,
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT tenant_trace_privacy_settings_policy_chk
        CHECK (capture_policy IN ('none', 'metadata-only', 'sampled-redacted', 'full-secure'))
);

CREATE INDEX IF NOT EXISTS idx_tenant_trace_privacy_settings_updated_at
    ON public.tenant_trace_privacy_settings (updated_at DESC);

-- +goose Down
DROP TABLE IF EXISTS public.tenant_trace_privacy_settings;
