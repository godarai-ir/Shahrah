-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS public.shahrah_sql_migrations;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.shahrah_sql_migrations (
    filename TEXT PRIMARY KEY,
    checksum TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd
