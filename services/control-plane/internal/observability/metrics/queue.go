// Package metrics provides a bounded queue for model metric events.
package metrics

import (
	"sync/atomic"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var metricsQueueDroppedTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "shahrah_model_metrics_queue_dropped_total",
	Help: "Number of model metric events dropped because the in-memory queue was full.",
})

// Queue is a concurrency-safe bounded queue for metric events.
type Queue struct {
	ch        chan platform.ModelMetricEvent
	dropped   atomic.Int64
	enqueued  atomic.Int64
	dequeued  atomic.Int64
	highWater atomic.Int64
}

// NewQueue creates a queue with the given capacity.
func NewQueue(capacity int) *Queue {
	if capacity <= 0 {
		capacity = 8192
	}

	return &Queue{ch: make(chan platform.ModelMetricEvent, capacity)}
}

// Enqueue adds an event to the queue if space is available.
func (q *Queue) Enqueue(ev platform.ModelMetricEvent) bool {
	select {
	case q.ch <- ev:
		q.enqueued.Add(1)
		q.updateHighWater()

		return true
	default:
		q.dropped.Add(1)
		metricsQueueDroppedTotal.Inc()

		return false
	}
}

// DequeueBatch removes up to count events from the queue.
func (q *Queue) DequeueBatch(count int) []platform.ModelMetricEvent {
	if count <= 0 {
		count = 128
	}

	out := make([]platform.ModelMetricEvent, 0, count)
	for i := 0; i < count; i++ {
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
