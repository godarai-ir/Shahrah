package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/providernative"
	"github.com/Delpak/shahrah/services/control-plane/internal/router"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const videoTestRatePerSecond = 0.05

// videoTestStore is an in-memory videoResourceStore with the repository's
// claim and guarded-patch semantics.
type videoTestStore struct {
	mu      sync.Mutex
	records map[string]*platform.ProviderNativeResourceMapping
	stale   map[string]bool
}

func newVideoTestStore() *videoTestStore {
	return &videoTestStore{records: map[string]*platform.ProviderNativeResourceMapping{}, stale: map[string]bool{}}
}

func videoTestKey(tenantID, kind, publicID string) string {
	return tenantID + "|" + kind + "|" + publicID
}

func cloneVideoMapping(mapping *platform.ProviderNativeResourceMapping) *platform.ProviderNativeResourceMapping {
	encoded, _ := json.Marshal(mapping)
	copied := &platform.ProviderNativeResourceMapping{}
	_ = json.Unmarshal(encoded, copied)
	copied.UpstreamID = mapping.UpstreamID

	return copied
}

func (s *videoTestStore) UpsertProviderNativeResourceMapping(_ context.Context, mapping platform.ProviderNativeResourceMapping) (*platform.ProviderNativeResourceMapping, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[videoTestKey(mapping.TenantID, mapping.ResourceKind, mapping.PublicID)] = cloneVideoMapping(&mapping)

	return cloneVideoMapping(&mapping), nil
}

func (s *videoTestStore) GetProviderNativeResourceMapping(_ context.Context, tenantID, _, kind, publicID string) (*platform.ProviderNativeResourceMapping, error) {
	return s.find(tenantID, kind, publicID)
}

func (s *videoTestStore) UpdateProviderNativeResourceStatus(_ context.Context, tenantID, _, kind, publicID, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[videoTestKey(tenantID, kind, publicID)]
	if !ok {
		return providernative.ErrResourceNotFound
	}

	record.Status = status

	return nil
}

func (s *videoTestStore) ClaimProviderNativeResourceStatus(_ context.Context, tenantID, _, kind, publicID string, from []string, to string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := videoTestKey(tenantID, kind, publicID)

	record, ok := s.records[key]
	if !ok {
		return false, nil
	}

	for _, status := range from {
		if record.Status == status {
			record.Status = to

			return true, nil
		}
	}

	if record.Status == to && s.stale[key] {
		delete(s.stale, key)

		return true, nil
	}

	return false, nil
}

func (s *videoTestStore) ListPendingProviderNativeResourceMappings(_ context.Context, provider, kind string, statuses []string, _ int) ([]platform.ProviderNativeResourceMapping, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]platform.ProviderNativeResourceMapping, 0)

	for _, record := range s.records {
		if record.Provider == provider && record.ResourceKind == kind && containsFoldValue(statuses, record.Status) {
			result = append(result, *cloneVideoMapping(record))
		}
	}

	return result, nil
}

func (s *videoTestStore) FindTenantProviderNativeResourceMapping(_ context.Context, tenantID, kind, publicID string) (*platform.ProviderNativeResourceMapping, error) {
	return s.find(tenantID, kind, publicID)
}

func (s *videoTestStore) ListTenantProviderNativeResourceMappings(_ context.Context, tenantID, kind, after string, ascending bool, limit int) ([]platform.ProviderNativeResourceMapping, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]platform.ProviderNativeResourceMapping, 0)

	for _, record := range s.records {
		if _, deleted := record.Metadata[videoMetaDeletedAt]; record.TenantID == tenantID && record.ResourceKind == kind && !deleted {
			result = append(result, *cloneVideoMapping(record))
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if ascending {
			return result[i].PublicID < result[j].PublicID
		}

		return result[i].PublicID > result[j].PublicID
	})

	for index := range result {
		if after != "" && result[index].PublicID == after {
			result = result[index+1:]

			break
		}
	}

	if len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}

func (s *videoTestStore) PatchProviderNativeResourceState(_ context.Context, tenantID, _, kind, publicID, expected, status string, patch map[string]interface{}) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[videoTestKey(tenantID, kind, publicID)]
	if !ok || (expected != "" && record.Status != expected) {
		return false, nil
	}

	if status != "" {
		record.Status = status
	}

	for key, value := range patch {
		record.Metadata[key] = value
	}

	return true, nil
}

func (s *videoTestStore) find(tenantID, kind, publicID string) (*platform.ProviderNativeResourceMapping, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[videoTestKey(tenantID, kind, publicID)]
	if !ok {
		return nil, providernative.ErrResourceNotFound
	}

	return cloneVideoMapping(record), nil
}

