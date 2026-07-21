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

func newAudioSpeechTestApp(t *testing.T, modelName string, provider string, deploymentModel string, tunnel *captureTunnel, logs *testLogs, allowedTypes []string, emitMetric ...func(input metricsobs.BuildInput)) (*fiber.App, *tracing.Queue) {
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
		allowedTypes = []string{string(platform.ModelTypeTextToSpeech)}
	}

	h := NewAudioSpeechHandler(rt, tunnel, &testBudgeter{}, &testWallets{}, logs, zap.NewNop(), emitMetric...)
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
	app.Post("/v1/audio/speech", h.Handle)

	return app, q
}

func TestAudioSpeechHandlerJSONSuccess(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "audio/wav", "x-request-id": "provider_speech_1"},
		Body:       []byte("fake-wav-bytes"),
	}}
	logs := &testLogs{}
	app, _ := newAudioSpeechTestApp(t, "openai:acct:gpt-4o-mini-tts", "openai", "gpt-4o-mini-tts", tunnel, logs, nil)

	req := httptest.NewRequest("POST", "/v1/audio/speech", bytes.NewBufferString(`{
		"model":"openai:acct:gpt-4o-mini-tts",
		"input":"سلام، به گدارAI خوش آمدید.",
		"voice":"alloy",
		"response_format":"wav",
		"speed":1.1,
		"instructions":"با لحن آرام بخوان."
	}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "audio/wav", resp.Header.Get("Content-Type"))
	assert.Equal(t, "provider_speech_1", resp.Header.Get("X-Request-Id"))
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, []byte("fake-wav-bytes"), bodyBytes)

	assert.Equal(t, "/audio/speech", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])
	assert.False(t, tunnel.lastReq.Stream)
	assert.Equal(t, "application/json", tunnel.lastReq.Headers["Content-Type"])
	assert.Equal(t, "builtin", tunnel.lastReq.RequestMetadata["voice_type"])
	assert.Equal(t, "alloy", tunnel.lastReq.RequestMetadata["voice"])
	assert.Equal(t, "wav", tunnel.lastReq.RequestMetadata["response_format"])

	var outbound map[string]interface{}
	require.NoError(t, json.Unmarshal(tunnel.lastReq.Body, &outbound))
	assert.Equal(t, "gpt-4o-mini-tts", outbound["model"])
	assert.Equal(t, "alloy", outbound["voice"])

	require.Len(t, logs.loggedSpends, 1)
	assert.Greater(t, logs.loggedSpends[0].TotalTokens, 0)
	require.Len(t, logs.loggedEntries, 1)
	rawEntry, err := json.Marshal(logs.loggedEntries[0])
	require.NoError(t, err)
	assert.NotContains(t, string(rawEntry), "سلام، به گدارAI خوش آمدید")
	assert.NotContains(t, string(rawEntry), "با لحن آرام")
	assert.NotContains(t, string(rawEntry), "fake-wav-bytes")
	assert.Contains(t, string(rawEntry), "body_bytes")
}

func TestAudioSpeechHandlerCustomVoiceObject(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "audio/mpeg"},
		Body:       []byte("mp3"),
	}}
	app, _ := newAudioSpeechTestApp(t, "gpt-4o-mini-tts", "openai", "gpt-4o-mini-tts", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/audio/speech", bytes.NewBufferString(`{"model":"gpt-4o-mini-tts","input":"hello","voice":{"id":"voice_secret_123"}}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "custom", tunnel.lastReq.RequestMetadata["voice_type"])
	assert.NotContains(t, tunnel.lastReq.RequestMetadata, "voice_secret_123")
	rawMetadata, err := json.Marshal(tunnel.lastReq.RequestMetadata)
	require.NoError(t, err)
	assert.NotContains(t, string(rawMetadata), "voice_secret_123")
}

