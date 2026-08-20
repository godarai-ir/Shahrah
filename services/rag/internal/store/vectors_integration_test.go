package store

import (
	"context"
	"os"
	"testing"

	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSearchChunksIncludesPersianLexicalCandidate(t *testing.T) {
	const embeddingDimensions = 1536

	connection := os.Getenv("RAG_TEST_VECTOR_DATABASE_URL")
	if connection == "" {
		t.Skip("RAG_TEST_VECTOR_DATABASE_URL is not configured")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connection)
	if err != nil {
		t.Fatalf("connect vector integration database: %v", err)
	}
	defer pool.Close()

	if _, truncateErr := pool.Exec(ctx, `TRUNCATE rag_chunks`); truncateErr != nil {
		t.Fatalf("truncate vector fixtures: %v", truncateErr)
	}

	tenantID := uuid.NewString()
	knowledgeBaseID := uuid.NewString()

	queryVector := make([]float32, embeddingDimensions)
	queryVector[0] = 1

	fixtures := []struct {
		title, indexText string
		first, second    float32
	}{
		{title: "dense-first", indexText: "متن نامرتبط نخست", first: 1},
		{title: "dense-second", indexText: "متن نامرتبط دوم", first: 0.9, second: 0.1},
		{title: "lexical", indexText: "شماره کالای 123 آماده ارسال است", first: -1},
	}

	for index, fixture := range fixtures {
		embedding := make([]float32, embeddingDimensions)
		embedding[0] = fixture.first
		embedding[1] = fixture.second

		_, insertErr := pool.Exec(ctx, `
			INSERT INTO rag_chunks (
				tenant_id, knowledge_base_id, document_id, document_version_id, title,
				chunk_index, content, index_content, normalization_revision, content_sha256,
				token_count, embedding, active
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::vector,true)`,
			tenantID, knowledgeBaseID, uuid.NewString(), uuid.NewString(), fixture.title,
			index, fixture.indexText, fixture.indexText, "fa-ir-test",
			uuid.NewString(), 10, vectorLiteral(embedding),
		)
		if insertErr != nil {
			t.Fatalf("insert vector fixture %d: %v", index, insertErr)
		}
	}

	repository := &Store{vectors: pool}

	candidates, err := repository.SearchChunks(
		ctx, tenantID, []string{knowledgeBaseID}, domain.NormalizePersian("كالا ۱۲۳"), queryVector, 2, 10, 60, 0.2,
	)
	if err != nil {
		t.Fatalf("search hybrid chunks: %v", err)
	}

	foundLexical := false

	for _, candidate := range candidates {
		if candidate.Title == "lexical" {
			foundLexical = true
		}
	}

	if !foundLexical {
		t.Fatalf("Persian lexical candidate was absent: %#v", candidates)
	}
}
