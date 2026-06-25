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

const TypeChatPrompt = "chat_prompt"

var (
	ErrInvalidManifest = errors.New("invalid prompt manifest")
	ErrMissingVariable = errors.New("missing prompt variable")
	namePattern        = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)
	variablePattern    = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)
)

type FQN struct {
	Type       string
	TenantSlug string
	PromptSlug string
	Version    int
}

func NormalizeSlug(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "-")
	return value
}

func BuildFQN(tenantSlug, promptSlug string, version int) string {
	return fmt.Sprintf("%s:%s/%s:%d", TypeChatPrompt, strings.TrimSpace(tenantSlug), strings.TrimSpace(promptSlug), version)
}

func ParseFQN(raw string) (FQN, error) {
	raw = strings.TrimSpace(raw)
	typePart, rest, ok := strings.Cut(raw, ":")
	if !ok || typePart != TypeChatPrompt {
		return FQN{}, fmt.Errorf("prompt_version_fqn must start with %s:", TypeChatPrompt)
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

func ValidateManifest(manifest platform.PromptManifest) error {
	if manifest.Type == "" {
		manifest.Type = TypeChatPrompt
	}
	if manifest.Type != TypeChatPrompt {
		return fmt.Errorf("%w: type must be %s", ErrInvalidManifest, TypeChatPrompt)
	}
	if !namePattern.MatchString(strings.TrimSpace(manifest.Name)) {
		return fmt.Errorf("%w: name must contain only letters, numbers, hyphens, and underscores", ErrInvalidManifest)
	}
	if len(manifest.Messages) == 0 {
		return fmt.Errorf("%w: messages must not be empty", ErrInvalidManifest)
	}
	for i, msg := range manifest.Messages {
		if strings.TrimSpace(msg.Role) == "" {
			return fmt.Errorf("%w: message[%d].role is required", ErrInvalidManifest, i)
		}
		if err := contractsopenai.ValidateChatMessageContent(msg.Content); err != nil {
			return fmt.Errorf("%w: message[%d].content: %v", ErrInvalidManifest, i, err)
		}
	}
	if manifest.ModelSettings != nil {
		if manifest.ModelSettings.Temperature != nil && (*manifest.ModelSettings.Temperature < 0 || *manifest.ModelSettings.Temperature > 2) {
			return fmt.Errorf("%w: model_settings.temperature must be between 0 and 2", ErrInvalidManifest)
		}
		if manifest.ModelSettings.TopP != nil && (*manifest.ModelSettings.TopP < 0 || *manifest.ModelSettings.TopP > 1) {
			return fmt.Errorf("%w: model_settings.top_p must be between 0 and 1", ErrInvalidManifest)
		}
		if manifest.ModelSettings.MaxTokens != nil && *manifest.ModelSettings.MaxTokens < 1 {
			return fmt.Errorf("%w: model_settings.max_tokens must be at least 1", ErrInvalidManifest)
		}
	}
	if manifest.Cache != nil && manifest.Cache.Enabled {
		cacheType := strings.TrimSpace(strings.ToLower(manifest.Cache.Type))
		if cacheType != "simple" && cacheType != "smart" {
			return fmt.Errorf("%w: cache.type must be simple or smart", ErrInvalidManifest)
		}
		if manifest.Cache.TTL != nil && *manifest.Cache.TTL < 1 {
			return fmt.Errorf("%w: cache.ttl must be at least 1", ErrInvalidManifest)
		}
		if cacheType == "simple" && manifest.Cache.SimilarityThreshold != nil {
			return fmt.Errorf("%w: cache.similarity_threshold is only valid for smart cache", ErrInvalidManifest)
		}
		if cacheType == "smart" && manifest.Cache.SimilarityThreshold != nil {
			if *manifest.Cache.SimilarityThreshold < 0.8 || *manifest.Cache.SimilarityThreshold > 1 {
				return fmt.Errorf("%w: cache.similarity_threshold must be between 0.8 and 1", ErrInvalidManifest)
			}
		}
	}
	if manifest.Guardrails != nil {
		scope := strings.TrimSpace(strings.ToLower(manifest.Guardrails.Scope))
		if scope != "" && scope != "all" && scope != "last" {
			return fmt.Errorf("%w: guardrails.scope must be all or last", ErrInvalidManifest)
		}
		if manifest.ModelSettings != nil && manifest.ModelSettings.Stream != nil && *manifest.ModelSettings.Stream && len(manifest.Guardrails.Output) > 0 {
			return fmt.Errorf("%w: output guardrails require streaming to be disabled", ErrInvalidManifest)
		}
	}
	if manifest.Structured != nil && manifest.Structured.Enabled {
		format := strings.TrimSpace(strings.ToLower(manifest.Structured.Format))
		if format != "" && format != "json_schema" && format != "json_object" {
			return fmt.Errorf("%w: structured_output.format must be json_schema or json_object", ErrInvalidManifest)
		}
		if format == "json_schema" && len(manifest.Structured.Schema) == 0 {
			return fmt.Errorf("%w: structured_output.schema is required for json_schema format", ErrInvalidManifest)
		}
	}
	return nil
}

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
