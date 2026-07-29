// Package proxy hosts OpenAI-compatible request handlers for the control plane.
package proxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// Error codes matching LiteLLM exception_mapping.md.
const (
	ErrBadRequest         = 400
	ErrAuthentication     = 401
	ErrPermissionDenied   = 403
	ErrNotFound           = 404
	ErrTimeout            = 408
	ErrRateLimit          = 429
	ErrInternal           = 500
	ErrServiceUnavailable = 503
)

// ErrorTypeMap maps HTTP status codes to OpenAI error types.
var ErrorTypeMap = map[int]string{
	ErrBadRequest:         "invalid_request_error",
	ErrAuthentication:     "authentication_error",
	ErrPermissionDenied:   "permission_denied_error",
	ErrNotFound:           "invalid_request_error",
	ErrTimeout:            "timeout_error",
	ErrRateLimit:          "rate_limit_error",
	ErrInternal:           "server_error",
	ErrServiceUnavailable: "server_error",
}

// TransformProviderError transforms a provider error to OpenAI envelope format
// This version matches existing handler code signatures - takes 4 args including headers.
func TransformProviderError(c *fiber.Ctx, status int, body []byte, headers map[string]string) error {
	// Forward provider headers before writing error response
	forwardProviderHeaders(c, headers)

	errType := ErrorTypeMap[status]
	if errType == "" {
		errType = "internal_error"
	}

	message := parseGenericError(body)

	return WriteAPIError(c, status, message, errType)
}

// transformProviderError is an alias for TransformProviderError (for backward compatibility).
var transformProviderError = TransformProviderError

// parseGenericError tries to extract a message from any error format.
func parseGenericError(body []byte) string {
	if len(body) == 0 {
		return "Provider returned an empty response"
	}

	var generic map[string]interface{}
	if err := json.Unmarshal(body, &generic); err == nil {
		// Try top-level message
		if msg, ok := generic["message"].(string); ok && msg != "" {
			return msg
		}
		// Try nested error object
		if errObj, ok := generic["error"].(map[string]interface{}); ok {
			if msg, ok := errObj["message"].(string); ok && msg != "" {
				return msg
			}

			if msg, ok := errObj["error"].(string); ok && msg != "" {
				return msg
			}
		}
		// Try simple error string
		if msg, ok := generic["error"].(string); ok && msg != "" {
			return msg
		}

		if msg, ok := generic["error_message"].(string); ok && msg != "" {
			return msg
		}
	}

	if len(body) > 200 {
		return strings.TrimSpace(string(body[:200])) + "..."
	}

	return strings.TrimSpace(string(body))
}

// Aliases for backward compatibility with existing handler code.
var writeAPIError = WriteAPIError
var containsModel = ContainsModel

