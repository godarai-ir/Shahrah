package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	metricsobs "github.com/Delpak/shahrah/services/control-plane/internal/observability/metrics"
	"github.com/Delpak/shahrah/services/control-plane/internal/observability/tracing"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newAudioTranscriptionsTestApp(t *testing.T, modelName string, provider string, deploymentModel string, tunnel *captureTunnel, logs *testLogs, allowedTypes []string, emitMetric ...func(input metricsobs.BuildInput)) (*fiber.App, *tracing.Queue) {
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

	if allowedTypes == nil {
		allowedTypes = []string{string(platform.ModelTypeAudioTranscription)}
	}

	var metricEmitter func(input metricsobs.BuildInput)
	if len(emitMetric) > 0 {
		metricEmitter = emitMetric[0]
	}

	billing := &testBudgeter{}
	h := NewAudioTranscriptionsHandlerWithDependencies(BaseHandlerDependencies{
		Router:  rt,
		Tunnel:  tunnel,
		Billing: billing,
		BillingCapabilities: BillingCapabilities{
			MeteredUsage:       billing,
			MeteredUsagePolicy: CapabilityFailClosed,
		},
		Wallets:       &testWallets{},
		RequestLogger: logs,
		SpendLogger:   logs,
		Logger:        zap.NewNop(),
		EmitMetric:    metricEmitter,
	})
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

func TestAudioTranscriptionsHandlerSupportedProviderRouting(t *testing.T) {
	cases := []struct {
		name            string
		provider        string
		modelName       string
		deploymentModel string
		wantKind        string
	}{
		{name: "azure openai", provider: "azure-open-ai", modelName: "azure-open-ai:acct:whisper-1", deploymentModel: "whisper-1", wantKind: "openai_compatible"},
		{name: "groq", provider: "groq", modelName: "groq:acct:whisper-large-v3", deploymentModel: "whisper-large-v3", wantKind: "openai_compatible"},
		{name: "mistral", provider: "mistral-ai", modelName: "mistral-ai:acct:voxtral-mini-2602", deploymentModel: "voxtral-mini-2602", wantKind: "openai_compatible"},
		{name: "together", provider: "together-ai", modelName: "together-ai:acct:openai/whisper-large-v3", deploymentModel: "openai/whisper-large-v3", wantKind: "openai_compatible"},
		{name: "deepgram", provider: "deepgram", modelName: "deepgram:acct:nova-3-general", deploymentModel: "deepgram/nova-3-general", wantKind: "native"},
		{name: "cohere", provider: "cohere", modelName: "cohere:acct:cohere-transcribe-03-2026", deploymentModel: "cohere/cohere-transcribe-03-2026", wantKind: "native"},
		{name: "google vertex", provider: "google-vertex", modelName: "google-vertex:acct:google/chirp-3", deploymentModel: "google/chirp-3", wantKind: "native"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, contentType := buildMultipartBody(t, map[string]string{"model": tc.modelName}, map[string][]byte{"file": []byte("audio")})
			tunnel := &captureTunnel{resp: &platform.TunnelResponse{
				StatusCode: fiber.StatusOK,
				Headers:    map[string]string{"content-type": "application/json"},
				Body:       []byte(`{"text":"ok","usage":{"prompt_audio_seconds":1}}`),
			}}
			app, _ := newAudioTranscriptionsTestApp(t, tc.modelName, tc.provider, tc.deploymentModel, tunnel, nil, nil)

			req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body))
			req.Header.Set("Content-Type", contentType)

			resp, err := app.Test(req)
			defer closeResponseBody(t, resp)

			require.NoError(t, err)

			if resp.StatusCode != fiber.StatusOK {
				payload, _ := io.ReadAll(resp.Body)
				t.Logf("unexpected response: %s", payload)
			}

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.Equal(t, tc.wantKind, tunnel.lastReq.RequestMetadata["provider_adapter_kind"])
			assert.Equal(t, "/audio/transcriptions", tunnel.lastReq.RequestMetadata["provider_endpoint_path"])
		})
	}
}

