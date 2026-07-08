-- +goose Up
CREATE TABLE IF NOT EXISTS public.wallet_budget_window_settlements (
    tenant_id uuid NOT NULL REFERENCES public.tenants(tenant_id) ON DELETE CASCADE,
    window_id character varying(16) NOT NULL,
    settled_irr numeric(18,0) NOT NULL DEFAULT 0,
    last_observed_committed_irr numeric(18,0) NOT NULL DEFAULT 0,
    last_observed_at timestamp with time zone,
    last_settled_at timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, window_id)
);

CREATE INDEX IF NOT EXISTS idx_wallet_budget_window_settlements_observed
    ON public.wallet_budget_window_settlements (last_observed_at DESC);

-- +goose Down
DROP TABLE IF EXISTS public.wallet_budget_window_settlements;
