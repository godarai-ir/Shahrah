-- +goose Up
-- +goose StatementBegin
ALTER TABLE rag_chunks
    ADD COLUMN IF NOT EXISTS index_content text,
    ADD COLUMN IF NOT EXISTS normalization_revision text NOT NULL DEFAULT 'fa-ir-v1';

UPDATE rag_chunks SET index_content = content WHERE index_content IS NULL;

ALTER TABLE rag_chunks ALTER COLUMN index_content SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE rag_chunks
    DROP COLUMN IF EXISTS normalization_revision,
    DROP COLUMN IF EXISTS index_content;
-- +goose StatementEnd
