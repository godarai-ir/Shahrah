package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/gofiber/fiber/v2"
)

// writeAPIError writes a standardized OpenAI-compatible error response.
func writeAPIError(c *fiber.Ctx, status int, message string, errorType string) error {
	return c.Status(status).JSON(contractsopenai.APIError{
		Error: contractsopenai.ErrorInfo{
			Message: message,
			Type:    errorType,
			Code:    fmt.Sprintf("%d", status),
		},
	})
}

// transformProviderError converts non-OpenAI provider errors to OpenAI-compatible format.
func transformProviderError(c *fiber.Ctx, statusCode int, body []byte, headers map[string]string) error {
	// Forward provider headers even on error
	forwardProviderHeaders(c, headers)

	// Try to parse as OpenAI error first
	var apiErr contractsopenai.APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		return c.Status(statusCode).JSON(apiErr)
	}

	// Fall back to generic OpenAI-compatible error
	message := string(body)
	if len(message) > 500 {
		message = message[:500] + "..."
	}
	if message == "" {
		message = http.StatusText(statusCode)
	}

	errorType := classifyErrorType(statusCode)

	return c.Status(statusCode).JSON(contractsopenai.APIError{
		Error: contractsopenai.ErrorInfo{
			Message: message,
			Type:    errorType,
			Code:    fmt.Sprintf("%d", statusCode),
		},
	})
}

func classifyErrorType(statusCode int) string {
	switch {
	case statusCode == http.StatusTooManyRequests:
		return "rate_limit_error"
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return "authentication_error"
	case statusCode >= http.StatusInternalServerError:
		return "server_error"
	default:
		return "api_error"
	}
}

func containsModel(models []string, model string) bool {
	for _, candidate := range models {
		if candidate == model {
			return true
		}
	}
	return false
}

// forwardProviderHeaders forwards allowed provider response headers to the client.
func forwardProviderHeaders(c *fiber.Ctx, headers map[string]string) {
	allowedPrefixes := []string{
		"x-ratelimit-",
		"x-request-id",
		"x-trace-id",
		"x-correlation-id",
		"cf-ray",
		"traceparent",
	}

	for key, val := range headers {
		if isAllowedHeader(key, allowedPrefixes) {
			c.Set(key, val)
		}
	}
}

// isAllowedHeader checks if a header key matches allowed prefixes (case-insensitive).
func isAllowedHeader(key string, prefixes []string) bool {
	lower := strings.ToLower(key)
	for _, prefix := range prefixes {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}
	return false
}
