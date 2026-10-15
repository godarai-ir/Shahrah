-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.users
    ADD COLUMN IF NOT EXISTS first_name text,
    ADD COLUMN IF NOT EXISTS last_name text;

ALTER TABLE public.tenant_admin_invitations
    ADD COLUMN IF NOT EXISTS first_name text,
    ADD COLUMN IF NOT EXISTS last_name text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.tenant_admin_invitations
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS first_name;

ALTER TABLE public.users
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS first_name;
-- +goose StatementEnd
