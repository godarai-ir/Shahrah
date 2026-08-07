package server

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

const defaultGatewayMaxInFlight = 512

type backpressureMetrics interface {
	IncBackpressureInFlight()
	DecBackpressureInFlight()
	IncBackpressureRejected()
}

func newGatewayBackpressureMiddleware(limit int, metrics ...backpressureMetrics) fiber.Handler {
	if limit <= 0 {
		return nil
	}

	var recorder backpressureMetrics
	if len(metrics) > 0 {
		recorder = metrics[0]
	}

	sem := make(chan struct{}, limit)

	return func(c *fiber.Ctx) error {
		if !isGatewayBackpressurePath(c.Path()) || isBackpressureExemptPath(c.Path()) {
			return c.Next()
		}

		select {
		case sem <- struct{}{}:
			if recorder != nil {
				recorder.IncBackpressureInFlight()
			}

			defer func() {
				<-sem

				if recorder != nil {
					recorder.DecBackpressureInFlight()
				}
			}()

			return c.Next()
		default:
			if recorder != nil {
				recorder.IncBackpressureRejected()
			}

			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": map[string]interface{}{
					"message": "gateway overloaded",
					"type":    "overload_error",
				},
			})
		}
	}
}

func isGatewayBackpressurePath(path string) bool {
	return path == "/v1" || strings.HasPrefix(path, "/v1/")
}

func isBackpressureExemptPath(path string) bool {
	return path == "/health" ||
		path == "/ready" ||
		path == "/metrics" ||
		strings.HasPrefix(path, "/debug/pprof/")
}