func (s *videoTestStore) only(t *testing.T) *platform.ProviderNativeResourceMapping {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()

	require.Len(t, s.records, 1)

	for _, record := range s.records {
		return cloneVideoMapping(record)
	}

	return nil
}

// videoTestTunnel answers creates with a request_id, polls from a scripted
// table keyed by upstream ID, and content downloads with fixed bytes.
type videoTestTunnel struct {
	mu       sync.Mutex
	polls    map[string]string
	creates  int
	requests []platform.TunnelRequest
}

func (t *videoTestTunnel) Proxy(_ context.Context, req platform.TunnelRequest) (*platform.TunnelResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.requests = append(t.requests, req)

	switch req.Operation {
	case platform.ProviderOperationXAIVideosGet, fakeVideoPollOperation:
		body, ok := t.polls[req.OperationInput.ResourceID]
		if !ok {
			return &platform.TunnelResponse{StatusCode: fiber.StatusNotFound, Body: []byte(`{"error":"gone"}`)}, nil
		}

		return &platform.TunnelResponse{StatusCode: fiber.StatusOK, Body: []byte(body)}, nil
	default:
		t.creates++

		return &platform.TunnelResponse{StatusCode: fiber.StatusOK, Body: []byte(fmt.Sprintf(`{"request_id":"req_up_%d","job":"req_up_%d"}`, t.creates, t.creates))}, nil
	}
}

func (t *videoTestTunnel) ProxyRaw(_ context.Context, req platform.TunnelRequest, _ io.Reader, _ int64) (*platform.TunnelRawResponse, error) {
	t.mu.Lock()
	t.requests = append(t.requests, req)
	t.mu.Unlock()

	return &platform.TunnelRawResponse{
		StatusCode: fiber.StatusOK,
		Headers:    map[string]string{"Content-Type": "video/mp4", "Content-Length": "7"},
		Body:       io.NopCloser(strings.NewReader("MP4DATA")),
	}, nil
}

func (t *videoTestTunnel) Stream(context.Context, platform.TunnelRequest) (platform.TunnelStream, error) {
	return nil, errors.New("streaming is not used by videos")
}

func (t *videoTestTunnel) Healthy(context.Context) error { return nil }

func (t *videoTestTunnel) setPoll(upstreamID, body string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.polls[upstreamID] = body
}

func (t *videoTestTunnel) lastOf(operation platform.ProviderOperation) platform.TunnelRequest {
	t.mu.Lock()
	defer t.mu.Unlock()

	for index := len(t.requests) - 1; index >= 0; index-- {
		if t.requests[index].Operation == operation {
			return t.requests[index]
		}
	}

	return platform.TunnelRequest{}
}

func (t *videoTestTunnel) count() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return len(t.requests)
}

// videoTestBilling prices seconds at a flat rate and records reservations.
type videoTestBilling struct {
	testBudgeter
	mu       sync.Mutex
	profiles map[string]platform.ModelRuntimeProfileBinding
	reserved int
	commits  []platform.UsageCostBreakdown
	releases []string
}

func (b *videoTestBilling) profileFor(model string) (platform.ModelRuntimeProfileBinding, bool) {
	for _, binding := range b.profiles {
		if binding.MatchesModel(model) {
			return binding, true
		}
	}

	return platform.ModelRuntimeProfileBinding{}, false
}

func (b *videoTestBilling) PriceMeteredUsage(_ context.Context, model, dimension, quantity string, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	binding, ok := b.profileFor(model)
	if !ok || dimension != videoOutputSecondsDimension {
		return platform.UsageCostBreakdown{}, errors.New("pricing unavailable")
	}

	seconds, err := strconv.Atoi(quantity)
	if err != nil {
		return platform.UsageCostBreakdown{}, err
	}

	estimated.TotalCostUSD = float64(seconds) * videoTestRatePerSecond
	estimated.PricingRevision = binding.CatalogRevision
	estimated.ModelRuntimeProfile = &binding

	return estimated, nil
}

func (b *videoTestBilling) PriceVideoUsage(_ context.Context, _ string, seconds int, ticks *int64, estimated platform.UsageCostBreakdown) (platform.UsageCostBreakdown, error) {
	estimated.TotalCostUSD = float64(seconds) * videoTestRatePerSecond
	if ticks != nil {
		estimated.TotalCostUSD = float64(*ticks) / 1e10
	}

	return estimated, nil
}

func (b *videoTestBilling) ReserveBudget(_ context.Context, _ *platform.AuthContext, _ platform.UsageCostBreakdown, requestID string) (*platform.BudgetReservation, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.reserved++

	return &platform.BudgetReservation{ID: requestID}, nil
}

