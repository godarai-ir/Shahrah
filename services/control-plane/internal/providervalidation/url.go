package providervalidation

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// ErrInvalidBaseURL identifies an unsafe or provider-incompatible upstream URL.
var ErrInvalidBaseURL = errors.New("invalid provider base URL")

var geminiAPIVersionPattern = regexp.MustCompile(`^v[0-9]+(?:(?:alpha|beta)[0-9]*)?$`)

// ValidateBaseURL normalizes and validates provider-owned HTTPS base URLs.
func ValidateBaseURL(provider, raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%w: must be an absolute HTTPS URL", ErrInvalidBaseURL)
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return "", fmt.Errorf("%w: must use HTTPS", ErrInvalidBaseURL)
	}
	if parsed.User != nil {
		return "", fmt.Errorf("%w: userinfo is not allowed", ErrInvalidBaseURL)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("%w: query and fragment are not allowed", ErrInvalidBaseURL)
	}
	if parsed.Port() != "" && parsed.Port() != "443" {
		return "", fmt.Errorf("%w: only the default HTTPS port is allowed", ErrInvalidBaseURL)
	}

	hostname := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if hostname == "" || net.ParseIP(hostname) != nil {
		return "", fmt.Errorf("%w: host must be an approved provider domain", ErrInvalidBaseURL)
	}
	if hostname == "localhost" || strings.HasSuffix(hostname, ".local") || strings.HasSuffix(hostname, ".internal") {
		return "", fmt.Errorf("%w: local hosts are not allowed", ErrInvalidBaseURL)
	}
	if !providerHostAllowed(NormalizeProvider(provider), hostname) {
		return "", fmt.Errorf("%w: host is not allowed for provider %q", ErrInvalidBaseURL, strings.TrimSpace(provider))
	}

	parsed.Scheme = "https"
	parsed.Host = strings.ToLower(parsed.Host)

	return strings.TrimRight(parsed.String(), "/"), nil
}

func providerHostAllowed(provider, hostname string) bool {
	exact := map[string][]string{
		"ai21":             {"api.ai21.com"},
		"anthropic":        {"api.anthropic.com"},
		"cerebras":         {"api.cerebras.ai"},
		"cohere":           {"api.cohere.com", "api.cohere.ai"},
		"deepgram":         {"api.deepgram.com"},
		"deepinfra":        {"api.deepinfra.com"},
		"elevenlabs":       {"api.elevenlabs.io"},
		"google-gemini":    {"generativelanguage.googleapis.com"},
		"groq":             {"api.groq.com"},
		"mistral-ai":       {"api.mistral.ai"},
		"openai":           {"api.openai.com"},
		"openrouter":       {"openrouter.ai"},
		"perplexity-ai":    {"api.perplexity.ai"},
		"sambanova":        {"api.sambanova.ai"},
		"together-ai":      {"api.together.xyz", "api.together.ai"},
		"xai":              {"api.x.ai"},
		"azure-ai-foundry": nil,
		"azure-open-ai":    nil,
		"databricks":       nil,
		"google-vertex":    nil,
	}

	for _, allowed := range exact[provider] {
		if hostname == allowed {
			return true
		}
	}

	switch provider {
	case "azure-ai-foundry", "azure-open-ai":
		return strings.HasSuffix(hostname, ".openai.azure.com") ||
			strings.HasSuffix(hostname, ".services.ai.azure.com") ||
			strings.HasSuffix(hostname, ".models.ai.azure.com")
	case "databricks":
		return strings.HasSuffix(hostname, ".cloud.databricks.com") ||
			strings.HasSuffix(hostname, ".azuredatabricks.net") ||
			strings.HasSuffix(hostname, ".gcp.databricks.com")
	case "google-vertex":
		return strings.HasSuffix(hostname, "-aiplatform.googleapis.com")
	default:
		return false
	}
}

func appendURLPath(base, suffix string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(suffix, "/")
}

func ensureVersionedBase(base, version string) string {
	parsed, err := url.Parse(base)
	if err != nil {
		return appendURLPath(base, version)
	}

	path := strings.ToLower(strings.TrimRight(parsed.Path, "/"))
	if strings.HasSuffix(path, "/"+strings.ToLower(version)) {
		return strings.TrimRight(base, "/")
	}

	return appendURLPath(base, version)
}

func ensureGeminiVersionedBase(base, version string) (string, error) {
	version = strings.ToLower(strings.TrimSpace(version))
	if !geminiAPIVersionPattern.MatchString(version) {
		return "", fmt.Errorf("%w: invalid Gemini API version", ErrInvalidConfiguration)
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%w: invalid Gemini base URL", ErrInvalidConfiguration)
	}

	lastSegment := strings.ToLower(strings.TrimSpace(pathLastSegment(parsed.Path)))
	if geminiAPIVersionPattern.MatchString(lastSegment) {
		return strings.TrimRight(base, "/"), nil
	}

	return appendURLPath(base, version), nil
}

func azureOpenAIModelsURL(base, apiVersion string) (string, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("%w: invalid Azure OpenAI endpoint", ErrInvalidConfiguration)
	}

	lowerPath := strings.ToLower(strings.TrimRight(parsed.Path, "/"))
	if index := strings.Index(lowerPath, "/openai/deployments/"); index >= 0 {
		parsed.Path = parsed.Path[:index]
		parsed.RawPath = ""
		lowerPath = strings.ToLower(strings.TrimRight(parsed.Path, "/"))
	}

	if strings.HasSuffix(lowerPath, "/openai/v1") {
		parsed.Path = strings.TrimRight(parsed.Path, "/") + "/models"

		return parsed.String(), nil
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/openai/models"
	if strings.TrimSpace(apiVersion) == "" {
		apiVersion = "2024-10-21"
	}

	query := parsed.Query()
	query.Set("api-version", strings.TrimSpace(apiVersion))
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func ensureOpenRouterAPIBase(base string) string {
	parsed, err := url.Parse(base)
	if err != nil {
		return base
	}

	path := strings.ToLower(strings.TrimRight(parsed.Path, "/"))
	if path == "" {
		return appendURLPath(base, "api/v1")
	}
	if strings.HasSuffix(path, "/api/v1") {
		return strings.TrimRight(base, "/")
	}

	return base
}

func pathLastSegment(value string) string {
	value = strings.TrimRight(value, "/")
	if value == "" {
		return ""
	}

	if index := strings.LastIndex(value, "/"); index >= 0 {
		return value[index+1:]
	}

	return value
}
