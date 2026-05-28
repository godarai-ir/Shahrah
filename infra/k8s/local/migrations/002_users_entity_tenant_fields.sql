-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS tenant_id uuid,
    ADD COLUMN IF NOT EXISTS mobile text,
    ADD COLUMN IF NOT EXISTS active boolean;

UPDATE public.users
SET active = true
WHERE active IS NULL;

UPDATE public.users
SET active = false
WHERE active = true
  AND LOWER(COALESCE(role, '')) = 'deactivated';

UPDATE public.users
SET mobile = NULLIF(TRIM(user_alias), '')
WHERE mobile IS NULL
  AND TRIM(COALESCE(user_alias, '')) ~ '^\+?[0-9]{10,15}$';

ALTER TABLE public.users
    ALTER COLUMN active SET DEFAULT true;

INSERT INTO public.tenants (slug, domain, status)
SELECT 'default', 'default', 'active'
WHERE NOT EXISTS (SELECT 1 FROM public.tenants);

WITH default_tenant AS (
    SELECT tenant_id
    FROM public.tenants
    WHERE slug = 'default'
    LIMIT 1
), fallback_tenant AS (
    SELECT tenant_id FROM default_tenant
    UNION ALL
    SELECT tenant_id
    FROM public.tenants
    WHERE NOT EXISTS (SELECT 1 FROM default_tenant)
    LIMIT 1
)
UPDATE public.users u
SET tenant_id = (SELECT tenant_id FROM fallback_tenant LIMIT 1)
WHERE u.tenant_id IS NULL;

ALTER TABLE public.users
    ALTER COLUMN tenant_id SET NOT NULL,
    ALTER COLUMN active SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'users_tenant_id_fkey'
    ) THEN
        ALTER TABLE public.users
            ADD CONSTRAINT users_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON public.users (tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON public.users (tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_users_tenant_mobile ON public.users (tenant_id, mobile);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_tenant_mobile;
DROP INDEX IF EXISTS idx_users_tenant_email;
DROP INDEX IF EXISTS idx_users_tenant_id;

ALTER TABLE public.users
    DROP CONSTRAINT IF EXISTS users_tenant_id_fkey;

ALTER TABLE public.users
    DROP COLUMN IF EXISTS active,
    DROP COLUMN IF EXISTS mobile,
    DROP COLUMN IF EXISTS tenant_id;
-- +goose StatementEnd
