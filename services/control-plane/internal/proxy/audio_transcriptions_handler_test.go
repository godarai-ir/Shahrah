package proxy

import (
	"bytes"
	"encoding/json"
	"io"
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

func newAudioTranscriptionsTestApp(t *testing.T, modelName string, provider string, deploymentModel string, tunnel *captureTunnel, logs *testLogs, allowedTypes []string, emitMetric ...func(input metricsobs.BuildInput)) (*fiber.App, *tracing.Queue) {
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
	if allowedTypes == nil {
		allowedTypes = []string{string(platform.ModelTypeAudioTranscription)}
	}

	h := NewAudioTranscriptionsHandler(rt, tunnel, &testBudgeter{}, &testWallets{}, logs, zap.NewNop(), emitMetric...)
	q := tracing.NewQueue(64)
	s := tracing.NewSampler(1.0, 1500)

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
	app.Post("/v1/audio/transcriptions", h.Handle)

	return app, q
}

func TestAudioTranscriptionsHandlerMultipartSuccess(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{
		"model":                      "openai:acct:gpt-4o-transcribe",
		"language":                   "fa",
		"prompt":                     "کلمه تخصصی شاهراه",
		"response_format":            "verbose_json",
		"timestamp_granularities[]":  "segment",
		"temperature":                "0.2",
		"include[]":                  "logprobs",
		"known_speaker_names[]":      "speaker one",
		"known_speaker_references[]": "ref_1",
	}, map[string][]byte{"file": []byte("fake-audio-bytes")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json", "x-request-id": "provider_req_1"},
		Body:       []byte(`{"text":"سلام دنیا","segments":[{"id":0,"text":"سلام دنیا"}],"usage":{"type":"tokens","input_tokens":14,"output_tokens":45,"total_tokens":59}}`),
	}}
	logs := &testLogs{}
	app, _ := newAudioTranscriptionsTestApp(t, "openai:acct:gpt-4o-transcribe", "openai", "gpt-4o-transcribe", tunnel, logs, nil)

	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "provider_req_1", resp.Header.Get("X-Request-Id"))
	assert.Equal(t, "/audio/transcriptions", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])
	assert.False(t, tunnel.lastReq.Stream)
	assert.True(t, strings.HasPrefix(tunnel.lastReq.Headers["Content-Type"], "multipart/form-data"))

	outboundForm := readMultipartForm(t, tunnel.lastReq.Headers["Content-Type"], tunnel.lastReq.Body)
	assert.Equal(t, []string{"gpt-4o-transcribe"}, outboundForm.Value["model"])
	assert.Equal(t, []string{"fa"}, outboundForm.Value["language"])
	assert.Equal(t, []string{"verbose_json"}, outboundForm.Value["response_format"])
	assertMultipartFileBytes(t, outboundForm, "file", []byte("fake-audio-bytes"))
	assert.Equal(t, 1, tunnel.lastReq.RequestMetadata["file_count"])
	assert.Equal(t, 1, tunnel.lastReq.RequestMetadata["audio_file_count"])
	assert.Equal(t, "fa", tunnel.lastReq.RequestMetadata["language"])
	assert.Equal(t, []string{"segment"}, tunnel.lastReq.RequestMetadata["timestamp_granularities"])
	assert.Equal(t, 1, tunnel.lastReq.RequestMetadata["known_speaker_names_count"])
	assert.Equal(t, 1, tunnel.lastReq.RequestMetadata["known_speaker_references_count"])

	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, 14, logs.loggedSpends[0].InputTokens)
	assert.Equal(t, 45, logs.loggedSpends[0].OutputTokens)
	assert.Equal(t, 59, logs.loggedSpends[0].TotalTokens)
	require.Len(t, logs.loggedEntries, 1)
	rawEntry, err := json.Marshal(logs.loggedEntries[0])
	require.NoError(t, err)
	assert.NotContains(t, string(rawEntry), "fake-audio-bytes")
	assert.NotContains(t, string(rawEntry), "کلمه تخصصی شاهراه")
	assert.NotContains(t, string(rawEntry), "سلام دنیا")
	assert.Contains(t, string(rawEntry), "text_length")
}

