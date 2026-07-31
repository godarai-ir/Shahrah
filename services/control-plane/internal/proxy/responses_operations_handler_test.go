package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type responseOperationResources struct {
	mappings map[string]platform.ProviderNativeResourceMapping
	statuses map[string]string
}

func (s *responseOperationResources) UpsertProviderNativeResourceMapping(
	_ context.Context,
	mapping platform.ProviderNativeResourceMapping,
) (*platform.ProviderNativeResourceMapping, error) {
	if s.mappings == nil {
		s.mappings = make(map[string]platform.ProviderNativeResourceMapping)
	}
	s.mappings[mapping.PublicID] = mapping
	return &mapping, nil
}

func (s *responseOperationResources) GetProviderNativeResourceMapping(
	_ context.Context,
	tenantID string,
	provider string,
	_ string,
	publicID string,
) (*platform.ProviderNativeResourceMapping, error) {
	mapping, ok := s.mappings[publicID]
	if !ok {
		return nil, storage.ErrProviderNativeResourceNotFound
	}
	if mapping.Provider != "" && !strings.EqualFold(mapping.Provider, provider) {
		return nil, storage.ErrProviderNativeResourceNotFound
	}
	if mapping.TenantID != "" && mapping.TenantID != tenantID {
		return nil, storage.ErrProviderNativeResourceNotFound
	}
	return &mapping, nil
}

func (s *responseOperationResources) UpdateProviderNativeResourceStatus(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	publicID string,
	status string,
) error {
	if _, ok := s.mappings[publicID]; !ok {
		return storage.ErrProviderNativeResourceNotFound
	}
	if s.statuses == nil {
		s.statuses = make(map[string]string)
	}
	s.statuses[publicID] = status
	mapping := s.mappings[publicID]
	mapping.Status = status
	s.mappings[publicID] = mapping
	return nil
}

type responseOperationDeployments struct {
	deployments []platform.Deployment
}

func (s responseOperationDeployments) ListDeployments(
	_ context.Context,
	_ string,
) ([]platform.Deployment, error) {
	return append([]platform.Deployment(nil), s.deployments...), nil
}

