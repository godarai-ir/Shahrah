package responsecache

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArvanCloudBGEEmbedderIntegration(t *testing.T) {
	if os.Getenv("SHAHRAH_SMART_CACHE_EMBEDDING_INTEGRATION") != "1" {
		t.Skip("set SHAHRAH_SMART_CACHE_EMBEDDING_INTEGRATION=1 to call the real BGE-M3 endpoint")
	}

	baseURL := os.Getenv("SHAHRAH_SMART_CACHE_EMBEDDING_BASE_URL")

	apiKey := os.Getenv("SHAHRAH_SMART_CACHE_EMBEDDING_API_KEY")
	if baseURL == "" || apiKey == "" {
		t.Fatal("embedding base URL and API key are required for integration testing")
	}

	dimensions := 1024

	if raw := os.Getenv("SHAHRAH_SMART_CACHE_EMBEDDING_DIMENSIONS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		require.NoError(t, err)

		dimensions = parsed
	}

	embedder, err := NewArvanCloudBGEEmbedder(&http.Client{Timeout: 20 * time.Second}, ArvanCloudConfig{
		BaseURL: baseURL, APIKey: apiKey, Model: "Bge-m3", Revision: "v1", Dimensions: dimensions, MaxBatchSize: 3,
	})
	require.NoError(t, err)
	vectors, err := embedder.Embed(context.Background(), []string{
		"چگونه رمز عبور خود را بازیابی کنم؟",
		"روش بازنشانی گذرواژه چیست؟",
		"وضعیت آب و هوای تهران چگونه است؟",
	})
	require.NoError(t, err)
	require.Len(t, vectors, 3)

	for _, vector := range vectors {
		assert.Len(t, vector, dimensions)
	}

	similar, err := CosineSimilarity(vectors[0], vectors[1])
	require.NoError(t, err)
	unrelated, err := CosineSimilarity(vectors[0], vectors[2])
	require.NoError(t, err)
	assert.Greater(t, similar, unrelated)
}
