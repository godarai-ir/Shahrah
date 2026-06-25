package platform

import (
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

type PromptManifest struct {
	Type          string                        `json:"type"`
	Name          string                        `json:"name"`
	Description   string                        `json:"description,omitempty"`
	Messages      []contractsopenai.ChatMessage `json:"messages"`
	Variables     map[string]string             `json:"variables,omitempty"`
	Model         string                        `json:"model,omitempty"`
	Parameters    map[string]interface{}        `json:"parameters,omitempty"`
	Metadata      map[string]interface{}        `json:"metadata,omitempty"`
	VersionAlias  string                        `json:"version_alias,omitempty"`
	ModelSettings *PromptModelSettings          `json:"model_settings,omitempty"`
	Logging       *PromptLoggingConfig          `json:"logging,omitempty"`
	Cache         *PromptCacheConfig            `json:"cache,omitempty"`
	Guardrails    *PromptGuardrailConfig        `json:"guardrails,omitempty"`
	Structured    *StructuredOutputConfig       `json:"structured_output,omitempty"`
}

type PromptModelSettings struct {
	Stream      *bool    `json:"stream,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	MaxTokens   *int     `json:"max_tokens,omitempty"`
}

type PromptLoggingConfig struct {
	Enabled   bool `json:"enabled"`
	StoreLogs bool `json:"store_logs"`
}

type PromptCacheConfig struct {
	Enabled             bool     `json:"enabled"`
	Type                string   `json:"type,omitempty"`
	TTL                 *int     `json:"ttl,omitempty"`
	Namespace           string   `json:"namespace,omitempty"`
	SimilarityThreshold *float64 `json:"similarity_threshold,omitempty"`
}

type PromptGuardrailConfig struct {
	Input  []string `json:"input,omitempty"`
	Output []string `json:"output,omitempty"`
	Scope  string   `json:"scope,omitempty"`
}

type StructuredOutputConfig struct {
	Enabled bool                   `json:"enabled"`
	Format  string                 `json:"format,omitempty"`
	Strict  bool                   `json:"strict,omitempty"`
	Schema  map[string]interface{} `json:"schema,omitempty"`
}

type Prompt struct {
	ID            string         `json:"id"`
	TenantID      string         `json:"tenant_id,omitempty"`
	Name          string         `json:"name"`
	Slug          string         `json:"slug"`
	Description   string         `json:"description,omitempty"`
	LatestVersion int            `json:"latest_version"`
	Latest        *PromptVersion `json:"latest_version_detail,omitempty"`
	CreatedBy     string         `json:"created_by,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type PromptVersion struct {
	ID            string                        `json:"id"`
	TenantID      string                        `json:"tenant_id,omitempty"`
	PromptID      string                        `json:"prompt_id"`
	Version       int                           `json:"version"`
	FQN           string                        `json:"fqn"`
	VersionAlias  string                        `json:"version_alias,omitempty"`
	Name          string                        `json:"name"`
	Description   string                        `json:"description,omitempty"`
	Messages      []contractsopenai.ChatMessage `json:"messages"`
	Variables     map[string]string             `json:"variables,omitempty"`
	Model         string                        `json:"model,omitempty"`
	Parameters    map[string]interface{}        `json:"parameters,omitempty"`
	Metadata      map[string]interface{}        `json:"metadata,omitempty"`
	Manifest      PromptManifest                `json:"manifest,omitempty"`
	CommitMessage string                        `json:"commit_message,omitempty"`
	CreatedBy     string                        `json:"created_by,omitempty"`
	CreatedAt     time.Time                     `json:"created_at"`
}

type CreatePromptRequest struct {
	Manifest      PromptManifest `json:"manifest"`
	CommitMessage string         `json:"commit_message,omitempty"`
}

type CreatePromptVersionRequest struct {
	Manifest      PromptManifest `json:"manifest"`
	CommitMessage string         `json:"commit_message,omitempty"`
}

type RenderPromptVersionRequest struct {
	Variables map[string]string             `json:"variables,omitempty"`
	Messages  []contractsopenai.ChatMessage `json:"messages,omitempty"`
	Model     string                        `json:"model,omitempty"`
}

type RenderPromptVersionResponse struct {
	PromptID        string                        `json:"prompt_id"`
	PromptVersionID string                        `json:"prompt_version_id"`
	PromptFQN       string                        `json:"prompt_fqn"`
	Model           string                        `json:"model,omitempty"`
	Messages        []contractsopenai.ChatMessage `json:"messages"`
	Variables       map[string]string             `json:"variables,omitempty"`
}

type PromptListResponse struct {
	Data       []Prompt `json:"data"`
	Total      int      `json:"total"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	NextOffset *int     `json:"next_offset,omitempty"`
}

type PromptVersionListResponse struct {
	Data       []PromptVersion `json:"data"`
	Total      int             `json:"total"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
	NextOffset *int            `json:"next_offset,omitempty"`
}
