package store

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/jackc/pgx/v5"
)

// VectorChunk is one inactive chunk staged before atomic document activation.
type VectorChunk struct {
	TenantID              string
	KnowledgeBaseID       string
	DocumentID            string
	VersionID             string
	Title                 string
	Chunk                 domain.Chunk
	Embedding             []float32
	NormalizationRevision string
}

// SearchCandidate is a hybrid retrieval candidate before global reranking.
type SearchCandidate struct {
	KnowledgeBaseID string
	DocumentID      string
	VersionID       string
	Title           string
	Text            string
	IndexText       string
	Page            *int
	Heading         string
	ContentSHA256   string
	TokenCount      int
	RetrievalScore  float64
}

// ReplaceVersionChunks idempotently replaces an inactive version's staged chunks.
func (s *Store) ReplaceVersionChunks(ctx context.Context, chunks []VectorChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	tx, err := s.vectors.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin replace vector chunks: %w", err)
	}
	defer rollback(ctx, tx)

	first := chunks[0]
	if _, err := tx.Exec(ctx, `
		DELETE FROM rag_chunks
		WHERE tenant_id = $1 AND document_version_id = $2 AND active = false`, first.TenantID, first.VersionID); err != nil {
		return fmt.Errorf("delete staged vector chunks: %w", err)
	}

	batch := &pgx.Batch{}

	for _, item := range chunks {
		embedding1536, embedding1024, err := embeddingColumns(item.Embedding)
		if err != nil {
			return err
		}

		batch.Queue(`
			INSERT INTO rag_chunks (
				tenant_id, knowledge_base_id, document_id, document_version_id, title,
				chunk_index, content, index_content, normalization_revision, content_sha256,
				token_count, page, heading, embedding, embedding_1024, active
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14::vector,$15::vector,false)`,
			item.TenantID, item.KnowledgeBaseID, item.DocumentID, item.VersionID, item.Title,
			item.Chunk.Index, item.Chunk.Text, item.Chunk.IndexText, item.NormalizationRevision,
			item.Chunk.SHA256, item.Chunk.TokenCount, item.Chunk.Page, item.Chunk.Heading,
			embedding1536, embedding1024)
	}

	results := tx.SendBatch(ctx, batch)
	for range chunks {
		if _, err := results.Exec(); err != nil {
			closeErr := results.Close()
			if closeErr != nil {
				return errors.Join(fmt.Errorf("insert vector chunk: %w", err), fmt.Errorf("close vector chunk batch: %w", closeErr))
			}

			return fmt.Errorf("insert vector chunk: %w", err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("close vector chunk batch: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit vector chunks: %w", err)
	}

	return nil
}

// SearchChunks fuses dense and Persian-compatible trigram ranks before global reranking.
func (s *Store) SearchChunks(
	ctx context.Context,
	tenantID string,
	knowledgeBaseIDs []string,
	query string,
	embedding []float32,
	perKnowledgeBaseLimit, totalLimit, rrfRankConstant int,
	trigramThreshold float64,
) ([]SearchCandidate, error) {
	embeddingColumn, err := searchEmbeddingColumn(len(embedding))
	if err != nil {
		return nil, err
	}

	tx, err := s.vectors.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin vector search: %w", err)
	}
	defer rollback(ctx, tx)

	_, err = tx.Exec(ctx, `SET LOCAL hnsw.iterative_scan = strict_order`)
	if err != nil {
		return nil, fmt.Errorf("configure filtered vector search: %w", err)
	}

	_, err = tx.Exec(ctx, `SELECT set_config('pg_trgm.word_similarity_threshold', $1, true)`,
		strconv.FormatFloat(trigramThreshold, 'f', -1, 64))
	if err != nil {
		return nil, fmt.Errorf("configure lexical search: %w", err)
	}

	querySQL := fmt.Sprintf(`
		WITH dense_ranked AS (
			SELECT chunk_id, knowledge_base_id,
			       row_number() OVER (
			           PARTITION BY knowledge_base_id ORDER BY %s <=> $4::vector
			       ) AS dense_rank
			FROM rag_chunks
			WHERE tenant_id = $1 AND knowledge_base_id = ANY($2::uuid[]) AND active = true
			  AND %s IS NOT NULL
		), dense AS (
			SELECT chunk_id, knowledge_base_id, dense_rank
			FROM dense_ranked
			WHERE dense_rank <= $5
		), lexical_ranked AS (
			SELECT chunk_id, knowledge_base_id,
			       row_number() OVER (
			           PARTITION BY knowledge_base_id
			           ORDER BY word_similarity($3, index_content) DESC
			       ) AS lexical_rank
			FROM rag_chunks
			WHERE tenant_id = $1 AND knowledge_base_id = ANY($2::uuid[]) AND active = true
			  AND $3 <%% index_content
		), lexical AS (
			SELECT chunk_id, knowledge_base_id, lexical_rank
			FROM lexical_ranked
			WHERE lexical_rank <= $5
		), fused AS (
			SELECT COALESCE(d.chunk_id, l.chunk_id) AS chunk_id,
			       COALESCE(d.knowledge_base_id, l.knowledge_base_id) AS knowledge_base_id,
			       COALESCE(1.0 / ($7 + d.dense_rank), 0) +
			       COALESCE(1.0 / ($7 + l.lexical_rank), 0) AS retrieval_score
			FROM dense d
			FULL OUTER JOIN lexical l ON l.chunk_id = d.chunk_id
		), limited AS (
			SELECT chunk_id, retrieval_score,
			       row_number() OVER (
			           PARTITION BY knowledge_base_id ORDER BY retrieval_score DESC
			       ) AS kb_rank
			FROM fused
		)
		SELECT c.knowledge_base_id::text, c.document_id::text, c.document_version_id::text,
		       c.title, c.content, c.index_content, c.page, c.heading, c.content_sha256,
		       c.token_count, limited.retrieval_score
		FROM limited
		JOIN rag_chunks c ON c.chunk_id = limited.chunk_id
		WHERE limited.kb_rank <= $5
		ORDER BY limited.retrieval_score DESC
		LIMIT $6`, embeddingColumn, embeddingColumn)

	rows, err := tx.Query(ctx, querySQL,
		tenantID, knowledgeBaseIDs, query, vectorLiteral(embedding), perKnowledgeBaseLimit, totalLimit,
		rrfRankConstant)
	if err != nil {
		return nil, fmt.Errorf("search hybrid chunks: %w", err)
	}
	defer rows.Close()

	candidates := make([]SearchCandidate, 0, totalLimit)

	for rows.Next() {
		var item SearchCandidate
		if err := rows.Scan(
			&item.KnowledgeBaseID, &item.DocumentID, &item.VersionID, &item.Title,
			&item.Text, &item.IndexText, &item.Page, &item.Heading, &item.ContentSHA256, &item.TokenCount,
			&item.RetrievalScore,
		); err != nil {
			return nil, fmt.Errorf("scan vector candidate: %w", err)
		}

		candidates = append(candidates, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vector candidates: %w", err)
	}

	rows.Close()

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit vector search: %w", err)
	}

	return candidates, nil
}

