-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.organizations
    ADD COLUMN IF NOT EXISTS tenant_id uuid;

DO $$
DECLARE
    has_teams_org_unit_id boolean;
    has_teams_organization_id boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'teams' AND column_name = 'org_unit_id'
    ) INTO has_teams_org_unit_id;

    SELECT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'teams' AND column_name = 'organization_id'
    ) INTO has_teams_organization_id;

    IF has_teams_org_unit_id THEN
        EXECUTE '
            UPDATE public.organizations o
            SET tenant_id = t.tenant_id
            FROM (
                SELECT org_unit_id, MAX(tenant_id::text)::uuid AS tenant_id
                FROM public.teams
                WHERE org_unit_id IS NOT NULL
                  AND tenant_id IS NOT NULL
                GROUP BY org_unit_id
            ) t
            WHERE o.tenant_id IS NULL
              AND o.organization_id = t.org_unit_id
        ';
    ELSIF has_teams_organization_id THEN
        EXECUTE '
            UPDATE public.organizations o
            SET tenant_id = t.tenant_id
            FROM (
                SELECT organization_id, MAX(tenant_id::text)::uuid AS tenant_id
                FROM public.teams
                WHERE organization_id IS NOT NULL
                  AND tenant_id IS NOT NULL
                GROUP BY organization_id
            ) t
            WHERE o.tenant_id IS NULL
              AND o.organization_id = t.organization_id
        ';
    END IF;

    UPDATE public.organizations o
    SET tenant_id = u.tenant_id
    FROM (
        SELECT om.organization_id, MAX(u.tenant_id::text)::uuid AS tenant_id
        FROM public.organization_memberships om
        JOIN public.users u ON u.user_id = om.user_id
        WHERE u.tenant_id IS NOT NULL
        GROUP BY om.organization_id
    ) u
    WHERE o.tenant_id IS NULL
      AND o.organization_id = u.organization_id;

    UPDATE public.organizations o
    SET tenant_id = t.tenant_id
    FROM (
        SELECT tenant_id
        FROM public.tenants
        ORDER BY created_at ASC
        LIMIT 1
    ) t
    WHERE o.tenant_id IS NULL;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM public.organizations WHERE tenant_id IS NULL) THEN
        ALTER TABLE public.organizations
            ALTER COLUMN tenant_id SET NOT NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'organizations_tenant_id_fkey'
    ) THEN
        ALTER TABLE public.organizations
            ADD CONSTRAINT organizations_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_organizations_tenant_id ON public.organizations (tenant_id);
CREATE INDEX IF NOT EXISTS idx_organizations_tenant_alias ON public.organizations (tenant_id, organization_alias);

ALTER TABLE public.organization_memberships
    ADD COLUMN IF NOT EXISTS tenant_id uuid;

UPDATE public.organization_memberships om
SET tenant_id = o.tenant_id
FROM public.organizations o
WHERE om.tenant_id IS NULL
  AND om.organization_id = o.organization_id
  AND o.tenant_id IS NOT NULL;

UPDATE public.organization_memberships om
SET tenant_id = u.tenant_id
FROM public.users u
WHERE om.tenant_id IS NULL
  AND om.user_id = u.user_id
  AND u.tenant_id IS NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM public.organization_memberships WHERE tenant_id IS NULL) THEN
        ALTER TABLE public.organization_memberships
            ALTER COLUMN tenant_id SET NOT NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'organization_memberships_tenant_id_fkey'
    ) THEN
        ALTER TABLE public.organization_memberships
            ADD CONSTRAINT organization_memberships_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_org_memberships_tenant_org ON public.organization_memberships (tenant_id, organization_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_org_memberships_tenant_org;
DROP INDEX IF EXISTS idx_organizations_tenant_alias;
DROP INDEX IF EXISTS idx_organizations_tenant_id;

ALTER TABLE public.organization_memberships
    DROP CONSTRAINT IF EXISTS organization_memberships_tenant_id_fkey;

ALTER TABLE public.organization_memberships
    DROP COLUMN IF EXISTS tenant_id;

ALTER TABLE public.organizations
    DROP CONSTRAINT IF EXISTS organizations_tenant_id_fkey;

ALTER TABLE public.organizations
    DROP COLUMN IF EXISTS tenant_id;
-- +goose StatementEnd
