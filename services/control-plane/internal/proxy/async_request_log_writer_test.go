package proxy

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type requestLogSink struct {
	mu      sync.Mutex
	singles int
	batches [][]platform.RequestLogEntry
	block   chan struct{}
	err     error
}

func (s *requestLogSink) LogRequest(context.Context, platform.RequestLogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.singles++

	return s.err
}

func (s *requestLogSink) LogRequests(_ context.Context, entries []platform.RequestLogEntry) error {
	if s.block != nil {
		<-s.block
	}

	copied := append([]platform.RequestLogEntry(nil), entries...)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.batches = append(s.batches, copied)

	return s.err
}

type requestLogMetricsSink struct {
	mu       sync.Mutex
	queued   int
	dropped  int
	flushes  int
	failures int
	batches  []int
}

func (m *requestLogMetricsSink) IncRequestLogQueued() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.queued++
}

func (m *requestLogMetricsSink) IncRequestLogDropped() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.dropped++
}

func (m *requestLogMetricsSink) ObserveRequestLogFlush(batchSize int, _ time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.flushes++
	m.batches = append(m.batches, batchSize)

	if err != nil {
		m.failures++
	}
}

func (m *requestLogMetricsSink) snapshot() (queued, dropped, flushes, failures int, batches []int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.queued, m.dropped, m.flushes, m.failures, append([]int(nil), m.batches...)
}

func (s *requestLogSink) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := s.singles
	for _, batch := range s.batches {
		total += len(batch)
	}

	return total
}

func TestAsyncRequestLogWriterEnqueuesWithoutSynchronousWrite(t *testing.T) {
	sink := &requestLogSink{}
	writer := NewAsyncRequestLogWriter(sink, AsyncRequestLogConfig{
		QueueSize:     8,
		Workers:       1,
		BatchSize:     10,
		FlushInterval: time.Hour,
		DropOnFull:    true,
	}, zap.NewNop())

	require.NoError(t, writer.LogRequest(context.Background(), platform.RequestLogEntry{RequestID: "req-1"}))
	require.Equal(t, 0, sink.count())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, writer.Stop(ctx))
	require.Equal(t, 1, sink.count())
}

func TestAsyncRequestLogWriterFlushesByBatchSize(t *testing.T) {
	sink := &requestLogSink{}

	writer := NewAsyncRequestLogWriter(sink, AsyncRequestLogConfig{
		QueueSize:     8,
		Workers:       1,
		BatchSize:     2,
		FlushInterval: time.Hour,
		DropOnFull:    true,
	}, zap.NewNop())
	defer writer.Stop(context.Background())

	require.NoError(t, writer.LogRequest(context.Background(), platform.RequestLogEntry{RequestID: "req-1"}))
	require.NoError(t, writer.LogRequest(context.Background(), platform.RequestLogEntry{RequestID: "req-2"}))

	require.Eventually(t, func() bool {
		return sink.count() == 2
	}, time.Second, 10*time.Millisecond)
}

func TestAsyncRequestLogWriterFlushesByInterval(t *testing.T) {
	sink := &requestLogSink{}

	writer := NewAsyncRequestLogWriter(sink, AsyncRequestLogConfig{
		QueueSize:     8,
		Workers:       1,
		BatchSize:     10,
		FlushInterval: 10 * time.Millisecond,
		DropOnFull:    true,
	}, zap.NewNop())
	defer writer.Stop(context.Background())

	require.NoError(t, writer.LogRequest(context.Background(), platform.RequestLogEntry{RequestID: "req-1"}))

	require.Eventually(t, func() bool {
		return sink.count() == 1
	}, time.Second, 10*time.Millisecond)
}

func TestAsyncRequestLogWriterQueueFullBehavior(t *testing.T) {
	block := make(chan struct{})
	sink := &requestLogSink{block: block}
	writer := NewAsyncRequestLogWriter(sink, AsyncRequestLogConfig{
		QueueSize:     1,
		Workers:       1,
		BatchSize:     1,
		FlushInterval: time.Hour,
		DropOnFull:    false,
	}, zap.NewNop())

	defer func() {
		close(block)

		_ = writer.Stop(context.Background())
	}()

	require.NoError(t, writer.LogRequest(context.Background(), platform.RequestLogEntry{RequestID: "req-1"}))
	require.Eventually(t, func() bool {
		return len(writer.queue) == 0
	}, time.Second, 10*time.Millisecond)
	require.NoError(t, writer.LogRequest(context.Background(), platform.RequestLogEntry{RequestID: "req-2"}))
	require.ErrorIs(t, writer.LogRequest(context.Background(), platform.RequestLogEntry{RequestID: "req-3"}), ErrRequestLogQueueFull)
}

