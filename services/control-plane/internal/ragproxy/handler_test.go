package ragproxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestForwardRuntimePreservesOpenAIAuthenticationAndRateLimit(t *testing.T) {
	t.Parallel()

	const sessionToken = "private-chatbot-session-token"

	type observedRequest struct {
		path          string
		authorization string
		writeErr      error
	}

	observed := make(chan observedRequest, 1)

	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		writer.Header().Set(fiber.HeaderRetryAfter, "17")
		writer.WriteHeader(fiber.StatusTooManyRequests)

		_, writeErr := writer.Write([]byte(`{"error":{"type":"rate_limit_error"}}`))
		observed <- observedRequest{
			path: request.URL.Path, authorization: request.Header.Get(fiber.HeaderAuthorization), writeErr: writeErr,
		}
	}))
	defer upstream.Close()

	handler, err := New(upstream.URL, strings.Repeat("a", 32), time.Second)
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/v1/chatbots/:publicID/openai/chat/completions", handler.ForwardRuntime)

	request := httptest.NewRequest(
		fiber.MethodPost,
		"/v1/chatbots/bot-1/openai/chat/completions",
		strings.NewReader(`{"model":"bot-1","messages":[]}`),
	)
	request.Header.Set(fiber.HeaderAuthorization, "Bearer "+sessionToken)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := app.Test(request)
	require.NoError(t, err)

	defer func() { require.NoError(t, response.Body.Close()) }()

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	upstreamRequest := <-observed
	require.Equal(t, "/v1/chatbots/bot-1/openai/chat/completions", upstreamRequest.path)
	require.Equal(t, "Bearer "+sessionToken, upstreamRequest.authorization)
	require.NoError(t, upstreamRequest.writeErr)
	require.Equal(t, fiber.StatusTooManyRequests, response.StatusCode)
	require.Equal(t, "17", response.Header.Get(fiber.HeaderRetryAfter))
	require.JSONEq(t, `{"error":{"type":"rate_limit_error"}}`, string(body))
}
