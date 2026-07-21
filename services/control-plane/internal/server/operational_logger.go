package server

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type operationalLogEvent struct {
	StartedAt       time.Time
	Latency         time.Duration
	IP              string
	Method          string
	Path            string
	Status          int
	BytesIn         int
	BytesOut        int
	TraceID         string
	UserAgent       string
	Error           string
	RequestHeaders  map[string][]string
	ResponseHeaders map[string][]string
}

func newOperationalLoggerMiddleware(settings config.OperationalLogSettings, logger *zap.Logger) fiber.Handler {
	if !settings.Enabled {
		return nil
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	settings = normalizeOperationalLogSettings(settings)
	queue := make(chan operationalLogEvent, settings.QueueSize)

	var successCounter atomic.Uint64

	startOperationalLogWorker(queue, logger)

	return func(c *fiber.Ctx) error {
		startedAt := time.Now().UTC()
		err := c.Next()
		latency := time.Since(startedAt)

		status := c.Response().StatusCode()
		if status == 0 {
			status = fiber.StatusOK
		}

		if err != nil {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				status = fiberErr.Code
			} else if status < fiber.StatusBadRequest {
				status = fiber.StatusInternalServerError
			}
		}

		if !shouldLogOperationalRequest(settings, status, latency, err != nil, &successCounter) {
			return err
		}

		event := operationalLogEvent{
			StartedAt: startedAt,
			Latency:   latency,
			IP:        c.IP(),
			Method:    c.Method(),
			Path:      c.Path(),
			Status:    status,
			BytesIn:   len(c.Request().Body()),
			BytesOut:  len(c.Response().Body()),
			TraceID:   c.GetRespHeader(tracing.TraceIDHeader),
			UserAgent: sanitizeLogValue(c.Get(fiber.HeaderUserAgent)),
		}
		if err != nil {
			event.Error = sanitizeLogValue(err.Error())
		} else if ctxErr := c.Context().Err(); ctxErr != nil {
			event.Error = sanitizeLogValue(ctxErr.Error())
		}

		if settings.IncludeHeaders {
			event.RequestHeaders = requestHeaders(c)
			event.ResponseHeaders = responseHeaders(c)
		}

		select {
		case queue <- event:
		default:
			logger.Debug("operational access log dropped because queue is full")
		}

		return err
	}
}

func normalizeOperationalLogSettings(settings config.OperationalLogSettings) config.OperationalLogSettings {
	if settings.QueueSize <= 0 {
		settings.QueueSize = 4096
	}

	if settings.SuccessSampleRate < 0 {
		settings.SuccessSampleRate = 0
	}

	if settings.SuccessSampleRate > 1 {
		settings.SuccessSampleRate = 1
	}

	if settings.SlowRequestMS < 0 {
		settings.SlowRequestMS = 0
	}

	return settings
}

func shouldLogOperationalRequest(settings config.OperationalLogSettings, status int, latency time.Duration, hasError bool, successCounter *atomic.Uint64) bool {
	if hasError || status >= fiber.StatusInternalServerError {
		return true
	}

	if settings.LogClientErrors && status >= fiber.StatusBadRequest {
		return true
	}

	if settings.SlowRequestMS > 0 && latency >= time.Duration(settings.SlowRequestMS)*time.Millisecond {
		return true
	}

	return shouldSampleSuccess(settings.SuccessSampleRate, successCounter)
}

func shouldSampleSuccess(sampleRate float64, counter *atomic.Uint64) bool {
	switch {
	case sampleRate <= 0:
		return false
	case sampleRate >= 1:
		return true
	default:
		every := uint64(1 / sampleRate)
		if every == 0 {
			every = 1
		}

		return counter.Add(1)%every == 0
	}
}

func startOperationalLogWorker(queue <-chan operationalLogEvent, logger *zap.Logger) {
	go func() {
		for event := range queue {
			fields := []zap.Field{
				zap.Time("started_at", event.StartedAt),
				zap.Int64("latency_ms", event.Latency.Milliseconds()),
				zap.String("ip", event.IP),
				zap.String("method", event.Method),
				zap.String("path", event.Path),
				zap.Int("status", event.Status),
				zap.Int("bytes_in", event.BytesIn),
				zap.Int("bytes_out", event.BytesOut),
				zap.String("trace_id", event.TraceID),
				zap.String("user_agent", event.UserAgent),
			}
			if event.Error != "" {
				fields = append(fields, zap.String("error", event.Error))
			}

			if len(event.RequestHeaders) > 0 {
				fields = append(fields, zap.String("request_headers", formatHeaderLogValue(event.RequestHeaders)))
			}

			if len(event.ResponseHeaders) > 0 {
				fields = append(fields, zap.String("response_headers", formatHeaderLogValue(event.ResponseHeaders)))
			}

			if event.Status >= fiber.StatusInternalServerError {
				logger.Warn("http request failed", fields...)
				continue
			}

			logger.Info("http request", fields...)
		}
	}()
}
