// Package app implements managed RAG product rules independently of HTTP and storage details.
package app

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/Delpak/shahrah/services/rag/internal/store"
	"github.com/google/uuid"
)

// RetrievalStore is the tenant-safe persistence required by retrieval.
type RetrievalStore interface {
	RequireRAGEnabled(context.Context, string) error
	RequireQueryCapacity(context.Context, string) error
	ResolveKnowledgeBaseProfile(context.Context, string, []string) (domain.RAGProfile, error)
	SearchChunks(context.Context, string, []string, string, []float32, int, int, int, float64) ([]store.SearchCandidate, error)
}

// RuntimeStore adds immutable chatbot publication and session persistence.
type RuntimeStore interface {
	RetrievalStore
	ResolveSession(context.Context, store.PublishedChatbot, string) (string, error)
	LoadConversation(context.Context, store.PublishedChatbot, string) ([]store.ConversationMessage, error)
	SaveConversationTurn(context.Context, store.PublishedChatbot, string, string, string, bool, []platform.Citation) error
	DeleteSession(context.Context, store.PublishedChatbot, string) error
	RecordUsage(context.Context, store.UsageEvent) error
}

// ModelProvider is the fail-closed model boundary used by retrieval and generation.
type ModelProvider interface {
	Embed(context.Context, string, string, []string, int) ([][]float32, int, error)
	Rerank(context.Context, string, string, string, []string, int) ([]contractsopenai.RerankResult, int, error)
	Generate(context.Context, string, string, string, string) (string, string, GenerationUsage, error)
}

// GenerationUsage is authoritative generation metering returned by the gateway.
type GenerationUsage struct {
	InputTokens     int
	OutputTokens    int
	ProviderCostUSD float64
}

// Service implements the fixed versioned RAG pipeline.
type Service struct {
	store    RuntimeStore
	provider ModelProvider
}

// New creates a managed RAG application service.
func New(repository RuntimeStore, provider ModelProvider) (*Service, error) {
	if repository == nil || provider == nil {
		return nil, errors.New("RAG store and model provider are required")
	}

	return &Service{store: repository, provider: provider}, nil
}

// Retrieve runs embedding, hybrid search, deduplication, mandatory reranking, and context packing.
func (s *Service) Retrieve(
	ctx context.Context,
	tenantID, tenantSlug string,
	request platform.RetrievalRequest,
) (*platform.RetrievalResponse, error) {
	response, _, err := s.retrieve(ctx, tenantID, tenantSlug, request, nil, nil)

	return response, err
}

