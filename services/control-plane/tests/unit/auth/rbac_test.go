package auth_test

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckRouteAuthorization_ProxyAdminAllowedEverywhere(t *testing.T) {
	err := auth.CheckRouteAuthorization("POST", "/api/v1/wallet/topup", &platform.AuthContext{
		Role: "proxy_admin",
	})
	require.NoError(t, err)
}

func TestCheckRouteAuthorization_TenantAdminAllowedOnManagement(t *testing.T) {
	err := auth.CheckRouteAuthorization("GET", "/api/v1/users", &platform.AuthContext{
		Role: "tenant_admin",
	})
	require.NoError(t, err)
}

func TestCheckRouteAuthorization_SettingsListAllowedOnWalletReads(t *testing.T) {
	for _, route := range []string{
		"/api/v1/wallet",
		"/api/v1/wallet/transactions?limit=20",
		"/api/v1/wallet/topup/settings",
	} {
		err := auth.CheckRouteAuthorization("GET", route, &platform.AuthContext{
			Role:        "member",
			Permissions: []string{"settings:list"},
		})
		require.NoError(t, err, route)
	}
}

func TestCheckRouteAuthorization_WalletMutationsRequireManagePermission(t *testing.T) {
	err := auth.CheckRouteAuthorization("POST", "/api/v1/wallet/topup/intents", &platform.AuthContext{
		Role:        "member",
		Permissions: []string{"settings:list"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "billing manage permission required")

	err = auth.CheckRouteAuthorization("POST", "/api/v1/wallet/topup/intents", &platform.AuthContext{
		Role:        "member",
		Permissions: []string{"settings:manage"},
	})
	require.NoError(t, err)

	err = auth.CheckRouteAuthorization("POST", "/api/v1/wallet/topup/intents", &platform.AuthContext{
		Role:        "member",
		Permissions: []string{"billing:manage"},
	})
	require.NoError(t, err)
}

func TestCheckRouteAuthorization_TenantUserDeniedOnManagement(t *testing.T) {
	err := auth.CheckRouteAuthorization("GET", "/api/v1/users", &platform.AuthContext{
		Role: "member",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user manage permission required")
}

func TestCheckRouteAuthorization_TenantAdminAllowedOnKeyManagement(t *testing.T) {
	err := auth.CheckRouteAuthorization("POST", "/api/v1/keys/vat", &platform.AuthContext{
		Role: "tenant_admin",
	})
	require.NoError(t, err)
}

func TestCheckRouteAuthorization_TeamMemberDeniedWithoutPermission(t *testing.T) {
	err := auth.CheckRouteAuthorization("POST", "/api/v1/keys/vat", &platform.AuthContext{
		Role: "team_member",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient permissions")
}

func TestCheckRouteAuthorization_TeamMemberAllowedWithPermission(t *testing.T) {
	err := auth.CheckRouteAuthorization("POST", "/api/v1/keys/vat", &platform.AuthContext{
		Role:                  "team_member",
		TeamMemberPermissions: []string{"/key/generate"},
	})
	require.NoError(t, err)
}

func TestCheckRouteAuthorization_LLMRouteAllowedForAuthenticatedKey(t *testing.T) {
	err := auth.CheckRouteAuthorization("POST", "/v1/chat/completions", &platform.AuthContext{
		Role: "member",
	})
	require.NoError(t, err)
}

func TestCheckRouteAuthorization_TeamManagerAllowedOnTeamMembershipRoutes(t *testing.T) {
	err := auth.CheckRouteAuthorization("POST", "/api/v1/teams/team-1/members", &platform.AuthContext{
		Role:        "member",
		Permissions: []string{"team:manage"},
	})
	require.NoError(t, err)
}

func TestCheckRouteAuthorization_TeamMemberDeniedOnTeamMembershipRoutes(t *testing.T) {
	err := auth.CheckRouteAuthorization("POST", "/api/v1/teams/team-1/members", &platform.AuthContext{
		Role: "team_member",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "team manage permission required")
}

func TestCheckRouteAuthorization_TenantAdminAllowedOnRoleRoutes(t *testing.T) {
	err := auth.CheckRouteAuthorization("GET", "/api/v1/roles", &platform.AuthContext{
		Role: "tenant_admin",
	})
	require.NoError(t, err)

	err = auth.CheckRouteAuthorization("PATCH", "/api/v1/users/roles", &platform.AuthContext{
		Role: "tenant_admin",
	})
	require.NoError(t, err)
}

func TestCheckRouteAuthorization_MemberDeniedOnRoleRoutes(t *testing.T) {
	err := auth.CheckRouteAuthorization("POST", "/api/v1/roles", &platform.AuthContext{
		Role: "member",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient permissions for roles route")

	err = auth.CheckRouteAuthorization("PATCH", "/api/v1/users/roles", &platform.AuthContext{
		Role: "member",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user manage permission required")
}
