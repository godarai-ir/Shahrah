package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

const videoAuthorizationTenantID = "00000000-0000-0000-0000-000000000001"

type videoAuthorizationStore struct {
	jobs              []platform.VideoGenerationJob
	getScopes         []storage.VideoJobAccessScope
	idempotencyScopes []storage.VideoJobAccessScope
	saveCalls         int
}

func (s *videoAuthorizationStore) CreateVideoJob(_ context.Context, job *platform.VideoGenerationJob) (*platform.VideoGenerationJob, error) {
	s.jobs = append(s.jobs, *job)
	return job, nil
}

func (s *videoAuthorizationStore) GetVideoJob(_ context.Context, scope storage.VideoJobAccessScope, jobID string) (*platform.VideoGenerationJob, error) {
	s.getScopes = append(s.getScopes, scope)
	for i := range s.jobs {
		if s.jobs[i].ID == jobID && videoJobVisibleToScope(scope, s.jobs[i]) {
			job := s.jobs[i]
			return &job, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (s *videoAuthorizationStore) GetVideoJobByIdempotency(_ context.Context, scope storage.VideoJobAccessScope, key string) (*platform.VideoGenerationJob, error) {
	s.idempotencyScopes = append(s.idempotencyScopes, scope)
	for i := range s.jobs {
		if s.jobs[i].IdempotencyKey == key &&
			s.jobs[i].TenantID == scope.TenantID &&
			s.jobs[i].OwnerPrincipal == scope.OwnerPrincipal {
			job := s.jobs[i]
			return &job, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (s *videoAuthorizationStore) GetVideoJobByProviderID(context.Context, string, string) (*platform.VideoGenerationJob, error) {
	return nil, pgx.ErrNoRows
}

func (s *videoAuthorizationStore) ListVideoJobs(_ context.Context, scope storage.VideoJobAccessScope, _ int, _ time.Time) ([]platform.VideoGenerationJob, error) {
	result := make([]platform.VideoGenerationJob, 0, len(s.jobs))
	for i := range s.jobs {
		if videoJobVisibleToScope(scope, s.jobs[i]) {
			result = append(result, s.jobs[i])
		}
	}
	return result, nil
}

func (s *videoAuthorizationStore) SaveVideoJob(context.Context, *platform.VideoGenerationJob, int64) error {
	s.saveCalls++
	return nil
}

func (s *videoAuthorizationStore) MarkVideoSubmission(context.Context, string, int, string, map[string]interface{}, string) error {
	return nil
}

func videoJobVisibleToScope(scope storage.VideoJobAccessScope, job platform.VideoGenerationJob) bool {
	return job.TenantID == scope.TenantID &&
		(scope.AllowTenantWide || (scope.OwnerPrincipal != "" && job.OwnerPrincipal == scope.OwnerPrincipal))
}

type videoAuthorizationObjectStore struct {
	presignCalls int
	deleteCalls  int
}

func (s *videoAuthorizationObjectStore) PutInput(context.Context, string, string, string, string, io.Reader) (platform.VideoArtifact, error) {
	return platform.VideoArtifact{}, nil
}

func (s *videoAuthorizationObjectStore) CopyFromURL(context.Context, string, string, platform.VideoArtifact) (platform.VideoArtifact, error) {
	return platform.VideoArtifact{}, nil
}

func (s *videoAuthorizationObjectStore) PresignGet(context.Context, string, time.Duration) (string, error) {
	s.presignCalls++
	return "https://example.invalid/video", nil
}

func (s *videoAuthorizationObjectStore) Delete(context.Context, string) error {
	s.deleteCalls++
	return nil
}

func TestVideoHandlersHideCrossOwnerJobsWithoutSideEffects(t *testing.T) {
	t.Parallel()

	store := &videoAuthorizationStore{jobs: []platform.VideoGenerationJob{videoAuthorizationJob(
		"user:user-b",
		videoAuthorizationTenantID,
	)}}
	objects := &videoAuthorizationObjectStore{}
	app := videoAuthorizationApp(
		&platform.AuthContext{UserID: "user-a", TokenHash: "shared-token", Role: "member"},
		videoAuthorizationTenantID,
		store,
		objects,
	)

	tests := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/v1/videos/job-b"},
		{method: http.MethodGet, path: "/v1/videos/job-b/content"},
		{method: http.MethodPost, path: "/v1/videos/job-b/cancel"},
		{method: http.MethodDelete, path: "/v1/videos/job-b"},
	}
	for _, test := range tests {
		response, body := performVideoAuthorizationRequest(t, app, test.method, test.path, nil, nil)
		require.Equal(t, fiber.StatusNotFound, response.StatusCode, "%s %s: %s", test.method, test.path, body)
		require.NotContains(t, body, "private prompt")
	}

	response, body := performVideoAuthorizationRequest(t, app, http.MethodGet, "/v1/videos", nil, nil)
	require.Equal(t, fiber.StatusOK, response.StatusCode, body)
	require.NotContains(t, body, "private prompt")
	require.Contains(t, body, `"data":[]`)
	require.Zero(t, objects.presignCalls)
	require.Zero(t, objects.deleteCalls)
	require.Zero(t, store.saveCalls)
}

func TestVideoOwnerCanReadAcrossTokenRotationButSharedTokenDoesNotCrossActors(t *testing.T) {
	t.Parallel()

	store := &videoAuthorizationStore{jobs: []platform.VideoGenerationJob{videoAuthorizationJob(
		"user:user-a",
		videoAuthorizationTenantID,
	)}}
	ownerApp := videoAuthorizationApp(
		&platform.AuthContext{UserID: "user-a", TokenHash: "rotated-token", Role: "member"},
		videoAuthorizationTenantID,
		store,
		&videoAuthorizationObjectStore{},
	)
	response, body := performVideoAuthorizationRequest(t, ownerApp, http.MethodGet, "/v1/videos/job-b", nil, nil)
	require.Equal(t, fiber.StatusOK, response.StatusCode, body)
	require.Contains(t, body, "private prompt")
	require.NotContains(t, body, "owner_principal")
	require.NotContains(t, body, "user:user-a")

	otherActorApp := videoAuthorizationApp(
		&platform.AuthContext{UserID: "user-b", TokenHash: "rotated-token", Role: "member"},
		videoAuthorizationTenantID,
		store,
		&videoAuthorizationObjectStore{},
	)
	response, body = performVideoAuthorizationRequest(t, otherActorApp, http.MethodGet, "/v1/videos/job-b", nil, nil)
	require.Equal(t, fiber.StatusNotFound, response.StatusCode, body)
	require.NotContains(t, body, "private prompt")
}

func TestVideoTenantAndProxyAdminsCanReadTenantJobs(t *testing.T) {
	t.Parallel()

	for _, role := range []string{"tenant_admin", "proxy_admin"} {
		role := role
		t.Run(role, func(t *testing.T) {
			t.Parallel()
			store := &videoAuthorizationStore{jobs: []platform.VideoGenerationJob{videoAuthorizationJob(
				"user:user-a",
				videoAuthorizationTenantID,
			)}}
			app := videoAuthorizationApp(
				&platform.AuthContext{Role: role},
				videoAuthorizationTenantID,
				store,
				&videoAuthorizationObjectStore{},
			)
			response, body := performVideoAuthorizationRequest(t, app, http.MethodGet, "/v1/videos/job-b", nil, nil)
			require.Equal(t, fiber.StatusOK, response.StatusCode, body)
			require.True(t, store.getScopes[len(store.getScopes)-1].AllowTenantWide)
		})
	}
}

func TestVideoOrgAndTeamAdminsRemainOwnerScoped(t *testing.T) {
	t.Parallel()

	for _, role := range []string{"org_admin", "team_admin"} {
		role := role
		t.Run(role, func(t *testing.T) {
			t.Parallel()
			store := &videoAuthorizationStore{jobs: []platform.VideoGenerationJob{videoAuthorizationJob(
				"user:user-a",
				videoAuthorizationTenantID,
			)}}
			app := videoAuthorizationApp(
				&platform.AuthContext{Role: role, UserID: "user-b"},
				videoAuthorizationTenantID,
				store,
				&videoAuthorizationObjectStore{},
			)
			response, body := performVideoAuthorizationRequest(t, app, http.MethodGet, "/v1/videos/job-b", nil, nil)
			require.Equal(t, fiber.StatusNotFound, response.StatusCode, body)
			require.False(t, store.getScopes[len(store.getScopes)-1].AllowTenantWide)
		})
	}
}

func TestVideoAdminWithAmbiguousDelegatedIdentityIsRejected(t *testing.T) {
	t.Parallel()

	store := &videoAuthorizationStore{jobs: []platform.VideoGenerationJob{videoAuthorizationJob(
		"user:user-a",
		videoAuthorizationTenantID,
	)}}
	app := videoAuthorizationApp(
		&platform.AuthContext{Role: "tenant_admin", UserID: "user-a", VirtualUserID: "virtual-a"},
		videoAuthorizationTenantID,
		store,
		&videoAuthorizationObjectStore{},
	)
	response, body := performVideoAuthorizationRequest(t, app, http.MethodGet, "/v1/videos/job-b", nil, nil)
	require.Equal(t, fiber.StatusForbidden, response.StatusCode, body)
	require.Empty(t, store.getScopes)
}

func TestVideoOwnerScopePreservesTenantIsolation(t *testing.T) {
	t.Parallel()

	store := &videoAuthorizationStore{jobs: []platform.VideoGenerationJob{videoAuthorizationJob(
		"user:user-a",
		"00000000-0000-0000-0000-000000000099",
	)}}
	app := videoAuthorizationApp(
		&platform.AuthContext{UserID: "user-a", Role: "member"},
		videoAuthorizationTenantID,
		store,
		&videoAuthorizationObjectStore{},
	)
	response, body := performVideoAuthorizationRequest(t, app, http.MethodGet, "/v1/videos/job-b", nil, nil)
	require.Equal(t, fiber.StatusNotFound, response.StatusCode, body)
}

func TestVideoOwnerPrincipalPrecedenceAndValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		auth    *platform.AuthContext
		want    string
		wantErr bool
	}{
		{name: "user", auth: &platform.AuthContext{UserID: "user-a", TokenHash: "token-a"}, want: "user:user-a"},
		{name: "virtual user", auth: &platform.AuthContext{VirtualUserID: "virtual-a", TokenHash: "token-a"}, want: "virtual_user:virtual-a"},
		{name: "token fallback", auth: &platform.AuthContext{TokenHash: "token-a"}, want: "token:token-a"},
		{name: "ambiguous actor", auth: &platform.AuthContext{UserID: "user-a", VirtualUserID: "virtual-a"}, wantErr: true},
		{name: "missing owner", auth: &platform.AuthContext{}, wantErr: true},
		{name: "missing auth", auth: nil, wantErr: true},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := videoOwnerPrincipal(test.auth)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, test.want, got)
		})
	}
}

