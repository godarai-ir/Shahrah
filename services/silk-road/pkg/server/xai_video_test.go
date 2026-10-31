package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func xaiVideoTestPayload(apiBase string, operation platform.ProviderOperation, body string, contentHosts ...string) platform.TunnelRequest {
	if len(contentHosts) == 0 {
		contentHosts = []string{"vidgen.x.ai"}
	}

	binding := platform.ModelRuntimeProfileBinding{
		Provider: "xai", CatalogRevision: "xai-video-test", CanonicalID: "xai/grok-imagine-video", Slug: "grok-imagine-video",
		ProfileVersion: 4, EffectiveFrom: time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version: 4, Endpoints: []string{"videos"}, Transports: []string{"http"}, Lifecycle: "active",
			Thinking:   platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
			Modalities: platform.ModelRuntimeModalities{Input: []string{"text", "image"}, Output: []string{"video"}},
			VideoControls: platform.ModelRuntimeVideoControls{
				Operations:         []string{platform.VideoOperationGenerate},
				MinDurationSeconds: 1, MaxDurationSeconds: 15, DefaultDurationSeconds: 8,
				AspectRatios: []string{"16:9"}, Resolutions: []string{"480p", "720p"}, DefaultResolution: "480p",
				ContentVariants: []string{"video"}, ContentHosts: contentHosts,
			},
			Limits: platform.ModelRuntimeLimits{MessagesBytes: 32 << 20, BatchBytes: 25 << 20, FileBytes: 48 << 20},
		},
	}

	return platform.TunnelRequest{
		Operation:            operation,
		OperationInput:       platform.ProviderOperationInput{ResourceID: "req-upstream-1"},
		Deployment:           platform.Deployment{Provider: "xai", Model: "grok-imagine-video", APIBase: apiBase, APIKey: "xai-secret"},
		ModelRuntimeProfiles: []platform.ModelRuntimeProfileBinding{binding},
		Headers:              map[string]string{"Content-Type": "application/json"},
		Body:                 []byte(body),
	}
}

func TestXAIVideoGenerationUsesNativePathAfterValidation(t *testing.T) {
	var gotPath, gotAuthorization, gotBody string

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotAuthorization = r.URL.Path, r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		_, _ = io.WriteString(w, `{"request_id":"req-upstream-1"}`)
	}))
	defer upstream.Close()

	body := `{"model":"grok-imagine-video","prompt":"waves","duration":4,"resolution":"720p"}`

	response, err := newXAIAdapter().Dispatch(context.Background(), upstream.Client(), xaiVideoTestPayload(upstream.URL+"/v1", platform.ProviderOperationXAIVideosGenerate, body))
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	defer response.Body.Close()

	if gotPath != "/v1/videos/generations" || gotAuthorization != "Bearer xai-secret" || gotBody != body {
		t.Fatalf("path=%q auth=%q body=%q", gotPath, gotAuthorization, gotBody)
	}
}

func TestXAIVideoGenerationRejectsInvalidBodiesBeforeUpstream(t *testing.T) {
	called := false

	upstream := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer upstream.Close()

	for _, body := range []string{
		`{"model":"grok-imagine-video","prompt":"x","output":{"upload_url":"https://evil.example/put"}}`,
		`{"model":"grok-imagine-video","prompt":"x","duration":30}`,
		`{"model":"grok-imagine-video","prompt":"x","resolution":"1080p"}`,
	} {
		_, err := newXAIAdapter().Dispatch(context.Background(), upstream.Client(), xaiVideoTestPayload(upstream.URL+"/v1", platform.ProviderOperationXAIVideosGenerate, body))
		if err == nil || !strings.Contains(err.Error(), "xai_invalid_video_request") {
			t.Fatalf("body %s: expected validation error, got %v", body, err)
		}
	}

	if called {
		t.Fatal("an invalid video request reached the provider")
	}
}

