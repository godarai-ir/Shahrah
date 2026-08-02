package responsecache

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestArvanCloudBGEEmbedder(t *testing.T) {
	handlerResult := make(chan error, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var handlerErr error
		defer func() { handlerResult <- handlerErr }()

		if r.URL.Path != "/v1/embeddings" {
			handlerErr = fmt.Errorf("unexpected request path %q", r.URL.Path)
			http.Error(w, handlerErr.Error(), http.StatusBadRequest)

			return
		}

		if authorization := r.Header.Get("Authorization"); authorization != "Bearer secret" {
			handlerErr = fmt.Errorf("unexpected authorization header %q", authorization)
			http.Error(w, handlerErr.Error(), http.StatusBadRequest)

			return
		}

		var request struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			handlerErr = fmt.Errorf("decode embedding request: %w", err)
			http.Error(w, handlerErr.Error(), http.StatusBadRequest)

			return
		}

		if request.Model != "Bge-m3" {
			handlerErr = fmt.Errorf("unexpected model %q", request.Model)
			http.Error(w, handlerErr.Error(), http.StatusBadRequest)

			return
		}

		if !slices.Equal([]string{"one", "two"}, request.Input) {
			handlerErr = fmt.Errorf("unexpected input %#v", request.Input)
			http.Error(w, handlerErr.Error(), http.StatusBadRequest)

			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{
			map[string]interface{}{"index": 1, "embedding": []float32{0, 1}},
			map[string]interface{}{"index": 0, "embedding": []float32{1, 0}},
		}})
	}))
	defer server.Close()

	embedder, err := NewArvanCloudBGEEmbedder(server.Client(), ArvanCloudConfig{
		BaseURL: server.URL + "/v1/", APIKey: "secret", Model: "Bge-m3", Revision: "v1", Dimensions: 2,
	})
	require.NoError(t, err)
	vectors, err := embedder.Embed(context.Background(), []string{"one", "two"})
	require.NoError(t, err)
	require.NoError(t, <-handlerResult)
	assert.Equal(t, [][]float32{{1, 0}, {0, 1}}, vectors)
}

func TestArvanCloudBGEEmbedderRejectsDimensionMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"index":0,"embedding":[1]}]}`))
	}))
	defer server.Close()

	embedder, err := NewArvanCloudBGEEmbedder(server.Client(), ArvanCloudConfig{BaseURL: server.URL, APIKey: "secret", Model: "Bge-m3", Dimensions: 2})
	require.NoError(t, err)
	_, err = embedder.Embed(context.Background(), []string{"text"})
	assert.ErrorContains(t, err, "unexpected dimensions")
}

func TestArvanCloudBGEEmbedderRedactsCredentialBearingURLFromErrors(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, assert.AnError
	})}
	embedder, err := NewArvanCloudBGEEmbedder(client, ArvanCloudConfig{
		BaseURL: "https://example.test/secret-path/v1", APIKey: "secret", Model: "Bge-m3", Dimensions: 2,
	})
	require.NoError(t, err)
	_, err = embedder.Embed(context.Background(), []string{"text"})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "secret-path")
	assert.NotContains(t, err.Error(), "secret")
}

func TestCosineSimilarity(t *testing.T) {
	score, err := CosineSimilarity([]float32{1, 1}, []float32{2, 2})
	require.NoError(t, err)
	assert.InDelta(t, 1, score, 1e-9)
	score, err = CosineSimilarity([]float32{1, 0}, []float32{0, 1})
	require.NoError(t, err)
	assert.InDelta(t, 0, score, 1e-9)

	_, err = CosineSimilarity([]float32{float32(math.Inf(1))}, []float32{1})
	assert.Error(t, err)
}
