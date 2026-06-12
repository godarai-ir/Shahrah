-- +goose Up
ALTER TABLE public.tenant_provider_account_models
    ADD COLUMN input_modalities text[] DEFAULT '{}'::text[] NOT NULL,
    ADD COLUMN output_modalities text[] DEFAULT '{}'::text[] NOT NULL,
    ADD COLUMN public_pricing jsonb;

-- +goose Down
ALTER TABLE public.tenant_provider_account_models
    DROP COLUMN IF EXISTS public_pricing,
    DROP COLUMN IF EXISTS output_modalities,
    DROP COLUMN IF EXISTS input_modalities;
