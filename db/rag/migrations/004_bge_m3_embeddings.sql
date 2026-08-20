-- +goose Up
-- +goose StatementBegin
ALTER TABLE rag_chunks
    ADD COLUMN IF NOT EXISTS embedding_1024 vector(1024);

ALTER TABLE rag_chunks
    ALTER COLUMN embedding DROP NOT NULL;

ALTER TABLE rag_chunks
    DROP CONSTRAINT IF EXISTS rag_chunks_embedding_dimensions_check;

ALTER TABLE rag_chunks
    ADD CONSTRAINT rag_chunks_embedding_dimensions_check CHECK (
        (embedding IS NOT NULL AND embedding_1024 IS NULL)
        OR (embedding IS NULL AND embedding_1024 IS NOT NULL)
    );

CREATE INDEX IF NOT EXISTS idx_rag_chunks_embedding_1024_hnsw
    ON rag_chunks USING hnsw (embedding_1024 vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_rag_chunks_embedding_1024_hnsw;
ALTER TABLE rag_chunks DROP CONSTRAINT IF EXISTS rag_chunks_embedding_dimensions_check;
DELETE FROM rag_chunks WHERE embedding IS NULL;
ALTER TABLE rag_chunks ALTER COLUMN embedding SET NOT NULL;
ALTER TABLE rag_chunks DROP COLUMN IF EXISTS embedding_1024;
-- +goose StatementEnd
