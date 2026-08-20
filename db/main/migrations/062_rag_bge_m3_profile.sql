-- +goose Up
-- +goose StatementBegin
INSERT INTO rag_profiles (
    revision, embedding_model, embedding_dimensions, rerank_model,
    chunk_tokens, chunk_overlap_tokens, max_chunks, max_context_tokens,
    configuration, active
)
SELECT
    'rag-v3-fa-bge-m3', 'Bge-m3', 1024, rerank_model,
    chunk_tokens, chunk_overlap_tokens, max_chunks, max_context_tokens,
    configuration || '{"embedding_provider":"arvancloud","embedding_revision":"v1"}'::jsonb,
    false
FROM rag_profiles
WHERE revision = 'rag-v2-fa'
ON CONFLICT (revision) DO UPDATE SET
    embedding_model = EXCLUDED.embedding_model,
    embedding_dimensions = EXCLUDED.embedding_dimensions,
    rerank_model = EXCLUDED.rerank_model,
    chunk_tokens = EXCLUDED.chunk_tokens,
    chunk_overlap_tokens = EXCLUDED.chunk_overlap_tokens,
    max_chunks = EXCLUDED.max_chunks,
    max_context_tokens = EXCLUDED.max_context_tokens,
    configuration = EXCLUDED.configuration;

UPDATE rag_profiles SET active = false WHERE active = true;
UPDATE rag_profiles SET active = true WHERE revision = 'rag-v3-fa-bge-m3';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE rag_profiles SET active = false WHERE active = true;
UPDATE rag_profiles SET active = true WHERE revision = 'rag-v2-fa';
DELETE FROM rag_profiles WHERE revision = 'rag-v3-fa-bge-m3';
-- +goose StatementEnd
