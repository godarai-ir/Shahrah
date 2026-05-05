package tracing

import (
	"sync/atomic"
	"time"
)

type SpanEvent struct {
	TraceID    string
	TenantID   string
	Method     string
	Route      string
	StatusCode int
	StartedAt  time.Time
	DurationMS int64
	Sampled    bool
}

type Queue struct {
	ch        chan SpanEvent
	dropped   atomic.Int64
	enqueued  atomic.Int64
	dequeued  atomic.Int64
	highWater atomic.Int64
}

func NewQueue(capacity int) *Queue {
	if capacity <= 0 {
		capacity = defaultQueueCap
	}
	return &Queue{ch: make(chan SpanEvent, capacity)}
}

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

func (q *Queue) DequeueBatch(max int) []SpanEvent {
	if max <= 0 {
		max = defaultBatchSize
	}
	out := make([]SpanEvent, 0, max)
	for i := 0; i < max; i++ {
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

func (q *Queue) Len() int { return len(q.ch) }

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