func newResponseOperationsTestApp(
	t *testing.T,
	tunnel platform.TunnelTransport,
	resources *responseOperationResources,
) *fiber.App {
	t.Helper()

	deployment := platform.Deployment{
		ID:                  "dep_1",
		TenantID:            "tenant_a",
		ModelName:           "gpt-4o",
		Provider:            "openai",
		Model:               "gpt-4o",
		APIKey:              "test-key",
		ProviderAccountID:   "acct_1",
		ProviderAccountName: "primary-openai",
	}
	xaiDeployment := platform.Deployment{
		ID:                  "dep_xai_1",
		TenantID:            "tenant_a",
		ModelName:           "grok-4.5",
		Provider:            "xai",
		Model:               "grok-4.5",
		APIKey:              "xai-test-key",
		ProviderAccountID:   "xai_acct_1",
		ProviderAccountName: "primary-xai",
	}
	deployments := responseOperationDeployments{deployments: []platform.Deployment{deployment, xaiDeployment}}
	mem, err := cache.NewMemoryCache(zap.NewNop())
	require.NoError(t, err)
	rt, err := router.NewRouter(
		nil,
		config.RouterSettings{CooldownTime: 60, NumRetries: 1},
		mem,
		deployments,
		zap.NewNop(),
	)
	require.NoError(t, err)

	h := NewResponsesHandlerWithDependencies(BaseHandlerDependencies{
		Router:            rt,
		Tunnel:            tunnel,
		Billing:           &testBudgeter{},
		Wallets:           &testWallets{},
		ProviderResources: resources,
		Deployments:       deployments,
		RequestLogger:     &testLogs{},
		SpendLogger:       &testLogs{},
		Logger:            zap.NewNop(),
	})

	app := fiber.New()
	queue := tracing.NewQueue(64)
	app.Use(tracing.Middleware(queue, tracing.NewSampler(1, 1000), 1000))
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID:  "tenant_a",
			KeyID:     "key_1",
			KeyType:   "pat",
			Role:      "team_admin",
			TokenHash: "hash_1",
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{
			ID:     "tenant_a",
			Slug:   "a",
			Status: "active",
		})
		return c.Next()
	})
	app.Post("/v1/responses", h.Handle)
	app.Post("/v1/responses/compact", h.HandleCompact)
	app.Post("/v1/responses/input_tokens", h.HandleInputTokens)
	app.Get("/v1/responses/:response_id", h.HandleRetrieve)
	app.Delete("/v1/responses/:response_id", h.HandleDelete)
	app.Post("/v1/responses/:response_id/cancel", h.HandleCancel)
	app.Get("/v1/responses/:response_id/input_items", h.HandleInputItems)
	app.Post("/v1/batches", h.HandleBatchCreate)
	app.Get("/v1/batches", h.HandleBatchList)
	app.Get("/v1/batches/:batch_id", h.HandleBatchGet)
	app.Post("/v1/batches/:batch_id/cancel", h.HandleBatchCancel)
	app.Get("/v1/batches/:batch_id/requests", h.HandleXAIBatchRequestsList)
	app.Post("/v1/batches/:batch_id/requests", h.HandleXAIBatchRequestsAdd)
	app.Get("/v1/batches/:batch_id/results", h.HandleXAIBatchResults)
	app.Post("/v1/files", h.HandleFileCreate)
	app.Get("/v1/files", h.HandleFileList)
	app.Get("/v1/files/:file_id", h.HandleFileGet)
	app.Put("/v1/files/:file_id", h.HandleFileUpdate)
	app.Delete("/v1/files/:file_id", h.HandleFileDelete)
	app.Get("/v1/files/:file_id/content", h.HandleFileContent)
	app.Post("/v1/files/:file_id/public-url", h.HandleFilePublicURLCreate)
	app.Post("/v1/files/:file_id/public-url/revoke", h.HandleFilePublicURLRevoke)
	return app
}

func TestOpenAIFileCreateUsesRawTunnelAndRecordsAffinity(t *testing.T) {
	resources := &responseOperationResources{}
	tunnel := &rawSpeechTunnel{
		captureTunnel: &captureTunnel{},
		rawResp: &platform.TunnelRawResponse{
			StatusCode: fiber.StatusOK,
			Headers:    map[string]string{"content-type": "application/json"},
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"file_public","status":"processed"}`))),
		},
	}
	app := newResponseOperationsTestApp(t, tunnel, resources)
	requestBody, contentType := buildMultipartBody(
		t,
		map[string]string{"purpose": "batch"},
		map[string][]byte{"file": []byte(`{"custom_id":"one"}`)},
	)
	req := httptest.NewRequest(http.MethodPost, "/v1/files", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer closeResponseBody(t, resp)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, platform.ProviderOperationOpenAIFilesCreate, tunnel.lastReq.Operation)
	assert.Equal(t, requestBody, tunnel.rawBody)
	mapping, ok := resources.mappings["file_public"]
	require.True(t, ok)
	assert.Equal(t, openAIFileResourceKind, mapping.ResourceKind)
	assert.Equal(t, "acct_1", mapping.ProviderAccountID)
}

func TestXAIFileCreateUsesSelectedProviderAndRecordsAffinity(t *testing.T) {
	resources := &responseOperationResources{}
	tunnel := &rawSpeechTunnel{
		captureTunnel: &captureTunnel{},
		rawResp: &platform.TunnelRawResponse{
			StatusCode: fiber.StatusOK,
			Headers:    map[string]string{"content-type": "application/json"},
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"id":"file_xai","status":"processed"}`))),
		},
	}
	app := newResponseOperationsTestApp(t, tunnel, resources)
	requestBody, contentType := buildMultipartBody(t, map[string]string{"purpose": "assistants"}, map[string][]byte{"file": []byte("hello")})
	req := httptest.NewRequest(http.MethodPost, "/v1/files", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-godarai-provider", "xai")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "xai", tunnel.lastReq.Deployment.Provider)
	assert.Equal(t, "xai_acct_1", tunnel.lastReq.Deployment.ProviderAccountID)
	assert.Equal(t, requestBody, tunnel.rawBody)
	mapping, ok := resources.mappings["file_xai"]
	require.True(t, ok)
	assert.Equal(t, "xai", mapping.Provider)
	assert.Equal(t, "xai_acct_1", mapping.ProviderAccountID)
}

