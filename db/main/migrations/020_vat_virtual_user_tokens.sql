-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.virtual_users (
    virtual_user_id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    team_id uuid,
    org_unit_id uuid,
    alias character varying(255) NOT NULL,
    display_name character varying(255),
    metadata jsonb DEFAULT '{}'::jsonb,
    active boolean NOT NULL DEFAULT true,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    CONSTRAINT virtual_users_pkey PRIMARY KEY (virtual_user_id)
);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'virtual_users_tenant_id_fkey'
    ) THEN
        ALTER TABLE public.virtual_users
            ADD CONSTRAINT virtual_users_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'virtual_users_team_id_fkey'
    ) THEN
        ALTER TABLE public.virtual_users
            ADD CONSTRAINT virtual_users_team_id_fkey
            FOREIGN KEY (team_id) REFERENCES public.teams(team_id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'virtual_users_org_unit_id_fkey'
    ) THEN
        ALTER TABLE public.virtual_users
            ADD CONSTRAINT virtual_users_org_unit_id_fkey
            FOREIGN KEY (org_unit_id) REFERENCES public.organizations(organization_id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_virtual_users_tenant_alias
    ON public.virtual_users (tenant_id, alias);
CREATE INDEX IF NOT EXISTS idx_virtual_users_tenant_id
    ON public.virtual_users (tenant_id);
CREATE INDEX IF NOT EXISTS idx_virtual_users_team_id
    ON public.virtual_users (team_id);
CREATE INDEX IF NOT EXISTS idx_virtual_users_org_unit_id
    ON public.virtual_users (org_unit_id);

ALTER TABLE public.verification_tokens
    ADD COLUMN IF NOT EXISTS virtual_user_id uuid,
    ADD COLUMN IF NOT EXISTS rotated_from_token_id uuid,
    ADD COLUMN IF NOT EXISTS rotation_grace_expires_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS revoked_reason text;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'verification_tokens_virtual_user_id_fkey'
    ) THEN
        ALTER TABLE public.verification_tokens
            ADD CONSTRAINT verification_tokens_virtual_user_id_fkey
            FOREIGN KEY (virtual_user_id) REFERENCES public.virtual_users(virtual_user_id) ON DELETE SET NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'verification_tokens_rotated_from_token_id_fkey'
    ) THEN
        ALTER TABLE public.verification_tokens
            ADD CONSTRAINT verification_tokens_rotated_from_token_id_fkey
            FOREIGN KEY (rotated_from_token_id) REFERENCES public.verification_tokens(id) ON DELETE SET NULL;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_vt_tenant_virtual_user
    ON public.verification_tokens (tenant_id, virtual_user_id);
CREATE INDEX IF NOT EXISTS idx_vt_tenant_key_type
    ON public.verification_tokens (tenant_id, key_type);
CREATE INDEX IF NOT EXISTS idx_vt_tenant_revoked_at
    ON public.verification_tokens (tenant_id, revoked_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_vt_tenant_revoked_at;
DROP INDEX IF EXISTS idx_vt_tenant_key_type;
DROP INDEX IF EXISTS idx_vt_tenant_virtual_user;

ALTER TABLE public.verification_tokens
    DROP CONSTRAINT IF EXISTS verification_tokens_rotated_from_token_id_fkey,
    DROP CONSTRAINT IF EXISTS verification_tokens_virtual_user_id_fkey;

ALTER TABLE public.verification_tokens
    DROP COLUMN IF EXISTS revoked_reason,
    DROP COLUMN IF EXISTS rotation_grace_expires_at,
    DROP COLUMN IF EXISTS rotated_from_token_id,
    DROP COLUMN IF EXISTS virtual_user_id;

DROP INDEX IF EXISTS idx_virtual_users_org_unit_id;
DROP INDEX IF EXISTS idx_virtual_users_team_id;
DROP INDEX IF EXISTS idx_virtual_users_tenant_id;
DROP INDEX IF EXISTS uq_virtual_users_tenant_alias;
DROP TABLE IF EXISTS public.virtual_users;
-- +goose StatementEnd