func (b *videoTestBilling) CommitBudget(_ context.Context, _ *platform.BudgetReservation, breakdown platform.UsageCostBreakdown) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.commits = append(b.commits, breakdown)

	return nil
}

func (b *videoTestBilling) ReleaseBudget(_ context.Context, _ *platform.BudgetReservation, reason string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.releases = append(b.releases, reason)

	return nil
}

func videoTestProfile(provider, model string, operations []string, controls func(*platform.ModelRuntimeVideoControls)) platform.ModelRuntimeProfileBinding {
	video := platform.ModelRuntimeVideoControls{
		Operations: operations, MinDurationSeconds: 1, MaxDurationSeconds: 15, DefaultDurationSeconds: 8,
		MinExtensionSeconds: 2, MaxExtensionSeconds: 10, DefaultExtensionSeconds: 6, MaxEditSourceSeconds: 8.7,
		AspectRatios: []string{"16:9", "9:16", "1:1"}, Resolutions: []string{"480p", "720p"}, DefaultResolution: "480p",
		Sizes: map[string]platform.ModelRuntimeVideoSize{
			"1280x720": {AspectRatio: "16:9", Resolution: "720p"},
			"720x1280": {AspectRatio: "9:16", Resolution: "720p"},
		},
		ReferenceModes: []string{platform.VideoReferenceFirstFrame}, MaxInputImageBytes: 1024,
		ContentVariants: []string{"video"}, ContentHosts: []string{"vidgen.x.ai"},
		ExtensionFields: []string{"aspect_ratio", "resolution"},
	}
	if controls != nil {
		controls(&video)
	}

	return platform.ModelRuntimeProfileBinding{
		Provider: provider, CatalogRevision: provider + "-video-revision", CanonicalID: provider + "/" + model, Slug: model,
		ProfileVersion: 4, EffectiveFrom: time.Now().UTC(),
		Profile: platform.ModelRuntimeProfile{
			Version: 4, Endpoints: []string{"videos"}, Transports: []string{"http"}, Lifecycle: "active",
			Thinking:        platform.ModelRuntimeThinkingProfile{DefaultMode: "disabled", AllowedModes: []string{"disabled"}},
			Modalities:      platform.ModelRuntimeModalities{Input: []string{"text", "image"}, Output: []string{"video"}},
			RequestControls: platform.ModelRuntimeRequestControls{InferenceGeoValues: []string{"global"}},
			VideoControls:   video,
			Limits:          platform.ModelRuntimeLimits{MessagesBytes: 32 << 20, BatchBytes: 25 << 20, FileBytes: 48 << 20},
		},
	}
}

// videoTestClock lets tests step past the GET refresh throttle.
type videoTestClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *videoTestClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now
}

func (c *videoTestClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.now = c.now.Add(d)
}

type videoTestEnv struct {
	clock   *videoTestClock
	app     *fiber.App
	handler *VideosHandler
	store   *videoTestStore
	tunnel  *videoTestTunnel
	billing *videoTestBilling
	logs    *testLogs
}

type videoTestModel struct {
	name, provider, deploymentModel string
	binding                         platform.ModelRuntimeProfileBinding
}

