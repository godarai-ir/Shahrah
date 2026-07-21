-- +goose Up
-- The legacy hierarchical budget model was replaced by budget-limit rules.
-- Refuse to discard unexpected data even though all runtime references are gone.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.budgets) THEN
        RAISE EXCEPTION 'refusing to drop public.budgets because it contains data';
    END IF;
END
$$;
-- +goose StatementEnd

DROP TABLE public.budgets;

-- +goose Down
CREATE TABLE public.budgets (
    budget_id uuid DEFAULT gen_random_uuid() NOT NULL,
    max_budget numeric(10,4),
    soft_budget numeric(10,4),
    tpm_limit bigint,
    rpm_limit bigint,
    max_parallel_requests integer,
    budget_duration character varying(50),
    budget_reset_at timestamp with time zone,
    model_max_budget jsonb,
    created_at timestamp with time zone DEFAULT now(),
    created_by character varying(255),
    updated_at timestamp with time zone DEFAULT now(),
    updated_by character varying(255),
    CONSTRAINT budgets_pkey PRIMARY KEY (budget_id)
);
