package server

import (
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const defaultGatewayMaxInFlight = 512

func newGatewayBackpressureMiddleware() fiber.Handler {
	limit := gatewayMaxInFlightFromEnv()
	if limit <= 0 {
		return nil
	}

	sem := make(chan struct{}, limit)
	return func(c *fiber.Ctx) error {
		if isBackpressureExemptPath(c.Path()) {
			return c.Next()
		}

		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			return c.Next()
		default:
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": map[string]interface{}{
					"message": "gateway overloaded",
					"type":    "overload_error",
				},
			})
		}
	}
}

func gatewayMaxInFlightFromEnv() int {
	raw := strings.TrimSpace(os.Getenv("GATEWAY_MAX_IN_FLIGHT_REQUESTS"))
	if raw == "" {
		return defaultGatewayMaxInFlight
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return defaultGatewayMaxInFlight
	}

	return value
}

func isBackpressureExemptPath(path string) bool {
	return path == "/health" ||
		path == "/ready" ||
		path == "/metrics" ||
		strings.HasPrefix(path, "/debug/pprof/")
}