func newVideoTestEnv(t *testing.T, features ProviderFeatureConfig, models ...videoTestModel) *videoTestEnv {
	t.Helper()

	if len(models) == 0 {
		models = []videoTestModel{{
			name: "grok-imagine-video", provider: "xai", deploymentModel: "grok-imagine-video",
			binding: videoTestProfile("xai", "grok-imagine-video", []string{"generate", "edit", "extend"}, nil),
		}}
	}

	mem, err := testcache.New(t, zap.NewNop())
	require.NoError(t, err)

	deployments := make([]config.ModelDeployment, 0, len(models))
	billing := &videoTestBilling{profiles: map[string]platform.ModelRuntimeProfileBinding{}}
	allowed := map[string][]string{}

	for _, model := range models {
		deployments = append(deployments, config.ModelDeployment{
			ModelName:        model.name,
			ProviderEndpoint: config.ProviderEndpointConfig{Provider: model.provider, Model: model.deploymentModel, APIKey: "test-key"},
		})
		billing.profiles[model.name] = model.binding
		allowed[model.name] = []string{string(platform.ModelTypeVideoGeneration)}
	}

	rt, err := router.NewRouter(deployments, config.RouterSettings{CooldownTime: 60, NumRetries: 1}, mem, nil, zap.NewNop())
	require.NoError(t, err)

	store, tunnel, logs := newVideoTestStore(), &videoTestTunnel{polls: map[string]string{}}, &testLogs{}
	features.Configured = true
	handler := NewVideosHandler(BaseHandlerDependencies{
		Router: rt, Tunnel: tunnel, RawTunnel: tunnel, Billing: billing,
		BillingCapabilities: BillingCapabilities{
			MeteredUsage: billing, MeteredUsagePolicy: CapabilityFailClosed,
			BudgetReservations: billing, BudgetReservationPolicy: CapabilityFailClosed,
			VideoUsage: billing,
		},
		RequestLogger: logs, SpendLogger: logs, Logger: zap.NewNop(), ProviderFeatures: features,
	}, store)

	clock := &videoTestClock{now: time.Now()}
	handler.now = clock.Now

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		tenantID := c.Get("X-Test-Tenant", "tenant_a")
		c.Locals(auth.AuthContextKey{}, &platform.AuthContext{
			TenantID: tenantID, KeyID: "key_1", KeyType: "pat", Role: "member", TokenHash: "hash_1",
			AllowedModelTypes: allowed,
		})
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: tenantID, Slug: tenantID, Status: "active"})

		return c.Next()
	})
	app.Post("/v1/videos", handler.HandleCreate)
	app.Get("/v1/videos", handler.HandleList)
	app.Post("/v1/videos/generations", handler.HandleGenerations)
	app.Post("/v1/videos/edits", handler.HandleEdits)
	app.Post("/v1/videos/extensions", handler.HandleExtensions)
	app.Get("/v1/videos/:video_id", handler.HandleGet)
	app.Delete("/v1/videos/:video_id", handler.HandleDelete)
	app.Get("/v1/videos/:video_id/content", handler.HandleContent)

	return &videoTestEnv{clock: clock, app: app, handler: handler, store: store, tunnel: tunnel, billing: billing, logs: logs}
}

func (e *videoTestEnv) do(t *testing.T, method, path, body, tenantID string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	if tenantID != "" {
		req.Header.Set("X-Test-Tenant", tenantID)
	}

	resp, err := e.app.Test(req, 5000)
	require.NoError(t, err)

	defer closeResponseBody(t, resp)

	received, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(received)
}

func (e *videoTestEnv) create(t *testing.T, body string) map[string]interface{} {
	t.Helper()

	status, received := e.do(t, "POST", "/v1/videos", body, "")
	require.Equal(t, fiber.StatusOK, status, received)

	var video map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(received), &video))

	return video
}

func (e *videoTestEnv) spendCount() int {
	e.logs.mu.RLock()
	defer e.logs.mu.RUnlock()

	return len(e.logs.loggedSpends)
}

var videoTestEnabled = ProviderFeatureConfig{VideosEnabled: true}

func TestVideosCanonicalCreateTranslatesAndReserves(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)

	video := env.create(t, `{"model":"grok-imagine-video","prompt":"a lighthouse at dawn","seconds":"4","size":"1280x720"}`)

	assert.Equal(t, "video", video["object"])
	assert.Equal(t, "queued", video["status"])
	assert.Equal(t, "4", video["seconds"])
	assert.Equal(t, "1280x720", video["size"])
	assert.True(t, strings.HasPrefix(video["id"].(string), "video_"))
	assert.Equal(t, 1, env.billing.reserved)

	sent := env.tunnel.lastOf(platform.ProviderOperationXAIVideosGenerate)

	var upstream map[string]interface{}
	require.NoError(t, json.Unmarshal(sent.Body, &upstream))
	assert.Equal(t, map[string]interface{}{
		"model": "grok-imagine-video", "prompt": "a lighthouse at dawn", "duration": float64(4),
		"aspect_ratio": "16:9", "resolution": "720p",
	}, upstream)
	require.Len(t, sent.ModelRuntimeProfiles, 1)

	record := env.store.only(t)
	assert.Equal(t, "req_up_1", record.UpstreamID)
	assert.InDelta(t, 4*videoTestRatePerSecond, record.EstimatedUsage.TotalCostUSD, 1e-9)
}

