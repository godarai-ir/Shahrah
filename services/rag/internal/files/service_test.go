package files

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type shortWriter struct {
	buffer bytes.Buffer
	limit  int
}

func (w *shortWriter) Write(payload []byte) (int, error) {
	return w.buffer.Write(payload[:min(w.limit, len(payload))])
}

type emptyWriter struct{}

func (emptyWriter) Write([]byte) (int, error) { return 0, nil }

func TestWriteFullCompletesPartialWrites(t *testing.T) {
	t.Parallel()

	writer := &shortWriter{limit: 2}
	require.NoError(t, writeFull(writer, []byte("payload")))
	require.Equal(t, "payload", writer.buffer.String())
}

func TestWriteFullRejectsNoProgress(t *testing.T) {
	t.Parallel()

	require.ErrorIs(t, writeFull(emptyWriter{}, []byte("payload")), io.ErrShortWrite)
}
