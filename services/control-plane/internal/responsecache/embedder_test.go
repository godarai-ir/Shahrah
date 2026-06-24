package responsecache

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestArvanCloudBGEEmbedder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/embeddings", r.URL.Path)
		assert.Equal(t, "Bearer secret", r.Header.Get("Authorization"))
		var request struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		assert.Equal(t, "Bge-m3", request.Model)
		assert.Equal(t, []string{"one", "two"}, request.Input)
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