func TestVideosRejectInvalidRequestsBeforeReserving(t *testing.T) {
	lite := videoTestProfile("xai", "grok-imagine-video-1.5-lite", []string{"generate"}, nil)

	for name, test := range map[string]struct{ path, body, want string }{
		"duration out of range": {"/v1/videos", `{"model":"grok-imagine-video-1.5-lite","prompt":"x","seconds":"30"}`, "duration must be between 1 and 15"},
		"numeric fraction":      {"/v1/videos", `{"model":"grok-imagine-video-1.5-lite","prompt":"x","seconds":"4.5"}`, "whole number"},
		"unknown size":          {"/v1/videos", `{"model":"grok-imagine-video-1.5-lite","prompt":"x","size":"1920x1080"}`, "size"},
		"unlisted extension":    {"/v1/videos", `{"model":"grok-imagine-video-1.5-lite","prompt":"x","keyframes":[]}`, "keyframes"},
		"edit not supported":    {"/v1/videos/edits", `{"model":"grok-imagine-video-1.5-lite","prompt":"x","video":{"url":"https://example.com/a.mp4"}}`, "does not support video edit"},
		"upload url rejected":   {"/v1/videos/generations", `{"model":"grok-imagine-video-1.5-lite","prompt":"x","output":{"upload_url":"https://evil.example/put"}}`, "output"},
		"voice clip rejected":   {"/v1/videos/generations", `{"model":"grok-imagine-video-1.5-lite","prompt":"x","reference_audios":[{"url":"https://x/a.wav"}]}`, "voice_id"},
		"inline image too big":  {"/v1/videos", `{"model":"grok-imagine-video-1.5-lite","prompt":"x","input_reference":{"image_url":"data:image/png;base64,` + strings.Repeat("A", 4096) + `"}}`, "at most 1024 bytes"},
	} {
		t.Run(name, func(t *testing.T) {
			env := newVideoTestEnv(t, videoTestEnabled, videoTestModel{"grok-imagine-video-1.5-lite", "xai", "grok-imagine-video-1.5-lite", lite})

			status, body := env.do(t, "POST", test.path, test.body, "")

			assert.Equal(t, fiber.StatusBadRequest, status, body)
			assert.Contains(t, body, test.want)
			assert.Zero(t, env.billing.reserved)
			assert.Zero(t, env.tunnel.count())
		})
	}
}

func TestVideosSettleOnceFromProviderReportedCost(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)
	video := env.create(t, `{"model":"grok-imagine-video","prompt":"waves","seconds":"4"}`)
	id := video["id"].(string)

	env.tunnel.setPoll("req_up_1", `{"status":"pending","progress":40}`)
	status, body := env.do(t, "GET", "/v1/videos/"+id, "", "")
	require.Equal(t, fiber.StatusOK, status)
	assert.Contains(t, body, `"status":"in_progress"`)
	assert.Contains(t, body, `"progress":40`)
	assert.Empty(t, env.billing.commits)

	env.tunnel.setPoll("req_up_1", `{"status":"done","progress":100,"video":{"url":"https://vidgen.x.ai/v.mp4","duration":4,"respect_moderation":true},"usage":{"cost_in_usd_ticks":1800000000}}`)
	env.clock.advance(minVideoRefreshInterval)
	status, body = env.do(t, "GET", "/v1/videos/"+id, "", "")
	require.Equal(t, fiber.StatusOK, status)
	assert.Contains(t, body, `"status":"completed"`)
	assert.NotContains(t, body, "req_up_")
	assert.NotContains(t, body, "vidgen.x.ai")

	env.handler.reconcileVideos(context.Background())
	_, _ = env.do(t, "GET", "/v1/videos/"+id, "", "")

	require.Len(t, env.billing.commits, 1)
	assert.InDelta(t, 0.18, env.billing.commits[0].TotalCostUSD, 1e-9)
	assert.Empty(t, env.billing.releases)
	assert.Equal(t, 1, env.spendCount())
	assert.Equal(t, "provider_reported_charge", env.store.only(t).Metadata[videoMetaSettlement])
}

func TestVideosReleaseFailedAndModeratedJobs(t *testing.T) {
	for name, poll := range map[string]string{
		"failed":    `{"status":"failed","error":{"code":"internal_error","message":"boom"}}`,
		"moderated": `{"status":"done","video":{"url":"","duration":4,"respect_moderation":false}}`,
	} {
		t.Run(name, func(t *testing.T) {
			env := newVideoTestEnv(t, videoTestEnabled)
			video := env.create(t, `{"model":"grok-imagine-video","prompt":"waves"}`)
			env.tunnel.setPoll("req_up_1", poll)

			status, body := env.do(t, "GET", "/v1/videos/"+video["id"].(string), "", "")

			require.Equal(t, fiber.StatusOK, status)
			assert.Contains(t, body, `"status":"failed"`)
			assert.Empty(t, env.billing.commits)
			assert.Len(t, env.billing.releases, 1)
			assert.Zero(t, env.spendCount())

			if name == "moderated" {
				assert.Contains(t, body, "content_policy_violation")
			}
		})
	}
}

