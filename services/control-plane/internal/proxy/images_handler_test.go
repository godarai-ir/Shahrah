package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
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
	var b budgeter = &testBudgeter{}
	if strings.EqualFold(provider, "xai") {
		b = &xaiImageBudgeter{}
	}

	return newImagesTestAppWithBudgeter(t, modelName, provider, deploymentModel, tunnel, b, logs, allowedTypes, emitMetric...)
}

type xaiImageBudgeter struct{ testBudgeter }

func (b *xaiImageBudgeter) FinalizeImageUsage(ctx context.Context, model string, usage *contractsopenai.Usage, providerCostUSDTicks *int64, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	finalized, err := b.testBudgeter.FinalizeImageUsage(ctx, model, usage, providerCostUSDTicks, estimated)
	if err != nil {
		return finalized, err
	}

	binding := xaiImageTestRuntimeProfile()
	finalized.ModelRuntimeProfile = &binding
	finalized.PricingRevision = binding.CatalogRevision

	return finalized, nil
}

func xaiImageTestRuntimeProfile() platform.ModelRuntimeProfileBinding {
	return platform.ModelRuntimeProfileBinding{
		Provider:        "xai",
		CatalogRevision: "xai-image-test-revision",
		CanonicalID:     "xai/grok-imagine-image",
		Slug:            "grok-imagine-image",
		ProfileVersion:  3,
		EffectiveFrom:   time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version:    3,
			Aliases:    []string{"grok-imagine-image-2026-03-02"},
			Endpoints:  []string{"image_generations", "image_edits", "files", "file_public_urls"},
			Transports: []string{"http"},
			Lifecycle:  "active",
			Thinking: platform.ModelRuntimeThinkingProfile{
				DefaultMode:  "disabled",
				AllowedModes: []string{"disabled"},
			},
			Sampling: platform.ModelRuntimeSamplingProfile{},
			Tools:    platform.ModelRuntimeToolsProfile{},
			Modalities: platform.ModelRuntimeModalities{
				Input:  []string{"text", "image"},
				Output: []string{"image"},
			},
			RequestControls: platform.ModelRuntimeRequestControls{
				ServiceTierValues:  []string{"standard", "batch"},
				InferenceGeoValues: []string{"global"},
			},
			ImageControls: platform.ModelRuntimeImageControls{
				AspectRatios:           []string{"1:1", "16:9", "auto"},
				Resolutions:            []string{"1k", "2k"},
				ResponseFormats:        []string{"url", "b64_json"},
				MaximumOutputs:         10,
				MaximumReferenceImages: 3,
				EditTransport:          "json",
				FilesInput:             true,
				StorageOptions:         true,
			},
			UsageAccounting: platform.ModelRuntimeUsageAccounting{
				RequiredResponseFields: []string{"cost_in_usd_ticks"},
				ExactProviderCost:      true,
			},
			Limits: platform.ModelRuntimeLimits{
				MaxInputCharacters: 1024,
				MessagesBytes:      32 * 1024 * 1024,
				BatchBytes:         25 * 1024 * 1024,
				FileBytes:          48 * 1024 * 1024,
			},
		},
	}
}

func TestXAIImageFileOutputsAreRecordedForLaterFilesOperations(t *testing.T) {
	resources := &responseOperationResources{}
	h := &ImagesHandler{resources: resources}
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		h.recordXAIImageFileOutputs(c, "tenant_a", platform.Deployment{
			ID: "dep_xai", Provider: "xai", ProviderAccountID: "acct_xai",
		}, []byte(`{"data":[{"url":"https://imgen.x.ai/temp","file_output":{"file_id":"file_generated","filename":"result.png","expires_at":1780000000,"public_url":"https://files-cdn.x.ai/token/result.png","public_url_expires_at":1779990000}}]}`))

		return c.SendStatus(fiber.StatusOK)
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", http.NoBody))

	require.NoError(t, err)
	defer closeResponseBody(t, resp)

	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	mapping, ok := resources.mappings["file_generated"]
	require.True(t, ok)
	assert.Equal(t, "xai", mapping.Provider)
	assert.Equal(t, "acct_xai", mapping.ProviderAccountID)
	assert.Equal(t, "result.png", mapping.Metadata["filename"])
	assert.Equal(t, "https://files-cdn.x.ai/token/result.png", mapping.Metadata["public_url"])
	require.NotNil(t, mapping.ExpiresAt)
	assert.Equal(t, int64(1780000000), mapping.ExpiresAt.Unix())
}

