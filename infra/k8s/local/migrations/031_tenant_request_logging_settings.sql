-- +goose Up
CREATE TABLE IF NOT EXISTS public.tenant_request_logging_settings (
    tenant_id uuid PRIMARY KEY REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    logging_mode text NOT NULL,
    updated_by_user_id uuid NULL REFERENCES public.users(user_id) ON DELETE SET NULL,
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    CONSTRAINT tenant_request_logging_settings_mode_chk
        CHECK (logging_mode IN ('none', 'metadata-only', 'redacted-content', 'full-content'))
);

CREATE INDEX IF NOT EXISTS idx_tenant_request_logging_settings_updated_at
    ON public.tenant_request_logging_settings (updated_at DESC);

-- +goose Down
DROP TABLE IF EXISTS public.tenant_request_logging_settings;