func TestVideoCreateRejectsMissingOrAmbiguousOwner(t *testing.T) {
	t.Parallel()

	for _, authInfo := range []*platform.AuthContext{
		{Role: "member"},
		{Role: "member", UserID: "user-a", VirtualUserID: "virtual-a"},
	} {
		app := videoAuthorizationApp(
			authInfo,
			videoAuthorizationTenantID,
			&videoAuthorizationStore{},
			&videoAuthorizationObjectStore{},
		)
		response, body := performVideoAuthorizationRequest(t, app, http.MethodPost, "/v1/videos", nil, nil)
		require.Equal(t, fiber.StatusForbidden, response.StatusCode, body)
	}
}

func TestVideoCreateIdempotencyLookupRemainsOwnerScopedForAdmin(t *testing.T) {
	t.Parallel()

	job := videoAuthorizationJob("user:user-a", videoAuthorizationTenantID)
	job.IdempotencyKey = "same-key"
	store := &videoAuthorizationStore{jobs: []platform.VideoGenerationJob{job}}
	app := videoAuthorizationApp(
		&platform.AuthContext{Role: "tenant_admin", UserID: "user-a", TokenHash: "token-a"},
		videoAuthorizationTenantID,
		store,
		&videoAuthorizationObjectStore{},
	)
	headers := map[string]string{
		"Content-Type":    "application/json",
		"Idempotency-Key": "same-key",
	}
	body := `{"model":"video-model","prompt":"prompt","seconds":5,"resolution":"720p","aspect_ratio":"16:9","n":1,"service_tier":"standard"}`
	response, responseBody := performVideoAuthorizationRequest(
		t,
		app,
		http.MethodPost,
		"/v1/videos",
		strings.NewReader(body),
		headers,
	)
	require.Equal(t, fiber.StatusOK, response.StatusCode, responseBody)
	require.Len(t, store.idempotencyScopes, 1)
	require.False(t, store.idempotencyScopes[0].AllowTenantWide)
	require.Equal(t, "user:user-a", store.idempotencyScopes[0].OwnerPrincipal)
}

