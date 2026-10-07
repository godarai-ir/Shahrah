package auth_test

import (
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	rejectionBlockedKey = "sk-blocked"
	rejectionExpiredKey = "sk-expired"
	rejectionMemberKey  = "sk-member"
	rejectionAdminKey   = "sk-admin"
)

type rejectionTestApp struct {
	app   *fiber.App
	calls *atomic.Int32
}

func newRejectionTestApp(t *testing.T) rejectionTestApp {
	t.Helper()

	expiredAt := time.Now().Add(-time.Hour)

	store := auth.NewInMemoryKeyStore()
	store.AddKey(auth.HashKey(rejectionBlockedKey), &platform.AuthContext{KeyAlias: "blocked", Role: "tenant_admin", Blocked: true})
	store.AddKey(auth.HashKey(rejectionExpiredKey), &platform.AuthContext{KeyAlias: "expired", Role: "tenant_admin", ExpiresAt: &expiredAt})
	store.AddKey(auth.HashKey(rejectionMemberKey), &platform.AuthContext{KeyAlias: "member", Role: "member"})
	store.AddKey(auth.HashKey(rejectionAdminKey), &platform.AuthContext{KeyAlias: "admin", Role: "tenant_admin"})

	calls := &atomic.Int32{}

	app := fiber.New()
	app.Use(auth.NewAuthMiddleware(store, "", zap.NewNop()).Middleware())
	app.All("/*", func(c *fiber.Ctx) error {
		calls.Add(1)

		return c.SendStatus(fiber.StatusOK)
	})

	return rejectionTestApp{app: app, calls: calls}
}

func (a rejectionTestApp) do(t *testing.T, method, path, apiKey string) int {
	t.Helper()

	req, err := http.NewRequest(method, path, http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := a.app.Test(req)
	defer closeResponseBody(t, resp)

	require.NoError(t, err)

	return resp.StatusCode
}

func TestAuthMiddleware_RejectedKeysNeverReachHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		apiKey     string
		wantStatus int
	}{
		{name: "blocked key", method: http.MethodGet, path: "/v1/models", apiKey: rejectionBlockedKey, wantStatus: http.StatusForbidden},
		{name: "expired key", method: http.MethodGet, path: "/v1/models", apiKey: rejectionExpiredKey, wantStatus: http.StatusUnauthorized},
		{name: "unknown key", method: http.MethodGet, path: "/v1/models", apiKey: "sk-unknown", wantStatus: http.StatusUnauthorized},
		{name: "forbidden member route", method: http.MethodDelete, path: "/api/v1/roles/x", apiKey: rejectionMemberKey, wantStatus: http.StatusForbidden},
		{name: "forbidden member route upper case", method: http.MethodDelete, path: "/API/v1/roles/x", apiKey: rejectionMemberKey, wantStatus: http.StatusForbidden},
		{name: "forbidden member route mixed case", method: http.MethodDelete, path: "/Api/V1/Roles/x", apiKey: rejectionMemberKey, wantStatus: http.StatusForbidden},
		{name: "forbidden member route trailing slash", method: http.MethodGet, path: "/api/v1/users/", apiKey: rejectionMemberKey, wantStatus: http.StatusForbidden},
		{name: "forbidden member route upper case trailing slash", method: http.MethodGet, path: "/API/V1/USERS/", apiKey: rejectionMemberKey, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newRejectionTestApp(t)

			status := app.do(t, tt.method, tt.path, tt.apiKey)

			assert.Equal(t, tt.wantStatus, status)
			assert.Zero(t, app.calls.Load(), "handler must not run for a rejected request")
		})
	}
}

func TestAuthMiddleware_AuthorizedKeysReachHandler(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		apiKey string
	}{
		{name: "tenant admin management route", method: http.MethodDelete, path: "/api/v1/roles/x", apiKey: rejectionAdminKey},
		{name: "tenant admin upper case route", method: http.MethodDelete, path: "/API/v1/roles/x", apiKey: rejectionAdminKey},
		{name: "member gateway route", method: http.MethodGet, path: "/v1/models", apiKey: rejectionMemberKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newRejectionTestApp(t)

			status := app.do(t, tt.method, tt.path, tt.apiKey)

			assert.Equal(t, http.StatusOK, status)
			assert.Equal(t, int32(1), app.calls.Load())
		})
	}
}

func TestCheckRouteAuthorization_NormalizesCaseAndTrailingSlash(t *testing.T) {
	member := &platform.AuthContext{Role: "member"}

	for _, route := range []string{
		"/API/v1/roles",
		"/api/V1/ROLES/x",
		"/api/v1/roles/",
		"/Api/V1/Users//",
	} {
		require.Error(t, auth.CheckRouteAuthorization(http.MethodDelete, route, member), route)
	}
}

func TestRequiredPlatformScope_IgnoresPathCase(t *testing.T) {
	assert.Equal(t,
		auth.RequiredPlatformScope(http.MethodPost, "/platform/v1/tenants/t1"),
		auth.RequiredPlatformScope(http.MethodPost, "/PLATFORM/V1/TENANTS/t1"),
	)
	assert.Equal(t,
		auth.RequiredPlatformScope(http.MethodPost, "/platform/v1/support/tickets"),
		auth.RequiredPlatformScope(http.MethodPost, "/platform/v1/Support/Tickets"),
	)
}
