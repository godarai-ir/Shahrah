package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	onBehalfOfTenantID = "11111111-1111-1111-1111-111111111111"
	onBehalfOfUserID   = "22222222-2222-2222-2222-222222222222"
	serviceKey         = "sk-librechat-service"
	plainVATKey        = "sk-plain-vat"
	masterKey          = "sk-master"
)

type fakeOnBehalfOfResolver struct {
	users map[string]*platform.AuthContext
	err   error
}

func (f fakeOnBehalfOfResolver) ResolveOnBehalfOfUserAuth(_ context.Context, tenantID, userID string) (*platform.AuthContext, error) {
	if f.err != nil {
		return nil, f.err
	}

	user, ok := f.users[tenantID+"/"+userID]
	if !ok {
		return nil, apperrors.ErrNotFound
	}

	return user, nil
}

func onBehalfOfApp(t *testing.T, resolver OnBehalfOfResolver) *fiber.App {
	t.Helper()

	keys := NewInMemoryKeyStore()
	keys.AddKey(HashKey(serviceKey), &platform.AuthContext{
		TenantID: onBehalfOfTenantID, KeyID: "key-service", KeyType: platform.KeyTypeVAT, Role: platform.LegacyKeyRoleServiceAccount,
		TeamID: "team-service", VirtualUserID: "vu-librechat", OnBehalfOfEnabled: true, CreditProduct: platform.CreditProductLibreChat,
		EnforceModelAllowlist: true, AllowedModels: []string{"a/model-1", "a/model-2"},
	})
	keys.AddKey(HashKey(plainVATKey), &platform.AuthContext{
		TenantID: onBehalfOfTenantID, KeyID: "key-plain", KeyType: platform.KeyTypeVAT, Role: platform.LegacyKeyRoleServiceAccount,
	})

	middleware := NewAuthMiddlewareWithOptions(keys, masterKey, zap.NewNop(), AuthMiddlewareOptions{OnBehalfOf: resolver})

	app := fiber.New()
	app.Use(middleware.Middleware())

	echo := func(c *fiber.Ctx) error {
		return c.JSON(GetAuthInfo(c))
	}
	app.Post("/v1/chat/completions", echo)
	app.Get("/api/v1/teams", echo)

	return app
}

func doOnBehalfOf(t *testing.T, app *fiber.App, method, path, key, onBehalfOf string) (int, map[string]interface{}) {
	t.Helper()

	req := httptest.NewRequest(method, path, http.NoBody)
	req.Header.Set("Authorization", "Bearer "+key)

	if onBehalfOf != "" {
		req.Header.Set("x-godarai-on-behalf-of", onBehalfOf)
	}

	resp, err := app.Test(req)
	require.NoError(t, err)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &body))

	return resp.StatusCode, body
}

func errorType(body map[string]interface{}) string {
	errBody, ok := body["error"].(map[string]interface{})
	if !ok {
		return ""
	}

	value, ok := errBody["type"].(string)
	if !ok {
		return ""
	}

	return value
}

func defaultResolver() fakeOnBehalfOfResolver {
	return fakeOnBehalfOfResolver{users: map[string]*platform.AuthContext{
		onBehalfOfTenantID + "/" + onBehalfOfUserID: {
			TenantID: onBehalfOfTenantID, UserID: onBehalfOfUserID, TeamID: "team-member", OrgUnitID: "ou-1", OrganizationID: "ou-1",
			EnforceModelAllowlist: true, AllowedModels: []string{"a/model-2", "a/model-3"},
		},
	}}
}

func TestOnBehalfOfAttributesUsageToMember(t *testing.T) {
	t.Parallel()

	status, body := doOnBehalfOf(t, onBehalfOfApp(t, defaultResolver()), fiber.MethodPost, "/v1/chat/completions", serviceKey, onBehalfOfUserID)
	require.Equal(t, fiber.StatusOK, status, body)
	require.Equal(t, onBehalfOfUserID, body["user_id"])
	require.Equal(t, "team-member", body["team_id"])
	require.Equal(t, "ou-1", body["org_unit_id"])
	require.Equal(t, "key-service", body["key_id"])
	require.Equal(t, "key-service", body["on_behalf_of_key_id"])
	require.Equal(t, platform.CreditProductLibreChat, body["credit_product"])
	require.Nil(t, body["virtual_user_id"], "the service virtual user must not receive member usage")
	require.Equal(t, []interface{}{"a/model-2"}, body["allowed_models"], "only models allowed for both key and member")
}

