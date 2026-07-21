-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.tags) THEN
        RAISE EXCEPTION 'refusing to drop public.tags because it contains data';
    END IF;
END
$$;
-- +goose StatementEnd

DROP TABLE public.tags;

-- +goose Down
CREATE TABLE public.tags (
    tag_name character varying(255) NOT NULL,
    budget_id uuid,
    spend numeric(10,4) DEFAULT 0.0,
    max_budget numeric(10,4),
    tpm_limit bigint,
    rpm_limit bigint,
    models text[] DEFAULT '{}'::text[],
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now()
);

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (tag_name);

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_budget_id_fkey FOREIGN KEY (budget_id) REFERENCES public.budgets(budget_id);
