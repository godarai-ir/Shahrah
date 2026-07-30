package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestForwardProviderHeadersIncludesMockMarker(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		forwardProviderHeaders(c, map[string]string{
			"X-Shahrah-Mock-Provider": "true",
		})

		return c.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", http.NoBody))
	if err != nil {
		t.Fatalf("test request: %v", err)
	}

	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			t.Errorf("close response body: %v", closeErr)
		}
	}()

	if got := response.Header.Get("X-Shahrah-Mock-Provider"); got != "true" {
		t.Fatalf("expected mock provider marker, got %q", got)
	}
}

func TestForwardProviderHeadersIncludesAllDocumentedAnthropicRateLimits(
	t *testing.T,
) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		forwardProviderHeaders(c, map[string]string{
			"anthropic-ratelimit-tokens-limit":           "10000",
			"anthropic-ratelimit-tokens-remaining":       "9000",
			"anthropic-ratelimit-tokens-reset":           "2026-07-29T12:00:00Z",
			"anthropic-priority-input-tokens-remaining":  "8000",
			"anthropic-priority-output-tokens-remaining": "7000",
			"anthropic-fast-input-tokens-remaining":      "6000",
			"anthropic-fast-output-tokens-remaining":     "5000",
		})
		return c.SendStatus(fiber.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{
		"Anthropic-Ratelimit-Tokens-Limit":           "10000",
		"Anthropic-Ratelimit-Tokens-Remaining":       "9000",
		"Anthropic-Ratelimit-Tokens-Reset":           "2026-07-29T12:00:00Z",
		"Anthropic-Priority-Input-Tokens-Remaining":  "8000",
		"Anthropic-Priority-Output-Tokens-Remaining": "7000",
		"Anthropic-Fast-Input-Tokens-Remaining":      "6000",
		"Anthropic-Fast-Output-Tokens-Remaining":     "5000",
	} {
		if got := response.Header.Get(name); got != want {
			t.Fatalf("%s=%q want %q", name, got, want)
		}
	}
}

func TestForwardProviderHeadersIncludesAllDocumentedOpenAIRateLimits(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		forwardProviderHeaders(c, map[string]string{
			"x-ratelimit-limit-requests":     "500",
			"x-ratelimit-limit-tokens":       "500000",
			"x-ratelimit-remaining-requests": "499",
			"x-ratelimit-remaining-tokens":   "499000",
			"x-ratelimit-reset-requests":     "120ms",
			"x-ratelimit-reset-tokens":       "1s",
		})
		return c.SendStatus(fiber.StatusNoContent)
	})

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	for name, want := range map[string]string{
		"X-Ratelimit-Limit-Requests":     "500",
		"X-Ratelimit-Limit-Tokens":       "500000",
		"X-Ratelimit-Remaining-Requests": "499",
		"X-Ratelimit-Remaining-Tokens":   "499000",
		"X-Ratelimit-Reset-Requests":     "120ms",
		"X-Ratelimit-Reset-Tokens":       "1s",
	} {
		if got := response.Header.Get(name); got != want {
			t.Fatalf("%s=%q want %q", name, got, want)
		}
	}
}
