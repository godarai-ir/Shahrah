-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.teams
    ADD COLUMN IF NOT EXISTS tenant_id uuid;

DO $$
DECLARE
    has_org_unit_id boolean;
    has_organization_id boolean;
    has_team_memberships boolean;
    has_team_memberships_tenant_id boolean;
    has_organizations_tenant_id boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'teams' AND column_name = 'org_unit_id'
    ) INTO has_org_unit_id;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'teams' AND column_name = 'organization_id'
    ) INTO has_organization_id;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'team_memberships'
    ) INTO has_team_memberships;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'team_memberships' AND column_name = 'tenant_id'
    ) INTO has_team_memberships_tenant_id;

    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'organizations' AND column_name = 'tenant_id'
    ) INTO has_organizations_tenant_id;

    IF has_team_memberships AND has_team_memberships_tenant_id THEN
        EXECUTE '
            UPDATE public.teams t
            SET tenant_id = tm.tenant_id
            FROM (
                SELECT team_id, MAX(tenant_id::text)::uuid AS tenant_id
                FROM public.team_memberships
                WHERE tenant_id IS NOT NULL
                GROUP BY team_id
            ) tm
            WHERE t.tenant_id IS NULL
              AND t.team_id = tm.team_id
        ';
    END IF;

    IF has_org_unit_id AND has_organizations_tenant_id THEN
        EXECUTE '
            UPDATE public.teams t
            SET tenant_id = o.tenant_id
            FROM public.organizations o
            WHERE t.tenant_id IS NULL
              AND t.org_unit_id = o.organization_id
              AND o.tenant_id IS NOT NULL
        ';
    ELSIF has_organization_id AND has_organizations_tenant_id THEN
        EXECUTE '
            UPDATE public.teams t
            SET tenant_id = o.tenant_id
            FROM public.organizations o
            WHERE t.tenant_id IS NULL
              AND t.organization_id = o.organization_id
              AND o.tenant_id IS NOT NULL
        ';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM public.teams WHERE tenant_id IS NULL) THEN
        ALTER TABLE public.teams
            ALTER COLUMN tenant_id SET NOT NULL;
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'teams_tenant_id_fkey'
    ) THEN
        ALTER TABLE public.teams
            ADD CONSTRAINT teams_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES public.tenants(tenant_id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_teams_tenant_id ON public.teams (tenant_id);
CREATE INDEX IF NOT EXISTS idx_teams_tenant_alias ON public.teams (tenant_id, team_alias);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_teams_tenant_alias;
DROP INDEX IF EXISTS idx_teams_tenant_id;

ALTER TABLE public.teams
    DROP CONSTRAINT IF EXISTS teams_tenant_id_fkey;

ALTER TABLE public.teams
    DROP COLUMN IF EXISTS tenant_id;
-- +goose StatementEnd
