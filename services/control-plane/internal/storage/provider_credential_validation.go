package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const providerCredentialValidationTimeout = 6 * time.Second

type providerCredentialValidationResult struct {
	Success bool
	Message string
}

func validateProviderAPIKey(ctx context.Context, provider, apiKey string, providerConfig map[string]interface{}, client *http.Client) providerCredentialValidationResult {
	probe, ok, probeErr := providerCredentialProbeFor(provider, providerConfig)
	if probeErr != nil {
		return providerCredentialValidationResult{
			Success: false,
			Message: "provider api_base is not allowed for credential validation",
		}
	}

	if !ok {
		return providerCredentialValidationResult{
			Success: false,
			Message: fmt.Sprintf("provider %q is not supported for direct connection validation", strings.TrimSpace(provider)),
		}
	}

	trimmedKey := strings.TrimSpace(apiKey)
	if trimmedKey == "" || isMaskedCredentialValue(trimmedKey) {
		return providerCredentialValidationResult{
			Success: false,
			Message: "API key is missing or masked; enter a new provider API key before testing",
		}
	}

	ctx, cancel := context.WithTimeout(ctx, providerCredentialValidationTimeout)
	defer cancel()

	httpClient := client
	if httpClient == nil {
		httpClient = &http.Client{Timeout: providerCredentialValidationTimeout}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probe.url, http.NoBody)
	if err != nil {
		return providerCredentialValidationResult{
			Success: false,
			Message: "provider validation endpoint is not configured correctly",
		}
	}

	for name, value := range probe.headers(trimmedKey) {
		req.Header.Set(name, value)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return providerCredentialValidationResult{Success: false, Message: "provider validation timed out"}
		}

		return providerCredentialValidationResult{Success: false, Message: "network error while contacting provider validation endpoint"}
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			return
		}
	}()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if readErr != nil {
		return providerCredentialValidationResult{Success: false, Message: "failed to read provider validation response"}
	}

	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return providerCredentialValidationResult{
			Success: true,
			Message: "provider accepted the credential on its models endpoint",
		}
	}

	return providerCredentialValidationResult{
		Success: false,
		Message: classifyProviderValidationStatus(resp.StatusCode, body, trimmedKey),
	}
}

type providerCredentialProbe struct {
	url     string
	headers func(apiKey string) map[string]string
}

func providerCredentialProbeFor(provider string, providerConfig map[string]interface{}) (providerCredentialProbe, bool, error) {
	normalized := strings.TrimSpace(strings.ToLower(provider))
	switch normalized {
	case "openai":
		return openAIModelsProbe(providerConfig)
	case "groq":
		return bearerModelsProbe(configString("groq", providerConfig, "api_base", "https://api.groq.com/openai/v1"))
	case "mistral", "mistral-ai":
		return bearerModelsProbe(configString("mistral-ai", providerConfig, "api_base", "https://api.mistral.ai/v1"))
	case "sambanova", "samba-nova":
		return bearerModelsProbe(configString("sambanova", providerConfig, "api_base", "https://api.sambanova.ai/v1"))
	case "togetherai", "together-ai", "together":
		return bearerModelsProbe(configString("togetherai", providerConfig, "api_base", "https://api.together.ai/v1"))
	default:
		return providerCredentialProbe{}, false, nil
	}
}

func openAIModelsProbe(providerConfig map[string]interface{}) (providerCredentialProbe, bool, error) {
	apiBase, err := configString("openai", providerConfig, "api_base", "https://api.openai.com/v1")
	if err != nil {
		return providerCredentialProbe{}, false, err
	}
	return providerCredentialProbe{
		url: strings.TrimRight(apiBase, "/") + "/models",
		headers: func(apiKey string) map[string]string {
			headers := map[string]string{"Authorization": "Bearer " + apiKey}
			if value := providerConfigText(providerConfig, "organization"); value != "" {
				headers["OpenAI-Organization"] = value
			}
			if value := providerConfigText(providerConfig, "project"); value != "" {
				headers["OpenAI-Project"] = value
			}
			return headers
		},
	}, true, nil
}

func providerConfigText(config map[string]interface{}, key string) string {
	value, _ := config[key].(string)
	return strings.TrimSpace(value)
}