// PurgeTenantVectors removes every vector for a deleted tenant.
func (s *Store) PurgeTenantVectors(ctx context.Context, tenantID string) error {
	if _, err := s.vectors.Exec(ctx, `DELETE FROM rag_chunks WHERE tenant_id = $1`, tenantID); err != nil {
		return fmt.Errorf("purge tenant vectors: %w", err)
	}

	return nil
}

// PurgeKnowledgeBaseVectors removes vectors after the metadata collection is durably hidden.
func (s *Store) PurgeKnowledgeBaseVectors(ctx context.Context, tenantID, knowledgeBaseID string) error {
	if _, err := s.vectors.Exec(ctx, `
		DELETE FROM rag_chunks WHERE tenant_id = $1 AND knowledge_base_id = $2`, tenantID, knowledgeBaseID); err != nil {
		return fmt.Errorf("purge knowledge base vectors: %w", err)
	}

	return nil
}

// PurgeDocumentVectors removes vectors after the metadata document is durably hidden.
func (s *Store) PurgeDocumentVectors(ctx context.Context, tenantID, documentID string) error {
	if _, err := s.vectors.Exec(ctx, `
		DELETE FROM rag_chunks WHERE tenant_id = $1 AND document_id = $2`, tenantID, documentID); err != nil {
		return fmt.Errorf("purge document vectors: %w", err)
	}

	return nil
}

func (s *Store) restoreVectorActivation(
	ctx context.Context,
	tenantID, documentID, failedVersionID string,
	previousVersionID *string,
) error {
	tx, err := s.vectors.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin vector activation compensation: %w", err)
	}
	defer rollback(ctx, tx)

	if _, err := tx.Exec(ctx, `
		UPDATE rag_chunks SET active = false
		WHERE tenant_id = $1 AND document_id = $2 AND document_version_id = $3`,
		tenantID, documentID, failedVersionID); err != nil {
		return fmt.Errorf("deactivate failed vector version: %w", err)
	}

	if previousVersionID != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE rag_chunks SET active = true
			WHERE tenant_id = $1 AND document_id = $2 AND document_version_id = $3`,
			tenantID, documentID, *previousVersionID); err != nil {
			return fmt.Errorf("reactivate previous vector version: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit vector activation compensation: %w", err)
	}

	return nil
}

func vectorLiteral(vector []float32) string {
	var builder strings.Builder
	builder.Grow(len(vector)*10 + 2)
	builder.WriteByte('[')

	for i, value := range vector {
		if i > 0 {
			builder.WriteByte(',')
		}

		builder.WriteString(strconv.FormatFloat(float64(value), 'g', -1, 32))
	}

	builder.WriteByte(']')

	return builder.String()
}

func embeddingColumns(embedding []float32) (any, any, error) {
	switch len(embedding) {
	case 1536:
		return vectorLiteral(embedding), nil, nil
	case 1024:
		return nil, vectorLiteral(embedding), nil
	default:
		return nil, nil, fmt.Errorf("unsupported RAG embedding dimensions: %d", len(embedding))
	}
}

func searchEmbeddingColumn(dimensions int) (string, error) {
	switch dimensions {
	case 1536:
		return "embedding", nil
	case 1024:
		return "embedding_1024", nil
	default:
		return "", fmt.Errorf("unsupported RAG embedding dimensions: %d", dimensions)
	}
}