func TestXAIFilePublicURLUsesOwnedFileProvider(t *testing.T) {
	resources := &responseOperationResources{mappings: map[string]platform.ProviderNativeResourceMapping{
		"file_xai": {
			PublicID: "file_xai", TenantID: "tenant_a", Provider: "xai",
			ProviderAccountID: "xai_acct_1", ResourceKind: openAIFileResourceKind,
			UpstreamID: "file_xai_upstream", Status: "processed",
		},
	}}
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"public_url":"https://files-cdn.x.ai/token/file.png"}`),
	}}
	app := newResponseOperationsTestApp(t, tunnel, resources)
	req := httptest.NewRequest(http.MethodPost, "/v1/files/file_xai/public-url", bytes.NewBufferString(`{"expires_after":3600}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, platform.ProviderOperationXAIFilePublicURLCreate, tunnel.lastReq.Operation)
	assert.Equal(t, "file_xai_upstream", tunnel.lastReq.OperationInput.ResourceID)
	assert.Equal(t, "xai", tunnel.lastReq.Deployment.Provider)
	assert.JSONEq(t, `{"expires_after":3600}`, string(tunnel.lastReq.Body))
	assert.Equal(t, "https://files-cdn.x.ai/token/file.png", resources.mappings["file_xai"].Metadata["public_url"])
}

func TestXAIFilePublicURLRejectsKnownIneligibleFileMetadata(t *testing.T) {
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)
	shortExpiry := now.Add(90 * time.Minute)
	tests := []struct {
		name       string
		mapping    platform.ProviderNativeResourceMapping
		expires    *int
		wantReason string
	}{
		{
			name:       "too large",
			mapping:    platform.ProviderNativeResourceMapping{Metadata: map[string]interface{}{"bytes": int64(50*1024*1024 + 1)}},
			wantReason: "50 MiB",
		},
		{
			name:       "unsupported content type",
			mapping:    platform.ProviderNativeResourceMapping{Metadata: map[string]interface{}{"content_type": "text/plain"}},
			wantReason: "content_type",
		},
		{
			name:       "url outlives file",
			mapping:    platform.ProviderNativeResourceMapping{ExpiresAt: &shortExpiry},
			expires:    pointerTo(2 * 60 * 60),
			wantReason: "remaining lifetime",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Contains(t, validateXAIFilePublicURLRequest(&test.mapping, test.expires, now), test.wantReason)
		})
	}
	allowed := platform.ProviderNativeResourceMapping{Metadata: map[string]interface{}{
		"bytes":        int64(50 * 1024 * 1024),
		"content_type": "image/png",
	}}
	assert.Empty(t, validateXAIFilePublicURLRequest(&allowed, pointerTo(3600), now))
}

func pointerTo(value int) *int { return &value }

