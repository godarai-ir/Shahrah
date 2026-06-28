// Package tracing provides request tracing, privacy, and export configuration.
package tracing

import (
	"sync/atomic"
	"time"
)

// SpanEvent is the internal tracing event payload.
type SpanEvent struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
	Name         string
	Kind         string
	TenantID     string
	Model        string
	Method       string
	Route        string
	StatusCode   int
	StartedAt    time.Time
	DurationMS   int64
	Sampled      bool
	ErrorCode    string
	Attributes   map[string]interface{}
}

// Queue buffers span events before export.
type Queue struct {
	ch        chan SpanEvent
	dropped   atomic.Int64
	enqueued  atomic.Int64
	dequeued  atomic.Int64
	highWater atomic.Int64
}

// NewQueue creates a tracing queue with the given capacity.
func NewQueue(capacity int) *Queue {
	if capacity <= 0 {
		capacity = defaultQueueCap
	}

	return &Queue{ch: make(chan SpanEvent, capacity)}
}

// Enqueue appends a span event if capacity is available.
func (q *Queue) Enqueue(ev SpanEvent) bool {
	select {
	case q.ch <- ev:
		q.enqueued.Add(1)
		q.updateHighWater()

		return true
	default:
		q.dropped.Add(1)
		return false
	}
}

func (q *Queue) DequeueBatch(limit int) []SpanEvent {
	if limit <= 0 {
		limit = defaultBatchSize
	}

	out := make([]SpanEvent, 0, limit)
	for i := 0; i < limit; i++ {
		select {
		case ev := <-q.ch:
			out = append(out, ev)

			q.dequeued.Add(1)
		default:
			return out
		}
	}

	return out
}

// Len returns the current queue depth.
func (q *Queue) Len() int { return len(q.ch) }

// Metrics returns queue counters for observability.
func (q *Queue) Metrics() map[string]int64 {
	return map[string]int64{
		"enqueued":   q.enqueued.Load(),
		"dequeued":   q.dequeued.Load(),
		"dropped":    q.dropped.Load(),
		"high_water": q.highWater.Load(),
		"depth":      int64(len(q.ch)),
	}
}

func (q *Queue) updateHighWater() {
	cur := int64(len(q.ch))
	for {
		prev := q.highWater.Load()
		if cur <= prev {
			return
		}

		if q.highWater.CompareAndSwap(prev, cur) {
			return
		}
	}
}
