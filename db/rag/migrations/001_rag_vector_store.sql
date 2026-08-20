-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS rag_chunks (
    chunk_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    knowledge_base_id uuid NOT NULL,
    document_id uuid NOT NULL,
    document_version_id uuid NOT NULL,
    title text NOT NULL,
    chunk_index integer NOT NULL,
    content text NOT NULL,
    content_sha256 text NOT NULL,
    token_count integer NOT NULL,
    page integer,
    heading text NOT NULL DEFAULT '',
    embedding vector(1536) NOT NULL,
    active boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (document_version_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_rag_chunks_tenant_kb_active
    ON rag_chunks (tenant_id, knowledge_base_id, active);

CREATE INDEX IF NOT EXISTS idx_rag_chunks_document_version
    ON rag_chunks (tenant_id, document_version_id);

CREATE INDEX IF NOT EXISTS idx_rag_chunks_embedding_hnsw
    ON rag_chunks USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rag_chunks;
-- +goose StatementEnd
