// Package prompts validates and renders prompt manifests and FQNs.
package prompts

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// TypeChatPrompt is the only prompt manifest type supported by this package.
const TypeChatPrompt = "chat_prompt"

var (
	// ErrInvalidManifest indicates the prompt manifest failed validation.
	ErrInvalidManifest = errors.New("invalid prompt manifest")
	// ErrMissingVariable indicates rendering could not resolve one or more variables.
	ErrMissingVariable = errors.New("missing prompt variable")
	namePattern        = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)
	variablePattern    = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)
)

// FQN identifies a prompt version by type, tenant, slug, and version.
type FQN struct {
	Type       string
	TenantSlug string
	PromptSlug string
	Version    int
}

// NormalizeSlug converts a display slug into a normalized prompt slug.
func NormalizeSlug(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "-")

	return value
}

// BuildFQN formats a prompt version FQN string.
func BuildFQN(tenantSlug, promptSlug string, version int) string {
	return fmt.Sprintf("%s:%s/%s:%d", TypeChatPrompt, strings.TrimSpace(tenantSlug), strings.TrimSpace(promptSlug), version)
}

// ParseFQN parses a prompt version FQN string.
func ParseFQN(raw string) (FQN, error) {
	raw = strings.TrimSpace(raw)

	typePart, rest, ok := strings.Cut(raw, ":")
	if !ok || typePart != TypeChatPrompt {
		return FQN{}, fmt.Errorf("prompt_version_fqn must start with %s", TypeChatPrompt)
	}

	pathPart, versionPart, ok := strings.Cut(rest, ":")
	if !ok {
		return FQN{}, errors.New("prompt_version_fqn must include a version")
	}

	tenantSlug, promptSlug, ok := strings.Cut(pathPart, "/")
	if !ok || strings.TrimSpace(tenantSlug) == "" || strings.TrimSpace(promptSlug) == "" {
		return FQN{}, errors.New("prompt_version_fqn must be chat_prompt:{tenant_slug}/{prompt_name}:{version}")
	}

	version, err := strconv.Atoi(strings.TrimSpace(versionPart))
	if err != nil || version < 1 {
		return FQN{}, errors.New("prompt_version_fqn version must be a positive integer")
	}

	return FQN{Type: typePart, TenantSlug: tenantSlug, PromptSlug: promptSlug, Version: version}, nil
}

// ValidateManifest checks a prompt manifest for structural and semantic errors.
func ValidateManifest(manifest platform.PromptManifest) error {
	if manifest.Type == "" {
		manifest.Type = TypeChatPrompt
	}

	if err := validateManifestType(manifest.Type); err != nil {
		return err
	}

	if err := validateManifestName(manifest.Name); err != nil {
		return err
	}

	if err := validateManifestMessages(manifest.Messages); err != nil {
		return err
	}

	if err := validateManifestModelSettings(manifest.ModelSettings); err != nil {
		return err
	}

	if err := validateManifestCache(manifest.Cache); err != nil {
		return err
	}

	if err := validateManifestGuardrails(manifest.Guardrails, manifest.ModelSettings); err != nil {
		return err
	}

	return validateManifestStructured(manifest.Structured)
}

func validateManifestType(value string) error {
	if value != TypeChatPrompt {
		return fmt.Errorf("%w: type must be %s", ErrInvalidManifest, TypeChatPrompt)
	}

	return nil
}

func validateManifestName(value string) error {
	if !namePattern.MatchString(strings.TrimSpace(value)) {
		return fmt.Errorf("%w: name must contain only letters, numbers, hyphens, and underscores", ErrInvalidManifest)
	}

	return nil
}

func validateManifestMessages(messages []contractsopenai.ChatMessage) error {
	if len(messages) == 0 {
		return fmt.Errorf("%w: messages must not be empty", ErrInvalidManifest)
	}

	for i, msg := range messages {
		if strings.TrimSpace(msg.Role) == "" {
			return fmt.Errorf("%w: message[%d].role is required", ErrInvalidManifest, i)
		}

		if err := contractsopenai.ValidateChatMessageContent(msg.Content); err != nil {
			return fmt.Errorf("%w: message[%d].content: %v", ErrInvalidManifest, i, err)
		}
	}

	return nil
}

func validateManifestModelSettings(settings *platform.PromptModelSettings) error {
	if settings == nil {
		return nil
	}

	if settings.Temperature != nil && (*settings.Temperature < 0 || *settings.Temperature > 2) {
		return fmt.Errorf("%w: model_settings.temperature must be between 0 and 2", ErrInvalidManifest)
	}

	if settings.TopP != nil && (*settings.TopP < 0 || *settings.TopP > 1) {
		return fmt.Errorf("%w: model_settings.top_p must be between 0 and 1", ErrInvalidManifest)
	}

	if settings.MaxTokens != nil && *settings.MaxTokens < 1 {
		return fmt.Errorf("%w: model_settings.max_tokens must be at least 1", ErrInvalidManifest)
	}

	return nil
}