// newVideoContentServers starts a TLS media host and an API host whose poll
// points at it. The media handler is supplied by each test.
func newVideoContentServers(t *testing.T, media http.HandlerFunc, poll string) (*httptest.Server, *httptest.Server) {
	t.Helper()

	mediaServer := httptest.NewTLSServer(media)
	t.Cleanup(mediaServer.Close)

	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/videos/req-upstream-1" {
			http.NotFound(w, r)

			return
		}

		_, _ = io.WriteString(w, strings.ReplaceAll(poll, "MEDIA", mediaServer.URL))
	}))
	t.Cleanup(api.Close)

	return api, mediaServer
}

func mediaHost(t *testing.T, server *httptest.Server) string {
	t.Helper()

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse media url: %v", err)
	}

	return parsed.Hostname()
}

func TestXAIVideoContentStreamsFromAllowlistedHostWithoutCredentials(t *testing.T) {
	var mediaAuthorization string

	api, media := newVideoContentServers(t, func(w http.ResponseWriter, r *http.Request) {
		mediaAuthorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Set-Cookie", "tracking=1")
		_, _ = io.WriteString(w, "MP4DATA")
	}, `{"status":"done","video":{"url":"MEDIA/v.mp4","duration":4,"respect_moderation":true}}`)

	payload := xaiVideoTestPayload(api.URL+"/v1", platform.ProviderOperationXAIVideosContent, "", mediaHost(t, media))

	response, err := newXAIAdapter().(rawProviderAdapter).DispatchRaw(context.Background(), media.Client(), payload, nil, 0)
	if err != nil {
		t.Fatalf("content: %v", err)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	if string(body) != "MP4DATA" || response.Headers.Get("Content-Type") != "video/mp4" {
		t.Fatalf("body=%q headers=%v", body, response.Headers)
	}

	if mediaAuthorization != "" || response.Headers.Get("Set-Cookie") != "" {
		t.Fatalf("credentials or extra headers crossed the media boundary: auth=%q headers=%v", mediaAuthorization, response.Headers)
	}
}

func TestXAIVideoContentFailsClosed(t *testing.T) {
	redirect := func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://169.254.169.254/latest", http.StatusFound)
	}
	serve := func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "MP4DATA") }

	t.Run("host not allowlisted", func(t *testing.T) {
		api, media := newVideoContentServers(t, serve, `{"status":"done","video":{"url":"MEDIA/v.mp4","respect_moderation":true}}`)

		_, err := newXAIAdapter().(rawProviderAdapter).DispatchRaw(context.Background(), media.Client(), xaiVideoTestPayload(api.URL+"/v1", platform.ProviderOperationXAIVideosContent, ""), nil, 0)
		if err == nil || !strings.Contains(err.Error(), "xai_video_content_host_not_allowed") {
			t.Fatalf("expected host refusal, got %v", err)
		}
	})

	t.Run("redirect refused", func(t *testing.T) {
		api, media := newVideoContentServers(t, redirect, `{"status":"done","video":{"url":"MEDIA/v.mp4","respect_moderation":true}}`)

		_, err := newXAIAdapter().(rawProviderAdapter).DispatchRaw(context.Background(), media.Client(), xaiVideoTestPayload(api.URL+"/v1", platform.ProviderOperationXAIVideosContent, "", mediaHost(t, media)), nil, 0)
		if err == nil || !strings.Contains(err.Error(), "redirect_refused") {
			t.Fatalf("expected redirect refusal, got %v", err)
		}
	})

	for name, poll := range map[string]string{
		"pending":   `{"status":"pending","progress":10}`,
		"moderated": `{"status":"done","video":{"url":"MEDIA/v.mp4","respect_moderation":false}}`,
	} {
		t.Run(name, func(t *testing.T) {
			api, media := newVideoContentServers(t, serve, poll)

			response, err := newXAIAdapter().(rawProviderAdapter).DispatchRaw(context.Background(), media.Client(), xaiVideoTestPayload(api.URL+"/v1", platform.ProviderOperationXAIVideosContent, "", mediaHost(t, media)), nil, 0)
			if err != nil {
				t.Fatalf("content: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusConflict {
				t.Fatalf("status=%d, want 409", response.StatusCode)
			}
		})
	}
}