func TestAsyncRequestLogWriterReportsQueueAndFlushMetrics(t *testing.T) {
	sink := &requestLogSink{err: errors.New("storage unavailable")}
	metrics := &requestLogMetricsSink{}
	writer := NewManagedAsyncRequestLogWriter(sink, AsyncRequestLogConfig{}, zap.NewNop(), metrics)

	writer.flushBatch(t.Context(), []platform.RequestLogEntry{{RequestID: "req-1"}})

	queued, dropped, flushes, failures, batches := metrics.snapshot()
	require.Equal(t, 0, queued)
	require.Equal(t, 0, dropped)
	require.Equal(t, 1, flushes)
	require.Equal(t, 1, failures)
	require.Equal(t, []int{1}, batches)
}

func TestAsyncRequestLogWriterReportsDroppedEntries(t *testing.T) {
	block := make(chan struct{})
	sink := &requestLogSink{block: block}
	metrics := &requestLogMetricsSink{}
	writer := NewAsyncRequestLogWriter(sink, AsyncRequestLogConfig{
		QueueSize:     1,
		Workers:       1,
		BatchSize:     1,
		FlushInterval: time.Hour,
		DropOnFull:    false,
	}, zap.NewNop(), metrics)

	defer func() {
		close(block)
		require.NoError(t, writer.Stop(t.Context()))
	}()

	require.NoError(t, writer.LogRequest(t.Context(), platform.RequestLogEntry{RequestID: "req-1"}))
	require.Eventually(t, func() bool { return len(writer.queue) == 0 }, time.Second, 10*time.Millisecond)
	require.NoError(t, writer.LogRequest(t.Context(), platform.RequestLogEntry{RequestID: "req-2"}))
	require.ErrorIs(t, writer.LogRequest(t.Context(), platform.RequestLogEntry{RequestID: "req-3"}), ErrRequestLogQueueFull)

	queued, dropped, _, _, _ := metrics.snapshot()
	require.Equal(t, 2, queued)
	require.Equal(t, 1, dropped)
}

func TestAsyncRequestLogWriterMaterializesLazyPayloadForGenericSink(t *testing.T) {
	sink := &requestLogSink{}
	writer := NewManagedAsyncRequestLogWriter(sink, AsyncRequestLogConfig{}, zap.NewNop())
	entries := []platform.RequestLogEntry{{
		RequestBody: buildJSONRequestLogPayload(
			[]byte(`{"model":"original","messages":[]}`),
			map[string]interface{}{"model": "override"},
		),
	}}

	writer.flushBatch(t.Context(), entries)

	sink.mu.Lock()
	defer sink.mu.Unlock()

	require.Len(t, sink.batches, 1)
	require.Equal(t, "override", sink.batches[0][0].RequestBody["model"])
	require.NotContains(t, sink.batches[0][0].RequestBody, observability.LazyJSONPayloadField)
}

func TestAsyncRequestLogWriterStopIsSafeDuringConcurrentEnqueue(t *testing.T) {
	sink := &requestLogSink{}
	writer := NewAsyncRequestLogWriter(sink, AsyncRequestLogConfig{
		QueueSize:     256,
		Workers:       1,
		BatchSize:     32,
		FlushInterval: time.Hour,
		DropOnFull:    true,
	}, zap.NewNop())

	var producers sync.WaitGroup

	start := make(chan struct{})

	for index := 0; index < 32; index++ {
		producers.Add(1)
		go func() {
			defer producers.Done()

			<-start

			for request := 0; request < 100; request++ {
				_ = writer.LogRequest(t.Context(), platform.RequestLogEntry{RequestID: "request"})
			}
		}()
	}

	close(start)
	require.NoError(t, writer.Stop(t.Context()))
	producers.Wait()
}
