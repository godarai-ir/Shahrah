-- +goose Up
-- +goose StatementBegin
INSERT INTO rag_profiles (
    revision, embedding_model, embedding_dimensions, rerank_model,
    chunk_tokens, chunk_overlap_tokens, max_chunks, max_context_tokens,
    configuration, active
)
VALUES (
    'rag-v2-fa', 'text-embedding-3-small', 1536, 'rerank-v4.0-pro',
    700, 100, 8, 8000,
    '{
      "retrieval_per_knowledge_base": 10,
      "candidate_limit": 50,
      "minimum_rerank_score": 0.05,
      "conversation_turns": 6,
      "conversation_tokens": 6000,
      "tokenizer": "cl100k_base",
      "normalization_revision": "fa-ir-v2",
      "retrieval_mode": "dense_trigram_rrf",
      "rrf_rank_constant": 60,
      "trigram_word_similarity_threshold": 0.2,
      "ocr_model": "mistral-ocr-4-0",
      "ocr_languages": ["fa", "en"],
      "ocr_trigger": ["empty_text", "low_extraction_quality"],
      "culture_validation": ["currency_amount", "numeric_date"],
      "display_text_preserved": true,
      "canonical_digits": "ascii",
      "remove_arabic_diacritics_from_index": true,
      "remove_bidi_controls_from_index": true
    }'::jsonb,
    false
)
ON CONFLICT (revision) DO NOTHING;

UPDATE rag_profiles SET active = false WHERE active = true;
UPDATE rag_profiles SET active = true WHERE revision = 'rag-v2-fa';

ALTER TABLE document_versions
    ADD COLUMN IF NOT EXISTS normalization_revision text NOT NULL DEFAULT 'fa-ir-v1';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE rag_profiles SET active = false WHERE active = true;
UPDATE rag_profiles SET active = true WHERE revision = 'rag-v1';
ALTER TABLE document_versions DROP COLUMN IF EXISTS normalization_revision;
DELETE FROM rag_profiles WHERE revision = 'rag-v2-fa';
-- +goose StatementEnd
