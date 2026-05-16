package auth

import (
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const (
	roleProxyAdmin  = "proxy_admin"
	roleTenantAdmin = "tenant_admin"
	roleOrgAdmin    = "org_admin"
	roleTeamAdmin   = "team_admin"
	roleTeamMember  = "team_member"
	roleTenantUser  = "tenant_user"
)

var defaultTeamMemberPermissions = map[string]struct{}{
	"/key/info":   {},
	"/key/health": {},
}

// IsProxyAdmin returns true when the caller is a platform admin.
func IsProxyAdmin(role string) bool {
	return normalizeRole(role) == roleProxyAdmin
}

// IsTenantAdminOrHigher returns true for proxy admin and tenant admin.
func IsTenantAdminOrHigher(authInfo *platform.AuthContext) bool {
	if authInfo == nil {
		return false
	}
	switch normalizeRole(authInfo.Role) {
	case roleProxyAdmin, roleTenantAdmin:
		return true
	default:
		return false
	}
}

// IsOrgAdminOrHigher returns true for proxy admin, tenant admin, and org admin.
func IsOrgAdminOrHigher(authInfo *platform.AuthContext) bool {
	if authInfo == nil {
		return false
	}
	switch normalizeRole(authInfo.Role) {
	case roleProxyAdmin, roleTenantAdmin, roleOrgAdmin:
		return true
	default:
		return false
	}
}

// IsTeamAdminOrHigher returns true for proxy admin, tenant admin, org admin, and team admin.
func IsTeamAdminOrHigher(authInfo *platform.AuthContext) bool {
	if authInfo == nil {
		return false
	}
	switch normalizeRole(authInfo.Role) {
	case roleProxyAdmin, roleTenantAdmin, roleOrgAdmin, roleTeamAdmin:
		return true
	default:
		return false
	}
}

// IsOrgAdminOrHigherForOrgUnit returns true when caller can manage the given OrgUnit.
// For org_admin: checks if they are member of the OrgUnit
// For tenant_admin/proxy_admin: always returns true
func IsOrgAdminOrHigherForOrgUnit(authInfo *platform.AuthContext, orgUnitID string) bool {
	if authInfo == nil {
		return false
	}
	normalizedRole := normalizeRole(authInfo.Role)
	// proxy_admin and tenant_admin can manage any OrgUnit
	if normalizedRole == roleProxyAdmin || normalizedRole == roleTenantAdmin {
		return true
	}
	// org_admin can only manage their assigned OrgUnits
	if normalizedRole == roleOrgAdmin {
		return authInfo.OrgUnitID == orgUnitID
	}
	return false
}

// CheckRouteAuthorization validates whether the current auth context can access the given route.
func CheckRouteAuthorization(method string, route string, authInfo *platform.AuthContext) error {
	if authInfo == nil {
		return fmt.Errorf("authentication required")
	}
	if IsProxyAdmin(authInfo.Role) {
		return nil
	}

	normalizedRoute := normalizeRoute(route)
	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))

	if !strings.HasPrefix(normalizedRoute, "/api/v1/") {
		return nil
	}

	switch {
	case normalizedRoute == "/api/v1/me/permissions":
		return nil
	case strings.HasPrefix(normalizedRoute, "/api/v1/org-units"):
		return requireTenantAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/teams/") && strings.Contains(normalizedRoute, "/members"):
		return requireTeamAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/virtual-users"):
		return requireTeamAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/wallet"):
		return requireTenantAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/models"):
		return requireTenantAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/tenant-provider-accounts"):
		if normalizedMethod == "DELETE" && isProviderAccountDeleteRoute(normalizedRoute) {
			return nil
		}
		return requireTenantAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/access-principals"):
		return requireTeamAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/roles"):
		return requireTenantAdmin(authInfo)
	case normalizedRoute == "/api/v1/users/roles":
		return requireTenantAdmin(authInfo)
	case normalizedRoute == "/api/v1/users/invitations/accept":
		return requireTenantAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/users"):
		return requireTenantAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/keys"):
		return checkKeyRouteAuthorization(normalizedMethod, normalizedRoute, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/personal-access-tokens"):
		return checkPATRouteAuthorization(authInfo)
	default:
		return requireTenantAdmin(authInfo)
	}
}

