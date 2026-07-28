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
	billingservice "github.com/Delpak/shahrah/services/control-plane/internal/billing"
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

type imageBillingTestStore struct {
	testWallets
	wallet *platform.Wallet
	rate   *platform.ExchangeRate
}

func (s *imageBillingTestStore) GetWallet(
	_ context.Context,
	tenantID string,
) (*platform.Wallet, error) {
	if s.wallet == nil {
		return &platform.Wallet{TenantID: tenantID, BalanceIRR: 1_000_000}, nil
	}
	return s.wallet, nil
}

func (s *imageBillingTestStore) GetLatestExchangeRate(
	_ context.Context,
) (*platform.ExchangeRate, error) {
	if s.rate == nil {
		return &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"}, nil
	}
	return s.rate, nil
}

func (s *imageBillingTestStore) UpsertExchangeRate(
	_ context.Context,
	rate platform.ExchangeRate,
) (*platform.ExchangeRate, error) {
	s.rate = &rate
	return &rate, nil
}

func (b *rateLimitCountingBudgeter) CheckRateLimit(_ context.Context, _ *platform.AuthContext, _ string, estimatedTokens int) (*cache.RateLimitResult, error) {
	b.calls++
	b.lastEstimatedTokens = estimatedTokens

	return &cache.RateLimitResult{Allowed: true, Current: 1, Limit: 100, ResetAt: time.Now().Add(time.Minute)}, nil
}

func TestImagesHandlerJSONGeneration(t *testing.T) {
	providerBody := []byte(`{"created":1,"data":[{"b64_json":"AAAA","revised_prompt":"safe","mime_type":"image/png"}],"usage":{"input_tokens":12,"output_tokens":272,"total_tokens":284,"cost_in_usd_ticks":12345}}`)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       providerBody,
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
	assert.Equal(t, "provider_reported", logs.loggedSpends[0].RatingStatus)
	assert.Equal(t, "provider_reported", logs.loggedSpends[0].PricingSource)
	assert.Equal(t, 1, logs.loggedSpends[0].OutputCount)
	require.NotNil(t, logs.loggedSpends[0].RatingSnapshot)
	require.Len(t, logs.loggedSpends[0].CostComponents, 1)
	assert.Equal(t, "provider-reported:request", logs.loggedSpends[0].CostComponents[0].RateKey)
	assert.Equal(t, "0.0000012345", logs.loggedSpends[0].CostComponents[0].ComponentCostUSD)
	require.Len(t, logs.requestEntries(), 1)
	assert.Equal(t, "openai:acct:gpt-image-1", logs.requestEntries()[0].RequestBody["model"])
	assert.Equal(t, "paint a bridge", logs.requestEntries()[0].RequestBody["prompt"])
	assert.Equal(t, "generations", logs.requestEntries()[0].RequestBody["operation"])
	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, providerBody, responseBody)
}