func TestAudioTranscriptionsHandlerValidationErrors(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newAudioTranscriptionsTestApp(t, "gpt-4o-transcribe", "openai", "gpt-4o-transcribe", tunnel, nil, nil)

	cases := []struct {
		name        string
		fields      map[string]string
		files       map[string][]byte
		wantParam   string
		wantMessage string
	}{
		{
			name:      "missing file",
			fields:    map[string]string{"model": "gpt-4o-transcribe"},
			files:     nil,
			wantParam: "file",
		},
		{
			name:      "bad response format",
			fields:    map[string]string{"model": "gpt-4o-transcribe", "response_format": "docx"},
			files:     map[string][]byte{"file": []byte("audio")},
			wantParam: "response_format",
		},
		{
			name:        "timestamps require verbose json",
			fields:      map[string]string{"model": "gpt-4o-transcribe", "response_format": "json", "timestamp_granularities[]": "word"},
			files:       map[string][]byte{"file": []byte("audio")},
			wantParam:   "response_format",
			wantMessage: "timestamp_granularities requires response_format=verbose_json",
		},
		{
			name:      "bad temperature",
			fields:    map[string]string{"model": "gpt-4o-transcribe", "temperature": "3"},
			files:     map[string][]byte{"file": []byte("audio")},
			wantParam: "temperature",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, contentType := buildMultipartBody(t, tc.fields, tc.files)
			req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body))
			req.Header.Set("Content-Type", contentType)

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Contains(t, string(bodyBytes), `"param":"`+tc.wantParam+`"`)
			if tc.wantMessage != "" {
				assert.Contains(t, string(bodyBytes), tc.wantMessage)
			}
		})
	}
}

func TestAudioTranscriptionsHandlerRequiresMultipart(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newAudioTranscriptionsTestApp(t, "gpt-4o-transcribe", "openai", "gpt-4o-transcribe", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewBufferString(`{"model":"gpt-4o-transcribe"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), `"param":"content_type"`)
}

func TestAudioTranscriptionsHandlerDeniedModelType(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{"model": "gpt-4o-mini"}, map[string][]byte{"file": []byte("audio")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newAudioTranscriptionsTestApp(t, "gpt-4o-mini", "openai", "gpt-4o-mini", tunnel, nil, []string{"chat"})

	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	assert.Empty(t, tunnel.lastReq.RequestID)
}

func TestAudioTranscriptionsHandlerProviderErrorPassthrough(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{"model": "gpt-4o-transcribe"}, map[string][]byte{"file": []byte("audio")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusBadRequest,
		Headers:    map[string]string{"content-type": "application/json", "x-request-id": "provider_bad_1"},
		Body:       []byte(`{"error":{"message":"bad audio","type":"invalid_request_error"}}`),
	}}
	app, _ := newAudioTranscriptionsTestApp(t, "gpt-4o-transcribe", "openai", "gpt-4o-transcribe", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "provider_bad_1", resp.Header.Get("X-Request-Id"))
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "bad audio")
}

func TestAudioTranscriptionsHandlerStreamRelaysProviderEvents(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{"model": "gpt-4o-transcribe", "stream": "true"}, map[string][]byte{"file": []byte("audio")})
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("event: transcript.text.delta\n")}
	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"delta\":\"سلام\"}\n\n"), IsDone: true}
	close(stream.chunks)

	tunnel := &captureTunnel{stream: stream}
	logs := &testLogs{}
	app, _ := newAudioTranscriptionsTestApp(t, "gpt-4o-transcribe", "openai", "gpt-4o-transcribe", tunnel, logs, nil)

	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.True(t, tunnel.lastReq.Stream)
	assert.Equal(t, "/audio/transcriptions", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "transcript.text.delta")
	assert.Contains(t, string(bodyBytes), `"delta":"سلام"`)
	assert.True(t, stream.closed)
	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, "success", logs.loggedSpends[0].Status)
}

func TestAudioTranscriptionsHandlerTraceRedaction(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{
		"model":  "gpt-4o-transcribe",
		"prompt": "secret speech prompt",
	}, map[string][]byte{"file": []byte("secret-audio-bytes")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"text":"secret transcript text","usage":{"input_tokens":1,"output_tokens":2,"total_tokens":3}}`),
	}}
	app, q := newAudioTranscriptionsTestApp(t, "gpt-4o-transcribe", "openai", "gpt-4o-transcribe", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	events := q.DequeueBatch(20)
	for _, ev := range events {
		raw, err := json.Marshal(ev.Attributes)
		require.NoError(t, err)
		text := string(raw)
		assert.NotContains(t, text, "secret-audio-bytes")
		assert.NotContains(t, text, "secret speech prompt")
		assert.NotContains(t, text, "secret transcript text")
	}
}
