-- +goose Up
ALTER TABLE public.prompt_versions
    ADD COLUMN IF NOT EXISTS manifest jsonb NOT NULL DEFAULT '{}'::jsonb;

UPDATE public.prompt_versions
SET manifest = jsonb_strip_nulls(
    jsonb_build_object(
        'type', 'chat_prompt',
        'name', name,
        'description', description,
        'messages', messages,
        'variables', variables,
        'model', model,
        'parameters', parameters,
        'metadata', metadata,
        'version_alias', version_alias
    )
)
WHERE manifest = '{}'::jsonb;

ALTER TABLE public.prompt_versions
    ADD CONSTRAINT prompt_versions_manifest_object
    CHECK (jsonb_typeof(manifest) = 'object');

-- +goose Down
ALTER TABLE public.prompt_versions
    DROP CONSTRAINT IF EXISTS prompt_versions_manifest_object;

ALTER TABLE public.prompt_versions
    DROP COLUMN IF EXISTS manifest;
