package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newImagesTestApp(t *testing.T, modelName string, provider string, deploymentModel string, tunnel *captureTunnel, logs *testLogs, allowedTypes []string, emitMetric ...func(input metricsobs.BuildInput)) (*fiber.App, *tracing.Queue) {
	return newImagesTestAppWithBudgeter(t, modelName, provider, deploymentModel, tunnel, &testBudgeter{}, logs, allowedTypes, emitMetric...)
}

func newImagesTestAppWithBudgeter(t *testing.T, modelName string, provider string, deploymentModel string, tunnel *captureTunnel, b budgeter, logs *testLogs, allowedTypes []string, emitMetric ...func(input metricsobs.BuildInput)) (*fiber.App, *tracing.Queue) {
	t.Helper()

	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter([]config.ModelDeployment{{
		ModelName: modelName,
		ProviderEndpoint: config.ProviderEndpointConfig{
			Provider: provider,
			Model:    deploymentModel,
			APIKey:   "test-key",
		},
	}}, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	if logs == nil {
		logs = &testLogs{}
	}

	if b == nil {
		b = &testBudgeter{}
	}

	h := NewImagesHandler(rt, tunnel, b, &testWallets{}, logs, zap.NewNop(), emitMetric...)
	q := tracing.NewQueue(64)
	s := tracing.NewSampler(1.0, 1500)

	if allowedTypes == nil {
		allowedTypes = []string{"image_generation"}
	}

	app := fiber.New()
	app.Use(tracing.Middleware(q, s, 1500))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "team_admin",
			TokenHash: "hash_1",
			AllowedModelTypes: map[string][]string{
				modelName: allowedTypes,
			},
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: "tenant_a", Slug: "a", Status: "active"})

		return c.Next()
	})
	app.Post("/v1/images/generations", h.HandleGenerations)
	app.Post("/v1/images/edits", h.HandleEdits)
	app.Post("/v1/images/variations", h.HandleVariations)

	return app, q
}

type rateLimitCountingBudgeter struct {
	testBudgeter
	calls               int
	lastEstimatedTokens int
}

func (b *rateLimitCountingBudgeter) CheckRateLimit(_ context.Context, _ *platform.AuthContext, _ string, estimatedTokens int) (*cache.RateLimitResult, error) {
	b.calls++
	b.lastEstimatedTokens = estimatedTokens

	return &cache.RateLimitResult{Allowed: true, Current: 1, Limit: 100, ResetAt: time.Now().Add(time.Minute)}, nil
}

func TestImagesHandlerJSONGeneration(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"created":1,"data":[{"b64_json":"AAAA","revised_prompt":"safe","mime_type":"image/png"}],"usage":{"input_tokens":12,"output_tokens":272,"total_tokens":284,"cost_in_usd_ticks":12345}}`),
	}}
	logs := &testLogs{}
	app, _ := newImagesTestApp(t, "openai:acct:gpt-image-1", "openai", "gpt-image-1", tunnel, logs, nil)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewBufferString(`{"model":"openai:acct:gpt-image-1","prompt":"paint a bridge","n":2,"size":"1024x1024","quality":"high"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, fiber.MIMEApplicationJSON, resp.Header.Get(fiber.HeaderContentType))
	assert.Equal(t, "/images/generations", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])
	assert.Equal(t, "application/json", tunnel.lastReq.Headers["Content-Type"])

	var outbound map[string]interface{}
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	assert.Equal(t, "gpt-image-1", outbound["model"])
	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, 12, logs.loggedSpends[0].InputTokens)
	assert.Equal(t, 272, logs.loggedSpends[0].OutputTokens)
	assert.Equal(t, 284, logs.loggedSpends[0].TotalTokens)
	assert.InDelta(t, 0.0000012345, logs.loggedSpends[0].CostUSD, 1e-12)
	assert.Equal(t, int64(12345), logs.loggedSpends[0].RequestMetadata["provider_cost_in_usd_ticks"])
}

