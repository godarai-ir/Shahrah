-- +goose Up
ALTER TABLE public.virtual_model_targets
    ADD COLUMN IF NOT EXISTS weight double precision DEFAULT 0 NOT NULL;

-- +goose Down
ALTER TABLE public.virtual_model_targets
    DROP COLUMN IF EXISTS weight;
