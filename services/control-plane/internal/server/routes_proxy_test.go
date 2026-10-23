package server

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/proxy"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

var updateProxyRouteGolden = flag.Bool("update-proxy-routes", false, "rewrite testdata/proxy_routes.golden")

// proxyRouteScenarios cover every nil-handler branch in registerProxyRoutes,
// including the fallbacks that keep TokenBazaar, xAI and OCR routes reachable
// without the direct provider handlers.
func proxyRouteScenarios() []struct {
	name string
	deps Dependencies
} {
	all := Dependencies{
		BudgetMeHandler:                   func(*fiber.Ctx) error { return nil },
		ProxyHandler:                      &proxy.ChatCompletionHandler{},
		CompletionsHandler:                &proxy.CompletionsHandler{},
		ResponsesHandler:                  &proxy.ResponsesHandler{},
		EmbeddingsHandler:                 &proxy.EmbeddingsHandler{},
		ModerationsHandler:                &proxy.ModerationsHandler{},
		RerankHandler:                     &proxy.RerankHandler{},
		ImagesHandler:                     &proxy.ImagesHandler{},
		AudioTranscriptionsHandler:        &proxy.AudioTranscriptionsHandler{},
		AudioSpeechHandler:                &proxy.AudioSpeechHandler{},
		MistralOCRHandler:                 &proxy.MistralOCRHandler{},
		ZAILayoutParsingHandler:           &proxy.ZAILayoutParsingHandler{},
		MistralNativeHandler:              &proxy.MistralNativeHandler{},
		ModelsHandler:                     &proxy.ModelsHandler{},
		AnthropicNativeHandler:            &proxy.AnthropicNativeHandler{},
		TokenBazaarMessagesHandler:        &proxy.TokenBazaarMessagesHandler{},
		TokenBazaarGenerateContentHandler: &proxy.TokenBazaarGenerateContentHandler{},
		GeminiNativeHandler:               &proxy.GeminiNativeHandler{},
	}

	return []struct {
		name string
		deps Dependencies
	}{
		{name: "all", deps: all},
		{name: "none", deps: Dependencies{}},
		{name: "responses-only", deps: Dependencies{ResponsesHandler: &proxy.ResponsesHandler{}}},
		{name: "anthropic-only", deps: Dependencies{AnthropicNativeHandler: &proxy.AnthropicNativeHandler{}}},
		{name: "mistral-native-only", deps: Dependencies{MistralNativeHandler: &proxy.MistralNativeHandler{}}},
		{name: "tokenbazaar-only", deps: Dependencies{
			TokenBazaarMessagesHandler:        &proxy.TokenBazaarMessagesHandler{},
			TokenBazaarGenerateContentHandler: &proxy.TokenBazaarGenerateContentHandler{},
		}},
		{name: "zai-only", deps: Dependencies{ZAILayoutParsingHandler: &proxy.ZAILayoutParsingHandler{}}},
		{name: "models-only", deps: Dependencies{ModelsHandler: &proxy.ModelsHandler{}}},
	}
}

func renderProxyRoutes(deps Dependencies) string {
	app := fiber.New()
	registerProxyRoutes(app, deps)

	var b strings.Builder

	for _, route := range app.GetRoutes(true) {
		fmt.Fprintf(&b, "%s %s handlers=%d\n", route.Method, route.Path, len(route.Handlers))
	}

	return b.String()
}

// TestProxyRouteTableIsStable pins the exact registered route table (method,
// path, order and handler count) so refactors of registerProxyRoutes cannot
// change routing. Regenerate deliberately with -update-proxy-routes.
func TestProxyRouteTableIsStable(t *testing.T) {
	var b strings.Builder

	for _, scenario := range proxyRouteScenarios() {
		fmt.Fprintf(&b, "## %s\n%s\n", scenario.name, renderProxyRoutes(scenario.deps))
	}

	golden := filepath.Join("testdata", "proxy_routes.golden")
	if *updateProxyRouteGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0o750))
		require.NoError(t, os.WriteFile(golden, []byte(b.String()), 0o600))
	}

	want, err := os.ReadFile(golden)
	require.NoError(t, err)
	require.Equal(t, strings.ReplaceAll(string(want), "\r\n", "\n"), b.String())
}
