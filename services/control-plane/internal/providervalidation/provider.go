// Package providervalidation implements non-billable provider credential
// probes shared by tenant BYOK onboarding and platform-managed credentials.
package providervalidation

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const (
	defaultAnthropicBase  = "https://api.anthropic.com"
	defaultCerebrasBase   = "https://api.cerebras.ai/v1"
	defaultCohereBase     = "https://api.cohere.com"
	defaultDeepInfraBase  = "https://api.deepinfra.com/v1/openai"
	defaultGeminiBase     = "https://generativelanguage.googleapis.com/v1beta"
	defaultGroqBase       = "https://api.groq.com/openai/v1"
	defaultMistralBase    = "https://api.mistral.ai/v1"
	defaultOpenAIBase     = "https://api.openai.com/v1"
	defaultOpenRouterBase = "https://openrouter.ai/api/v1"
	defaultTogetherBase   = "https://api.together.xyz/v1"
	defaultXAIBase        = "https://api.x.ai/v1"
)

// Contract identifies one canonical provider authentication contract.
type Contract struct {
	Provider string
	AuthMode string
}

type probeBuilder func(Request) (probeSpec, error)

var probeRegistry = map[Contract]probeBuilder{
	{Provider: "anthropic", AuthMode: "api_key"}:                buildAnthropicProbe,
	{Provider: "azure-open-ai", AuthMode: "api_key"}:            buildAzureOpenAIProbe,
	{Provider: "cerebras", AuthMode: "api_key"}:                 buildBearerModelsProbe(defaultCerebrasBase),
	{Provider: "cohere", AuthMode: "api_key"}:                   buildCohereProbe,
	{Provider: "databricks", AuthMode: "personal_access_token"}: buildDatabricksProbe,
	{Provider: "deepinfra", AuthMode: "api_key"}:                buildBearerModelsProbe(defaultDeepInfraBase),
	{Provider: "google-gemini", AuthMode: "api_key"}:            buildGeminiProbe,
	{Provider: "groq", AuthMode: "api_key"}:                     buildBearerModelsProbe(defaultGroqBase),
	{Provider: "mistral-ai", AuthMode: "api_key"}:               buildBearerModelsProbe(defaultMistralBase),
	{Provider: "openai", AuthMode: "api_key"}:                   buildOpenAIProbe,
	{Provider: "openrouter", AuthMode: "api_key"}:               buildOpenRouterProbe,
	{Provider: "together-ai", AuthMode: "api_key"}:              buildBearerModelsProbe(defaultTogetherBase),
	{Provider: "xai", AuthMode: "api_key"}:                      buildBearerModelsProbe(defaultXAIBase),
}

// NormalizeProvider returns the database/catalog slug for known aliases.
func NormalizeProvider(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "azure":
		return "azure-open-ai"
	case "bedrock":
		return "aws-bedrock"
	case "gemini", "google-ai-studio":
		return "google-gemini"
	case "vertex", "vertex-ai", "vertexai", "google-vertexai":
		return "google-vertex"
	case "mistral":
		return "mistral-ai"
	case "samba-nova":
		return "sambanova"
	case "together", "togetherai":
		return "together-ai"
	case "perplexity", "perplexityai":
		return "perplexity-ai"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

// Supports reports whether a documented non-billable probe exists.
func Supports(provider, authMode string) bool {
	_, ok := probeRegistry[canonicalContract(provider, authMode)]

	return ok
}

// Contracts returns the supported contracts in stable order for diagnostics
// and catalog-alignment tests.
func Contracts() []Contract {
	items := make([]Contract, 0, len(probeRegistry))
	for contract := range probeRegistry {
		items = append(items, contract)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Provider == items[j].Provider {
			return items[i].AuthMode < items[j].AuthMode
		}

		return items[i].Provider < items[j].Provider
	})

	return items
}

func canonicalContract(provider, authMode string) Contract {
	return Contract{
		Provider: NormalizeProvider(provider),
		AuthMode: strings.ToLower(strings.TrimSpace(authMode)),
	}
}

func buildProbe(request Request) (probeSpec, bool, error) {
	builder, ok := probeRegistry[canonicalContract(request.Provider, request.AuthMode)]
	if !ok {
		return probeSpec{}, false, nil
	}

	spec, err := builder(request)
	if err != nil {
		return probeSpec{}, true, err
	}

	return spec, true, nil
}

func buildBearerModelsProbe(fallback string) probeBuilder {
	return func(request Request) (probeSpec, error) {
		base, err := validatedConfiguredBase(request, fallback, "base_url", "api_base")
		if err != nil {
			return probeSpec{}, err
		}

		apiKey, err := requiredSecret(request, "api_key")
		if err != nil {
			return probeSpec{}, err
		}

		return probeSpec{
			url:     appendURLPath(base, "models"),
			headers: bearerHeaders(apiKey),
		}, nil
	}
}

func buildOpenAIProbe(request Request) (probeSpec, error) {
	spec, err := buildBearerModelsProbe(defaultOpenAIBase)(request)
	if err != nil {
		return probeSpec{}, err
	}

	setOptionalHeader(spec.headers, "OpenAI-Organization", configString(request.Config, "organization"))
	setOptionalHeader(spec.headers, "OpenAI-Project", configString(request.Config, "project"))

	return spec, nil
}