func TestAudioSpeechHandlerValidationErrors(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "audio/mpeg"}, Body: []byte("mp3")}}
	app, _ := newAudioSpeechTestApp(t, "gpt-4o-mini-tts", "openai", "gpt-4o-mini-tts", tunnel, nil, nil)

	longInput := strings.Repeat("a", 4097)
	cases := []struct {
		name      string
		body      string
		wantParam string
	}{
		{name: "missing model", body: `{"input":"hello","voice":"alloy"}`, wantParam: "model"},
		{name: "missing input", body: `{"model":"gpt-4o-mini-tts","voice":"alloy"}`, wantParam: "input"},
		{name: "long input", body: `{"model":"gpt-4o-mini-tts","input":"` + longInput + `","voice":"alloy"}`, wantParam: "input"},
		{name: "missing voice", body: `{"model":"gpt-4o-mini-tts","input":"hello"}`, wantParam: "voice"},
		{name: "bad response format", body: `{"model":"gpt-4o-mini-tts","input":"hello","voice":"alloy","response_format":"ogg"}`, wantParam: "response_format"},
		{name: "bad speed", body: `{"model":"gpt-4o-mini-tts","input":"hello","voice":"alloy","speed":9}`, wantParam: "speed"},
		{name: "bad stream format", body: `{"model":"gpt-4o-mini-tts","input":"hello","voice":"alloy","stream_format":"chunks"}`, wantParam: "stream_format"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/v1/audio/speech", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Contains(t, string(bodyBytes), `"param":"`+tc.wantParam+`"`)
		})
	}
}

func TestAudioSpeechHandlerRequiresJSON(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "audio/mpeg"}, Body: []byte("mp3")}}
	app, _ := newAudioSpeechTestApp(t, "gpt-4o-mini-tts", "openai", "gpt-4o-mini-tts", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/audio/speech", bytes.NewBufferString(`model=gpt-4o-mini-tts`))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), `"param":"content_type"`)
}

func TestAudioSpeechHandlerDeniedModelType(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "audio/mpeg"}, Body: []byte("mp3")}}
	app, _ := newAudioSpeechTestApp(t, "gpt-4o-mini", "openai", "gpt-4o-mini", tunnel, nil, []string{"chat"})

	req := httptest.NewRequest("POST", "/v1/audio/speech", bytes.NewBufferString(`{"model":"gpt-4o-mini","input":"hello","voice":"alloy"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	assert.Empty(t, tunnel.lastReq.RequestID)
}

func TestAudioSpeechHandlerProviderErrorPassthrough(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusBadRequest,
		Headers:    map[string]string{"content-type": "application/json", "x-request-id": "provider_bad_speech"},
		Body:       []byte(`{"error":{"message":"bad voice","type":"invalid_request_error"}}`),
	}}
	app, _ := newAudioSpeechTestApp(t, "gpt-4o-mini-tts", "openai", "gpt-4o-mini-tts", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/audio/speech", bytes.NewBufferString(`{"model":"gpt-4o-mini-tts","input":"hello","voice":"alloy"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "provider_bad_speech", resp.Header.Get("X-Request-Id"))
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "bad voice")
}

func TestAudioSpeechHandlerStreamRelaysProviderEvents(t *testing.T) {
	stream := newFakeTunnelStream(map[string]string{"content-type": "text/event-stream"}, 2, 0)
	stream.chunks <- platform.TunnelChunk{Data: []byte("event: speech.audio.delta\n")}

	stream.chunks <- platform.TunnelChunk{Data: []byte("data: {\"audio\":\"AAAA\"}\n\n"), IsDone: true}

	close(stream.chunks)

	tunnel := &captureTunnel{stream: stream}
	logs := &testLogs{}
	app, _ := newAudioSpeechTestApp(t, "gpt-4o-mini-tts", "openai", "gpt-4o-mini-tts", tunnel, logs, nil)

	req := httptest.NewRequest("POST", "/v1/audio/speech", bytes.NewBufferString(`{"model":"gpt-4o-mini-tts","input":"hello","voice":"alloy","stream_format":"sse"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, int((5 * time.Second).Milliseconds()))
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))
	assert.True(t, tunnel.lastReq.Stream)
	assert.Equal(t, "/audio/speech", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "speech.audio.delta")
	assert.Contains(t, string(bodyBytes), `"audio":"AAAA"`)
	assert.True(t, stream.closed)
	require.Len(t, logs.loggedSpends, 1)
	assert.Equal(t, "success", logs.loggedSpends[0].Status)
}

func TestAudioSpeechHandlerTraceRedaction(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "audio/mpeg"},
		Body:       []byte("secret-audio-output"),
	}}
	app, q := newAudioSpeechTestApp(t, "gpt-4o-mini-tts", "openai", "gpt-4o-mini-tts", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/audio/speech", bytes.NewBufferString(`{"model":"gpt-4o-mini-tts","input":"secret speech input","voice":"alloy","instructions":"secret speaking style"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	events := q.DequeueBatch(20)
	for _, ev := range events {
		raw, err := json.Marshal(ev.Attributes)
		require.NoError(t, err)

		text := string(raw)
		assert.NotContains(t, text, "secret speech input")
		assert.NotContains(t, text, "secret speaking style")
		assert.NotContains(t, text, "secret-audio-output")
	}
}
