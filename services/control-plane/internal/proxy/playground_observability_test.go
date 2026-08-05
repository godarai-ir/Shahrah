package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestPlaygroundObservabilityRequiresTrustedDelegatedAuthentication(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name      string
		delegated bool
		wantRunID string
	}{
		{name: "trusted delegated request", delegated: true, wantRunID: "run-1"},
		{name: "untrusted public request", delegated: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Get("/", func(c *fiber.Ctx) error {
				if testCase.delegated {
					auth.MarkDelegatedRequest(c)
				}

				if err := parsePlaygroundObservability(c); err != nil {
					return err
				}

				return c.JSON(enrichPlaygroundObservability(c, map[string]interface{}{"endpoint": "/v1/chat/completions"}))
			})

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.Header.Set(apiheaders.HeaderPlaygroundObservability, `{"source":"tenant_admin_playground","client_run_id":"run-1","tags":{"experiment":"a"}}`)
			resp, err := app.Test(req)
			require.NoError(t, err)
			t.Cleanup(func() { closeResponseBody(t, resp) })
			require.Equal(t, fiber.StatusOK, resp.StatusCode)

			var got map[string]interface{}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))

			if testCase.wantRunID == "" {
				require.NotContains(t, got, "playground_client_run_id")
				require.NotContains(t, got, "playground_tags")

				return
			}

			require.Equal(t, testCase.wantRunID, got["playground_client_run_id"])
			require.Equal(t, map[string]interface{}{"experiment": "a"}, got["playground_tags"])
		})
	}
}
