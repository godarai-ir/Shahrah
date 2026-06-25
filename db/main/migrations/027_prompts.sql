-- +goose Up
CREATE TABLE IF NOT EXISTS public.prompts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    description text NOT NULL DEFAULT '',
    latest_version integer NOT NULL DEFAULT 0,
    created_by text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT prompts_pkey PRIMARY KEY (id),
    CONSTRAINT prompts_tenant_fk FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT prompts_slug_format CHECK (char_length(slug) BETWEEN 1 AND 256 AND slug ~ '^[A-Za-z0-9_-]+$'),
    CONSTRAINT prompts_latest_version_nonnegative CHECK (latest_version >= 0)
);

CREATE TABLE IF NOT EXISTS public.prompt_versions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    prompt_id uuid NOT NULL,
    version integer NOT NULL,
    fqn text NOT NULL,
    version_alias text NOT NULL DEFAULT '',
    name text NOT NULL,
    description text NOT NULL DEFAULT '',
    messages jsonb NOT NULL,
    variables jsonb NOT NULL DEFAULT '{}'::jsonb,
    model text NOT NULL DEFAULT '',
    parameters jsonb NOT NULL DEFAULT '{}'::jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    commit_message text NOT NULL DEFAULT '',
    created_by text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT prompt_versions_pkey PRIMARY KEY (id),
    CONSTRAINT prompt_versions_tenant_fk FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    CONSTRAINT prompt_versions_prompt_fk FOREIGN KEY (prompt_id) REFERENCES public.prompts(id) ON DELETE CASCADE,
    CONSTRAINT prompt_versions_version_positive CHECK (version > 0),
    CONSTRAINT prompt_versions_messages_array CHECK (jsonb_typeof(messages) = 'array'),
    CONSTRAINT prompt_versions_variables_object CHECK (jsonb_typeof(variables) = 'object'),
    CONSTRAINT prompt_versions_parameters_object CHECK (jsonb_typeof(parameters) = 'object'),
    CONSTRAINT prompt_versions_metadata_object CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_prompts_tenant_slug
    ON public.prompts (tenant_id, slug);

CREATE UNIQUE INDEX IF NOT EXISTS uq_prompt_versions_tenant_prompt_version
    ON public.prompt_versions (tenant_id, prompt_id, version);

CREATE UNIQUE INDEX IF NOT EXISTS uq_prompt_versions_tenant_fqn
    ON public.prompt_versions (tenant_id, fqn);

CREATE INDEX IF NOT EXISTS idx_prompts_tenant_updated ON public.prompts (tenant_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_prompt_versions_prompt_version ON public.prompt_versions (prompt_id, version DESC);

-- +goose Down
DROP TABLE IF EXISTS public.prompt_versions;
DROP TABLE IF EXISTS public.prompts;