func (s *Service) retrieve(
	ctx context.Context,
	tenantID, tenantSlug string,
	request platform.RetrievalRequest,
	chatbotID *string,
	requiredProfileRevision *string,
) (*platform.RetrievalResponse, domain.RAGProfile, error) {
	startedAt := time.Now()
	query := domain.NormalizePersian(request.Query)
	knowledgeBaseIDs := uniqueStrings(request.KnowledgeBaseIDs)

	if err := s.store.RequireRAGEnabled(ctx, tenantID); err != nil {
		return nil, domain.RAGProfile{}, err
	}

	if err := s.store.RequireQueryCapacity(ctx, tenantID); err != nil {
		return nil, domain.RAGProfile{}, err
	}

	if query == "" {
		return nil, domain.RAGProfile{}, fmt.Errorf("query is required")
	}

	if utf8.RuneCountInString(query) > platform.RAGMaxQueryCharacters {
		return nil, domain.RAGProfile{}, fmt.Errorf("query exceeds %d characters", platform.RAGMaxQueryCharacters)
	}

	if len(knowledgeBaseIDs) < 1 || len(knowledgeBaseIDs) > 5 {
		return nil, domain.RAGProfile{}, fmt.Errorf("knowledge_base_ids must contain between 1 and 5 unique identifiers")
	}

	for _, id := range knowledgeBaseIDs {
		if _, err := uuid.Parse(id); err != nil {
			return nil, domain.RAGProfile{}, fmt.Errorf("invalid knowledge base identifier")
		}
	}

	profile, err := s.store.ResolveKnowledgeBaseProfile(ctx, tenantID, knowledgeBaseIDs)
	if err != nil {
		return nil, domain.RAGProfile{}, err
	}

	if requiredProfileRevision != nil && profile.Revision != *requiredProfileRevision {
		return nil, domain.RAGProfile{}, store.ErrProfileMismatch
	}

	response := &platform.RetrievalResponse{
		RequestID:          "rag_" + uuid.NewString(),
		RAGProfileRevision: profile.Revision,
		Chunks:             []platform.RetrievalChunk{},
	}

	vectors, embeddingTokens, err := s.provider.Embed(
		ctx, tenantSlug, profile.EmbeddingModel, []string{query}, profile.EmbeddingDimensions,
	)
	if err != nil {
		return nil, profile, fmt.Errorf("query embedding unavailable: %w", err)
	}

	response.Usage.EmbeddingInputTokens = embeddingTokens

	if len(vectors) != 1 {
		return nil, profile, s.recordFailedRetrieval(
			ctx, tenantID, knowledgeBaseIDs, chatbotID, response,
			errors.New("query embedding unavailable: invalid vector count"), startedAt,
		)
	}

	candidates, err := s.store.SearchChunks(
		ctx, tenantID, knowledgeBaseIDs, query, vectors[0],
		profile.Configuration.RetrievalPerKnowledgeBase, profile.Configuration.CandidateLimit,
		profile.Configuration.RRFRankConstant, profile.Configuration.TrigramWordSimilarityThreshold,
	)
	if err != nil {
		return nil, profile, s.recordFailedRetrieval(
			ctx, tenantID, knowledgeBaseIDs, chatbotID, response,
			fmt.Errorf("hybrid retrieval unavailable: %w", err), startedAt,
		)
	}

	candidates = deduplicateCandidates(candidates)

	if len(candidates) == 0 {
		usageErr := s.store.RecordUsage(context.WithoutCancel(ctx), store.UsageEvent{
			TenantID: tenantID, RequestID: response.RequestID, KnowledgeBaseIDs: knowledgeBaseIDs,
			ChatbotID: chatbotID,
			Operation: "retrieval", Status: "completed", EmbeddingTokens: embeddingTokens,
			LatencyMS: int(time.Since(startedAt).Milliseconds()),
		})
		if usageErr != nil {
			return nil, profile, fmt.Errorf("record retrieval usage: %w", usageErr)
		}

		return response, profile, nil
	}

	documents := make([]string, len(candidates))
	for i := range candidates {
		documents[i] = candidates[i].IndexText
	}

	ranked, rerankUnits, err := s.provider.Rerank(
		ctx, tenantSlug, profile.RerankModel, query, documents, min(profile.MaxChunks, len(documents)),
	)
	if err != nil {
		response.Usage.RerankUnits = rerankUnits

		return nil, profile, s.recordFailedRetrieval(
			ctx, tenantID, knowledgeBaseIDs, chatbotID, response,
			fmt.Errorf("rerank unavailable: %w", err), startedAt,
		)
	}

	response.Usage.RerankUnits = rerankUnits

	if err := validateRerankResults(ranked, len(candidates), min(profile.MaxChunks, len(documents))); err != nil {
		return nil, profile, s.recordFailedRetrieval(
			ctx, tenantID, knowledgeBaseIDs, chatbotID, response,
			fmt.Errorf("rerank unavailable: %w", err), startedAt,
		)
	}

	contextTokens := 0

	for _, rank := range ranked {
		if rank.RelevanceScore < profile.Configuration.MinimumRerankScore || len(response.Chunks) >= profile.MaxChunks {
			continue
		}

		candidate := candidates[rank.Index]
		if contextTokens+candidate.TokenCount > profile.MaxContextTokens {
			continue
		}

		contextTokens += candidate.TokenCount
		response.Chunks = append(response.Chunks, platform.RetrievalChunk{
			CitationID:      fmt.Sprintf("%d", len(response.Chunks)+1),
			KnowledgeBaseID: candidate.KnowledgeBaseID,
			DocumentID:      candidate.DocumentID,
			DocumentVersion: candidate.VersionID,
			Title:           candidate.Title,
			Page:            candidate.Page,
			Heading:         candidate.Heading,
			Text:            candidate.Text,
			RetrievalScore:  candidate.RetrievalScore,
			RerankScore:     rank.RelevanceScore,
		})
	}

	usageErr := s.store.RecordUsage(context.WithoutCancel(ctx), store.UsageEvent{
		TenantID: tenantID, RequestID: response.RequestID, KnowledgeBaseIDs: knowledgeBaseIDs,
		ChatbotID: chatbotID,
		Operation: "retrieval", Status: "completed", EmbeddingTokens: embeddingTokens,
		RerankUnits: rerankUnits, LatencyMS: int(time.Since(startedAt).Milliseconds()),
	})
	if usageErr != nil {
		return nil, profile, fmt.Errorf("record retrieval usage: %w", usageErr)
	}

	return response, profile, nil
}

func (s *Service) recordFailedRetrieval(
	ctx context.Context,
	tenantID string,
	knowledgeBaseIDs []string,
	chatbotID *string,
	response *platform.RetrievalResponse,
	cause error,
	startedAt time.Time,
) error {
	usageErr := s.store.RecordUsage(context.WithoutCancel(ctx), store.UsageEvent{
		TenantID: tenantID, RequestID: response.RequestID, KnowledgeBaseIDs: knowledgeBaseIDs,
		ChatbotID: chatbotID, Operation: "retrieval", Status: "failed",
		EmbeddingTokens: response.Usage.EmbeddingInputTokens, RerankUnits: response.Usage.RerankUnits,
		LatencyMS: int(time.Since(startedAt).Milliseconds()),
	})
	if usageErr != nil {
		return errors.Join(cause, fmt.Errorf("record failed retrieval usage: %w", usageErr))
	}

	return cause
}

func validateRerankResults(results []contractsopenai.RerankResult, candidateCount, topN int) error {
	if len(results) == 0 && candidateCount > 0 {
		return errors.New("provider returned no results")
	}

	if len(results) > topN {
		return fmt.Errorf("provider returned %d results, maximum is %d", len(results), topN)
	}

	seen := make(map[int]struct{}, len(results))
	for _, result := range results {
		if result.Index < 0 || result.Index >= candidateCount {
			return fmt.Errorf("provider returned out-of-range index %d", result.Index)
		}

		if _, exists := seen[result.Index]; exists {
			return fmt.Errorf("provider returned duplicate index %d", result.Index)
		}

		if math.IsNaN(result.RelevanceScore) || math.IsInf(result.RelevanceScore, 0) {
			return fmt.Errorf("provider returned non-finite score for index %d", result.Index)
		}

		seen[result.Index] = struct{}{}
	}

	return nil
}

func deduplicateCandidates(input []store.SearchCandidate) []store.SearchCandidate {
	seen := make(map[string]struct{}, len(input))

	output := make([]store.SearchCandidate, 0, len(input))
	for _, candidate := range input {
		key := strings.TrimSpace(candidate.ContentSHA256)
		if key == "" {
			key = candidate.DocumentID + ":" + candidate.Text
		}

		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}

		output = append(output, candidate)
	}

	return output
}

func uniqueStrings(input []string) []string {
	seen := make(map[string]struct{}, len(input))

	output := make([]string, 0, len(input))
	for _, item := range input {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		output = append(output, item)
	}

	return output
}