func TestVideosReleaseReservationWhenUpstreamJobIsGone(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)
	video := env.create(t, `{"model":"grok-imagine-video","prompt":"waves","seconds":"2"}`)

	status, body := env.do(t, "GET", "/v1/videos/"+video["id"].(string), "", "")

	require.Equal(t, fiber.StatusOK, status)
	assert.Contains(t, body, "upstream_expired")
	assert.Empty(t, env.billing.commits)
	assert.Equal(t, []string{"video_upstream_expired"}, env.billing.releases)
}

func TestVideosClientPollsAreThrottledToTheProvider(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)
	video := env.create(t, `{"model":"grok-imagine-video","prompt":"waves"}`)
	path := "/v1/videos/" + video["id"].(string)

	env.tunnel.setPoll("req_up_1", `{"status":"pending","progress":10}`)

	polls := func() int {
		count := 0

		for _, req := range env.tunnel.requests {
			if req.Operation == platform.ProviderOperationXAIVideosGet {
				count++
			}
		}

		return count
	}

	for i := 0; i < 3; i++ {
		_, body := env.do(t, "GET", path, "", "")
		assert.Contains(t, body, `"status":"in_progress"`)
	}

	assert.Equal(t, 1, polls())

	env.clock.advance(minVideoRefreshInterval)
	_, _ = env.do(t, "GET", path, "", "")
	assert.Equal(t, 2, polls())
}

func TestVideosInterruptedSettlementIsNeverChargedTwice(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)
	video := env.create(t, `{"model":"grok-imagine-video","prompt":"waves"}`)
	id := video["id"].(string)

	record := env.store.only(t)
	record.Status = videoStatusSettling
	_, err := env.store.UpsertProviderNativeResourceMapping(context.Background(), *record)
	require.NoError(t, err)

	env.store.stale[videoTestKey("tenant_a", videoResourceKind, id)] = true

	env.handler.reconcileVideos(context.Background())

	assert.Empty(t, env.billing.commits)
	assert.Equal(t, []string{"video_settlement_interrupted"}, env.billing.releases)
	assert.Zero(t, env.spendCount())
	assert.Equal(t, videoStatusCompleted, env.store.only(t).Status)
	assert.Equal(t, "interrupted", env.store.only(t).Metadata[videoMetaSettlement])
}

func TestVideosAreTenantIsolated(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)
	video := env.create(t, `{"model":"grok-imagine-video","prompt":"waves"}`)
	id := video["id"].(string)

	for _, request := range []struct{ method, path string }{
		{"GET", "/v1/videos/" + id}, {"GET", "/v1/videos/" + id + "/content"}, {"DELETE", "/v1/videos/" + id},
	} {
		foreignStatus, foreignBody := env.do(t, request.method, request.path, "", "tenant_b")
		unknownStatus, unknownBody := env.do(t, request.method, strings.Replace(request.path, id, "video_unknown", 1), "", "tenant_b")

		assert.Equal(t, fiber.StatusNotFound, foreignStatus)
		assert.Equal(t, unknownBody, foreignBody, "foreign and unknown IDs must be indistinguishable")
		assert.Equal(t, unknownStatus, foreignStatus)
	}

	status, body := env.do(t, "POST", "/v1/videos/edits", `{"model":"grok-imagine-video","prompt":"x","video":{"id":"`+id+`"}}`, "tenant_b")
	assert.Equal(t, fiber.StatusBadRequest, status, body)

	_, list := env.do(t, "GET", "/v1/videos", "", "tenant_b")
	assert.Contains(t, list, `"data":[]`)
}

func TestVideosListPaginatesAndHidesDeleted(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)
	ids := make([]string, 0, 3)

	for range 3 {
		ids = append(ids, env.create(t, `{"model":"grok-imagine-video","prompt":"waves"}`)["id"].(string))
	}

	status, _ := env.do(t, "DELETE", "/v1/videos/"+ids[0], "", "")
	require.Equal(t, fiber.StatusOK, status)

	_, body := env.do(t, "GET", "/v1/videos?limit=1", "", "")

	var page map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(body), &page))
	assert.Equal(t, true, page["has_more"])
	assert.Len(t, page["data"], 1)

	_, all := env.do(t, "GET", "/v1/videos?limit=10", "", "")
	assert.NotContains(t, all, ids[0])

	missing, _ := env.do(t, "GET", "/v1/videos/"+ids[0], "", "")
	assert.Equal(t, fiber.StatusNotFound, missing)
}

func TestVideosDeletedPendingJobStillSettles(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)
	video := env.create(t, `{"model":"grok-imagine-video","prompt":"waves","seconds":"4"}`)

	status, _ := env.do(t, "DELETE", "/v1/videos/"+video["id"].(string), "", "")
	require.Equal(t, fiber.StatusOK, status)

	env.tunnel.setPoll("req_up_1", `{"status":"done","video":{"url":"https://vidgen.x.ai/v.mp4","duration":4,"respect_moderation":true},"usage":{"cost_in_usd_ticks":2000000000}}`)
	env.handler.reconcileVideos(context.Background())

	require.Len(t, env.billing.commits, 1)
	assert.Contains(t, env.store.only(t).Metadata, videoMetaDeletedAt)
}