func videoAuthorizationJob(ownerPrincipal, tenantID string) platform.VideoGenerationJob {
	return platform.VideoGenerationJob{
		ID:               "job-b",
		TenantID:         tenantID,
		OwnerPrincipal:   ownerPrincipal,
		GenerationStatus: platform.VideoGenerationSucceeded,
		DeliveryStatus:   platform.VideoDeliveryReady,
		Request:          platform.VideoCreateRequest{Prompt: "private prompt"},
		Artifacts: []platform.VideoArtifact{{
			ID:        "artifact-b",
			ObjectKey: "tenants/tenant-b/videos/job-b/output.mp4",
		}},
		CreatedAt: time.Now().UTC(),
	}
}

func videoAuthorizationApp(
	authInfo *platform.AuthContext,
	tenantID string,
	store *videoAuthorizationStore,
	objects *videoAuthorizationObjectStore,
) *fiber.App {
	handler := NewVideoHandler(nil, nil, store, nil, objects, nil)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.AuthContextKey{}, authInfo)
		c.Locals(tenant.ContextKey{}, &platform.Tenant{ID: tenantID})
		return c.Next()
	})
	app.Post("/v1/videos", handler.Create)
	app.Get("/v1/videos", handler.List)
	app.Get("/v1/videos/:id", handler.Get)
	app.Get("/v1/videos/:id/content", handler.Content)
	app.Post("/v1/videos/:id/cancel", handler.Cancel)
	app.Delete("/v1/videos/:id", handler.Delete)
	return app
}

func performVideoAuthorizationRequest(
	t *testing.T,
	app *fiber.App,
	method string,
	path string,
	body io.Reader,
	headers map[string]string,
) (*http.Response, string) {
	t.Helper()
	if body == nil {
		body = bytes.NewReader(nil)
	}
	request := httptest.NewRequest(method, path, body)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	response, err := app.Test(request, -1)
	require.NoError(t, err)
	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	return response, string(responseBody)
}
