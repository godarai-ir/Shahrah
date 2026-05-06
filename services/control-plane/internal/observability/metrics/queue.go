package metrics

import (
	"sync/atomic"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

type Queue struct {
	ch        chan platform.ModelMetricEvent
	dropped   atomic.Int64
	enqueued  atomic.Int64
	dequeued  atomic.Int64
	highWater atomic.Int64
}

func NewQueue(capacity int) *Queue {
	if capacity <= 0 {
		capacity = 8192
	}
	return &Queue{ch: make(chan platform.ModelMetricEvent, capacity)}
}

func (q *Queue) Enqueue(ev platform.ModelMetricEvent) bool {
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

func (q *Queue) DequeueBatch(max int) []platform.ModelMetricEvent {
	if max <= 0 {
		max = 128
	}
	out := make([]platform.ModelMetricEvent, 0, max)
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