func buildAnthropicProbe(request Request) (probeSpec, error) {
	base, err := validatedConfiguredBase(request, defaultAnthropicBase, "base_url", "api_base")
	if err != nil {
		return probeSpec{}, err
	}

	apiKey, err := requiredSecret(request, "api_key")
	if err != nil {
		return probeSpec{}, err
	}

	version := configString(request.Config, "api_version")
	if version == "" {
		version = "2023-06-01"
	}
	if containsHeaderBreak(version) {
		return probeSpec{}, fmt.Errorf("%w: invalid Anthropic API version", ErrInvalidConfiguration)
	}

	base = ensureVersionedBase(base, "v1")

	return probeSpec{
		url: appendURLPath(base, "models") + "?limit=1",
		headers: map[string]string{
			"x-api-key":         apiKey,
			"anthropic-version": version,
		},
	}, nil
}

func buildAzureOpenAIProbe(request Request) (probeSpec, error) {
	base, err := validatedConfiguredBase(request, "", "endpoint", "base_url", "api_base")
	if err != nil {
		return probeSpec{}, err
	}
	if base == "" {
		return probeSpec{}, fmt.Errorf("%w: Azure OpenAI endpoint is required", ErrInvalidConfiguration)
	}

	apiKey, err := requiredSecret(request, "api_key")
	if err != nil {
		return probeSpec{}, err
	}

	probeURL, err := azureOpenAIModelsURL(base, configString(request.Config, "api_version"))
	if err != nil {
		return probeSpec{}, err
	}

	return probeSpec{
		url:     probeURL,
		headers: map[string]string{"api-key": apiKey},
	}, nil
}

func buildCohereProbe(request Request) (probeSpec, error) {
	base, err := validatedConfiguredBase(request, defaultCohereBase, "base_url", "api_base")
	if err != nil {
		return probeSpec{}, err
	}

	apiKey, err := requiredSecret(request, "api_key")
	if err != nil {
		return probeSpec{}, err
	}

	base = ensureVersionedBase(base, "v1")
	headers := bearerHeaders(apiKey)
	setOptionalHeader(headers, "X-Client-Name", configString(request.Config, "client_name"))

	return probeSpec{url: appendURLPath(base, "models"), headers: headers}, nil
}

func buildDatabricksProbe(request Request) (probeSpec, error) {
	base, err := validatedConfiguredBase(request, "", "serving_endpoint_url", "endpoint", "base_url", "api_base")
	if err != nil {
		return probeSpec{}, err
	}
	if base == "" {
		return probeSpec{}, fmt.Errorf("%w: Databricks serving endpoint URL is required", ErrInvalidConfiguration)
	}

	token, err := requiredSecret(request, "access_token", "api_key")
	if err != nil {
		return probeSpec{}, err
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return probeSpec{}, fmt.Errorf("%w: invalid Databricks endpoint", ErrInvalidConfiguration)
	}

	parsed.Path = "/api/2.0/serving-endpoints"
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return probeSpec{url: parsed.String(), headers: bearerHeaders(token)}, nil
}

func buildGeminiProbe(request Request) (probeSpec, error) {
	base, err := validatedConfiguredBase(request, defaultGeminiBase, "base_url", "api_base")
	if err != nil {
		return probeSpec{}, err
	}

	apiKey, err := requiredSecret(request, "api_key")
	if err != nil {
		return probeSpec{}, err
	}

	version := configString(request.Config, "api_version")
	if version == "" {
		version = "v1beta"
	}

	base, err = ensureGeminiVersionedBase(base, version)
	if err != nil {
		return probeSpec{}, err
	}

	return probeSpec{
		url:     appendURLPath(base, "models") + "?pageSize=1",
		headers: map[string]string{"x-goog-api-key": apiKey},
	}, nil
}

func buildOpenRouterProbe(request Request) (probeSpec, error) {
	base, err := validatedConfiguredBase(request, defaultOpenRouterBase, "base_url", "api_base")
	if err != nil {
		return probeSpec{}, err
	}

	apiKey, err := requiredSecret(request, "api_key")
	if err != nil {
		return probeSpec{}, err
	}

	base = ensureOpenRouterAPIBase(base)
	headers := bearerHeaders(apiKey)
	setOptionalHeader(headers, "HTTP-Referer", configString(request.Config, "http_referer"))
	setOptionalHeader(headers, "X-OpenRouter-Title", configString(request.Config, "app_title"))

	// The models catalog can be read without authenticating a key. /key is the
	// documented non-billable endpoint that proves the supplied key itself.
	return probeSpec{url: appendURLPath(base, "key"), headers: headers}, nil
}

func validatedConfiguredBase(request Request, fallback string, keys ...string) (string, error) {
	raw := configString(request.Config, keys...)
	if raw == "" {
		raw = fallback
	}
	if raw == "" {
		return "", nil
	}

	normalized, err := ValidateBaseURL(request.Provider, raw)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidConfiguration, err)
	}

	return normalized, nil
}

func configString(config map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := config[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}

func requiredSecret(request Request, keys ...string) (string, error) {
	for _, key := range keys {
		if value := strings.TrimSpace(request.Secrets[key]); value != "" {
			if isMaskedSecret(value) {
				return "", fmt.Errorf("%w: credential is masked", ErrMissingCredential)
			}

			return value, nil
		}
	}

	return "", fmt.Errorf("%w: required credential field %q is missing", ErrMissingCredential, keys[0])
}

func bearerHeaders(secret string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + secret}
}

func setOptionalHeader(headers map[string]string, name, value string) {
	if value != "" && !containsHeaderBreak(value) {
		headers[name] = value
	}
}

func containsHeaderBreak(value string) bool {
	return strings.ContainsAny(value, "\r\n")
}

func isMaskedSecret(value string) bool {
	return strings.Contains(value, "•") ||
		strings.Contains(value, "*") ||
		strings.Contains(value, "…")
}