func TestVideosContentStreamsCompletedVideos(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)
	id := env.create(t, `{"model":"grok-imagine-video","prompt":"waves"}`)["id"].(string)

	status, body := env.do(t, "GET", "/v1/videos/"+id+"/content", "", "")
	assert.Equal(t, fiber.StatusConflict, status, body)

	env.tunnel.setPoll("req_up_1", `{"status":"done","video":{"url":"https://vidgen.x.ai/v.mp4","duration":8,"respect_moderation":true},"usage":{"cost_in_usd_ticks":4000000000}}`)
	env.handler.reconcileVideos(context.Background())

	status, body = env.do(t, "GET", "/v1/videos/"+id+"/content?variant=thumbnail", "", "")
	assert.Equal(t, fiber.StatusBadRequest, status)
	assert.Contains(t, body, "unsupported_variant")

	req := httptest.NewRequest("GET", "/v1/videos/"+id+"/content", http.NoBody)
	resp, err := env.app.Test(req, 5000)
	require.NoError(t, err)

	defer closeResponseBody(t, resp)

	received, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "MP4DATA", string(received))
	assert.Equal(t, "video/mp4", resp.Header.Get("Content-Type"))
	assert.Equal(t, int64(7), resp.ContentLength)

	raw := env.tunnel.lastOf(platform.ProviderOperationXAIVideosContent)
	assert.Equal(t, "req_up_1", raw.OperationInput.ResourceID)
}

func TestVideosNativeShapeRoundTrip(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)

	status, body := env.do(t, "POST", "/v1/videos/generations", `{"model":"grok-imagine-video","prompt":"waves","duration":3}`, "")
	require.Equal(t, fiber.StatusOK, status, body)

	var created map[string]string
	require.NoError(t, json.Unmarshal([]byte(body), &created))
	assert.True(t, strings.HasPrefix(created["request_id"], "video_"))

	env.tunnel.setPoll("req_up_1", `{"status":"pending","progress":0}`)
	_, polled := env.do(t, "GET", "/v1/videos/"+created["request_id"], "", "")
	assert.Contains(t, polled, `"status":"pending"`)

	env.clock.advance(minVideoRefreshInterval)
	env.tunnel.setPoll("req_up_1", `{"status":"done","video":{"url":"https://vidgen.x.ai/v.mp4","duration":3,"respect_moderation":true},"usage":{"cost_in_usd_ticks":1500000000}}`)
	_, done := env.do(t, "GET", "/v1/videos/"+created["request_id"], "", "")
	assert.Contains(t, done, `"status":"done"`)
	assert.Contains(t, done, "/v1/videos/"+created["request_id"]+"/content")
	assert.NotContains(t, done, "vidgen.x.ai")
}

func TestVideosEditResolvesOwnCompletedSource(t *testing.T) {
	env := newVideoTestEnv(t, videoTestEnabled)
	id := env.create(t, `{"model":"grok-imagine-video","prompt":"waves"}`)["id"].(string)

	status, body := env.do(t, "POST", "/v1/videos/edits", `{"model":"grok-imagine-video","prompt":"add a boat","video":{"id":"`+id+`"}}`, "")
	assert.Equal(t, fiber.StatusBadRequest, status, body)

	env.tunnel.setPoll("req_up_1", `{"status":"done","video":{"url":"https://vidgen.x.ai/source.mp4","duration":6,"respect_moderation":true},"usage":{"cost_in_usd_ticks":3000000000}}`)
	env.handler.reconcileVideos(context.Background())

	status, body = env.do(t, "POST", "/v1/videos/edits", `{"model":"grok-imagine-video","prompt":"add a boat","video":{"id":"`+id+`"}}`, "")
	require.Equal(t, fiber.StatusOK, status, body)

	var upstream map[string]interface{}
	require.NoError(t, json.Unmarshal(env.tunnel.lastOf(platform.ProviderOperationXAIVideosEdit).Body, &upstream))
	assert.Equal(t, map[string]interface{}{"url": "https://vidgen.x.ai/source.mp4"}, upstream["video"])
	assert.NotContains(t, body, "vidgen.x.ai")
}

