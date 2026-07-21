-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.projects) THEN
        RAISE EXCEPTION 'refusing to drop public.projects because it contains data';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM public.verification_tokens
        WHERE project_id IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'refusing to drop verification_tokens.project_id because it contains data';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE ONLY public.verification_tokens
    DROP CONSTRAINT verification_tokens_project_id_fkey;

ALTER TABLE ONLY public.verification_tokens
    DROP COLUMN project_id;

DROP TABLE public.projects;

-- +goose Down
CREATE TABLE public.projects (
    project_id uuid DEFAULT gen_random_uuid() NOT NULL,
    project_alias character varying(255),
    description character varying(255),
    team_id uuid,
    budget_id uuid,
    metadata jsonb DEFAULT '{}'::jsonb,
    models text[] DEFAULT '{}'::text[],
    spend numeric(10,4) DEFAULT 0.0,
    model_spend jsonb DEFAULT '{}'::jsonb,
    model_rpm_limit jsonb DEFAULT '{}'::jsonb,
    model_tpm_limit jsonb DEFAULT '{}'::jsonb,
    blocked boolean DEFAULT false,
    object_permission_id uuid,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255)
);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (project_id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budgets(budget_id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_object_permission_id_fkey FOREIGN KEY (object_permission_id) REFERENCES public.object_permissions(object_permission_id);

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id);

ALTER TABLE ONLY public.verification_tokens
    ADD COLUMN project_id uuid;

ALTER TABLE ONLY public.verification_tokens
    ADD CONSTRAINT verification_tokens_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id);
