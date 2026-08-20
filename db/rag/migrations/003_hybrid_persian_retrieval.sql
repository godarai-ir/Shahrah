-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_rag_chunks_index_content_trgm
    ON rag_chunks USING gin (index_content gin_trgm_ops)
    WHERE active = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_rag_chunks_index_content_trgm;
-- pg_trgm may be shared by another feature and is intentionally retained.
-- +goose StatementEnd
