// Package tracing provides request tracing, privacy, and export configuration.
package tracing

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// TraceIDHeader is the public trace identifier header.
	TraceIDHeader      = apiheaders.HeaderTraceID
	TraceParentHeader  = "traceparent"
	TraceStateHeader   = "tracestate"
	LocalTraceIDKey    = "shahrah.trace_id"
	LocalStartTimeKey  = "shahrah.trace_start"
	LocalTenantIDKey   = "shahrah.trace_tenant_id"
	LocalModelKey      = "shahrah.trace_model"
	LocalSpanSeqKey    = "shahrah.trace_span_seq"
	LocalRootSpanIDKey = "shahrah.trace_root_span_id"
	LocalSampledKey    = "shahrah.trace_sampled"
	LocalStepBufferKey = "shahrah.trace_step_buffer"
	LocalQueueKey      = "shahrah.trace_queue"
	LocalPrivacyKey    = "shahrah.trace_privacy"
	defaultUnknownPath = "unknown"
)

// Middleware installs tracing context and request-span capture.
func Middleware(queue *Queue, sampler *Sampler, slowMS int64, privacy ...*PrivacyManager) fiber.Handler {
	var privacyManager *PrivacyManager
	if len(privacy) > 0 {
		privacyManager = privacy[0]
	}

	if privacyManager == nil {
		cfg := DefaultConfig()
		privacyManager = NewPrivacyManager(cfg)
	}

	return func(c *fiber.Ctx) error {
		traceID := resolveTraceID(c)
		c.Locals(LocalTraceIDKey, traceID)
		c.Locals(LocalStartTimeKey, time.Now())
		c.Locals(LocalSpanSeqKey, int64(0))
		c.Locals(LocalRootSpanIDKey, rootSpanID(traceID))
		c.Locals(LocalSampledKey, false)
		c.Locals(LocalStepBufferKey, []SpanEvent{})
		c.Locals(LocalQueueKey, queue)
		c.Locals(LocalPrivacyKey, privacyManager)

		c.Set(TraceIDHeader, traceID)
		propagateW3CHeaders(c)

		err := c.Next()

		start := time.Time{}
		if startedAt, ok := c.Locals(LocalStartTimeKey).(time.Time); ok {
			start = startedAt
		}

		if start.IsZero() {
			start = time.Now()
		}

		durationMS := time.Since(start).Milliseconds()
		statusCode := c.Response().StatusCode()

		tenantID := resolveTraceTenantID(c)

		shouldSample := sampler.ShouldSample(DecisionInput{
			TenantID:   tenantID,
			RequestID:  traceID,
			StatusCode: statusCode,
			DurationMS: durationMS,
		})

		c.Locals(LocalSampledKey, shouldSample)

		if shouldSample {
			model := ""
			if scopedModel, ok := c.Locals(LocalModelKey).(string); ok {
				model = scopedModel
			}

			root := SpanEvent{
				TraceID:      traceID,
				SpanID:       RootSpanIDFromContext(c),
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
			}
			_ = queue.Enqueue(root)
			flushStepBuffer(c, queue)
		}

		return err
	}
}

func rootSpanID(traceID string) string {
	return fmt.Sprintf("%s:0", traceID)
}

func nextSpanID(c *fiber.Ctx, traceID string) string {
	cur := int64(0)
	if currentSeq, ok := c.Locals(LocalSpanSeqKey).(int64); ok {
		cur = currentSeq
	}

	cur++
	c.Locals(LocalSpanSeqKey, cur)

	return fmt.Sprintf("%s:%d", traceID, cur)
}

// RecordStepSpan records a child span for the active request.
func RecordStepSpan(c *fiber.Ctx, queue *Queue, parentSpanID, name, kind string, startedAt time.Time, statusCode int, attrs map[string]interface{}) {
	if c == nil || queue == nil || startedAt.IsZero() {
		return
	}

	traceID := TraceIDFromContext(c)
	if strings.TrimSpace(traceID) == "" {
		return
	}

	tenantID := resolveTraceTenantID(c)

	if strings.TrimSpace(parentSpanID) == "" {
		parentSpanID = RootSpanIDFromContext(c)
	}

	spanID := nextSpanID(c, traceID)

	appendStepSpan(c, SpanEvent{
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
		Sampled:      false,
		Attributes:   attrs,
	})
}

