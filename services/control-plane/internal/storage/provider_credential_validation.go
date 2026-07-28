package storage

import (
	"context"
	"net/http"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/providervalidation"
)

type providerCredentialValidationResult = providervalidation.Result

func validateProviderCredential(
	ctx context.Context,
	provider string,
	authMode string,
	providerConfig map[string]interface{},
	secrets map[string]string,
	client *http.Client,
) providerCredentialValidationResult {
	return providervalidation.NewProber(client).Probe(ctx, providervalidation.Request{
		Provider: provider,
		AuthMode: authMode,
		Config:   providerConfig,
		Secrets:  secrets,
	})
}

func providerCredentialSecrets(credentials map[string]interface{}) map[string]string {
	secrets := make(map[string]string)

	for key, value := range credentials {
		if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
			secrets[key] = strings.TrimSpace(text)
		}
	}

	if raw, ok := credentials["raw"].(map[string]interface{}); ok {
		for key, value := range raw {
			if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
				secrets[key] = strings.TrimSpace(text)
			}
		}
	}

	return secrets
}

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
