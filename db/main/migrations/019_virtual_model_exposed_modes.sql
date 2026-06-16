-- +goose Up
ALTER TABLE public.virtual_models
    ADD COLUMN IF NOT EXISTS exposed_modes text[] NOT NULL DEFAULT '{}'::text[];

-- +goose Down
ALTER TABLE public.virtual_models
    DROP COLUMN IF EXISTS exposed_modes;
