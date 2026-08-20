package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRAGProfileValidateAcceptsCompleteConfiguration(t *testing.T) {
	t.Parallel()

	require.NoError(t, validTestRAGProfile().Validate())
}

func TestRAGProfileValidateRejectsMissingConfigurationWithoutDefaults(t *testing.T) {
	t.Parallel()

	profile := validTestRAGProfile()
	profile.Configuration.NormalizationRevision = ""

	require.ErrorContains(t, profile.Validate(), "identifiers")
}

func TestRAGProfileValidateRejectsUnsupportedTokenizer(t *testing.T) {
	t.Parallel()

	profile := validTestRAGProfile()
	profile.Configuration.Tokenizer = "unknown"

	require.ErrorContains(t, profile.Validate(), "unsupported RAG tokenizer")
}

func validTestRAGProfile() RAGProfile {
	return RAGProfile{
		Revision: "test", EmbeddingModel: "embedding", EmbeddingDimensions: 1536,
		RerankModel: "rerank", ChunkTokens: 700, ChunkOverlapTokens: 100,
		MaxChunks: 8, MaxContextTokens: 8000,
		Configuration: RAGProfileConfiguration{
			RetrievalPerKnowledgeBase: 10, CandidateLimit: 50, MinimumRerankScore: 0.05,
			ConversationTurns: 6, ConversationTokens: 6000, Tokenizer: "cl100k_base",
			NormalizationRevision: "fa-ir-test", RetrievalMode: "dense_trigram_rrf",
			RRFRankConstant: 60, TrigramWordSimilarityThreshold: 0.2, OCRModel: "ocr",
		},
	}
}
