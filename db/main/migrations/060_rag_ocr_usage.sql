-- +goose Up
-- +goose StatementBegin
ALTER TABLE rag_usage_events
    ADD COLUMN IF NOT EXISTS ocr_pages bigint NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rag_usage_events DROP COLUMN IF EXISTS ocr_pages;
-- +goose StatementEnd
