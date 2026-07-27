-- +goose Up
ALTER TABLE public.tenant_provider_account_models
    ADD COLUMN catalog_available boolean DEFAULT true NOT NULL,
    ADD COLUMN catalog_checked_at timestamp with time zone,
    ADD COLUMN catalog_block_reason character varying(100);

CREATE INDEX idx_tenant_provider_account_models_runtime_available
    ON public.tenant_provider_account_models (account_id, enabled, catalog_available);

ALTER TABLE public.model_deployments
    ADD COLUMN catalog_managed boolean DEFAULT false NOT NULL,
    ADD COLUMN catalog_available boolean DEFAULT true NOT NULL,
    ADD COLUMN catalog_checked_at timestamp with time zone,
    ADD COLUMN catalog_block_reason character varying(100);

-- +goose Down
DROP INDEX IF EXISTS public.idx_tenant_provider_account_models_runtime_available;
ALTER TABLE public.model_deployments
    DROP COLUMN IF EXISTS catalog_block_reason,
    DROP COLUMN IF EXISTS catalog_checked_at,
    DROP COLUMN IF EXISTS catalog_available,
    DROP COLUMN IF EXISTS catalog_managed;
ALTER TABLE public.tenant_provider_account_models
    DROP COLUMN IF EXISTS catalog_block_reason,
    DROP COLUMN IF EXISTS catalog_checked_at,
    DROP COLUMN IF EXISTS catalog_available;
