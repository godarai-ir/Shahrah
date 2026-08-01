package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/gofiber/fiber/v2"
)

func TestGeminiOperationFromRequestUsesFiniteNativeRoutes(t *testing.T) {
	tests := []struct {
		method, path, body string
		operation          platform.ProviderOperation
		model              string
		stream             bool
	}{
		{http.MethodPost, "/v1beta/models/gemini-3.6-flash:generateContent", `{}`, platform.ProviderOperationGeminiGenerateContent, "gemini-3.6-flash", false},
		{http.MethodPost, "/v1beta/models/gemini-3.6-flash:streamGenerateContent", `{}`, platform.ProviderOperationGeminiStreamGenerateContent, "gemini-3.6-flash", true},
		{http.MethodPost, "/v1beta/models/gemini-embedding-2:embedContent", `{}`, platform.ProviderOperationGeminiEmbedContent, "gemini-embedding-2", false},
		{http.MethodPost, "/v1beta/interactions", `{"model":"gemini-3.5-flash"}`, platform.ProviderOperationGeminiInteractionsCreate, "gemini-3.5-flash", false},
		{http.MethodPost, "/v1beta/interactions", `{"model":"gemini-3.5-flash","stream":true}`, platform.ProviderOperationGeminiInteractionsCreate, "gemini-3.5-flash", true},
		{http.MethodDelete, "/v1beta/cachedContents/cache-1", ``, platform.ProviderOperationGeminiCachesDelete, "", false},
	}
	for _, test := range tests {
		app := fiber.New()
		app.All("/*", func(c *fiber.Ctx) error {
			operation, _, model, stream, err := geminiOperationFromRequest(c)
			if err != nil {
				return err
			}

			if operation != test.operation || model != test.model || stream != test.stream {
				t.Fatalf("%s: got operation=%s model=%s stream=%v", test.path, operation, model, stream)
			}

			return c.SendStatus(http.StatusNoContent)
		})

		request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
		request.Header.Set("Content-Type", "application/json")

		response, err := app.Test(request)
		if err != nil || response.StatusCode != http.StatusNoContent {
			t.Fatalf("%s failed: status=%v err=%v", test.path, response.StatusCode, err)
		}

		_ = response.Body.Close()
	}
}

func TestParseGeminiNativeUsagePreservesModalitiesThinkingTierAndSearch(t *testing.T) {
	usage := parseGeminiNativeUsage([]byte(`{
		"usageMetadata": {
			"promptTokenCount": 12,
			"candidatesTokenCount": 7,
			"thoughtsTokenCount": 3,
			"totalTokenCount": 22,
			"cachedContentTokenCount": 4,
			"trafficType": "PRIORITY",
			"promptTokensDetails": [{"modality":"IMAGE","tokenCount":5}],
			"candidatesTokensDetails": [{"modality":"AUDIO","tokenCount":6}]
		},
		"candidates": [{"groundingMetadata":{"webSearchQueries":["one","two"]}}]
	}`))
	if usage == nil || usage.PromptTokens != 12 || usage.CompletionTokens != 10 || usage.CacheReadInputTokens != 4 || usage.ServiceTier != "priority" {
		t.Fatalf("unexpected Gemini usage: %+v", usage)
	}

	if usage.PromptTokensDetails == nil || usage.PromptTokensDetails.ImageTokens != 5 || usage.CompletionTokensDetails == nil || usage.CompletionTokensDetails.AudioTokens != 6 {
		t.Fatalf("missing modality details: %+v", usage)
	}

	if usage.ServerToolUse["google_search_queries"] != 2 {
		t.Fatalf("missing Google Search usage: %+v", usage.ServerToolUse)
	}
}

func TestGeminiOperationFromRequestRejectsArbitraryActions(t *testing.T) {
	app := fiber.New()
	app.All("/*", func(c *fiber.Ctx) error {
		_, _, _, _, err := geminiOperationFromRequest(c)
		if err == nil {
			t.Fatal("arbitrary operation was accepted")
		}

		return c.Status(http.StatusBadRequest).SendString(err.Error())
	})

	response, err := app.Test(httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-3.6-flash:arbitrary", http.NoBody))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusBadRequest || !strings.Contains(string(body), "unsupported Gemini model operation") {
		t.Fatalf("unexpected response: %d %s", response.StatusCode, body)
	}
}

func TestGeminiOperationFromRequestRejectsWrongMethodForModelAction(t *testing.T) {
	app := fiber.New()
	app.All("/*", func(c *fiber.Ctx) error {
		_, _, _, _, err := geminiOperationFromRequest(c)
		if err == nil {
			t.Fatal("GET model action was accepted")
		}

		return c.Status(http.StatusBadRequest).SendString(err.Error())
	})

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/v1beta/models/gemini-3.6-flash:generateContent", http.NoBody))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusBadRequest || !strings.Contains(string(body), "require POST") {
		t.Fatalf("unexpected response: %d %s", response.StatusCode, body)
	}
}

func TestGoogleGeminiAllowlistIsExact(t *testing.T) {
	if len(platform.SupportedGoogleGeminiModelIDs) != 9 {
		t.Fatalf("allowlist has %d models", len(platform.SupportedGoogleGeminiModelIDs))
	}

	for _, unsupported := range []string{"gemini-latest", "gemini-3.1-pro-preview-customtools", "gemma-3-27b-it", "veo-3.1"} {
		if isAllowlistedGoogleGeminiModel(unsupported) {
			t.Fatalf("unexpected allowlisted model %s", unsupported)
		}
	}
}