func validateManifestCache(cache *platform.PromptCacheConfig) error {
	if cache == nil || !cache.Enabled {
		return nil
	}

	cacheType := strings.TrimSpace(strings.ToLower(cache.Type))
	if cacheType != "simple" && cacheType != "smart" {
		return fmt.Errorf("%w: cache.type must be simple or smart", ErrInvalidManifest)
	}

	if cache.TTL != nil && *cache.TTL < 1 {
		return fmt.Errorf("%w: cache.ttl must be at least 1", ErrInvalidManifest)
	}

	if cacheType == "simple" && cache.SimilarityThreshold != nil {
		return fmt.Errorf("%w: cache.similarity_threshold is only valid for smart cache", ErrInvalidManifest)
	}

	if cacheType == "smart" && cache.SimilarityThreshold != nil {
		if *cache.SimilarityThreshold < 0.8 || *cache.SimilarityThreshold > 1 {
			return fmt.Errorf("%w: cache.similarity_threshold must be between 0.8 and 1", ErrInvalidManifest)
		}
	}

	return nil
}

func validateManifestGuardrails(guardrails *platform.PromptGuardrailConfig, settings *platform.PromptModelSettings) error {
	if guardrails == nil {
		return nil
	}

	scope := strings.TrimSpace(strings.ToLower(guardrails.Scope))
	if scope != "" && scope != "all" && scope != "last" {
		return fmt.Errorf("%w: guardrails.scope must be all or last", ErrInvalidManifest)
	}

	if settings != nil && settings.Stream != nil && *settings.Stream && len(guardrails.Output) > 0 {
		return fmt.Errorf("%w: output guardrails require streaming to be disabled", ErrInvalidManifest)
	}

	return nil
}

func validateManifestStructured(structured *platform.StructuredOutputConfig) error {
	if structured == nil || !structured.Enabled {
		return nil
	}

	format := strings.TrimSpace(strings.ToLower(structured.Format))
	if format != "" && format != "json_schema" && format != "json_object" {
		return fmt.Errorf("%w: structured_output.format must be json_schema or json_object", ErrInvalidManifest)
	}

	if format == "json_schema" && len(structured.Schema) == 0 {
		return fmt.Errorf("%w: structured_output.schema is required for json_schema format", ErrInvalidManifest)
	}

	return nil
}

// Render expands a prompt version with variables and optional extra messages.
func Render(version platform.PromptVersion, variables map[string]string, appendMessages []contractsopenai.ChatMessage, requestModel string) (platform.RenderPromptVersionResponse, error) {
	if len(version.Messages) == 0 && len(version.Manifest.Messages) > 0 {
		version.Messages = version.Manifest.Messages
	}

	if strings.TrimSpace(version.Model) == "" && strings.TrimSpace(version.Manifest.Model) != "" {
		version.Model = version.Manifest.Model
	}

	if variables == nil {
		variables = map[string]string{}
	}

	rendered := make([]contractsopenai.ChatMessage, 0, len(version.Messages)+len(appendMessages))
	missing := map[string]struct{}{}

	for _, msg := range version.Messages {
		next := msg

		content, err := renderContent(msg.Content, variables, missing)
		if err != nil {
			return platform.RenderPromptVersionResponse{}, err
		}

		next.Content = content
		rendered = append(rendered, next)
	}

	if len(missing) > 0 {
		keys := make([]string, 0, len(missing))
		for key := range missing {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		return platform.RenderPromptVersionResponse{}, fmt.Errorf("%w: %s", ErrMissingVariable, strings.Join(keys, ", "))
	}

	rendered = append(rendered, appendMessages...)

	model := strings.TrimSpace(requestModel)
	if model == "" {
		model = strings.TrimSpace(version.Model)
	}

	return platform.RenderPromptVersionResponse{
		PromptID: version.PromptID, PromptVersionID: version.ID, PromptFQN: version.FQN,
		Model: model, Messages: rendered, Variables: variables,
	}, nil
}

func renderContent(content interface{}, variables map[string]string, missing map[string]struct{}) (interface{}, error) {
	switch v := content.(type) {
	case nil:
		return nil, nil
	case string:
		return renderString(v, variables, missing), nil
	case []contractsopenai.ChatContentPart:
		out := make([]contractsopenai.ChatContentPart, len(v))
		copy(out, v)

		for i := range out {
			if out[i].Type == "text" {
				out[i].Text = renderString(out[i].Text, variables, missing)
			}
		}

		return out, nil
	default:
		parts, ok, err := contractsopenai.ChatContentParts(v)
		if err != nil || !ok {
			if err != nil {
				return nil, err
			}

			return content, nil
		}

		for i := range parts {
			if parts[i].Type == "text" {
				parts[i].Text = renderString(parts[i].Text, variables, missing)
			}
		}

		return parts, nil
	}
}

func renderString(input string, variables map[string]string, missing map[string]struct{}) string {
	return variablePattern.ReplaceAllStringFunc(input, func(match string) string {
		groups := variablePattern.FindStringSubmatch(match)
		if len(groups) < 2 {
			return match
		}

		key := groups[1]

		value, ok := variables[key]
		if !ok {
			missing[key] = struct{}{}
			return match
		}

		return value
	})
}
