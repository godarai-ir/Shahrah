-- +goose Up
-- +goose StatementBegin
ALTER TABLE document_versions
    ADD COLUMN IF NOT EXISTS detected_language text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE document_versions DROP COLUMN IF EXISTS detected_language;
-- +goose StatementEnd
