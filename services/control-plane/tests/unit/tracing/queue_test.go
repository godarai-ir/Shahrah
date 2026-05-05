package tracing_test

import (
	"testing"

	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/stretchr/testify/assert"
)

func TestQueueDropsWhenFull(t *testing.T) {
	q := tracing.NewQueue(1)
	ok1 := q.Enqueue(tracing.SpanEvent{TraceID: "1"})
	ok2 := q.Enqueue(tracing.SpanEvent{TraceID: "2"})
	assert.True(t, ok1)
	assert.False(t, ok2)

	m := q.Metrics()
	assert.EqualValues(t, 1, m["dropped"])
	assert.EqualValues(t, 1, m["depth"])
}

func TestQueueDequeueBatch(t *testing.T) {
	q := tracing.NewQueue(10)
	_ = q.Enqueue(tracing.SpanEvent{TraceID: "1"})
	_ = q.Enqueue(tracing.SpanEvent{TraceID: "2"})
	b := q.DequeueBatch(10)
	assert.Len(t, b, 2)
	assert.Equal(t, "1", b[0].TraceID)
	assert.Equal(t, "2", b[1].TraceID)
}