func TestImagesHandlerWalletPreflightSupportsPerImagePricing(t *testing.T) {
	t.Setenv("IMAGE_DIMENSIONAL_BILLING_MODE", "enforce")
	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)
	modelName := "xai:acct:grok-imagine-image"
	price := &platform.ModelPrice{
		Status: "priced",
		Profiles: []platform.PricingProfile{{
			Region:   "global",
			Currency: "USD",
			Rates: []platform.PricingRate{{
				Key:         "output:image",
				Usage:       "output",
				Unit:        "image",
				Amount:      "0.02",
				Quantity:    1,
				ServiceTier: "standard",
			}},
		}},
	}
	deployments := completionsDeploymentStore{deployments: []platform.Deployment{{
		ID:               "dep-image",
		TenantID:         "tenant_a",
		ModelName:        modelName,
		Provider:         "xai",
		Model:            "grok-imagine-image",
		APIKey:           "test-key",
		RegionName:       "global",
		CredentialSource: platform.CredentialSourceGodarManagedKey,
		BillingMode:      platform.BillingModeWalletBilled,
		EffectivePricing: price,
		PricingSource:    "public",
	}}}
	rt, err := router.NewRouter(
		nil,
		config.RouterSettings{CooldownTime: 60, NumRetries: 1},
		mem,
		deployments,
		zap.NewNop(),
	)
	require.NoError(t, err)

	store := &imageBillingTestStore{
		wallet: &platform.Wallet{TenantID: "tenant_a", BalanceIRR: 1_000_000},
		rate:   &platform.ExchangeRate{USDtoIRR: 850000, Source: "test"},
	}
	billing := billingservice.NewService(mem, store, 850000, "test")
	logs := &testLogs{}
	providerBody := []byte(`{"created":1,"data":[{"url":"one"},{"url":"two"}]}`)
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       providerBody,
	}}
	handler := NewImagesHandler(rt, tunnel, billing, store, logs, zap.NewNop())
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "team_admin",
			TokenHash: "hash_1",
			AllowedModelTypes: map[string][]string{
				modelName: {"image_generation"},
			},
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:     "tenant_a",
			Slug:   "a",
			Status: "active",
		})
		return c.Next()
	})
	app.Post("/v1/images/generations", handler.HandleGenerations)

	req := httptest.NewRequest(
		"POST",
		"/v1/images/generations",
		bytes.NewBufferString(`{"model":"xai:acct:grok-imagine-image","prompt":"two images","n":2}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, providerBody, body)
	require.Len(t, logs.loggedSpends, 1)
	assert.InDelta(t, 0.04, logs.loggedSpends[0].CostUSD, 1e-12)
	assert.Equal(t, 2, logs.loggedSpends[0].OutputCount)
	assert.Equal(t, "rated", logs.loggedSpends[0].RatingStatus)
	require.Len(t, logs.loggedSpends[0].CostComponents, 1)
	assert.Equal(t, "image", logs.loggedSpends[0].CostComponents[0].Unit)
	assert.Equal(t, "0.04", logs.loggedSpends[0].CostComponents[0].ComponentCostUSD)
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

	require.NotNil(t, usage.InputTextTokens)
	assert.Equal(t, 5, *usage.InputTextTokens)
	assert.Equal(t, 1767, usage.OutputTokens)
	assert.Equal(t, 1772, usage.TotalTokens)
}

func TestEstimatedImageUsageIncludesEditInputImages(t *testing.T) {
	usage := estimatedImageUsage("openai:default:gpt-image-2", map[string]interface{}{
		"prompt_length": 20,
		"image_count":   2,
		"size":          "1024x1024",
		"quality":       "low",
	})

	require.NotNil(t, usage.InputTextTokens)
	require.NotNil(t, usage.InputImageTokens)
	assert.Equal(t, 5, *usage.InputTextTokens)
	assert.Equal(t, 3072, *usage.InputImageTokens)
	assert.Equal(t, 3077, usage.InputTokens)
	assert.Equal(t, 196, usage.OutputTokens)
	assert.Equal(t, 3273, usage.TotalTokens)
}

func TestEstimatedImageUsageUsesDallE2VariationUnits(t *testing.T) {
	usage := estimatedImageUsage("openai:default:dall-e-2", map[string]interface{}{
		"image_count": 1,
		"n":           2,
		"size":        "512x512",
	})

	require.NotNil(t, usage.InputTextTokens)
	require.NotNil(t, usage.InputImageTokens)
	assert.Equal(t, 100, *usage.InputTextTokens)
	assert.Equal(t, 1024, *usage.InputImageTokens)
	assert.Equal(t, 1124, usage.InputTokens)
	assert.Equal(t, 1800, usage.OutputTokens)
	assert.Equal(t, 2924, usage.TotalTokens)
}

func TestImageUsageFromResponseNormalizesOpenAIImageDetails(t *testing.T) {
	estimated := platform.ImageUsage{
		InputTextTokens:  imageTestIntPointer(1),
		InputImageTokens: imageTestIntPointer(1),
		OutputImageCount: 3,
		Resolution:       "4k",
		Quality:          "high",
	}
	usage := imageUsageFromResponse([]byte(`{
		"data":[{"url":"one"},{"url":"two"}],
		"usage":{
			"input_tokens":50,
			"input_tokens_details":{"text_tokens":10,"image_tokens":40},
			"output_tokens":50,
			"total_tokens":100
		}
	}`), estimated)

	require.NotNil(t, usage.InputTextTokens)
	require.NotNil(t, usage.InputImageTokens)
	assert.Equal(t, 10, *usage.InputTextTokens)
	assert.Equal(t, 40, *usage.InputImageTokens)
	assert.Equal(t, 50, usage.InputTokens)
	assert.Equal(t, 50, usage.OutputTokens)
	assert.Equal(t, 100, usage.TotalTokens)
	assert.Equal(t, 2, usage.OutputImageCount)
	assert.Equal(t, "4k", usage.Resolution)
	assert.Equal(t, "high", usage.Quality)
	assert.Equal(t, platform.ImageUsageSourceProviderUsage, usage.Source)
}

func TestImageUsageFromResponseSplitsMissingDetailsProportionally(t *testing.T) {
	estimated := platform.ImageUsage{
		InputTextTokens:  imageTestIntPointer(100),
		InputImageTokens: imageTestIntPointer(200),
		OutputImageCount: 1,
	}
	usage := imageUsageFromResponse([]byte(`{
		"data":[{"url":"one"}],
		"usage":{"input_tokens":10,"output_tokens":1,"total_tokens":11}
	}`), estimated)

	require.NotNil(t, usage.InputTextTokens)
	require.NotNil(t, usage.InputImageTokens)
	assert.Equal(t, 4, *usage.InputTextTokens)
	assert.Equal(t, 6, *usage.InputImageTokens)
}

func TestImageUsageFromResponseParsesProviderTicksExactly(t *testing.T) {
	usage := imageUsageFromResponse(
		[]byte(`{"data":[{"url":"one"}],"usage":{"cost_in_usd_ticks":200000000}}`),
		platform.ImageUsage{OutputImageCount: 1},
	)

	require.NotNil(t, usage.ProviderCostUSDTicks)
	assert.Equal(t, int64(200_000_000), *usage.ProviderCostUSDTicks)
	assert.Equal(t, platform.ImageUsageSourceProviderReported, usage.Source)
}

func TestImageUsageFromResponseDoesNotPromoteEstimateWithoutProviderFacts(t *testing.T) {
	estimated := platform.ImageUsage{
		InputTokens:      100,
		OutputTokens:     200,
		TotalTokens:      300,
		OutputImageCount: 2,
		Source:           platform.ImageUsageSourceEstimated,
	}

	for _, body := range [][]byte{
		[]byte(`not-json`),
		[]byte(`{"created":1}`),
		[]byte(`{"created":1,"usage":{}}`),
	} {
		usage := imageUsageFromResponse(body, estimated)

		assert.Equal(t, platform.ImageUsageSourceEstimated, usage.Source)
		assert.Zero(t, usage.InputTokens)
		assert.Zero(t, usage.OutputTokens)
		assert.Zero(t, usage.TotalTokens)
		assert.Equal(t, 2, usage.OutputImageCount)
	}
}

func TestCanonicalImageResolutionUsesExplicitKClassOnly(t *testing.T) {
	assert.Equal(t, "2k", canonicalImageResolution(map[string]interface{}{
		"resolution": "2K",
		"image_size": "4k",
		"size":       "1024x1024",
	}))
	assert.Equal(t, "4k", canonicalImageResolution(map[string]interface{}{
		"resolution": "1024x1024",
		"image_size": "4K",
		"size":       "2k",
	}))
	assert.Empty(t, canonicalImageResolution(map[string]interface{}{
		"size": "4096x4096",
	}))
}

func imageTestIntPointer(value int) *int {
	return &value
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
	logs := &testLogs{}
	app, _ := newImagesTestApp(t, "grok-imagine-image", "xai", "grok-imagine-image", tunnel, logs, nil)

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
	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, 1, logs.loggedSpends[0].OutputCount)
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
	logs := &testLogs{}
	app, _ := newImagesTestApp(t, "openai:acct:dall-e-2", "openai", "dall-e-2", tunnel, logs, nil)

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
	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, 1, logs.loggedSpends[0].OutputCount)
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
