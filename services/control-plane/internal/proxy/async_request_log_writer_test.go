package proxy

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type requestLogSink struct {
	mu      sync.Mutex
	singles int
	batches [][]platform.RequestLogEntry
	block   chan struct{}
}

func (s *requestLogSink) LogRequest(context.Context, platform.RequestLogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.singles++

	return nil
}

func (s *requestLogSink) LogRequests(_ context.Context, entries []platform.RequestLogEntry) error {
	if s.block != nil {
		<-s.block
	}

	copied := append([]platform.RequestLogEntry(nil), entries...)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.batches = append(s.batches, copied)

	return nil
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