func TestXAIFileUpdateUsesFixedOperationAndOwnedAccount(t *testing.T) {
	resources := &responseOperationResources{mappings: map[string]platform.ProviderNativeResourceMapping{
		"file_xai": {
			PublicID: "file_xai", TenantID: "tenant_a", Provider: "xai",
			ProviderAccountID: "xai_acct_1", ResourceKind: openAIFileResourceKind,
			UpstreamID: "file_xai_upstream", Status: "processed",
		},
	}}
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"id":"file_xai_upstream","filename":"renamed.png"}`),
	}}
	app := newResponseOperationsTestApp(t, tunnel, resources)
	req := httptest.NewRequest(http.MethodPut, "/v1/files/file_xai", bytes.NewBufferString(`{"filename":"renamed.png"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, platform.ProviderOperationXAIFilesUpdate, tunnel.lastReq.Operation)
	assert.Equal(t, "file_xai_upstream", tunnel.lastReq.OperationInput.ResourceID)
	assert.Equal(t, "xai_acct_1", tunnel.lastReq.Deployment.ProviderAccountID)
	assert.JSONEq(t, `{"filename":"renamed.png"}`, string(tunnel.lastReq.Body))
	assert.Equal(t, "renamed.png", resources.mappings["file_xai"].Metadata["filename"])
	_, createdWithUpstreamID := resources.mappings["file_xai_upstream"]
	assert.False(t, createdWithUpstreamID)
}