func TestImagesHandlerGenerationRequiresPrompt(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newImagesTestApp(t, "gpt-image-1", "openai", "gpt-image-1", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewBufferString(`{"model":"gpt-image-1"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), `"param":"prompt"`)
}

func TestEstimatedImageUsageUsesGPTImage2TokenTable(t *testing.T) {
	usage := estimatedImageUsage("openai:default:gpt-image-2", map[string]interface{}{
		"prompt_length":  20,
		"size":           "1024x1024",
		"quality":        "medium",
		"partial_images": 2,
	})

	require.NotNil(t, usage)
	assert.Equal(t, 5, usage.PromptTokens)
	assert.Equal(t, 1967, usage.CompletionTokens)
	assert.Equal(t, 1972, usage.TotalTokens)
}

func TestEstimatedImageUsageIncludesEditInputImages(t *testing.T) {
	usage := estimatedImageUsage("openai:default:gpt-image-2", map[string]interface{}{
		"prompt_length": 20,
		"image_count":   2,
		"size":          "1024x1024",
		"quality":       "low",
	})

	require.NotNil(t, usage)
	assert.Equal(t, 3077, usage.PromptTokens)
	assert.Equal(t, 196, usage.CompletionTokens)
	assert.Equal(t, 3273, usage.TotalTokens)
}

func TestEstimatedImageUsageUsesDallE2VariationUnits(t *testing.T) {
	usage := estimatedImageUsage("openai:default:dall-e-2", map[string]interface{}{
		"image_count": 1,
		"n":           2,
		"size":        "512x512",
	})

	require.NotNil(t, usage)
	assert.Equal(t, 1124, usage.PromptTokens)
	assert.Equal(t, 1800, usage.CompletionTokens)
	assert.Equal(t, 2924, usage.TotalTokens)
}

func TestImagesHandlerJSONEditRejectsEmptyImage(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newImagesTestApp(t, "gpt-image-1", "openai", "gpt-image-1", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/edits", bytes.NewBufferString(`{"model":"gpt-image-1","prompt":"edit this","image":""}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), `"param":"image"`)
}

func TestImagesHandlerJSONEdit(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{"created":1,"data":[{"b64_json":"AAAA","mime_type":"image/png"}]}`)}}
	app, _ := newImagesTestApp(t, "grok-imagine-image", "xai", "grok-imagine-image", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/edits", bytes.NewBufferString(`{"model":"grok-imagine-image","prompt":"make it brighter","image":{"url":"https://example.com/private/path/image.png","type":"image_url"},"output_format":"webp","output_compression":80,"moderation":"low"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "/images/edits", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])
	assert.Equal(t, "image_url", tunnel.lastReq.RequestMetadata["image_source_type"])
	assert.Equal(t, "webp", tunnel.lastReq.RequestMetadata["output_format"])
	assert.Equal(t, 80, tunnel.lastReq.RequestMetadata["output_compression"])
	assert.Equal(t, "low", tunnel.lastReq.RequestMetadata["moderation"])
}

func TestImagesHandlerMultipartEdit(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{"model": "openai:acct:gpt-image-1", "prompt": "edit this", "size": "1024x1024", "input_fidelity": "high"}, map[string][]byte{"image[]": []byte("fake-image-bytes"), "mask": []byte("fake-mask-bytes")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{"created":1,"data":[{"url":"https://cdn.example/edit.png"}]}`)}}
	app, _ := newImagesTestApp(t, "openai:acct:gpt-image-1", "openai", "gpt-image-1", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/edits", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.True(t, strings.HasPrefix(tunnel.lastReq.Headers["Content-Type"], "multipart/form-data"))
	outboundForm := readMultipartForm(t, tunnel.lastReq.Headers["Content-Type"], tunnel.lastReq.Body)
	assert.Equal(t, []string{"gpt-image-1"}, outboundForm.Value["model"])
	assert.Equal(t, []string{"edit this"}, outboundForm.Value["prompt"])
	assertMultipartFileBytes(t, outboundForm, "image[]", []byte("fake-image-bytes"))
	assertMultipartFileBytes(t, outboundForm, "mask", []byte("fake-mask-bytes"))
	assert.Equal(t, 2, tunnel.lastReq.RequestMetadata["file_count"])
	assert.Equal(t, 1, tunnel.lastReq.RequestMetadata["image_count"])
	assert.Equal(t, true, tunnel.lastReq.RequestMetadata["mask_present"])
	assert.Equal(t, "high", tunnel.lastReq.RequestMetadata["input_fidelity"])
}

func TestImagesHandlerMultipartVariation(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{"model": "openai:acct:dall-e-2", "n": "1", "response_format": "url"}, map[string][]byte{"image": []byte("fake-image-bytes")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{"created":1,"data":[{"url":"https://cdn.example/variant.png"}]}`)}}
	app, _ := newImagesTestApp(t, "openai:acct:dall-e-2", "openai", "dall-e-2", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/variations", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "/images/variations", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])
	assert.True(t, strings.HasPrefix(tunnel.lastReq.Headers["Content-Type"], "multipart/form-data"))
	outboundForm := readMultipartForm(t, tunnel.lastReq.Headers["Content-Type"], tunnel.lastReq.Body)
	assert.Equal(t, []string{"dall-e-2"}, outboundForm.Value["model"])
	assertMultipartFileBytes(t, outboundForm, "image", []byte("fake-image-bytes"))
	assert.Equal(t, 1, tunnel.lastReq.RequestMetadata["image_count"])
	assert.Equal(t, "url", tunnel.lastReq.RequestMetadata["response_format"])
}