func bearerModelsProbe(apiBase string, err error) (providerCredentialProbe, bool, error) {
	if err != nil {
		return providerCredentialProbe{}, false, err
	}

	return providerCredentialProbe{
		url: strings.TrimRight(apiBase, "/") + "/models",
		headers: func(apiKey string) map[string]string {
			return map[string]string{"Authorization": "Bearer " + apiKey}
		},
	}, true, nil
}

func classifyProviderValidationStatus(status int, body []byte, apiKey string) string {
	detail := sanitizedProviderError(body, apiKey)

	switch status {
	case http.StatusUnauthorized:
		return appendProviderDetail("provider rejected the API key as invalid", detail)
	case http.StatusForbidden:
		return appendProviderDetail("provider accepted the credential format but denied access", detail)
	case http.StatusPaymentRequired, http.StatusTooManyRequests:
		return appendProviderDetail("provider credential is valid but quota, billing, or rate limit blocked validation", detail)
	case http.StatusNotFound:
		return appendProviderDetail("provider validation endpoint was not found; check provider base URL configuration", detail)
	default:
		if status >= http.StatusInternalServerError {
			return appendProviderDetail("provider validation endpoint returned a server error", detail)
		}

		return appendProviderDetail(fmt.Sprintf("provider validation failed with HTTP %d", status), detail)
	}
}

func appendProviderDetail(message, detail string) string {
	if strings.TrimSpace(detail) == "" {
		return message
	}

	return message + ": " + detail
}

func sanitizedProviderError(body []byte, apiKey string) string {
	text := strings.TrimSpace(string(body))
	if text == "" {
		return ""
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err == nil {
		text = firstProviderErrorMessage(payload)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	text = strings.ReplaceAll(text, apiKey, "[redacted]")
	text = bearerTokenPattern.ReplaceAllString(text, "Bearer [redacted]")

	text = apiKeyPattern.ReplaceAllString(text, "[redacted]")
	if len(text) > 180 {
		text = strings.TrimSpace(text[:180]) + "..."
	}

	return text
}

func firstProviderErrorMessage(payload map[string]interface{}) string {
	if message, ok := payload["message"].(string); ok {
		return message
	}

	if errorValue, ok := payload["error"]; ok {
		switch typed := errorValue.(type) {
		case string:
			return typed
		case map[string]interface{}:
			if message, ok := typed["message"].(string); ok {
				return message
			}

			if detail, ok := typed["detail"].(string); ok {
				return detail
			}

			if code, ok := typed["code"].(string); ok {
				return code
			}
		}
	}

	if detail, ok := payload["detail"].(string); ok {
		return detail
	}

	return ""
}

var (
	bearerTokenPattern = regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._~+/=-]+`)
	apiKeyPattern      = regexp.MustCompile(`(?i)(sk|gsk|mistral|together)[-_][A-Za-z0-9._~+/=-]{8,}`)
)

func credentialHintFromMap(credentials map[string]interface{}) string {
	if apiKey, ok := credentials["api_key"].(string); ok {
		return maskCredentialHint(apiKey)
	}

	if secretRef, ok := credentials["secret_ref"].(string); ok && strings.TrimSpace(secretRef) != "" {
		return "secret_ref:" + maskCredentialHint(secretRef)
	}

	if _, ok := credentials["raw"]; ok {
		return "configured"
	}

	return ""
}

func maskCredentialHint(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	if isMaskedCredentialValue(trimmed) {
		return trimmed
	}

	prefixLen := 4
	if strings.HasPrefix(trimmed, "sk-") || strings.HasPrefix(trimmed, "gsk_") || strings.HasPrefix(trimmed, "gsk-") {
		prefixLen = 3
	}

	if len(trimmed) <= prefixLen+4 {
		return "••••"
	}

	return trimmed[:prefixLen] + "••••••••" + trimmed[len(trimmed)-4:]
}

func isMaskedCredentialValue(value string) bool {
	trimmed := strings.TrimSpace(value)
	return strings.Contains(trimmed, "•") || strings.Contains(trimmed, "*") || strings.Contains(trimmed, "…")
}

func configString(provider string, config map[string]interface{}, key, fallback string) (string, error) {
	if value, ok := config[key].(string); ok && strings.TrimSpace(value) != "" {
		normalized, err := ValidateProviderAPIBase(provider, value)
		if err != nil {
			return "", err
		}

		return normalized, nil
	}

	return fallback, nil
}
