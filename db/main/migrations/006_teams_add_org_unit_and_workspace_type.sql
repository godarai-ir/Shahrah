-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.teams
    ADD COLUMN IF NOT EXISTS org_unit_id uuid,
    ADD COLUMN IF NOT EXISTS workspace_type character varying(64);

UPDATE public.teams
SET org_unit_id = organization_id
WHERE org_unit_id IS NULL
  AND organization_id IS NOT NULL;

UPDATE public.teams
SET workspace_type = 'general'
WHERE workspace_type IS NULL OR TRIM(workspace_type) = '';

ALTER TABLE public.teams
    ALTER COLUMN workspace_type SET DEFAULT 'general';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'teams_org_unit_id_fkey'
    ) THEN
        ALTER TABLE public.teams
            ADD CONSTRAINT teams_org_unit_id_fkey
            FOREIGN KEY (org_unit_id) REFERENCES public.organizations(organization_id);
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_teams_org_unit_id ON public.teams (org_unit_id);
CREATE INDEX IF NOT EXISTS idx_teams_workspace_type ON public.teams (workspace_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_teams_workspace_type;
DROP INDEX IF EXISTS idx_teams_org_unit_id;

ALTER TABLE public.teams
    DROP CONSTRAINT IF EXISTS teams_org_unit_id_fkey;

ALTER TABLE public.teams
    DROP COLUMN IF EXISTS workspace_type,
    DROP COLUMN IF EXISTS org_unit_id;
-- +goose StatementEnd
