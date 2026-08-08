package responsecache

import (
	"context"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
	"github.com/Delpak/shahrah/packages/contracts/platform"
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

	store, err := testcache.New(t, zap.NewNop())
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
	req := contractsopenai.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []contractsopenai.ChatMessage{{Role: "user", Content: "hello"}},
		Metadata: map[string]interface{}{"source": "tenant_admin_playground", "client_run_id": "run-1", "comparison_id": "comparison-1"},
	}
	scope := Scope{TenantID: "tenant-1", KeyID: "key-1", Model: "gpt-4"}
	result, pending, err := service.Lookup(context.Background(), scope, req, cfg)
	require.NoError(t, err)
	assert.Equal(t, "miss", result.Status)
	require.NoError(t, service.Store(context.Background(), pending, []byte(`{"id":"cached"}`), "trace-original", "openai", "gpt-4", platform.UsageCostBreakdown{InputTokens: 10, OutputTokens: 5, TotalCostUSD: 0.01, TotalCostIRR: 8500, ExchangeRate: 850000}, "test-pricing-v1"))

	secondReq := req
	secondReq.Metadata = map[string]interface{}{"source": "tenant_admin_playground", "client_run_id": "run-2", "comparison_id": "comparison-2"}
	result, _, err = service.Lookup(context.Background(), scope, secondReq, cfg)
	require.NoError(t, err)
	require.NotNil(t, result.Entry)
	assert.Equal(t, "run-1", req.Metadata["client_run_id"], "fingerprinting must not mutate request metadata")
	assert.Equal(t, "trace-original", result.Entry.TraceID)
	assert.Equal(t, 0.01, result.Entry.CostSnapshot.CostUSD)
	assert.Equal(t, "test-pricing-v1", result.Entry.CostSnapshot.PricingVersion)
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
	require.NoError(t, service.Store(context.Background(), pending, []byte(`{"id":"cached"}`), "trace", "openai", "gpt-4", platform.UsageCostBreakdown{TotalCostUSD: 0.01}, "test-pricing-v1"))

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

func TestResponsesSimpleCachePreservesUnknownFieldsInFingerprint(t *testing.T) {
	service := newTestService(t, nil)
	cfg, err := service.ParseHeader(`{"type":"simple","ttl":60}`)
	require.NoError(t, err)

	req := contractsopenai.ResponseRequest{Model: "gpt-4", Input: "hello"}
	scope := Scope{TenantID: "tenant-1", KeyID: "key-1", Model: "gpt-4"}
	firstBody := []byte(`{"model":"gpt-4","input":"hello","future_field":{"mode":"strict"}}`)
	result, pending, err := service.LookupResponse(context.Background(), scope, req, firstBody, cfg)
	require.NoError(t, err)
	assert.Equal(t, "miss", result.Status)
	require.NoError(t, service.Store(
		context.Background(),
		pending,
		[]byte(`{"id":"resp_cached"}`),
		"trace-original",
		"openai",
		"gpt-4",
		platform.UsageCostBreakdown{TotalCostUSD: 0.01},
		"test-pricing-v1",
	))

	reorderedBody := []byte(`{"future_field":{"mode":"strict"},"input":"hello","model":"gpt-4"}`)
	result, _, err = service.LookupResponse(context.Background(), scope, req, reorderedBody, cfg)
	require.NoError(t, err)
	require.NotNil(t, result.Entry)
	assert.Equal(t, "trace-original", result.Entry.TraceID)

	changedBody := []byte(`{"model":"gpt-4","input":"hello","future_field":{"mode":"relaxed"}}`)
	result, _, err = service.LookupResponse(context.Background(), scope, req, changedBody, cfg)
	require.NoError(t, err)
	assert.Nil(t, result.Entry)
}

func TestResponsesSmartCacheMatchesOnlyFinalUserText(t *testing.T) {
	embedder := &testEmbedder{vectors: map[string][]float32{
		"reset my password": {1, 0},
		"password recovery": {0.99, 0.01},
	}}
	service := newTestService(t, embedder)
	cfg, err := service.ParseHeader(`{"type":"smart","ttl":60,"similarity_threshold":0.9}`)
	require.NoError(t, err)

	scope := Scope{TenantID: "tenant-1", KeyID: "key-1", Model: "gpt-4"}
	baseReq := contractsopenai.ResponseRequest{Model: "gpt-4", Input: "reset my password", Instructions: "helpful"}
	baseBody := []byte(`{"model":"gpt-4","input":"reset my password","instructions":"helpful"}`)
	_, pending, err := service.LookupResponse(context.Background(), scope, baseReq, baseBody, cfg)
	require.NoError(t, err)
	require.NoError(t, service.Store(
		context.Background(),
		pending,
		[]byte(`{"id":"resp_cached"}`),
		"trace-original",
		"openai",
		"gpt-4",
		platform.UsageCostBreakdown{TotalCostUSD: 0.01},
		"test-pricing-v1",
	))

	paraphraseReq := baseReq
	paraphraseReq.Input = "password recovery"
	paraphraseBody := []byte(`{"model":"gpt-4","input":"password recovery","instructions":"helpful"}`)
	result, _, err := service.LookupResponse(context.Background(), scope, paraphraseReq, paraphraseBody, cfg)
	require.NoError(t, err)
	require.NotNil(t, result.Entry)
	assert.True(t, result.Semantic)
	assert.GreaterOrEqual(t, result.Similarity, 0.9)

	changedInstructions := []byte(`{"model":"gpt-4","input":"password recovery","instructions":"different"}`)
	result, _, err = service.LookupResponse(context.Background(), scope, paraphraseReq, changedInstructions, cfg)
	require.NoError(t, err)
	assert.Nil(t, result.Entry)
}

