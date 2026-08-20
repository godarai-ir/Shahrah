package app

import (
	"context"
	"errors"
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/rag/internal/domain"
	"github.com/Delpak/shahrah/services/rag/internal/store"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type retrievalStoreStub struct {
	candidates []store.SearchCandidate
	usage      *[]store.UsageEvent
}

func (s retrievalStoreStub) RequireRAGEnabled(context.Context, string) error    { return nil }
func (s retrievalStoreStub) RequireQueryCapacity(context.Context, string) error { return nil }

func (s retrievalStoreStub) ResolveKnowledgeBaseProfile(context.Context, string, []string) (domain.RAGProfile, error) {
	return testRAGProfile(), nil
}
func (s retrievalStoreStub) SearchChunks(context.Context, string, []string, string, []float32, int, int, int, float64) ([]store.SearchCandidate, error) {
	return s.candidates, nil
}
func (s retrievalStoreStub) ResolveSession(context.Context, store.PublishedChatbot, string) (string, error) {
	return uuid.NewString(), nil
}
func (s retrievalStoreStub) LoadConversation(context.Context, store.PublishedChatbot, string) ([]store.ConversationMessage, error) {
	return nil, nil
}
func (s retrievalStoreStub) SaveConversationTurn(context.Context, store.PublishedChatbot, string, string, string, bool, []platform.Citation) error {
	return nil
}
func (s retrievalStoreStub) DeleteSession(context.Context, store.PublishedChatbot, string) error {
	return nil
}
func (s retrievalStoreStub) RecordUsage(_ context.Context, event store.UsageEvent) error {
	if s.usage != nil {
		*s.usage = append(*s.usage, event)
	}

	return nil
}

type providerStub struct {
	rerankErr     error
	rerankResults []contractsopenai.RerankResult
}

func (providerStub) Embed(context.Context, string, string, []string, int) ([][]float32, int, error) {
	return [][]float32{make([]float32, testRAGProfile().EmbeddingDimensions)}, 3, nil
}

func testRAGProfile() domain.RAGProfile {
	return domain.RAGProfile{
		Revision: "test-profile", EmbeddingModel: "embedding", EmbeddingDimensions: 1536,
		RerankModel: "rerank", ChunkTokens: 700, ChunkOverlapTokens: 100,
		MaxChunks: 8, MaxContextTokens: 8000,
		Configuration: domain.RAGProfileConfiguration{
			RetrievalPerKnowledgeBase: 10, CandidateLimit: 50, MinimumRerankScore: 0.05,
			ConversationTurns: 6, ConversationTokens: 6000, Tokenizer: "cl100k_base",
			NormalizationRevision: "fa-ir-test", RetrievalMode: "dense_trigram_rrf",
			RRFRankConstant: 60, TrigramWordSimilarityThreshold: 0.2, OCRModel: "ocr",
		},
	}
}
func (p providerStub) Rerank(context.Context, string, string, string, []string, int) ([]contractsopenai.RerankResult, int, error) {
	if p.rerankErr != nil {
		return nil, 0, p.rerankErr
	}

	if p.rerankResults != nil {
		return p.rerankResults, 1, nil
	}

	return []contractsopenai.RerankResult{{Index: 0, RelevanceScore: 0.9}}, 1, nil
}

func TestRetrieveRejectsMalformedRerankIndices(t *testing.T) {
	t.Parallel()

	kbID := uuid.NewString()
	usage := make([]store.UsageEvent, 0, 1)
	service, err := New(retrievalStoreStub{candidates: []store.SearchCandidate{{
		KnowledgeBaseID: kbID, DocumentID: uuid.NewString(), VersionID: uuid.NewString(), Text: "evidence", TokenCount: 1,
	}}, usage: &usage}, providerStub{rerankResults: []contractsopenai.RerankResult{{Index: 1, RelevanceScore: 0.9}}})
	require.NoError(t, err)

	_, err = service.Retrieve(context.Background(), uuid.NewString(), "tenant", platform.RetrievalRequest{
		Query: "question", KnowledgeBaseIDs: []string{kbID},
	})
	require.ErrorContains(t, err, "out-of-range")
	require.Len(t, usage, 1)
	require.Equal(t, "failed", usage[0].Status)
	require.Equal(t, 3, usage[0].EmbeddingTokens)
	require.Equal(t, 1, usage[0].RerankUnits)
}
func (providerStub) Generate(context.Context, string, string, string, string) (string, string, GenerationUsage, error) {
	return "", "", GenerationUsage{}, nil
}

func TestRetrieveDeduplicatesAndReranks(t *testing.T) {
	t.Parallel()

	kbID := uuid.NewString()
	service, err := New(retrievalStoreStub{candidates: []store.SearchCandidate{
		{KnowledgeBaseID: kbID, DocumentID: uuid.NewString(), VersionID: uuid.NewString(), Text: "evidence", ContentSHA256: "same", TokenCount: 1},
		{KnowledgeBaseID: kbID, DocumentID: uuid.NewString(), VersionID: uuid.NewString(), Text: "duplicate", ContentSHA256: "same", TokenCount: 1},
	}}, providerStub{})
	require.NoError(t, err)

	response, err := service.Retrieve(context.Background(), uuid.NewString(), "tenant", platform.RetrievalRequest{
		Query: "question", KnowledgeBaseIDs: []string{kbID},
	})
	require.NoError(t, err)
	require.Len(t, response.Chunks, 1)
	require.Equal(t, 3, response.Usage.EmbeddingInputTokens)
	require.Equal(t, testRAGProfile().Revision, response.RAGProfileRevision)
}

func TestRetrieveFailsClosedWhenRerankFails(t *testing.T) {
	t.Parallel()

	kbID := uuid.NewString()
	service, err := New(retrievalStoreStub{candidates: []store.SearchCandidate{{
		KnowledgeBaseID: kbID, DocumentID: uuid.NewString(), VersionID: uuid.NewString(), Text: "evidence", TokenCount: 1,
	}}}, providerStub{rerankErr: errors.New("offline")})
	require.NoError(t, err)

	_, err = service.Retrieve(context.Background(), uuid.NewString(), "tenant", platform.RetrievalRequest{
		Query: "question", KnowledgeBaseIDs: []string{kbID},
	})
	require.ErrorContains(t, err, "rerank unavailable")
}
