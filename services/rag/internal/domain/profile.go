package domain

import (
	"errors"
	"fmt"
	"strings"
)

// RAGProfile is an immutable, versioned runtime configuration loaded from metadata storage.
type RAGProfile struct {
	Revision            string
	EmbeddingModel      string
	EmbeddingDimensions int
	RerankModel         string
	ChunkTokens         int
	ChunkOverlapTokens  int
	MaxChunks           int
	MaxContextTokens    int
	Configuration       RAGProfileConfiguration
}

// RAGProfileConfiguration contains provider-independent pipeline settings stored as JSON.
type RAGProfileConfiguration struct {
	RetrievalPerKnowledgeBase      int     `json:"retrieval_per_knowledge_base"`
	CandidateLimit                 int     `json:"candidate_limit"`
	MinimumRerankScore             float64 `json:"minimum_rerank_score"`
	ConversationTurns              int     `json:"conversation_turns"`
	ConversationTokens             int     `json:"conversation_tokens"`
	Tokenizer                      string  `json:"tokenizer"`
	NormalizationRevision          string  `json:"normalization_revision"`
	RetrievalMode                  string  `json:"retrieval_mode"`
	RRFRankConstant                int     `json:"rrf_rank_constant"`
	TrigramWordSimilarityThreshold float64 `json:"trigram_word_similarity_threshold"`
	OCRModel                       string  `json:"ocr_model"`
}

// Validate rejects incomplete profiles so the pipeline fails closed instead of using hidden defaults.
func (p RAGProfile) Validate() error {
	if strings.TrimSpace(p.Revision) == "" || strings.TrimSpace(p.EmbeddingModel) == "" ||
		strings.TrimSpace(p.RerankModel) == "" {
		return errors.New("RAG profile revision and provider models are required")
	}

	if p.EmbeddingDimensions <= 0 || p.ChunkTokens <= 0 || p.ChunkOverlapTokens < 0 ||
		p.ChunkOverlapTokens >= p.ChunkTokens || p.MaxChunks <= 0 || p.MaxContextTokens <= 0 {
		return errors.New("RAG profile numeric limits are invalid")
	}

	configuration := p.Configuration
	if configuration.RetrievalPerKnowledgeBase <= 0 || configuration.CandidateLimit <= 0 ||
		configuration.CandidateLimit < configuration.RetrievalPerKnowledgeBase ||
		configuration.ConversationTurns <= 0 || configuration.ConversationTokens <= 0 ||
		configuration.RRFRankConstant <= 0 {
		return errors.New("RAG profile pipeline limits are invalid")
	}

	if configuration.MinimumRerankScore < 0 || configuration.MinimumRerankScore > 1 ||
		configuration.TrigramWordSimilarityThreshold <= 0 || configuration.TrigramWordSimilarityThreshold > 1 {
		return errors.New("RAG profile score thresholds are invalid")
	}

	if strings.TrimSpace(configuration.Tokenizer) == "" ||
		strings.TrimSpace(configuration.NormalizationRevision) == "" ||
		strings.TrimSpace(configuration.RetrievalMode) == "" || strings.TrimSpace(configuration.OCRModel) == "" {
		return errors.New("RAG profile pipeline identifiers are required")
	}

	if configuration.Tokenizer != "cl100k_base" {
		return fmt.Errorf("unsupported RAG tokenizer %q", configuration.Tokenizer)
	}

	return nil
}

// Chunk is a deterministic text unit produced during ingestion.
type Chunk struct {
	Index      int
	Text       string
	IndexText  string
	TokenCount int
	Page       *int
	Heading    string
	SHA256     string
}
