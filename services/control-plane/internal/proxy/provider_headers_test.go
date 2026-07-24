package proxy

import (
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

	response, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("test request: %v", err)
	}
	if got := response.Header.Get("X-Shahrah-Mock-Provider"); got != "true" {
		t.Fatalf("expected mock provider marker, got %q", got)
	}
}
