package responsecache

import (
	"context"
	"testing"
	"time"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	basecache "github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type testEmbedder struct {
	vectors map[string][]float32
	err     error
}

func (e *testEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	if e.err != nil {
		return nil, e.err
	}
	result := make([][]float32, len(texts))
	for i, text := range texts {
		result[i] = e.vectors[text]
	}
	return result, nil
}
func (e *testEmbedder) Identity() EmbeddingIdentity {
	return EmbeddingIdentity{Provider: "arvancloud", Model: "Bge-m3", Revision: "v1", Dimensions: 2}
}

func newTestService(t *testing.T, embedder Embedder) *Service {
	t.Helper()
	store, err := basecache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	service, err := NewService(store, embedder, Config{Enabled: true, DefaultTTL: time.Minute, MinTTL: time.Second, MaxTTL: time.Hour, MaxResponseBytes: 1024, MaxSemanticCandidates: 10})
	require.NoError(t, err)
	return service
}

func TestParseHeaderValidation(t *testing.T) {
	service := newTestService(t, nil)
	cfg, err := service.ParseHeader(`{"type":"smart","namespace":"tenant-1"}`)
	require.NoError(t, err)
	assert.Equal(t, 0.9, cfg.SimilarityThreshold)
	assert.Equal(t, 60, cfg.TTL)
	for _, raw := range []string{
		`{"type":"unknown"}`, `{"type":"simple","similarity_threshold":0.9}`,
		`{"type":"smart","unknown":true}`, `{"type":"smart"} {"type":"smart"}`,
	} {
		_, err := service.ParseHeader(raw)
		assert.Error(t, err, raw)
	}
}

func TestSimpleCacheExactHitAndIsolation(t *testing.T) {
	service := newTestService(t, nil)
	cfg, _ := service.ParseHeader(`{"type":"simple","ttl":60}`)
	req := contractsopenai.ChatCompletionRequest{Model: "gpt-4", Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}}}
	scope := Scope{TenantID: "tenant-1", KeyID: "key-1", Model: "gpt-4"}
	result, pending, err := service.Lookup(context.Background(), scope, req, cfg)
	require.NoError(t, err)
	assert.Equal(t, "miss", result.Status)
	require.NoError(t, service.Store(context.Background(), pending, []byte(`{"id":"cached"}`), "trace-original", "openai", "gpt-4"))
	result, _, err = service.Lookup(context.Background(), scope, req, cfg)
	require.NoError(t, err)
	require.NotNil(t, result.Entry)
	assert.Equal(t, "trace-original", result.Entry.TraceID)
	result, _, err = service.Lookup(context.Background(), Scope{TenantID: "tenant-1", KeyID: "key-2", Model: "gpt-4"}, req, cfg)
	require.NoError(t, err)
	assert.Nil(t, result.Entry)
}

func TestSmartCacheMatchesOnlyFinalUserText(t *testing.T) {
	embedder := &testEmbedder{vectors: map[string][]float32{
		"reset my password": {1, 0}, "password recovery": {0.99, 0.01}, "weather": {0, 1},
	}}
	service := newTestService(t, embedder)
	cfg, _ := service.ParseHeader(`{"type":"smart","ttl":60,"similarity_threshold":0.9}`)
	scope := Scope{TenantID: "tenant-1", KeyID: "key-1", Model: "gpt-4"}
	base := contractsopenai.ChatCompletionRequest{Model: "gpt-4", Messages: []contractsopenai.ChatMessage{{Role: "system", Content: "helpful"}, {Role: "user", Content: "reset my password"}}}
	_, pending, err := service.Lookup(context.Background(), scope, base, cfg)
	require.NoError(t, err)
	require.NoError(t, service.Store(context.Background(), pending, []byte(`{"id":"cached"}`), "trace", "openai", "gpt-4"))
	paraphrase := base
	paraphrase.Messages = append([]contractsopenai.ChatMessage(nil), base.Messages...)
	paraphrase.Messages[1].Content = "password recovery"
	result, _, err := service.Lookup(context.Background(), scope, paraphrase, cfg)
	require.NoError(t, err)
	require.NotNil(t, result.Entry)
	assert.GreaterOrEqual(t, result.Similarity, 0.9)
	unrelated := paraphrase
	unrelated.Messages = append([]contractsopenai.ChatMessage(nil), paraphrase.Messages...)
	unrelated.Messages[1].Content = "weather"
	result, _, err = service.Lookup(context.Background(), scope, unrelated, cfg)
	require.NoError(t, err)
	assert.Nil(t, result.Entry)
}

func TestSmartCacheBypassesUnsafeRequests(t *testing.T) {
	service := newTestService(t, &testEmbedder{})
	cfg, _ := service.ParseHeader(`{"type":"smart"}`)
	req := contractsopenai.ChatCompletionRequest{Model: "gpt-4", Stream: true, Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}}}
	result, _, err := service.Lookup(context.Background(), Scope{TenantID: "t", KeyID: "k", Model: "gpt-4"}, req, cfg)
	require.NoError(t, err)
	assert.Equal(t, "bypass", result.Status)
}

func TestAcquireMissWaitsAndRespectsCancellation(t *testing.T) {
	service := newTestService(t, nil)
	cfg, _ := service.ParseHeader(`{"type":"simple"}`)
	req := contractsopenai.ChatCompletionRequest{Model: "gpt-4", Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}}}
	_, pending, err := service.Lookup(context.Background(), Scope{TenantID: "t", KeyID: "k", Model: "gpt-4"}, req, cfg)
	require.NoError(t, err)
	release, waited, err := service.AcquireMiss(context.Background(), pending)
	require.NoError(t, err)
	assert.False(t, waited)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, waited, err = service.AcquireMiss(ctx, pending)
	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, waited)
	release()
}
