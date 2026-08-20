package platform

import "time"

// RAGMaxQueryCharacters is the maximum Unicode code-point length accepted by retrieval and chatbot inputs.
const RAGMaxQueryCharacters = 8000

// RAGDocumentStatus is the lifecycle state of an ingested document version.
type RAGDocumentStatus string

const (
	RAGDocumentQueued    RAGDocumentStatus = "queued"
	RAGDocumentScanning  RAGDocumentStatus = "scanning"
	RAGDocumentParsing   RAGDocumentStatus = "parsing"
	RAGDocumentChunking  RAGDocumentStatus = "chunking"
	RAGDocumentEmbedding RAGDocumentStatus = "embedding"
	RAGDocumentIndexing  RAGDocumentStatus = "indexing"
	RAGDocumentReady     RAGDocumentStatus = "ready"
	RAGDocumentFailed    RAGDocumentStatus = "failed"
	RAGDocumentDeleting  RAGDocumentStatus = "deleting"
)

// KnowledgeBase is a tenant-scoped managed collection of documents.
type KnowledgeBase struct {
	ID                 string     `json:"id"`
	TenantID           string     `json:"tenant_id,omitempty"`
	Name               string     `json:"name"`
	Description        string     `json:"description,omitempty"`
	RAGProfileRevision string     `json:"rag_profile_revision"`
	DocumentCount      int        `json:"document_count"`
	ReadyDocumentCount int        `json:"ready_document_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
}

// CreateKnowledgeBaseRequest creates a managed knowledge base.
type CreateKnowledgeBaseRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// UpdateKnowledgeBaseRequest mutates editable knowledge base metadata.
type UpdateKnowledgeBaseRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// KnowledgeBaseList is a paginated knowledge base result.
type KnowledgeBaseList struct {
	Items  []KnowledgeBase `json:"items"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// KnowledgeBaseDocument describes one logical document and its active version.
type KnowledgeBaseDocument struct {
	ID                  string            `json:"id"`
	KnowledgeBaseID     string            `json:"knowledge_base_id"`
	Title               string            `json:"title"`
	SourceType          string            `json:"source_type"`
	MediaType           string            `json:"media_type"`
	Status              RAGDocumentStatus `json:"status"`
	ActiveVersionID     *string           `json:"active_version_id,omitempty"`
	ProcessingVersionID *string           `json:"processing_version_id,omitempty"`
	SizeBytes           int64             `json:"size_bytes"`
	PageCount           *int              `json:"page_count,omitempty"`
	ChunkCount          int               `json:"chunk_count"`
	DetectedLanguage    string            `json:"detected_language,omitempty"`
	FailureCode         string            `json:"failure_code,omitempty"`
	FailureMessage      string            `json:"failure_message,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

// CreateTextDocumentRequest ingests tenant-provided text without a file upload.
type CreateTextDocumentRequest struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// DocumentList is a paginated knowledge base document result.
type DocumentList struct {
	Items  []KnowledgeBaseDocument `json:"items"`
	Total  int                     `json:"total"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
}

// RAGUsage contains provider metering for a retrieval request.
type RAGUsage struct {
	EmbeddingInputTokens int `json:"embedding_input_tokens"`
	RerankUnits          int `json:"rerank_units"`
}

// RetrievalRequest asks the fixed RAG profile to retrieve grounded context.
type RetrievalRequest struct {
	Query            string   `json:"query"`
	KnowledgeBaseIDs []string `json:"knowledge_base_ids"`
}

// RetrievalChunk is a reranked, citation-addressable context unit.
type RetrievalChunk struct {
	CitationID      string  `json:"citation_id"`
	KnowledgeBaseID string  `json:"knowledge_base_id"`
	DocumentID      string  `json:"document_id"`
	DocumentVersion string  `json:"document_version_id"`
	Title           string  `json:"title"`
	Page            *int    `json:"page,omitempty"`
	Heading         string  `json:"heading,omitempty"`
	Text            string  `json:"text"`
	RetrievalScore  float64 `json:"retrieval_score"`
	RerankScore     float64 `json:"rerank_score"`
}

// RetrievalResponse is the stable managed retrieval contract.
type RetrievalResponse struct {
	RequestID          string           `json:"request_id"`
	RAGProfileRevision string           `json:"rag_profile_revision"`
	Chunks             []RetrievalChunk `json:"chunks"`
	Usage              RAGUsage         `json:"usage"`
}

// Chatbot is the mutable management representation of a chatbot draft.
type Chatbot struct {
	ID                 string         `json:"id"`
	TenantID           string         `json:"tenant_id,omitempty"`
	PublicID           string         `json:"public_id"`
	Name               string         `json:"name"`
	Description        string         `json:"description,omitempty"`
	SystemInstruction  string         `json:"system_instruction"`
	AnswerModel        string         `json:"answer_model"`
	Tone               string         `json:"tone"`
	DefaultLocale      string         `json:"default_locale"`
	AnswerLanguage     string         `json:"answer_language"`
	AccessMode         string         `json:"access_mode"`
	RetentionDays      int            `json:"retention_days"`
	KnowledgeBaseIDs   []string       `json:"knowledge_base_ids"`
	AllowedOrigins     []string       `json:"allowed_origins"`
	Branding           map[string]any `json:"branding"`
	PublishedVersionID *string        `json:"published_version_id,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          *time.Time     `json:"deleted_at,omitempty"`
}

// UpsertChatbotRequest creates or updates chatbot draft configuration.
type UpsertChatbotRequest struct {
	Name              string         `json:"name"`
	Description       string         `json:"description,omitempty"`
	SystemInstruction string         `json:"system_instruction,omitempty"`
	AnswerModel       string         `json:"answer_model"`
	Tone              string         `json:"tone,omitempty"`
	DefaultLocale     string         `json:"default_locale,omitempty"`
	AnswerLanguage    string         `json:"answer_language,omitempty"`
	AccessMode        string         `json:"access_mode,omitempty"`
	RetentionDays     int            `json:"retention_days"`
	KnowledgeBaseIDs  []string       `json:"knowledge_base_ids"`
	AllowedOrigins    []string       `json:"allowed_origins,omitempty"`
	Branding          map[string]any `json:"branding,omitempty"`
}

// ChatbotList is a paginated chatbot result.
type ChatbotList struct {
	Items  []Chatbot `json:"items"`
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// ChatbotPublication is an immutable published chatbot version.
type ChatbotPublication struct {
	VersionID          string    `json:"version_id"`
	ChatbotID          string    `json:"chatbot_id"`
	Revision           int       `json:"revision"`
	RAGProfileRevision string    `json:"rag_profile_revision"`
	PublishedAt        time.Time `json:"published_at"`
}

// ChatbotSessionTokenRequest creates a short-lived private bot token.
type ChatbotSessionTokenRequest struct {
	TTLSeconds int `json:"ttl_seconds,omitempty"`
}

// ChatbotSessionToken is returned only when a private session token is issued.
type ChatbotSessionToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ChatbotResponseRequest is the public chatbot answer input.
type ChatbotResponseRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
}

// Citation is a validated source reference included in an answer.
type Citation struct {
	ID         string `json:"id"`
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	Page       *int   `json:"page,omitempty"`
	Quote      string `json:"quote"`
}

// ChatbotResponse is emitted atomically after citation validation.
type ChatbotResponse struct {
	RequestID   string     `json:"request_id"`
	TraceID     string     `json:"trace_id"`
	SessionID   string     `json:"session_id"`
	Answer      string     `json:"answer"`
	Citations   []Citation `json:"citations"`
	Refusal     bool       `json:"refusal"`
	ActualModel string     `json:"actual_model"`
}

// ChatbotAnalytics contains privacy-safe aggregate product telemetry.
type ChatbotAnalytics struct {
	QueryCount           int64   `json:"query_count"`
	RefusalCount         int64   `json:"refusal_count"`
	CitationFailureCount int64   `json:"citation_failure_count"`
	AverageLatencyMS     float64 `json:"average_latency_ms"`
	ProviderCostUSD      float64 `json:"provider_cost_usd"`
	PlatformCostIRR      float64 `json:"platform_cost_irr"`
}
