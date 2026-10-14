package chatprovisioning

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const testToken = "0123456789abcdef0123456789abcdef"

type recordedRequest struct {
	Method string
	Path   string
	Auth   string
	Body   map[string]any
}

func newTestServer(t *testing.T, status int, response string) (*httptest.Server, *recordedRequest) {
	t.Helper()

	recorded := &recordedRequest{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded.Method = r.Method
		recorded.Path = r.URL.Path
		recorded.Auth = r.Header.Get("Authorization")

		_ = json.NewDecoder(r.Body).Decode(&recorded.Body)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))

	t.Cleanup(server.Close)

	return server, recorded
}

func TestHTTPClientRegisterSendsTheTenantAndDecodesReadiness(t *testing.T) {
	t.Parallel()

	server, recorded := newTestServer(t, http.StatusCreated, `{"ready":true}`)
	client, err := NewHTTPClient(server.URL, testToken, time.Second)
	require.NoError(t, err)

	registration, err := client.Register(context.Background(), "tenant-1", RegisterRequest{
		Slug: "acme", ClientID: "gdc_1", ClientSecret: "gds_secret",
	})

	require.NoError(t, err)
	require.True(t, registration.Ready)
	require.Equal(t, http.MethodPut, recorded.Method)
	require.Equal(t, "/tenants/tenant-1", recorded.Path)
	require.Equal(t, "Bearer "+testToken, recorded.Auth)
	require.Equal(t, map[string]any{
		"slug": "acme",
		"oidc": map[string]any{"clientId": "gdc_1", "clientSecret": "gds_secret"},
	}, recorded.Body)
}

func TestHTTPClientRegisterOmitsAnEmptySecret(t *testing.T) {
	t.Parallel()

	server, recorded := newTestServer(t, http.StatusOK, `{"ready":false,"reason":"DISCOVERY_FAILED"}`)
	client, err := NewHTTPClient(server.URL, testToken, time.Second)
	require.NoError(t, err)

	registration, err := client.Register(context.Background(), "tenant-1", RegisterRequest{Slug: "acme", ClientID: "gdc_1"})

	require.NoError(t, err)
	require.False(t, registration.Ready)
	require.Equal(t, "DISCOVERY_FAILED", registration.Reason)
	require.Equal(t, map[string]any{"clientId": "gdc_1"}, recorded.Body["oidc"])
}

func TestHTTPClientSetStatusPatchesTheTenant(t *testing.T) {
	t.Parallel()

	server, recorded := newTestServer(t, http.StatusOK, `{"status":"suspended"}`)
	client, err := NewHTTPClient(server.URL, testToken, time.Second)
	require.NoError(t, err)

	require.NoError(t, client.SetStatus(context.Background(), "tenant-1", "suspended"))
	require.Equal(t, http.MethodPatch, recorded.Method)
	require.Equal(t, map[string]any{"status": "suspended"}, recorded.Body)
}

func TestHTTPClientClassifiesFailures(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		status    int
		permanent bool
	}{
		{name: "validation", status: http.StatusBadRequest, permanent: true},
		{name: "unauthorized", status: http.StatusUnauthorized, permanent: true},
		{name: "slug taken", status: http.StatusConflict, permanent: true},
		{name: "rate limited", status: http.StatusTooManyRequests},
		{name: "request timeout", status: http.StatusRequestTimeout},
		{name: "server error", status: http.StatusInternalServerError},
		{name: "unavailable", status: http.StatusServiceUnavailable},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server, _ := newTestServer(t, tc.status, `{"error":{"code":"SLUG_TAKEN"}}`)
			client, err := NewHTTPClient(server.URL, testToken, time.Second)
			require.NoError(t, err)

			_, err = client.Register(context.Background(), "tenant-1", RegisterRequest{Slug: "acme", ClientID: "c"})

			apiErr, ok := AsAPIError(err)
			require.True(t, ok)
			require.Equal(t, tc.status, apiErr.Status)
			require.Equal(t, "SLUG_TAKEN", apiErr.Code)
			require.Equal(t, tc.permanent, apiErr.Permanent())
		})
	}
}

func TestHTTPClientDoesNotFollowRedirectsWithTheToken(t *testing.T) {
	t.Parallel()

	var leaked atomic.Bool

	elsewhere := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { leaked.Store(true) }))
	t.Cleanup(elsewhere.Close)

	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, elsewhere.URL, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(origin.Close)

	client, err := NewHTTPClient(origin.URL, testToken, time.Second)
	require.NoError(t, err)

	_, err = client.Register(context.Background(), "tenant-1", RegisterRequest{Slug: "acme", ClientID: "c"})

	require.Error(t, err)
	require.False(t, leaked.Load())
}

func TestHTTPClientErrorsNeverCarryTheSecretOrToken(t *testing.T) {
	t.Parallel()

	server, _ := newTestServer(t, http.StatusBadRequest, `{"error":{"code":"INVALID_OIDC","message":"gds_secret"}}`)
	client, err := NewHTTPClient(server.URL, testToken, time.Second)
	require.NoError(t, err)

	_, err = client.Register(context.Background(), "tenant-1", RegisterRequest{
		Slug: "acme", ClientID: "c", ClientSecret: "gds_secret",
	})

	require.Error(t, err)
	require.NotContains(t, err.Error(), "gds_secret")
	require.NotContains(t, err.Error(), testToken)
}

func TestHTTPClientReportsAnUnreachableAPIAsATransientError(t *testing.T) {
	t.Parallel()

	server, _ := newTestServer(t, http.StatusOK, `{}`)
	address := server.URL

	server.Close()

	client, err := NewHTTPClient(address, testToken, time.Second)
	require.NoError(t, err)

	_, err = client.Register(context.Background(), "tenant-1", RegisterRequest{Slug: "acme", ClientID: "c"})

	require.Error(t, err)

	_, isAPIError := AsAPIError(err)
	require.False(t, isAPIError)
}

func TestNewHTTPClientValidatesItsInputs(t *testing.T) {
	t.Parallel()

	for _, baseURL := range []string{"", "ftp://chat", "not a url", "http://"} {
		_, err := NewHTTPClient(baseURL, testToken, time.Second)
		require.Error(t, err, baseURL)
	}

	_, err := NewHTTPClient("http://chat:3081", "", time.Second)
	require.Error(t, err)
}