func TestImagesHandlerVariationEstimateUsesDeploymentModel(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{"model": "openai:acct:legacy-variation", "n": "1"}, map[string][]byte{"image": []byte("fake-image-bytes")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{"created":1,"data":[{"url":"https://cdn.example/variant.png"}]}`)}}
	budgeter := &rateLimitCountingBudgeter{}
	app, _ := newImagesTestAppWithBudgeter(t, "openai:acct:legacy-variation", "openai", "dall-e-2", tunnel, budgeter, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/variations", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, budgeter.calls)
	assert.Equal(t, 2124, budgeter.lastEstimatedTokens)
	outboundForm := readMultipartForm(t, tunnel.lastReq.Headers["Content-Type"], tunnel.lastReq.Body)
	assert.Equal(t, []string{"dall-e-2"}, outboundForm.Value["model"])
}

func TestImagesHandlerOpenAIVariationRejectsGPTImageModel(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{"model": "openai:acct:gpt-image-1", "n": "1"}, map[string][]byte{"image": []byte("fake-image-bytes")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	budgeter := &rateLimitCountingBudgeter{}
	app, _ := newImagesTestAppWithBudgeter(t, "openai:acct:gpt-image-1", "openai", "gpt-image-1", tunnel, budgeter, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/variations", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Empty(t, tunnel.lastReq.RequestID)
	assert.Zero(t, budgeter.calls)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "dall-e-2")
}

func TestImagesHandlerMissingModel(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newImagesTestApp(t, "gpt-image-1", "openai", "gpt-image-1", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewBufferString(`{"prompt":"missing model"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestImagesHandlerDeniedModelType(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newImagesTestApp(t, "gpt-4o-mini", "openai", "gpt-4o-mini", tunnel, nil, []string{"chat"})

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewBufferString(`{"model":"gpt-4o-mini","prompt":"no"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestImagesHandlerProviderErrorPassthrough(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusBadRequest,
		Headers:    map[string]string{"content-type": "application/json", "x-provider": "xai"},
		Body:       []byte(`{"error":{"message":"bad image","type":"invalid_request_error"}}`),
	}}
	app, _ := newImagesTestApp(t, "grok-imagine-image", "xai", "grok-imagine-image", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/generations", bytes.NewBufferString(`{"model":"grok-imagine-image","prompt":"x"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "bad image")
}

func TestImagesHandlerTraceRedaction(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"created":1,"data":[{"b64_json":"VERYSECRETBASE64","url":"https://private.example/full/path.png"}]}`),
	}}
	app, q := newImagesTestApp(t, "grok-imagine-image", "xai", "grok-imagine-image", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/images/edits", bytes.NewBufferString(`{"model":"grok-imagine-image","prompt":"secret prompt","image":{"url":"https://private.example/secret.png","type":"image_url"}}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	events := q.DequeueBatch(20)
	for _, ev := range events {
		raw, _ := json.Marshal(ev.Attributes)
		text := string(raw)
		assert.NotContains(t, text, "VERYSECRETBASE64")
		assert.NotContains(t, text, "https://private.example")
		assert.NotContains(t, text, "secret prompt")
	}
}

func buildMultipartBody(t *testing.T, fields map[string]string, files map[string][]byte) ([]byte, string) {
	t.Helper()

	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		require.NoError(t, w.WriteField(k, v))
	}

	for name, data := range files {
		part, err := w.CreateFormFile(name, name+".png")
		require.NoError(t, err)
		_, err = part.Write(data)
		require.NoError(t, err)
	}

	require.NoError(t, w.Close())

	if !strings.Contains(w.FormDataContentType(), "boundary=") {
		t.Fatalf("missing multipart boundary")
	}

	return buf.Bytes(), w.FormDataContentType()
}

func assertMultipartFileBytes(t *testing.T, form *multipart.Form, field string, expected []byte) {
	t.Helper()

	require.NotNil(t, form)
	files := form.File[field]
	require.Len(t, files, 1, "multipart file field %q", field)
	assert.Equal(t, expected, readMultipartFile(t, files[0]))
}

func readMultipartForm(t *testing.T, contentType string, body []byte) *multipart.Form {
	t.Helper()

	_, params, err := mime.ParseMediaType(contentType)
	require.NoError(t, err)

	boundary := params["boundary"]
	require.NotEmpty(t, boundary)

	form, err := multipart.NewReader(bytes.NewReader(body), boundary).ReadForm(int64(len(body)))
	require.NoError(t, err)

	return form
}

func readMultipartFile(t *testing.T, header *multipart.FileHeader) []byte {
	t.Helper()

	file, err := header.Open()
	require.NoError(t, err)

	defer file.Close()

	data, err := io.ReadAll(file)
	require.NoError(t, err)

	return data
}
