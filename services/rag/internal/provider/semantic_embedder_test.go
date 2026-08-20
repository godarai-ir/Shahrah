package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSemanticEmbedderUsesBatchesAndPreservesOrder(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		calls.Add(1)
		require.Equal(t, "Bearer secret", request.Header.Get("Authorization"))

		var payload struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		require.Equal(t, "Bge-m3", payload.Model)

		data := make([]map[string]any, len(payload.Input))
		for index := range payload.Input {
			vector := make([]float32, 1024)
			vector[0] = float32(index + 1)
			data[index] = map[string]any{"index": index, "embedding": vector}
		}

		require.NoError(t, json.NewEncoder(writer).Encode(map[string]any{
			"data": data, "model": "Bge-m3", "usage": map[string]int{"total_tokens": len(payload.Input)},
		}))
	}))
	t.Cleanup(server.Close)

	embedder, err := newSemanticEmbedder(SemanticEmbeddingConfig{
		BaseURL: server.URL, APIKey: "secret", Model: "Bge-m3", Revision: "v1",
		Dimensions: 1024, Timeout: time.Second, MaxBatchSize: 2,
	})
	require.NoError(t, err)

	vectors, tokens, err := embedder.embed(
		context.Background(), "Bge-m3", []string{"one", "two", "three"}, 1024,
	)
	require.NoError(t, err)
	require.Len(t, vectors, 3)
	require.Equal(t, 3, tokens)
	require.EqualValues(t, 2, calls.Load())
}

func TestSemanticEmbedderRejectsProfileIdentityMismatch(t *testing.T) {
	t.Parallel()

	embedder, err := newSemanticEmbedder(SemanticEmbeddingConfig{
		BaseURL: "https://example.com/v1", APIKey: "secret", Model: "Bge-m3", Revision: "v1",
		Dimensions: 1024, Timeout: time.Second, MaxBatchSize: 16,
	})
	require.NoError(t, err)

	_, _, err = embedder.embed(context.Background(), "another-model", []string{"text"}, 1536)
	require.ErrorContains(t, err, "does not match Semantic Cache identity")
}