func TestResponsesSmartCacheSupportsTextMessagesAndBypassesMultimodalInput(t *testing.T) {
	embedder := &testEmbedder{vectors: map[string][]float32{
		"reset my password": {1, 0},
		"password recovery": {0.99, 0.01},
	}}
	service := newTestService(t, embedder)
	cfg, err := service.ParseHeader(`{"type":"smart","ttl":60}`)
	require.NoError(t, err)

	scope := Scope{TenantID: "tenant-1", KeyID: "key-1", Model: "gpt-4"}
	baseBody := []byte(`{"model":"gpt-4","input":[{"role":"user","content":[{"type":"input_text","text":"reset my password"}]}]}`)
	baseReq := contractsopenai.ResponseRequest{Model: "gpt-4", Input: []interface{}{
		map[string]interface{}{"role": "user", "content": []interface{}{map[string]interface{}{"type": "input_text", "text": "reset my password"}}},
	}}
	_, pending, err := service.LookupResponse(context.Background(), scope, baseReq, baseBody, cfg)
	require.NoError(t, err)
	require.NoError(t, service.Store(
		context.Background(),
		pending,
		[]byte(`{"id":"resp_cached"}`),
		"trace-original",
		"openai",
		"gpt-4",
		platform.UsageCostBreakdown{},
		"test-pricing-v1",
	))

	paraphraseBody := []byte(`{"model":"gpt-4","input":[{"role":"user","content":[{"type":"input_text","text":"password recovery"}]}]}`)
	result, _, err := service.LookupResponse(context.Background(), scope, baseReq, paraphraseBody, cfg)
	require.NoError(t, err)
	require.NotNil(t, result.Entry)
	assert.True(t, result.Semantic)

	multimodalBody := []byte(`{"model":"gpt-4","input":[{"role":"user","content":[{"type":"input_text","text":"password recovery"},{"type":"input_image","image_url":"https://example.com/a.png"}]}]}`)
	result, _, err = service.LookupResponse(context.Background(), scope, baseReq, multimodalBody, cfg)
	require.NoError(t, err)
	assert.Equal(t, "bypass", result.Status)
}

func TestResponsesCacheBypassesRequestsThatUseProviderState(t *testing.T) {
	service := newTestService(t, &testEmbedder{})
	scope := Scope{TenantID: "tenant-1", KeyID: "key-1", Model: "gpt-4"}

	tests := []struct {
		name string
		req  contractsopenai.ResponseRequest
		body string
	}{
		{
			name: "previous response",
			req: contractsopenai.ResponseRequest{
				Model:              "gpt-4",
				Input:              "continue",
				PreviousResponseID: "resp_previous",
			},
			body: `{"model":"gpt-4","input":"continue","previous_response_id":"resp_previous"}`,
		},
		{
			name: "background mode",
			req:  contractsopenai.ResponseRequest{Model: "gpt-4", Input: "hello"},
			body: `{"model":"gpt-4","input":"hello","background":true}`,
		},
		{
			name: "conversation ID",
			req:  contractsopenai.ResponseRequest{Model: "gpt-4", Input: "hello"},
			body: `{"model":"gpt-4","input":"hello","conversation":"conv_1"}`,
		},
		{
			name: "conversation object",
			req:  contractsopenai.ResponseRequest{Model: "gpt-4", Input: "hello"},
			body: `{"model":"gpt-4","input":"hello","conversation":{"id":"conv_1"}}`,
		},
	}

	for _, cacheType := range []string{"simple", "smart"} {
		t.Run(cacheType, func(t *testing.T) {
			cfg, err := service.ParseHeader(`{"type":"` + cacheType + `","ttl":60}`)
			require.NoError(t, err)

			for _, test := range tests {
				t.Run(test.name, func(t *testing.T) {
					result, pending, lookupErr := service.LookupResponse(
						context.Background(),
						scope,
						test.req,
						[]byte(test.body),
						cfg,
					)
					require.NoError(t, lookupErr)
					assert.Equal(t, "bypass", result.Status)
					assert.Nil(t, pending)
				})
			}
		})
	}
}

func TestResponsesCacheAllowsDisabledBackgroundMode(t *testing.T) {
	service := newTestService(t, nil)
	cfg, err := service.ParseHeader(`{"type":"simple","ttl":60}`)
	require.NoError(t, err)

	result, pending, err := service.LookupResponse(
		context.Background(),
		Scope{TenantID: "tenant-1", KeyID: "key-1", Model: "gpt-4"},
		contractsopenai.ResponseRequest{Model: "gpt-4", Input: "hello"},
		[]byte(`{"model":"gpt-4","input":"hello","background":false}`),
		cfg,
	)
	require.NoError(t, err)
	assert.Equal(t, "miss", result.Status)
	assert.NotNil(t, pending)
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
