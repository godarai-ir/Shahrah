package proxy

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type spendLogSink struct {
	mu      sync.Mutex
	singles int
	batches [][]platform.SpendLog
	block   chan struct{}
	err     error
}

func (s *spendLogSink) LogSpend(context.Context, platform.SpendLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.singles++
	return s.err
}

func (s *spendLogSink) LogSpends(_ context.Context, logs []platform.SpendLog) error {
	if s.block != nil {
		<-s.block
	}
	copied := append([]platform.SpendLog(nil), logs...)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batches = append(s.batches, copied)
	return s.err
}

func (s *spendLogSink) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := s.singles
	for _, batch := range s.batches {
		total += len(batch)
	}
	return total
}

type spendLogMetricsSink struct {
	mu            sync.Mutex
	queued        int
	dropped       int
	flushes       int
	failures      int
	flushBatches  []int
	flushDuration []time.Duration
}

func (m *spendLogMetricsSink) IncSpendLogQueued() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queued++
}

func (m *spendLogMetricsSink) IncSpendLogDropped() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dropped++
}

func (m *spendLogMetricsSink) ObserveSpendLogFlush(batchSize int, duration time.Duration, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flushes++
	m.flushBatches = append(m.flushBatches, batchSize)
	m.flushDuration = append(m.flushDuration, duration)
	if err != nil {
		m.failures++
	}
}

func (m *spendLogMetricsSink) snapshot() (queued, dropped, flushes, failures int, batches []int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.queued, m.dropped, m.flushes, m.failures, append([]int(nil), m.flushBatches...)
}

func TestAsyncSpendLogWriterEnqueuesWithoutSynchronousWrite(t *testing.T) {
	sink := &spendLogSink{}
	metrics := &spendLogMetricsSink{}
	writer := NewAsyncSpendLogWriter(sink, AsyncSpendLogConfig{
		QueueSize:     8,
		Workers:       1,
		BatchSize:     10,
		FlushInterval: time.Hour,
		DropOnFull:    true,
	}, zap.NewNop(), metrics)

	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-1"}))
	require.Equal(t, 0, sink.count())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	require.NoError(t, writer.Stop(ctx))
	require.Equal(t, 1, sink.count())

	queued, dropped, flushes, failures, batches := metrics.snapshot()
	require.Equal(t, 1, queued)
	require.Equal(t, 0, dropped)
	require.Equal(t, 1, flushes)
	require.Equal(t, 0, failures)
	require.Equal(t, []int{1}, batches)
}

func TestAsyncSpendLogWriterFlushesByBatchSize(t *testing.T) {
	sink := &spendLogSink{}
	metrics := &spendLogMetricsSink{}
	writer := NewAsyncSpendLogWriter(sink, AsyncSpendLogConfig{
		QueueSize:     8,
		Workers:       1,
		BatchSize:     2,
		FlushInterval: time.Hour,
		DropOnFull:    true,
	}, zap.NewNop(), metrics)
	defer writer.Stop(context.Background())

	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-1"}))
	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-2"}))

	require.Eventually(t, func() bool {
		return sink.count() == 2
	}, time.Second, 10*time.Millisecond)

	queued, _, flushes, failures, batches := metrics.snapshot()
	require.Equal(t, 2, queued)
	require.Equal(t, 1, flushes)
	require.Equal(t, 0, failures)
	require.Equal(t, []int{2}, batches)
}

func TestAsyncSpendLogWriterFlushesByInterval(t *testing.T) {
	sink := &spendLogSink{}
	writer := NewAsyncSpendLogWriter(sink, AsyncSpendLogConfig{
		QueueSize:     8,
		Workers:       1,
		BatchSize:     10,
		FlushInterval: 10 * time.Millisecond,
		DropOnFull:    true,
	}, zap.NewNop(), nil)
	defer writer.Stop(context.Background())

	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-1"}))

	require.Eventually(t, func() bool {
		return sink.count() == 1
	}, time.Second, 10*time.Millisecond)
}

func TestAsyncSpendLogWriterQueueFullBehaviorIsNonBlocking(t *testing.T) {
	block := make(chan struct{})
	sink := &spendLogSink{block: block}
	metrics := &spendLogMetricsSink{}
	writer := NewAsyncSpendLogWriter(sink, AsyncSpendLogConfig{
		QueueSize:     1,
		Workers:       1,
		BatchSize:     1,
		FlushInterval: time.Hour,
		DropOnFull:    false,
	}, zap.NewNop(), metrics)
	defer func() {
		close(block)
		_ = writer.Stop(context.Background())
	}()

	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-1"}))
	require.Eventually(t, func() bool {
		return len(writer.queue) == 0
	}, time.Second, 10*time.Millisecond)
	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-2"}))

	start := time.Now()
	require.ErrorIs(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-3"}), ErrSpendLogQueueFull)
	require.Less(t, time.Since(start), 100*time.Millisecond)

	queued, dropped, _, _, _ := metrics.snapshot()
	require.Equal(t, 2, queued)
	require.Equal(t, 1, dropped)
}

func TestAsyncSpendLogWriterDropOnFullSuppressesQueueFullError(t *testing.T) {
	block := make(chan struct{})
	sink := &spendLogSink{block: block}
	writer := NewAsyncSpendLogWriter(sink, AsyncSpendLogConfig{
		QueueSize:     1,
		Workers:       1,
		BatchSize:     1,
		FlushInterval: time.Hour,
		DropOnFull:    true,
	}, zap.NewNop(), nil)
	defer func() {
		close(block)
		_ = writer.Stop(context.Background())
	}()

	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-1"}))
	require.Eventually(t, func() bool {
		return len(writer.queue) == 0
	}, time.Second, 10*time.Millisecond)
	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-2"}))
	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-3"}))
}

func TestAsyncSpendLogWriterRecordsFlushFailures(t *testing.T) {
	writeErr := errors.New("db unavailable")
	sink := &spendLogSink{err: writeErr}
	metrics := &spendLogMetricsSink{}
	writer := NewAsyncSpendLogWriter(sink, AsyncSpendLogConfig{
		QueueSize:     8,
		Workers:       1,
		BatchSize:     1,
		FlushInterval: time.Hour,
		DropOnFull:    true,
	}, zap.NewNop(), metrics)
	defer writer.Stop(context.Background())

	require.NoError(t, writer.LogSpend(context.Background(), platform.SpendLog{RequestID: "req-1"}))

	require.Eventually(t, func() bool {
		_, _, _, failures, _ := metrics.snapshot()
		return failures == 1
	}, time.Second, 10*time.Millisecond)
}
