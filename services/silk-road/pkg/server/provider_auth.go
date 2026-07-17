package server

import (
	"net/http"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// deploymentAPIKey is the single auth-material extraction point. Typed tunnel
// auth wins; APIKey remains a temporary compatibility path for older control
// plane nodes during rolling deployment.
func deploymentAPIKey(deployment platform.Deployment) string {
	if deployment.ProviderAuth != nil {
		switch deployment.ProviderAuth.Mode {
		case "api_key", "personal_access_token":
		default:
			return ""
		}
		return deployment.ProviderAuth.APIKey
	}
	return deployment.APIKey
}

func applyDocumentedProviderHeaders(req *http.Request, deployment platform.Deployment) {
	if req == nil || len(deployment.ProviderHeaders) == 0 {
		return
	}
	switch normalizeProvider(deployment.Provider) {
	case "openai":
		copyHeader(req, deployment, "OpenAI-Organization")
		copyHeader(req, deployment, "OpenAI-Project")
	case "openrouter":
		copyHeader(req, deployment, "HTTP-Referer")
		copyHeader(req, deployment, "X-OpenRouter-Title")
	case "cohere":
		copyHeader(req, deployment, "X-Client-Name")
	}
}

func copyHeader(req *http.Request, deployment platform.Deployment, name string) {
	if value := strings.TrimSpace(deployment.ProviderHeaders[name]); value != "" {
		req.Header.Set(name, value)
	}
}