// forwardProviderHeaders forwards relevant provider headers to the client
// Uses case-insensitive matching to handle providers that send headers in any case.
func forwardProviderHeaders(c *fiber.Ctx, headers map[string]string) {
	if headers == nil {
		return
	}

	// Build a normalized lookup map (lowercase keys)
	normalized := make(map[string]string, len(headers))
	for k, v := range headers {
		normalized[strings.ToLower(k)] = v
	}

	// Header mappings: lowercase lookup key -> proper mixed-case output key
	// This table handles both formats providers may send
	headerMappings := map[string]string{
		"request-id":                                  "Request-Id",
		"x-request-id":                                "X-Request-Id",
		"retry-after":                                 "Retry-After",
		"anthropic-ratelimit-requests-limit":          "Anthropic-Ratelimit-Requests-Limit",
		"anthropic-ratelimit-requests-remaining":      "Anthropic-Ratelimit-Requests-Remaining",
		"anthropic-ratelimit-requests-reset":          "Anthropic-Ratelimit-Requests-Reset",
		"anthropic-ratelimit-input-tokens-limit":      "Anthropic-Ratelimit-Input-Tokens-Limit",
		"anthropic-ratelimit-input-tokens-remaining":  "Anthropic-Ratelimit-Input-Tokens-Remaining",
		"anthropic-ratelimit-input-tokens-reset":      "Anthropic-Ratelimit-Input-Tokens-Reset",
		"anthropic-ratelimit-output-tokens-limit":     "Anthropic-Ratelimit-Output-Tokens-Limit",
		"anthropic-ratelimit-output-tokens-remaining": "Anthropic-Ratelimit-Output-Tokens-Remaining",
		"anthropic-ratelimit-output-tokens-reset":     "Anthropic-Ratelimit-Output-Tokens-Reset",
		"x-ratelimit-remaining-tokens":                "X-Ratelimit-Remaining-Tokens",
		"x-ratelimit-remaining-requests":              "X-Ratelimit-Remaining-Requests",
		"x-ratelimit-limit":                           "X-Ratelimit-Limit",
		"x-ratelimit-reset":                           "X-Ratelimit-Reset",
		"cf-ray":                                      "Cf-Ray",
		"x-correlation-id":                            "X-Correlation-Id",
		apiheaders.HeaderTraceID:                      "X-GodarAI-Trace-Id",
		"x-trace-id":                                  "X-Trace-Id",
		"traceparent":                                 "Traceparent",
		"x-shahrah-mock-provider":                     "X-Shahrah-Mock-Provider",
		// Also handle mixed-case input that some providers may send
		"x-ratelimit-remaining": "X-Ratelimit-Remaining",
	}

	for lookupKey, outputKey := range headerMappings {
		if val, ok := normalized[lookupKey]; ok {
			c.Set(outputKey, val)
		}
	}
}

// WriteAPIError writes an API error response in OpenAI format.
func WriteAPIError(c *fiber.Ctx, status int, message string, errType string) error {
	return c.Status(status).JSON(contractsopenai.APIError{
		Error: contractsopenai.ErrorInfo{
			Message: message,
			Type:    errType,
		},
	})
}

func writeAPIErrorDetail(c *fiber.Ctx, status int, message string, errType string, param string, code string) error {
	return c.Status(status).JSON(contractsopenai.APIError{
		Error: contractsopenai.ErrorInfo{
			Message: message,
			Type:    errType,
			Param:   param,
			Code:    code,
		},
	})
}

func writeRoutingDependencyUnavailable(
	c *fiber.Ctx,
	logger *zap.Logger,
	routeErr error,
	model string,
	tenantID string,
) (bool, error) {
	if !errors.Is(routeErr, router.ErrDependencyUnavailable) {
		return false, nil
	}

	emitDenyCounter(logger, denyReasonDependencyUnavailable, model, tenantID)

	return true, writeAPIErrorDetail(
		c,
		fiber.StatusServiceUnavailable,
		"Routing dependency unavailable",
		"dependency_unavailable",
		"",
		"dependency_unavailable",
	)
}

func writeUnsupportedModelModeError(c *fiber.Ctx, model string, requestedMode platform.ModelType, exposedModes []string) error {
	mode := strings.TrimSpace(string(requestedMode))
	model = strings.TrimSpace(model)
	message := fmt.Sprintf("Model %q does not expose requested mode %q. Exposed modes: %s.", model, mode, formatQuotedStringList(exposedModes))

	return c.Status(fiber.StatusBadRequest).JSON(contractsopenai.APIError{
		Error: contractsopenai.ErrorInfo{
			Message: message,
			Type:    "invalid_request_error",
			Param:   "model",
			Code:    "unsupported_model_mode",
			Details: map[string]interface{}{
				"requested_mode": mode,
				"model":          model,
				"exposed_modes":  exposedModes,
				"reason":         "requested mode is not available on this virtual model",
			},
		},
	})
}

func formatQuotedStringList(values []string) string {
	if len(values) == 0 {
		return "[]"
	}

	quoted := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		quoted = append(quoted, fmt.Sprintf("%q", value))
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}

// ContainsModel checks if a model is in the allowed models list.
func ContainsModel(allowed []string, model string) bool {
	for _, m := range allowed {
		if m == "*" || m == model {
			return true
		}
	}

	return false
}