func TestOnBehalfOfRejections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resolver OnBehalfOfResolver
		method   string
		path     string
		key      string
		user     string
		status   int
		errType  string
	}{
		{name: "key without capability", resolver: defaultResolver(), method: fiber.MethodPost, path: "/v1/chat/completions", key: plainVATKey, user: onBehalfOfUserID, status: fiber.StatusForbidden, errType: onBehalfOfNotAllowedType},
		{name: "management route", resolver: defaultResolver(), method: fiber.MethodGet, path: "/api/v1/teams", key: serviceKey, user: onBehalfOfUserID, status: fiber.StatusForbidden, errType: onBehalfOfNotAllowedType},
		{name: "master key", resolver: defaultResolver(), method: fiber.MethodPost, path: "/v1/chat/completions", key: masterKey, user: onBehalfOfUserID, status: fiber.StatusForbidden, errType: onBehalfOfNotAllowedType},
		{name: "resolver not configured", resolver: nil, method: fiber.MethodPost, path: "/v1/chat/completions", key: serviceKey, user: onBehalfOfUserID, status: fiber.StatusForbidden, errType: onBehalfOfNotAllowedType},
		{name: "unknown member", resolver: defaultResolver(), method: fiber.MethodPost, path: "/v1/chat/completions", key: serviceKey, user: "33333333-3333-3333-3333-333333333333", status: fiber.StatusForbidden, errType: onBehalfOfUserNotFoundType},
		{name: "malformed member id", resolver: defaultResolver(), method: fiber.MethodPost, path: "/v1/chat/completions", key: serviceKey, user: "not-a-uuid", status: fiber.StatusForbidden, errType: onBehalfOfUserNotFoundType},
		{name: "lookup failure", resolver: fakeOnBehalfOfResolver{err: errors.New("db down")}, method: fiber.MethodPost, path: "/v1/chat/completions", key: serviceKey, user: onBehalfOfUserID, status: fiber.StatusInternalServerError, errType: "server_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, body := doOnBehalfOf(t, onBehalfOfApp(t, tt.resolver), tt.method, tt.path, tt.key, tt.user)
			require.Equal(t, tt.status, status, body)
			require.Equal(t, tt.errType, errorType(body))
		})
	}
}

func TestServiceKeyWithoutHeaderKeepsItsOwnIdentity(t *testing.T) {
	t.Parallel()

	status, body := doOnBehalfOf(t, onBehalfOfApp(t, defaultResolver()), fiber.MethodPost, "/v1/chat/completions", serviceKey, "")
	require.Equal(t, fiber.StatusOK, status, body)
	require.Equal(t, "team-service", body["team_id"])
	require.Nil(t, body["on_behalf_of_key_id"])
}

func TestModelIntersection(t *testing.T) {
	t.Parallel()

	unrestricted := &platform.AuthContext{}
	wildcard := &platform.AuthContext{EnforceModelAllowlist: true, AllowedModels: []string{"*"}}
	keyList := &platform.AuthContext{
		EnforceModelAllowlist: true,
		AllowedModels:         []string{"m1", "m2"},
		AllowedModelTypes:     map[string][]string{"m1": {"chat"}, "m2": {"chat"}},
	}
	userList := &platform.AuthContext{
		EnforceModelAllowlist: true,
		AllowedModels:         []string{"m2", "m3"},
		AllowedModelTypes:     map[string][]string{"m2": {"chat"}, "m3": {"embedding"}},
	}
	disjoint := &platform.AuthContext{EnforceModelAllowlist: true, AllowedModels: []string{"other"}}
	caseVariant := &platform.AuthContext{EnforceModelAllowlist: true, AllowedModels: []string{"M1"}}

	tests := []struct {
		name         string
		key, user    *platform.AuthContext
		wantModels   []string
		wantEnforced bool
	}{
		{name: "both unrestricted", key: unrestricted, user: unrestricted, wantModels: nil, wantEnforced: false},
		{name: "key unrestricted uses member policy", key: unrestricted, user: userList, wantModels: []string{"m2", "m3"}, wantEnforced: true},
		{name: "member wildcard uses key policy", key: keyList, user: wildcard, wantModels: []string{"m1", "m2"}, wantEnforced: true},
		{name: "intersection", key: keyList, user: userList, wantModels: []string{"m2"}, wantEnforced: true},
		{name: "empty intersection denies all", key: keyList, user: disjoint, wantModels: []string{}, wantEnforced: true},
		{name: "matching is case sensitive", key: keyList, user: caseVariant, wantModels: []string{}, wantEnforced: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			merged := mergeOnBehalfOf(tt.key, tt.user)
			require.Equal(t, tt.wantEnforced, merged.EnforceModelAllowlist)
			require.Equal(t, tt.wantModels, merged.AllowedModels)

			for _, model := range merged.AllowedModels {
				require.True(t, modelPolicyAllows(tt.key, model), "%s must be allowed for the key", model)
				require.True(t, modelPolicyAllows(tt.user, model), "%s must be allowed for the member", model)
			}
		})
	}

	merged := mergeOnBehalfOf(keyList, userList)
	require.Equal(t, map[string][]string{"m2": {"chat"}}, merged.AllowedModelTypes)
}