func appendStepSpan(c *fiber.Ctx, ev SpanEvent) {
	if c == nil {
		return
	}

	buffer := []SpanEvent(nil)
	if currentBuffer, ok := c.Locals(LocalStepBufferKey).([]SpanEvent); ok {
		buffer = currentBuffer
	}

	buffer = append(buffer, ev)
	c.Locals(LocalStepBufferKey, buffer)
}

func flushStepBuffer(c *fiber.Ctx, queue *Queue) {
	if c == nil || queue == nil {
		return
	}

	buffer := []SpanEvent(nil)
	if currentBuffer, ok := c.Locals(LocalStepBufferKey).([]SpanEvent); ok {
		buffer = currentBuffer
	}

	privacy := PrivacyFromContext(c)
	policy := CapturePolicyNone

	if privacy != nil && len(buffer) > 0 {
		policy = privacy.ResolvePolicyForTenant(buffer[0].TenantID)
	}

	for _, ev := range buffer {
		if privacy != nil {
			ev.Attributes = privacy.Apply(ev.TraceID, ev.SpanID, ev.Name, ev.Attributes, policy)
		}

		ev.Sampled = true
		_ = queue.Enqueue(ev)
	}

	c.Locals(LocalStepBufferKey, []SpanEvent{})
}

func resolveTraceID(c *fiber.Ctx) string {
	if v := apiheaders.GetTraceID(c); v != "" {
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

func resolveTraceTenantID(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}

	if scopedTenantID, ok := c.Locals(LocalTenantIDKey).(string); ok {
		if tenantID := strings.TrimSpace(scopedTenantID); tenantID != "" {
			return tenantID
		}
	}

	if currentTenant := tenant.FromContext(c); currentTenant != nil {
		if tenantID := strings.TrimSpace(currentTenant.ID); tenantID != "" {
			return tenantID
		}
	}

	return strings.TrimSpace(apiheaders.GetTenant(c))
}

// TraceIDFromContext returns the active trace identifier.
func TraceIDFromContext(c *fiber.Ctx) string {
	if v := c.Locals(LocalTraceIDKey); v != nil {
		if traceID, ok := v.(string); ok && traceID != "" {
			return traceID
		}
	}

	return strings.TrimSpace(c.Get(TraceIDHeader))
}

// SetTenantForTrace stores the tenant identifier for the current trace.
func SetTenantForTrace(c *fiber.Ctx, tenantID string) {
	c.Locals(LocalTenantIDKey, tenantID)
}

// SetModelForTrace stores the active model for the current trace.
func SetModelForTrace(c *fiber.Ctx, model string) {
	if c == nil {
		return
	}

	c.Locals(LocalModelKey, strings.TrimSpace(model))
}

// DurationMS returns the elapsed request duration in milliseconds.
func DurationMS(c *fiber.Ctx) int64 {
	start := time.Time{}
	if startedAt, ok := c.Locals(LocalStartTimeKey).(time.Time); ok {
		start = startedAt
	}

	if start.IsZero() {
		return 0
	}

	return time.Since(start).Milliseconds()
}

// StatusCodeString returns the response status code as text.
func StatusCodeString(c *fiber.Ctx) string {
	return strconv.Itoa(c.Response().StatusCode())
}

// QueueFromContext returns the tracing queue stored on the request.
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

// RootSpanIDFromContext returns the root span id for the request.
func RootSpanIDFromContext(c *fiber.Ctx) string {
	if c == nil {
		return ""
	}

	if v := c.Locals(LocalRootSpanIDKey); v != nil {
		if root, ok := v.(string); ok && strings.TrimSpace(root) != "" {
			return root
		}
	}

	traceID := TraceIDFromContext(c)
	if strings.TrimSpace(traceID) == "" {
		return ""
	}

	return rootSpanID(traceID)
}

// PrivacyFromContext returns the privacy manager stored on the request.
func PrivacyFromContext(c *fiber.Ctx) *PrivacyManager {
	if c == nil {
		return nil
	}

	if v := c.Locals(LocalPrivacyKey); v != nil {
		if p, ok := v.(*PrivacyManager); ok {
			return p
		}
	}

	return nil
}
