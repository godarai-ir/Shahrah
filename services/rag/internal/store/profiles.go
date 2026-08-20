package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/jackc/pgx/v5"
)

// ErrProfileMismatch indicates that resources pinned to different profiles were combined.
var ErrProfileMismatch = errors.New("RAG profile revisions do not match")

// LoadActiveRAGProfile returns the single profile used for newly created knowledge bases.
func (s *Store) LoadActiveRAGProfile(ctx context.Context) (domain.RAGProfile, error) {
	return scanRAGProfile(s.metadata.QueryRow(ctx, `
		SELECT revision, embedding_model, embedding_dimensions, rerank_model,
		       chunk_tokens, chunk_overlap_tokens, max_chunks, max_context_tokens, configuration
		FROM rag_profiles WHERE active = true`))
}

// LoadRAGProfile returns an immutable profile revision.
func (s *Store) LoadRAGProfile(ctx context.Context, revision string) (domain.RAGProfile, error) {
	return scanRAGProfile(s.metadata.QueryRow(ctx, `
		SELECT revision, embedding_model, embedding_dimensions, rerank_model,
		       chunk_tokens, chunk_overlap_tokens, max_chunks, max_context_tokens, configuration
		FROM rag_profiles WHERE revision = $1`, revision))
}

// ResolveKnowledgeBaseProfile validates tenant ownership and returns the common pinned profile.
func (s *Store) ResolveKnowledgeBaseProfile(
	ctx context.Context,
	tenantID string,
	knowledgeBaseIDs []string,
) (domain.RAGProfile, error) {
	var (
		count     int
		revisions []string
	)

	err := s.metadata.QueryRow(ctx, `
		SELECT count(*)::int, array_agg(DISTINCT rag_profile_revision)
		FROM knowledge_bases
		WHERE tenant_id = $1 AND knowledge_base_id = ANY($2::uuid[]) AND deleted_at IS NULL`,
		tenantID, knowledgeBaseIDs).Scan(&count, &revisions)
	if err != nil {
		return domain.RAGProfile{}, fmt.Errorf("resolve knowledge base profile: %w", err)
	}

	if count != len(knowledgeBaseIDs) {
		return domain.RAGProfile{}, ErrNotFound
	}

	if len(revisions) != 1 {
		return domain.RAGProfile{}, ErrProfileMismatch
	}

	return s.LoadRAGProfile(ctx, revisions[0])
}

func scanRAGProfile(row rowScanner) (domain.RAGProfile, error) {
	var (
		profile       domain.RAGProfile
		configuration []byte
	)

	err := row.Scan(
		&profile.Revision, &profile.EmbeddingModel, &profile.EmbeddingDimensions, &profile.RerankModel,
		&profile.ChunkTokens, &profile.ChunkOverlapTokens, &profile.MaxChunks, &profile.MaxContextTokens,
		&configuration,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.RAGProfile{}, ErrNotFound
	}

	if err != nil {
		return domain.RAGProfile{}, fmt.Errorf("scan RAG profile: %w", err)
	}

	if err := json.Unmarshal(configuration, &profile.Configuration); err != nil {
		return domain.RAGProfile{}, fmt.Errorf("decode RAG profile %q: %w", profile.Revision, err)
	}

	if err := profile.Validate(); err != nil {
		return domain.RAGProfile{}, fmt.Errorf("validate RAG profile %q: %w", profile.Revision, err)
	}

	return profile, nil
}
