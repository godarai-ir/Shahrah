-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.end_users) THEN
        RAISE EXCEPTION 'refusing to drop public.end_users because it contains data';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM public.spend_logs
        WHERE end_user_id IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'refusing to drop spend_logs.end_user_id because it contains data';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE ONLY public.spend_logs
    DROP COLUMN end_user_id;

DROP TABLE public.end_users;

-- +goose Down
CREATE TABLE public.end_users (
    user_id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_alias character varying(255),
    team_id uuid,
    budget_id uuid,
    spend numeric(10,4) DEFAULT 0.0,
    max_budget numeric(10,4),
    tpm_limit bigint,
    rpm_limit bigint,
    blocked boolean DEFAULT false,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

ALTER TABLE ONLY public.end_users
    ADD CONSTRAINT end_users_pkey PRIMARY KEY (user_id);

CREATE INDEX idx_end_user_team ON public.end_users USING btree (team_id);

ALTER TABLE ONLY public.end_users
    ADD CONSTRAINT end_users_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budgets(budget_id);

ALTER TABLE ONLY public.end_users
    ADD CONSTRAINT end_users_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.teams(team_id);

ALTER TABLE ONLY public.spend_logs
    ADD COLUMN end_user_id uuid;