func TestXAIFileListRecordsReturnedFilesForOwnedFollowUpOperations(t *testing.T) {
	resources := &responseOperationResources{}
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"data":[{"id":"file_existing","bytes":42,"filename":"existing.pdf","expires_at":1780000000}],"pagination_token":"next"}`),
	}}
	app := newResponseOperationsTestApp(t, tunnel, resources)
	req := httptest.NewRequest(http.MethodGet, "/v1/files?limit=20&pagination_token=previous", nil)
	req.Header.Set("x-godarai-provider", "xai")
	req.Header.Set("x-godarai-provider-account", "xai_acct_1")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	mapping, ok := resources.mappings["file_existing"]
	require.True(t, ok)
	assert.Equal(t, "xai", mapping.Provider)
	assert.Equal(t, "xai_acct_1", mapping.ProviderAccountID)
	assert.Equal(t, int64(42), mapping.Metadata["bytes"])
	assert.Equal(t, "existing.pdf", mapping.Metadata["filename"])
}

func TestOpenAIBatchOperationsArePublicAndAccountBound(t *testing.T) {
	resources := &responseOperationResources{mappings: map[string]platform.ProviderNativeResourceMapping{
		"file_input": {
			PublicID: "file_input", TenantID: "tenant_a", Provider: "openai",
			ProviderAccountID: "acct_1", ResourceKind: openAIFileResourceKind,
			UpstreamID: "file_upstream", Status: "processed",
		},
	}}
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"id":"batch_public","status":"validating"}`),
	}}
	app := newResponseOperationsTestApp(t, tunnel, resources)

	req := httptest.NewRequest(http.MethodPost, "/v1/batches", bytes.NewBufferString(`{
		"input_file_id":"file_input",
		"endpoint":"/v1/responses",
		"completion_window":"24h",
		"future_field":{"preserve":true}
	}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer closeResponseBody(t, resp)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, platform.ProviderOperationOpenAIBatchesCreate, tunnel.lastReq.Operation)
	assert.Equal(t, "acct_1", tunnel.lastReq.Deployment.ProviderAccountID)

	var relayed map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &relayed))
	require.JSONEq(t, `"file_upstream"`, string(relayed["input_file_id"]))
	require.JSONEq(t, `{"preserve":true}`, string(relayed["future_field"]))
	mapping, ok := resources.mappings["batch_public"]
	require.True(t, ok)
	assert.Equal(t, "acct_1", mapping.ProviderAccountID)
	assert.Equal(t, openAIBatchResourceKind, mapping.ResourceKind)
}

func TestXAIBatchCreateFailsClosedUntilSettlementIsReady(t *testing.T) {
	resources := &responseOperationResources{}
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"batch_id":"batch_xai","name":"nightly","state":{"num_requests":0}}`),
	}}
	app := newResponseOperationsTestApp(t, tunnel, resources)
	req := httptest.NewRequest(http.MethodPost, "/v1/batches", bytes.NewBufferString(`{"name":"nightly"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-godarai-provider", "xai")
	req.Header.Set("x-godarai-provider-account", "xai_acct_1")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	assert.Empty(t, tunnel.lastReq.Operation)
	assert.Empty(t, resources.mappings)
}

func TestParseXAIBatchRequestsRejectsUnknownSurface(t *testing.T) {
	requests, err := parseXAIBatchRequests([]byte(`{"batch_requests":[
		{"batch_request_id":"one","batch_request":{"chat_get_completion":{"model":"grok-4.3"}}},
		{"batch_request_id":"two","batch_request":{"responses":{"model":"grok-4.3"}}},
		{"batch_request_id":"three","batch_request":{"image_generation":{"model":"grok-imagine-image"}}},
		{"batch_request_id":"four","batch_request":{"image_edit":{"model":"grok-imagine-image-quality"}}}
	]}`))
	require.NoError(t, err)
	require.Equal(t, []xaiBatchModelRequest{
		{Model: "grok-4.3", Endpoint: "chat_completions"},
		{Model: "grok-4.3", Endpoint: "responses"},
		{Model: "grok-imagine-image", Endpoint: "image_generations"},
		{Model: "grok-imagine-image-quality", Endpoint: "image_edits"},
	}, requests)

	_, err = parseXAIBatchRequests([]byte(`{"batch_requests":[{"batch_request":{"unknown":{"model":"grok-4.3"}}}]}`))
	require.ErrorContains(t, err, "unsupported xAI batch request type")
}

func TestResponsesResourceOperationsArePublicAndAccountBound(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		path      string
		operation platform.ProviderOperation
		assert    func(*testing.T, platform.ProviderOperationInput)
	}{
		{
			name:      "retrieve",
			method:    http.MethodGet,
			path:      "/v1/responses/resp_public?include=reasoning.encrypted_content&include_obfuscation=false&starting_after=7",
			operation: platform.ProviderOperationOpenAIResponsesGet,
			assert: func(t *testing.T, input platform.ProviderOperationInput) {
				require.Equal(t, []string{"reasoning.encrypted_content"}, input.Include)
				require.NotNil(t, input.IncludeObfuscation)
				assert.False(t, *input.IncludeObfuscation)
				require.NotNil(t, input.StartingAfter)
				assert.EqualValues(t, 7, *input.StartingAfter)
			},
		},
		{
			name:      "delete",
			method:    http.MethodDelete,
			path:      "/v1/responses/resp_public",
			operation: platform.ProviderOperationOpenAIResponsesDelete,
		},
		{
			name:      "cancel",
			method:    http.MethodPost,
			path:      "/v1/responses/resp_public/cancel",
			operation: platform.ProviderOperationOpenAIResponsesCancel,
		},
		{
			name:      "input items",
			method:    http.MethodGet,
			path:      "/v1/responses/resp_public/input_items?after=item_1&include=message.input_image.image_url&limit=50&order=asc",
			operation: platform.ProviderOperationOpenAIResponsesInputItems,
			assert: func(t *testing.T, input platform.ProviderOperationInput) {
				assert.Equal(t, "item_1", input.AfterID)
				assert.Equal(t, 50, input.Limit)
				assert.Equal(t, "asc", input.Order)
				assert.Equal(t, []string{"message.input_image.image_url"}, input.Include)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := &responseOperationResources{mappings: map[string]platform.ProviderNativeResourceMapping{
				"resp_public": {
					PublicID:          "resp_public",
					TenantID:          "tenant_a",
					Provider:          "openai",
					ProviderAccountID: "acct_1",
					ResourceKind:      responsesResourceKind,
					UpstreamID:        "resp_upstream",
					Status:            "completed",
				},
			}}
			tunnel := &captureTunnel{resp: &platform.TunnelResponse{
				StatusCode: http.StatusOK,
				Headers:    map[string]string{"x-request-id": "openai_req_1"},
				Body:       []byte(`{"id":"resp_upstream","object":"response","status":"completed"}`),
			}}
			app := newResponseOperationsTestApp(t, tunnel, resources)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-Client-Request-Id", "client_req_1")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.operation, tunnel.lastReq.Operation)
			assert.Equal(t, "resp_upstream", tunnel.lastReq.OperationInput.ResourceID)
			assert.Equal(t, "acct_1", tunnel.lastReq.Deployment.ProviderAccountID)
			assert.Equal(t, "client_req_1", tunnel.lastReq.Headers["X-Client-Request-Id"])
			if tt.assert != nil {
				tt.assert(t, tunnel.lastReq.OperationInput)
			}
		})
	}
}

func TestResponsesRetrieveSupportsSSEContinuation(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 1, 0)
	stream.chunks <- platform.TunnelChunk{
		Data:   []byte("event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"hi\"}\n\n"),
		IsDone: true,
	}
	close(stream.chunks)

	resources := &responseOperationResources{mappings: map[string]platform.ProviderNativeResourceMapping{
		"resp_public": {
			PublicID:          "resp_public",
			TenantID:          "tenant_a",
			Provider:          "openai",
			ProviderAccountID: "acct_1",
			ResourceKind:      responsesResourceKind,
			UpstreamID:        "resp_upstream",
		},
	}}
	tunnel := &captureTunnel{stream: stream}
	app := newResponseOperationsTestApp(t, tunnel, resources)

	req := httptest.NewRequest(http.MethodGet, "/v1/responses/resp_public?stream=true&starting_after=3", nil)
	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.True(t, tunnel.lastReq.Stream)
	require.NotNil(t, tunnel.lastReq.OperationInput.Stream)
	assert.True(t, *tunnel.lastReq.OperationInput.Stream)
	require.NotNil(t, tunnel.lastReq.OperationInput.StartingAfter)
	assert.EqualValues(t, 3, *tunnel.lastReq.OperationInput.StartingAfter)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "response.output_text.delta")
}

func TestResponsesBodyOperationsPreserveOfficialPayloads(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		body      string
		operation platform.ProviderOperation
	}{
		{
			name:      "compact",
			path:      "/v1/responses/compact",
			body:      `{"model":"gpt-4o","input":[{"role":"user","content":"hello"}],"future_compaction_field":true}`,
			operation: platform.ProviderOperationOpenAIResponsesCompact,
		},
		{
			name:      "input tokens",
			path:      "/v1/responses/input_tokens",
			body:      `{"conversation":"conv_1","input":"hello","future_token_field":{"enabled":true}}`,
			operation: platform.ProviderOperationOpenAIResponsesInputTokens,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tunnel := &captureTunnel{resp: &platform.TunnelResponse{
				StatusCode: http.StatusOK,
				Body:       []byte(`{"object":"response.input_tokens","input_tokens":12}`),
			}}
			app := newResponseOperationsTestApp(t, tunnel, &responseOperationResources{})
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.operation, tunnel.lastReq.Operation)

			var original, outbound map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(tt.body), &original))
			require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
			for key, value := range original {
				assert.Equal(t, value, outbound[key])
			}
		})
	}
}

func TestResponsesCreatePersistsProviderAccountAffinity(t *testing.T) {
	resources := &responseOperationResources{}
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: http.StatusOK,
		Body: []byte(`{
			"id":"resp_created",
			"object":"response",
			"model":"gpt-4o",
			"status":"completed",
			"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}
		}`),
	}}
	app := newResponseOperationsTestApp(t, tunnel, resources)
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"model":"gpt-4o","input":"hello"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Request-Id", "create_client_req")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "create_client_req", tunnel.lastReq.Headers["X-Client-Request-Id"])
	mapping, ok := resources.mappings["resp_created"]
	require.True(t, ok)
	assert.Equal(t, "acct_1", mapping.ProviderAccountID)
	assert.Equal(t, "resp_created", mapping.UpstreamID)
	assert.Equal(t, "completed", mapping.Status)
}

func TestResponsesContinuationMayOmitModelAndUsesPreviousResponseAccount(t *testing.T) {
	resources := &responseOperationResources{mappings: map[string]platform.ProviderNativeResourceMapping{
		"resp_public": {
			PublicID:          "resp_public",
			TenantID:          "tenant_a",
			Provider:          "openai",
			ProviderAccountID: "acct_1",
			ResourceKind:      responsesResourceKind,
			UpstreamID:        "resp_upstream",
			Metadata:          map[string]interface{}{"model": "gpt-4o"},
		},
	}}
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: http.StatusOK,
		Body: []byte(`{
			"id":"resp_continued",
			"object":"response",
			"model":"gpt-4o",
			"status":"completed",
			"usage":{"input_tokens":1,"output_tokens":1,"total_tokens":2}
		}`),
	}}
	app := newResponseOperationsTestApp(t, tunnel, resources)
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"previous_response_id":"resp_public"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "acct_1", tunnel.lastReq.Deployment.ProviderAccountID)

	var outbound map[string]interface{}
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	assert.NotContains(t, outbound, "model")
	assert.Equal(t, "resp_upstream", outbound["previous_response_id"])
}

func TestResponsesStreamingCreatePersistsProviderAccountAffinity(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{
		Data: []byte("event: response.created\r\ndata: " +
			`{"type":"response.created","response":{"id":"resp_streamed","model":"gpt-4o","status":"in_progress"}}`),
	}
	stream.chunks <- platform.TunnelChunk{
		Data:   []byte("\r\n\r\n"),
		IsDone: true,
	}
	close(stream.chunks)

	resources := &responseOperationResources{}
	app := newResponseOperationsTestApp(t, &captureTunnel{stream: stream}, resources)
	req := httptest.NewRequest(
		http.MethodPost,
		"/v1/responses",
		bytes.NewBufferString(`{"model":"gpt-4o","input":"hello","stream":true}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	mapping, ok := resources.mappings["resp_streamed"]
	require.True(t, ok)
	assert.Equal(t, "acct_1", mapping.ProviderAccountID)
	assert.Equal(t, "in_progress", mapping.Status)
}

func TestXAIResponseResourceMappingIsProviderAndTenantBound(t *testing.T) {
	resources := &responseOperationResources{mappings: map[string]platform.ProviderNativeResourceMapping{
		"resp_xai": {
			PublicID:          "resp_xai",
			TenantID:          "tenant_a",
			Provider:          "xai",
			ProviderAccountID: "xai_acct_1",
			ResourceKind:      responsesResourceKind,
			UpstreamID:        "resp_xai_upstream",
		},
	}}
	handler := &ResponsesHandler{resources: resources}

	mapping, err := handler.responseResourceMapping(context.Background(), "tenant_a", "resp_xai")
	require.NoError(t, err)
	assert.Equal(t, "xai", mapping.Provider)
	assert.Equal(t, "xai_acct_1", mapping.ProviderAccountID)

	_, err = handler.responseResourceMapping(context.Background(), "tenant_b", "resp_xai")
	require.ErrorIs(t, err, storage.ErrProviderNativeResourceNotFound)
}

func TestResolveResponseDeploymentPinsXAIProviderAccount(t *testing.T) {
	handler := &ResponsesHandler{deployments: responseOperationDeployments{deployments: []platform.Deployment{
		{ID: "openai_dep", Provider: "openai", Model: "grok-4.5", ProviderAccountID: "shared"},
		{ID: "xai_dep", Provider: "xai", Model: "grok-4.5", ProviderAccountID: "shared"},
	}}}

	deployment, err := handler.resolveResponseDeployment(
		context.Background(),
		"tenant_a",
		"xai",
		"shared",
		"grok-4.5",
	)
	require.NoError(t, err)
	assert.Equal(t, "xai_dep", deployment.ID)
}
