-- +goose Up
ALTER TABLE public.prompts
    DROP CONSTRAINT IF EXISTS prompts_slug_format,
    ADD CONSTRAINT prompts_slug_format CHECK (char_length(slug) BETWEEN 1 AND 256 AND slug ~ '^[A-Za-z0-9_-]+$');

-- +goose Down
ALTER TABLE public.prompts
    DROP CONSTRAINT IF EXISTS prompts_slug_format,
    ADD CONSTRAINT prompts_slug_format CHECK (slug ~ '^[A-Za-z0-9_-]{1,256}$');
