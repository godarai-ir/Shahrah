package tracing

import (
	"fmt"
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
	LocalModelKey      = "shahrah.trace_model"
	LocalSpanSeqKey    = "shahrah.trace_span_seq"
	LocalQueueKey      = "shahrah.trace_queue"
	defaultUnknownPath = "unknown"
)

func Middleware(queue *Queue, sampler *Sampler, slowMS int64) fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := resolveTraceID(c)
		c.Locals(LocalTraceIDKey, traceID)
		c.Locals(LocalStartTimeKey, time.Now())
		c.Locals(LocalSpanSeqKey, int64(0))
		c.Locals(LocalQueueKey, queue)

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
			model, _ := c.Locals(LocalModelKey).(string)
			spanID := nextSpanID(c, traceID)
			_ = queue.Enqueue(SpanEvent{
				TraceID:      traceID,
				SpanID:       spanID,
				ParentSpanID: "",
				Name:         "gateway.request",
				Kind:         "server",
				TenantID:     tenantID,
				Model:        strings.TrimSpace(model),
				Method:       c.Method(),
				Route:        normalizeRoute(c.Path()),
				StatusCode:   statusCode,
				StartedAt:    start,
				DurationMS:   durationMS,
				Sampled:      shouldSample || statusCode >= 500 || durationMS >= slowMS,
				Attributes: map[string]interface{}{
					"http.method": c.Method(),
					"http.route":  normalizeRoute(c.Path()),
				},
			})
		}

		return err
	}
}

func nextSpanID(c *fiber.Ctx, traceID string) string {
	cur, _ := c.Locals(LocalSpanSeqKey).(int64)
	cur++
	c.Locals(LocalSpanSeqKey, cur)
	return fmt.Sprintf("%s:%d", traceID, cur)
}

func RecordStepSpan(c *fiber.Ctx, queue *Queue, parentSpanID, name, kind string, startedAt time.Time, statusCode int, attrs map[string]interface{}) {
	if c == nil || queue == nil || startedAt.IsZero() {
		return
	}
	traceID := TraceIDFromContext(c)
	if strings.TrimSpace(traceID) == "" {
		return
	}
	tenantID, _ := c.Locals(LocalTenantIDKey).(string)
	if tenantID == "" {
		tenantID = c.Get("x-shahrah-tenant")
	}
	spanID := nextSpanID(c, traceID)
	_ = queue.Enqueue(SpanEvent{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		Name:         name,
		Kind:         kind,
		TenantID:     tenantID,
		Method:       c.Method(),
		Route:        normalizeRoute(c.Path()),
		StatusCode:   statusCode,
		StartedAt:    startedAt,
		DurationMS:   time.Since(startedAt).Milliseconds(),
		Sampled:      true,
		Attributes:   attrs,
	})
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

func SetModelForTrace(c *fiber.Ctx, model string) {
	if c == nil {
		return
	}
	c.Locals(LocalModelKey, strings.TrimSpace(model))
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

func QueueFromContext(c *fiber.Ctx) *Queue {
	if c == nil {
		return nil
	}
	if v := c.Locals(LocalQueueKey); v != nil {
		if q, ok := v.(*Queue); ok {
			return q
		}
	}
	return nil
}