func TestVideosRolloutGate(t *testing.T) {
	disabled := newVideoTestEnv(t, ProviderFeatureConfig{})
	status, _ := disabled.do(t, "POST", "/v1/videos", `{"model":"grok-imagine-video","prompt":"x"}`, "")
	assert.Equal(t, fiber.StatusNotFound, status)

	canary := newVideoTestEnv(t, ProviderFeatureConfig{VideosEnabled: true, VideoCanaryTenantIDs: []string{"tenant_a"}})
	status, _ = canary.do(t, "GET", "/v1/videos", "", "tenant_b")
	assert.Equal(t, fiber.StatusNotFound, status)

	status, _ = canary.do(t, "GET", "/v1/videos", "", "tenant_a")
	assert.Equal(t, fiber.StatusOK, status)
}

const fakeVideoPollOperation platform.ProviderOperation = "fakevideo_jobs_get"

// fakeVideoCodec is a second provider with a different native state machine,
// proving the seam: it plugs in with no route, worker, or handler change.
type fakeVideoCodec struct{}

func (fakeVideoCodec) createOperation(string) platform.ProviderOperation {
	return "fakevideo_jobs_create"
}

func (fakeVideoCodec) pollOperation() platform.ProviderOperation { return fakeVideoPollOperation }

func (fakeVideoCodec) contentOperation() platform.ProviderOperation {
	return "fakevideo_jobs_content"
}

func (fakeVideoCodec) buildCreateBody(request canonicalVideoRequest, _ platform.ModelRuntimeProfile) (map[string]interface{}, error) {
	return map[string]interface{}{"model": request.Model, "text": request.Prompt, "length": request.Seconds}, nil
}

func (fakeVideoCodec) validateBody(operation string, body []byte, profile platform.ModelRuntimeProfile) (platform.VideoRequestFacts, error) {
	var request struct {
		Text   string `json:"text"`
		Length int    `json:"length"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return platform.VideoRequestFacts{}, err
	}

	facts := platform.VideoRequestFacts{Operation: operation, HasPrompt: request.Text != "", DurationSeconds: request.Length}

	return facts, profile.ValidateVideoRequest(facts)
}

func (fakeVideoCodec) parseCreateResponse(body []byte) (string, error) {
	var response struct {
		Job string `json:"job"`
	}

	return response.Job, json.Unmarshal(body, &response)
}

func (fakeVideoCodec) parseJob(body []byte) (videoJobState, error) {
	var job struct {
		Phase   string `json:"phase"`
		Seconds int    `json:"seconds"`
	}
	if err := json.Unmarshal(body, &job); err != nil {
		return videoJobState{}, err
	}

	switch job.Phase {
	case "RENDERING":
		return videoJobState{Status: videoStatusInProgress, Progress: 50}, nil
	case "READY":
		return videoJobState{Status: videoStatusCompleted, Progress: 100, Seconds: job.Seconds}, nil
	default:
		return videoJobState{Status: videoStatusFailed, ErrorCode: "upstream_failed"}, nil
	}
}

func (fakeVideoCodec) sourceReference([]byte) (map[string]interface{}, error) {
	return nil, errors.New("fake provider has no edits")
}

func TestVideosSecondProviderPlugsInThroughTheCodecOnly(t *testing.T) {
	binding := videoTestProfile("fakevideo", "clip-1", []string{"generate"}, func(controls *platform.ModelRuntimeVideoControls) {
		controls.ExtensionFields = nil
	})
	env := newVideoTestEnv(t, videoTestEnabled, videoTestModel{"clip-1", "fakevideo", "clip-1", binding})
	env.handler.codecs["fakevideo"] = fakeVideoCodec{}

	id := env.create(t, `{"model":"clip-1","prompt":"a fox","seconds":"5"}`)["id"].(string)
	assert.Equal(t, "fakevideo_jobs_create", string(env.tunnel.lastOf("fakevideo_jobs_create").Operation))

	env.tunnel.setPoll("req_up_1", `{"phase":"RENDERING"}`)
	_, body := env.do(t, "GET", "/v1/videos/"+id, "", "")
	assert.Contains(t, body, `"status":"in_progress"`)

	env.tunnel.setPoll("req_up_1", `{"phase":"READY","seconds":5}`)
	env.handler.reconcileVideos(context.Background())

	require.Len(t, env.billing.commits, 1)
	assert.InDelta(t, 5*videoTestRatePerSecond, env.billing.commits[0].TotalCostUSD, 1e-9)
	assert.Equal(t, "estimated_settlement", env.store.only(t).Metadata[videoMetaSettlement])

	status, content := env.do(t, "GET", "/v1/videos/"+id+"/content", "", "")
	assert.Equal(t, fiber.StatusOK, status)
	assert.Equal(t, "MP4DATA", content)
}
