-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.spend_logs
    ADD COLUMN usage_breakdown_json jsonb,
    ADD COLUMN pricing_snapshot_json jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.spend_logs
    DROP COLUMN IF EXISTS usage_breakdown_json,
    DROP COLUMN IF EXISTS pricing_snapshot_json;
-- +goose StatementEnd