func newImagesTestAppWithBudgeter(t *testing.T, modelName string, provider string, deploymentModel string, tunnel *captureTunnel, b budgeter, logs *testLogs, allowedTypes []string, emitMetric ...func(input metricsobs.BuildInput)) (*fiber.App, *tracing.Queue) {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
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

	h := NewImagesHandler(rt, tunnel, b, logs, zap.NewNop(), emitMetric...)
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
	require.Len(t, logs.requestEntries(), 1)
	assert.Equal(t, "openai:acct:gpt-image-1", logs.requestEntries()[0].RequestBody["model"])
	assert.Equal(t, "paint a bridge", logs.requestEntries()[0].RequestBody["prompt"])
	assert.Equal(t, "generations", logs.requestEntries()[0].RequestBody["operation"])
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

func TestFillImageJSONMetadataCountsAllXAIEditReferences(t *testing.T) {
	metadata := map[string]interface{}{}
	fillImageJSONMetadata(metadata, map[string]interface{}{
		"image": map[string]interface{}{"file_id": "file_one"},
		"images": []interface{}{
			map[string]interface{}{"url": "https://example.test/two.png"},
			map[string]interface{}{"file_id": "file_three"},
		},
	}, imageOperation{Name: "edits"})
	assert.Equal(t, 3, metadata["image_count"])
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
	require.Len(t, tunnel.lastReq.ModelRuntimeProfiles, 1)
	assert.Equal(t, "xai", tunnel.lastReq.ModelRuntimeProfiles[0].Provider)
}

func TestImagesHandlerValidatesXAIImageControls(t *testing.T) {
	for name, test := range map[string]struct {
		route       string
		body        string
		wantMessage string
	}{
		"output count": {
			route:       "/v1/images/generations",
			body:        `{"model":"grok-imagine-image","prompt":"paint","n":11}`,
			wantMessage: "n exceeds model limit",
		},
		"aspect ratio": {
			route:       "/v1/images/generations",
			body:        `{"model":"grok-imagine-image","prompt":"paint","aspect_ratio":"7:5"}`,
			wantMessage: "unsupported aspect_ratio",
		},
		"resolution": {
			route:       "/v1/images/generations",
			body:        `{"model":"grok-imagine-image","prompt":"paint","resolution":"4k"}`,
			wantMessage: "unsupported resolution",
		},
		"reference count": {
			route:       "/v1/images/edits",
			body:        `{"model":"grok-imagine-image","prompt":"paint","image":["a","b","c","d"]}`,
			wantMessage: "reference images",
		},
	} {
		t.Run(name, func(t *testing.T) {
			tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Body: []byte(`{}`)}}
			app, _ := newImagesTestApp(t, "grok-imagine-image", "xai", "grok-imagine-image", tunnel, nil, nil)
			req := httptest.NewRequest("POST", test.route, bytes.NewBufferString(test.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
			received, readErr := io.ReadAll(resp.Body)
			require.NoError(t, readErr)
			assert.Contains(t, string(received), test.wantMessage)
			assert.Empty(t, tunnel.lastReq.RequestID)
		})
	}
}

func TestImagesHandlerRejectsUnsupportedXAIImageSurfaces(t *testing.T) {
	for name, test := range map[string]struct {
		route       string
		contentType string
		body        []byte
		wantMessage string
	}{
		"variations": {
			route:       "/v1/images/variations",
			contentType: "application/json",
			body:        []byte(`{"model":"grok-imagine-image","image":"ref"}`),
			wantMessage: "do not support image variations",
		},
		"multipart edits": {
			route:       "/v1/images/edits",
			contentType: "multipart/form-data; boundary=test",
			body: []byte(
				"--test\r\nContent-Disposition: form-data; name=\"model\"\r\n\r\ngrok-imagine-image\r\n" +
					"--test\r\nContent-Disposition: form-data; name=\"prompt\"\r\n\r\nedit\r\n" +
					"--test\r\nContent-Disposition: form-data; name=\"image\"; filename=\"a.png\"\r\n" +
					"Content-Type: image/png\r\n\r\na\r\n--test--\r\n",
			),
			wantMessage: "require application/json",
		},
	} {
		t.Run(name, func(t *testing.T) {
			tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Body: []byte(`{}`)}}
			app, _ := newImagesTestApp(t, "grok-imagine-image", "xai", "grok-imagine-image", tunnel, nil, nil)
			req := httptest.NewRequest("POST", test.route, bytes.NewReader(test.body))
			req.Header.Set("Content-Type", test.contentType)

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
			received, readErr := io.ReadAll(resp.Body)
			require.NoError(t, readErr)
			assert.Contains(t, string(received), test.wantMessage)
			assert.Empty(t, tunnel.lastReq.RequestID)
		})
	}
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

func TestImageUsageCostTicksPreservesInt64Precision(t *testing.T) {
	const expected int64 = 9_007_199_254_740_993

	ticks, ok := imageUsageCostTicks([]byte(`{"usage":{"cost_in_usd_ticks":9007199254740993}}`))
	require.True(t, ok)
	require.Equal(t, expected, ticks)
}
