-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.users
    DROP CONSTRAINT IF EXISTS users_email_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_tenant_email_ci
    ON public.users (tenant_id, LOWER(email))
    WHERE email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_tenant_mobile
    ON public.users (tenant_id, mobile)
    WHERE mobile IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_users_tenant_mobile;
DROP INDEX IF EXISTS uq_users_tenant_email_ci;

ALTER TABLE public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);
-- +goose StatementEnd
