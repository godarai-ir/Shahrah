package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddingColumnsRoutesSupportedDimensions(t *testing.T) {
	t.Parallel()

	embedding1536, embedding1024, err := embeddingColumns(make([]float32, 1536))
	require.NoError(t, err)
	require.NotNil(t, embedding1536)
	require.Nil(t, embedding1024)

	embedding1536, embedding1024, err = embeddingColumns(make([]float32, 1024))
	require.NoError(t, err)
	require.Nil(t, embedding1536)
	require.NotNil(t, embedding1024)
}

func TestEmbeddingColumnsRejectsUnknownDimensions(t *testing.T) {
	t.Parallel()

	_, _, err := embeddingColumns(make([]float32, 768))
	require.ErrorContains(t, err, "unsupported RAG embedding dimensions")
}
