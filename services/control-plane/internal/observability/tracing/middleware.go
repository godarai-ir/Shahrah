package tracing

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	TraceIDHeader      = "x-shahrah-trace-id"
	TraceParentHeader  = "traceparent"
	TraceStateHeader   = "tracestate"
	LocalTraceIDKey    = "shahrah.trace_id"
	LocalStartTimeKey  = "shahrah.trace_start"
	LocalTenantIDKey   = "shahrah.trace_tenant_id"
	defaultUnknownPath = "unknown"
)

func Middleware(queue *Queue, sampler *Sampler, slowMS int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := resolveTraceID(c)
		c.Locals(LocalTraceIDKey, traceID)
		c.Locals(LocalStartTimeKey, time.Now())

		c.Set(TraceIDHeader, traceID)
		propagateW3CHeaders(c)

		err := c.Next()

		start, _ := c.Locals(LocalStartTimeKey).(time.Time)
		if start.IsZero() {
			start = time.Now()
		}
		durationMS := time.Since(start).Milliseconds()
		statusCode := c.Response().StatusCode()
		tenantID, _ := c.Locals(LocalTenantIDKey).(string)
		if tenantID == "" {
			tenantID = c.Get("x-shahrah-tenant")
		}

		shouldSample := sampler.ShouldSample(DecisionInput{
			TenantID:   tenantID,
			RequestID:  traceID,
			StatusCode: statusCode,
			DurationMS: durationMS,
		})

		if shouldSample {
			_ = queue.Enqueue(SpanEvent{
				TraceID:    traceID,
				TenantID:   tenantID,
				Method:     c.Method(),
				Route:      normalizeRoute(c.Path()),
				StatusCode: statusCode,
				StartedAt:  start,
				DurationMS: durationMS,
				Sampled:    shouldSample || statusCode >= 500 || durationMS >= slowMS,
			})
		}

		return err
	}
}

func resolveTraceID(c *fiber.Ctx) string {
	if v := strings.TrimSpace(c.Get(TraceIDHeader)); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.Get("x-trace-id")); v != "" {
		return v
	}
	if tp := strings.TrimSpace(c.Get(TraceParentHeader)); tp != "" {
		parts := strings.Split(tp, "-")
		if len(parts) >= 2 && parts[1] != "" {
			return parts[1]
		}
	}
	return "trace_" + strings.ReplaceAll(uuid.NewString(), "-", "")
}

func propagateW3CHeaders(c *fiber.Ctx) {
	if tp := strings.TrimSpace(c.Get(TraceParentHeader)); tp != "" {
		c.Set(TraceParentHeader, tp)
	}
	if ts := strings.TrimSpace(c.Get(TraceStateHeader)); ts != "" {
		c.Set(TraceStateHeader, ts)
	}
}

func normalizeRoute(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return defaultUnknownPath
	}
	return path
}

func TraceIDFromContext(c *fiber.Ctx) string {
	if v := c.Locals(LocalTraceIDKey); v != nil {
		if traceID, ok := v.(string); ok && traceID != "" {
			return traceID
		}
	}
	return strings.TrimSpace(c.Get(TraceIDHeader))
}

func SetTenantForTrace(c *fiber.Ctx, tenantID string) {
	c.Locals(LocalTenantIDKey, tenantID)
}

func DurationMS(c *fiber.Ctx) int64 {
	start, _ := c.Locals(LocalStartTimeKey).(time.Time)
	if start.IsZero() {
		return 0
	}
	return time.Since(start).Milliseconds()
}

func StatusCodeString(c *fiber.Ctx) string {
	return strconv.Itoa(c.Response().StatusCode())
}