func TestAudioTranscriptionsHandlerGatesBedrockNovaSonic(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{"model": "aws-bedrock:acct:amazon.nova-2-sonic-v1:0"}, map[string][]byte{"file": []byte("audio")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newAudioTranscriptionsTestApp(t, "aws-bedrock:acct:amazon.nova-2-sonic-v1:0", "aws-bedrock", "amazon.nova-2-sonic-v1:0", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
	assert.Empty(t, tunnel.lastReq.RequestID)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Nova Sonic")
}

func TestAudioTranscriptionsHandlerRejectsNativeProviderStreaming(t *testing.T) {
	body, contentType := buildMultipartBody(t, map[string]string{"model": "deepgram:acct:nova-3-general", "stream": "true"}, map[string][]byte{"file": []byte("audio")})
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{StatusCode: fiber.StatusOK, Headers: map[string]string{"content-type": "application/json"}, Body: []byte(`{}`)}}
	app, _ := newAudioTranscriptionsTestApp(t, "deepgram:acct:nova-3-general", "deepgram", "deepgram/nova-3-general", tunnel, nil, nil)

	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", contentType)

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Empty(t, tunnel.lastReq.RequestID)

	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(bodyBytes), "Streaming audio transcriptions are not supported")
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

func TestValidateOpenAIAudioTranscriptionModelFields(t *testing.T) {
	t.Run("gpt-transcribe rejects language and languages", func(t *testing.T) {
		form := &multipart.Form{Value: map[string][]string{
			"language":    {"en"},
			"languages[]": {"en", "fa"},
		}}
		err := validateOpenAIAudioTranscriptionModelFields("gpt-transcribe", form)
		require.EqualError(t, err, "language and languages cannot be used together for gpt-transcribe")
	})

	t.Run("gpt-transcribe rejects unsafe keywords", func(t *testing.T) {
		form := &multipart.Form{Value: map[string][]string{
			"keywords[]": {"safe", "bad\nkeyword"},
		}}
		err := validateOpenAIAudioTranscriptionModelFields("gpt-transcribe", form)
		require.EqualError(t, err, "keywords values cannot contain <, >, carriage returns, or line feeds")
	})

	t.Run("OpenAI transcription models enforce the 25 MB file limit", func(t *testing.T) {
		form := &multipart.Form{File: map[string][]*multipart.FileHeader{
			"file": {{Filename: "large.wav", Size: 25*1024*1024 + 1}},
		}}
		err := validateOpenAIAudioTranscriptionModelFields("gpt-4o-mini-transcribe", form)
		require.EqualError(t, err, "file must not exceed 25 MB")
	})

	t.Run("diarize rejects prompt and mismatched speakers", func(t *testing.T) {
		promptForm := &multipart.Form{Value: map[string][]string{"prompt": {"not supported"}}}
		err := validateOpenAIAudioTranscriptionModelFields("gpt-4o-transcribe-diarize", promptForm)
		require.EqualError(t, err, "prompt is not supported by gpt-4o-transcribe-diarize")

		speakerForm := &multipart.Form{Value: map[string][]string{
			"known_speaker_names[]":      {"Alice", "Bob"},
			"known_speaker_references[]": {"ref-a"},
		}}
		err = validateOpenAIAudioTranscriptionModelFields("gpt-4o-transcribe-diarize", speakerForm)
		require.EqualError(t, err, "known_speaker_names and known_speaker_references must have the same length")
	})
}

func TestAudioTranscriptionMinuteReconciliation(t *testing.T) {
	minutes, ok := audioTranscriptionMinutes([]byte(`{"usage":{"seconds":90}}`))
	require.True(t, ok)
	assert.Equal(t, "1.5", minutes)

	minutes, ok = audioTranscriptionMinutes([]byte(`{"usage":{"prompt_audio_seconds":203,"prompt_tokens":4}}`))
	require.True(t, ok)
	assert.Equal(t, "3.3833333333333333", minutes)

	streamUsage := &audioTranscriptionStreamUsage{}
	streamUsage.Observe([]byte("event: transcript.text.done\n"))
	streamUsage.Observe([]byte("data: {\"usage\":{\"seconds\":45}}\n\n"))
	minutes, ok = streamUsage.Minutes()
	require.True(t, ok)
	assert.Equal(t, "0.75", minutes)
}

func TestMistralAudioTranscriptionsHandlerAcceptsOfficialJSONFileURLContract(t *testing.T) {
	tunnel := &captureTunnel{resp: &platform.TunnelResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       []byte(`{"model":"voxtral-mini-2602","text":"ok","usage":{"prompt_audio_seconds":12,"prompt_tokens":2,"completion_tokens":1,"total_tokens":3}}`),
	}}
	logs := &testLogs{}
	model := "mistral-ai:acct:voxtral-mini-2602"
	app, _ := newAudioTranscriptionsTestApp(t, model, "mistral-ai", "voxtral-mini-2602", tunnel, logs, nil)

	body := `{"model":"mistral-ai:acct:voxtral-mini-2602","file_url":"https://example.com/audio.wav","stream":false,"timestamp_granularities":["word"]}`
	req := httptest.NewRequest("POST", "/v1/audio/transcriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	if resp.StatusCode != fiber.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected response %d: %s", resp.StatusCode, payload)
	}

	assert.Equal(t, "application/json", tunnel.lastReq.Headers["Content-Type"])
	assert.JSONEq(t, `{"model":"voxtral-mini-2602","file_url":"https://example.com/audio.wav","stream":false,"timestamp_granularities":["word"]}`, string(tunnel.lastReq.Body))
	assert.Equal(t, "url", tunnel.lastReq.RequestMetadata["file_source"])
	require.Len(t, logs.loggedSpends, 1)
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
