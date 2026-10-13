-- +goose Up
-- +goose StatementBegin
-- The product label for gateway usage made with this client's access tokens,
-- for example librechat. It is chosen by tenant admins and stamped into the
-- access tokens the client receives.
ALTER TABLE public.oidc_clients
    ADD COLUMN IF NOT EXISTS credit_product character varying(32) NOT NULL DEFAULT 'oidc_app';

ALTER TABLE public.oidc_clients
    ADD CONSTRAINT chk_oidc_clients_credit_product
        CHECK (credit_product IN ('oidc_app', 'librechat'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.oidc_clients
    DROP CONSTRAINT IF EXISTS chk_oidc_clients_credit_product,
    DROP COLUMN IF EXISTS credit_product;
-- +goose StatementEnd
