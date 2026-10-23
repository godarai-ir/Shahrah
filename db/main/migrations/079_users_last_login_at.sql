-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS last_login_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.users DROP COLUMN IF EXISTS last_login_at;
-- +goose StatementEnd
