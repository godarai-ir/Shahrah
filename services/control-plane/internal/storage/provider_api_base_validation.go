package storage

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var ErrInvalidProviderAPIBase = errors.New("invalid provider api_base")

func ValidateProviderAPIBase(provider string, raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("%w: must be an absolute HTTPS URL", ErrInvalidProviderAPIBase)
	}

	if !strings.EqualFold(parsed.Scheme, "https") {
		return "", fmt.Errorf("%w: must use HTTPS", ErrInvalidProviderAPIBase)
	}

	if parsed.User != nil {
		return "", fmt.Errorf("%w: userinfo is not allowed", ErrInvalidProviderAPIBase)
	}

	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("%w: query and fragment are not allowed", ErrInvalidProviderAPIBase)
	}

	hostname := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if hostname == "" {
		return "", fmt.Errorf("%w: host is required", ErrInvalidProviderAPIBase)
	}

	if parsed.Port() != "" {
		if _, err := net.LookupPort("tcp", parsed.Port()); err != nil {
			return "", fmt.Errorf("%w: port is invalid", ErrInvalidProviderAPIBase)
		}
	}

	if net.ParseIP(hostname) != nil {
		return "", fmt.Errorf("%w: host must be an approved provider domain", ErrInvalidProviderAPIBase)
	}

	if !providerAPIBaseHostAllowed(provider, hostname) {
		return "", fmt.Errorf("%w: host is not allowed for provider %q", ErrInvalidProviderAPIBase, strings.TrimSpace(provider))
	}

	parsed.Scheme = "https"
	parsed.Host = strings.ToLower(parsed.Host)

	return strings.TrimRight(parsed.String(), "/"), nil
}

func validateProviderConfigAPIBase(provider string, config map[string]interface{}, keys ...string) error {
	for _, key := range keys {
		raw, ok := config[key].(string)
		if !ok {
			continue
		}

		normalized, err := ValidateProviderAPIBase(provider, raw)
		if err != nil {
			return err
		}

		if normalized == "" {
			delete(config, key)
		} else {
			config[key] = normalized
		}
	}

	return nil
}

func safeProviderAPIBase(provider string, raw string) string {
	normalized, err := ValidateProviderAPIBase(provider, raw)
	if err != nil {
		return ""
	}

	return normalized
}

func providerAPIBaseHostAllowed(provider string, hostname string) bool {
	for _, allowed := range allowedProviderAPIBaseHosts(provider) {
		if hostMatchesAllowed(hostname, allowed) {
			return true
		}
	}

	return false
}

func hostMatchesAllowed(hostname string, allowed string) bool {
	allowed = strings.TrimSpace(strings.ToLower(allowed))
	if allowed == "" {
		return false
	}

	if strings.HasPrefix(allowed, ".") {
		suffix := allowed
		exact := strings.TrimPrefix(allowed, ".")

		return hostname == exact || strings.HasSuffix(hostname, suffix)
	}

	if strings.HasPrefix(allowed, "*") {
		return strings.HasSuffix(hostname, strings.TrimPrefix(allowed, "*"))
	}

	return hostname == allowed
}

func allowedProviderAPIBaseHosts(provider string) []string {
	switch normalizeProviderAPIBaseProvider(provider) {
	case "openai":
		return []string{"api.openai.com"}
	case "groq":
		return []string{"api.groq.com"}
	case "mistral":
		return []string{"api.mistral.ai"}
	case "sambanova":
		return []string{"api.sambanova.ai"}
	case "togetherai":
		return []string{"api.together.ai"}
	case "anthropic":
		return []string{"api.anthropic.com"}
	case "cohere":
		return []string{"api.cohere.com", "api.cohere.ai"}
	case "gemini":
		return []string{"generativelanguage.googleapis.com"}
	case "openrouter":
		return []string{"openrouter.ai"}
	case "perplexity":
		return []string{"api.perplexity.ai"}
	case "deepinfra":
		return []string{"api.deepinfra.com"}
	case "fireworks":
		return []string{"api.fireworks.ai"}
	case "xai":
		return []string{"api.x.ai"}
	case "azure":
		return []string{".openai.azure.com"}
	case "vertexai":
		return []string{"*-aiplatform.googleapis.com"}
	default:
		return nil
	}
}

func normalizeProviderAPIBaseProvider(provider string) string {
	switch strings.TrimSpace(strings.ToLower(provider)) {
	case "", "openai":
		return "openai"
	case "mistral-ai":
		return "mistral"
	case "samba-nova":
		return "sambanova"
	case "together-ai", "together":
		return "togetherai"
	case "perplexityai":
		return "perplexity"
	case "vertex", "vertex-ai", "google-vertex", "google-vertexai":
		return "vertexai"
	default:
		return strings.TrimSpace(strings.ToLower(provider))
	}
}