func checkPATRouteAuthorization(authInfo *platform.AuthContext) error {
	if authInfo == nil {
		return fmt.Errorf("authentication required")
	}
	role := normalizeRole(authInfo.Role)
	switch role {
	case roleProxyAdmin, roleTenantAdmin, roleOrgAdmin, roleTeamAdmin, roleTeamMember, roleTenantUser:
		return nil
	default:
		return fmt.Errorf("insufficient permissions for personal access token route")
	}
}

func checkKeyRouteAuthorization(method string, route string, authInfo *platform.AuthContext) error {
	if IsTeamAdminOrHigher(authInfo) {
		return nil
	}

	if method == "GET" && route == "/api/v1/keys" {
		if hasTeamMemberPermission(authInfo, "/key/list") {
			return nil
		}
	}
	if method == "GET" && strings.HasPrefix(route, "/api/v1/keys/") {
		if hasTeamMemberPermission(authInfo, "/key/info") {
			return nil
		}
	}

	switch {
	case route == "/api/v1/keys" || route == "/api/v1/keys/pat" || route == "/api/v1/keys/vat":
		if method == "POST" && hasTeamMemberPermission(authInfo, "/key/generate") {
			return nil
		}
	case strings.HasSuffix(route, "/rotate"):
		if method == "POST" && hasTeamMemberPermission(authInfo, "/key/regenerate") {
			return nil
		}
	case strings.HasPrefix(route, "/api/v1/keys/"):
		switch method {
		case "PATCH", "PUT":
			if hasTeamMemberPermission(authInfo, "/key/update") {
				return nil
			}
		case "DELETE":
			if hasTeamMemberPermission(authInfo, "/key/delete") {
				return nil
			}
		}
	}

	return fmt.Errorf("insufficient permissions for key management route")
}

func requireTenantAdmin(authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}
	return fmt.Errorf("tenant admin privileges required")
}

func requireTeamAdmin(authInfo *platform.AuthContext) error {
	if IsTeamAdminOrHigher(authInfo) {
		return nil
	}
	return fmt.Errorf("team admin privileges required")
}

func hasTeamMemberPermission(authInfo *platform.AuthContext, permission string) bool {
	if authInfo == nil {
		return false
	}
	if IsTeamAdminOrHigher(authInfo) {
		return true
	}
	permission = strings.TrimSpace(permission)
	if permission == "" {
		return false
	}
	if len(authInfo.TeamMemberPermissions) == 0 {
		_, ok := defaultTeamMemberPermissions[permission]
		return ok
	}
	for _, item := range authInfo.TeamMemberPermissions {
		if strings.TrimSpace(item) == permission {
			return true
		}
	}
	return false
}

func normalizeRole(role string) string {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case "admin":
		return roleTenantAdmin
	case "user":
		return roleTenantUser
	default:
		return strings.TrimSpace(strings.ToLower(role))
	}
}

func normalizeRoute(route string) string {
	trimmed := strings.TrimSpace(route)
	if trimmed == "" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}

func isProviderAccountDeleteRoute(route string) bool {
	trimmed := strings.TrimSpace(route)
	if trimmed == "/api/v1/tenant-provider-accounts" {
		return false
	}
	if !strings.HasPrefix(trimmed, "/api/v1/tenant-provider-accounts/") {
		return false
	}
	rest := strings.TrimPrefix(trimmed, "/api/v1/tenant-provider-accounts/")
	if rest == "" {
		return false
	}
	return !strings.Contains(rest, "/")
}
