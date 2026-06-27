package auth

import (
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const (
	roleProxyAdmin     = "proxy_admin"
	roleTenantAdmin    = "tenant_admin"
	roleOrgAdmin       = "org_admin"
	roleTeamAdmin      = "team_admin"
	roleMember         = "member"
	roleReadOnlyMember = "readonly_member"
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
// For tenant_admin/proxy_admin: always returns true.
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
	case normalizedRoute == "/api/v1/subscription" || normalizedRoute == "/api/v1/entitlements":
		return checkSubscriptionRouteAuthorization(normalizedMethod, authInfo)
	case normalizedRoute == "/api/v1/subscription/change":
		return checkSubscriptionRouteAuthorization(normalizedMethod, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/gateway-policies/"):
		return checkPolicyRouteAuthorization(normalizedMethod, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/provider-catalog/"):
		return checkProviderCatalogRouteAuthorization(normalizedMethod, normalizedRoute, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/org-units"):
		return checkOrgUnitRouteAuthorization(normalizedMethod, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/teams/") && strings.Contains(normalizedRoute, "/members"):
		return requireTeamAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/virtual-users"):
		return requireTeamAdmin(authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/wallet"):
		return checkWalletRouteAuthorization(normalizedMethod, normalizedRoute, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/models"):
		return checkModelsRouteAuthorization(normalizedMethod, normalizedRoute, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/virtual-models"):
		return checkVirtualModelsRouteAuthorization(normalizedMethod, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/tenant-provider-accounts"):
		return checkProviderAccountRouteAuthorization(normalizedMethod, normalizedRoute, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/access-principals"):
		return checkAccessPrincipalRouteAuthorization(normalizedRoute, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/roles"):
		return checkRolesRouteAuthorization(normalizedMethod, authInfo)
	case normalizedRoute == "/api/v1/users/roles":
		return checkUserManagementRouteAuthorization(normalizedMethod, authInfo)
	case normalizedRoute == "/api/v1/users/invitations/accept":
		return nil
	case strings.HasPrefix(normalizedRoute, "/api/v1/users"):
		return checkUserManagementRouteAuthorization(normalizedMethod, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/keys"):
		return checkKeyRouteAuthorization(normalizedMethod, normalizedRoute, authInfo)
	case strings.HasPrefix(normalizedRoute, "/api/v1/personal-access-tokens"):
		return checkPATRouteAuthorization(authInfo)
	default:
		return requireTenantAdmin(authInfo)
	}
}

func checkSubscriptionRouteAuthorization(method string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD", "OPTIONS":
		if hasAnyPermission(authInfo, "settings:list", "settings:manage") {
			return nil
		}
	default:
		if hasPermission(authInfo, "settings:manage") {
			return nil
		}
	}

	return fmt.Errorf("settings permission required")
}

func checkPATRouteAuthorization(authInfo *platform.AuthContext) error {
	if authInfo == nil {
		return fmt.Errorf("authentication required")
	}

	role := normalizeRole(authInfo.Role)
	switch role {
	case roleProxyAdmin, roleTenantAdmin, roleOrgAdmin, roleTeamAdmin, roleMember, roleReadOnlyMember:
		return nil
	default:
		return fmt.Errorf("insufficient permissions for personal access token route")
	}
}

func checkWalletRouteAuthorization(method string, route string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))
	normalizedRoute := strings.TrimSpace(route)

	switch normalizedMethod {
	case "GET", "HEAD", "OPTIONS":
		if hasAnyPermission(authInfo, "settings:list", "settings:manage") {
			return nil
		}

		return fmt.Errorf("settings list permission required")
	case "PUT", "PATCH":
		if normalizedRoute == "/api/v1/wallet/topup/settings" && hasPermission(authInfo, "settings:manage") {
			return nil
		}
	case "POST":
		if (normalizedRoute == "/api/v1/wallet/topup" || normalizedRoute == "/api/v1/wallet/topup/intents") && hasPermission(authInfo, "settings:manage") {
			return nil
		}
	}

	return fmt.Errorf("tenant admin privileges required")
}

func checkPolicyRouteAuthorization(method string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD", "OPTIONS":
		if hasPermission(authInfo, "policy:list") || hasPermission(authInfo, "policy:manage") {
			return nil
		}
	default:
		if hasPermission(authInfo, "policy:manage") {
			return nil
		}
	}

	return fmt.Errorf("insufficient permissions for policy management route")
}

func checkAccessPrincipalRouteAuthorization(route string, authInfo *platform.AuthContext) error {
	if IsTeamAdminOrHigher(authInfo) {
		return nil
	}

	kind := extractQueryParam(route, "kind")
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "user":
		if hasPermission(authInfo, "user:list") || hasPermission(authInfo, "user:manage") {
			return nil
		}
	case "team":
		if hasPermission(authInfo, "team:read") || hasPermission(authInfo, "team:manage") {
			return nil
		}
	case "org_unit":
		if hasPermission(authInfo, "org_unit:read") || hasPermission(authInfo, "org_unit:list") || hasPermission(authInfo, "org_unit:manage") {
			return nil
		}
	default:
		if hasPermission(authInfo, "user:list") || hasPermission(authInfo, "team:read") || hasPermission(authInfo, "org_unit:read") {
			return nil
		}
	}

	return fmt.Errorf("insufficient permissions for access principal search")
}

func checkUserManagementRouteAuthorization(method string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD", "OPTIONS":
		if hasPermission(authInfo, "user:manage") {
			return nil
		}
	default:
		if hasPermission(authInfo, "user:manage") {
			return nil
		}
	}

	return fmt.Errorf("user manage permission required")
}

func checkRolesRouteAuthorization(method string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD", "OPTIONS":
		if hasPermission(authInfo, "role:list") || hasPermission(authInfo, "role:manage") || hasPermission(authInfo, "user:manage") {
			return nil
		}
	default:
		if hasPermission(authInfo, "role:manage") {
			return nil
		}
	}

	return fmt.Errorf("insufficient permissions for roles route")
}

func checkOrgUnitRouteAuthorization(method string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD", "OPTIONS":
		if hasPermission(authInfo, "org_unit:read") || hasPermission(authInfo, "org_unit:list") || hasPermission(authInfo, "org_unit:manage") {
			return nil
		}
	case "POST":
		if hasPermission(authInfo, "org_unit:create") || hasPermission(authInfo, "org_unit:manage") {
			return nil
		}
	case "PATCH", "PUT":
		if hasPermission(authInfo, "org_unit:manage") {
			return nil
		}
	case "DELETE":
		if hasPermission(authInfo, "org_unit:delete") || hasPermission(authInfo, "org_unit:manage") {
			return nil
		}
	}

	return fmt.Errorf("insufficient permissions for org unit route")
}

func checkModelsRouteAuthorization(method string, route string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))
	if normalizedMethod == "GET" || normalizedMethod == "HEAD" || normalizedMethod == "OPTIONS" {
		if hasAnyPermission(authInfo,
			"policy:list", "policy:manage",
			"provider_account:read", "provider_account:manage", "provider_account:create",
			"virtual_account:read", "virtual_account:manage",
		) {
			return nil
		}

		return fmt.Errorf("model list permission required")
	}

	if strings.HasSuffix(strings.TrimSpace(route), "/allowed") {
		if hasAnyPermission(authInfo,
			"policy:list", "policy:manage",
			"provider_account:read", "provider_account:manage", "provider_account:create",
			"virtual_account:read", "virtual_account:manage",
		) {
			return nil
		}

		return fmt.Errorf("model list permission required")
	}

	return fmt.Errorf("tenant admin privileges required")
}

func checkVirtualModelsRouteAuthorization(method string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	switch strings.ToUpper(strings.TrimSpace(method)) {
	case "GET", "HEAD", "OPTIONS":
		if hasAnyPermission(authInfo, "virtual_account:read", "virtual_account:manage", "policy:list", "policy:manage") {
			return nil
		}

		return fmt.Errorf("virtual account read permission required")
	default:
		if hasAnyPermission(authInfo, "virtual_account:manage", "virtual_account:delete") {
			return nil
		}

		return fmt.Errorf("virtual account manage permission required")
	}
}

func checkProviderCatalogRouteAuthorization(method string, route string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))

	normalizedRoute := strings.TrimSpace(route)
	if normalizedRoute == "/api/v1/provider-catalog/refresh" && normalizedMethod == "POST" {
		if hasAnyPermission(authInfo, "provider_account:manage", "provider_account:create", "settings:manage") {
			return nil
		}

		return fmt.Errorf("provider account manage permission required")
	}

	if normalizedMethod == "GET" || normalizedMethod == "HEAD" || normalizedMethod == "OPTIONS" {
		if hasAnyPermission(authInfo, "provider_account:read", "provider_account:manage", "provider_account:create") {
			return nil
		}

		return fmt.Errorf("provider account read permission required")
	}

	return fmt.Errorf("provider account manage permission required")
}

func checkProviderAccountRouteAuthorization(method string, route string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))
	normalizedRoute := strings.TrimSpace(route)

	if normalizedMethod == "GET" || normalizedMethod == "HEAD" || normalizedMethod == "OPTIONS" {
		if hasAnyPermission(authInfo, "provider_account:read", "provider_account:manage", "provider_account:create") {
			return nil
		}

		return fmt.Errorf("provider account read permission required")
	}

	if normalizedMethod == "POST" && normalizedRoute == "/api/v1/tenant-provider-accounts" {
		if hasAnyPermission(authInfo, "provider_account:create", "provider_account:manage") {
			return nil
		}

		return fmt.Errorf("provider account create permission required")
	}

	if normalizedMethod == "DELETE" && isProviderAccountDeleteRoute(normalizedRoute) {
		if hasAnyPermission(authInfo, "provider_account:delete", "provider_account:manage", "provider_account:create") {
			return nil
		}

		return fmt.Errorf("provider account delete permission required")
	}

	if hasAnyPermission(authInfo, "provider_account:manage", "provider_account:create") {
		return nil
	}

	return fmt.Errorf("provider account manage permission required")
}

func checkKeyRouteAuthorization(method string, route string, authInfo *platform.AuthContext) error {
	if IsTeamAdminOrHigher(authInfo) {
		return nil
	}

	switch {
	case route == "/api/v1/keys/vat":
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

func hasPermission(authInfo *platform.AuthContext, permission string) bool {
	if authInfo == nil {
		return false
	}

	permission = strings.TrimSpace(strings.ToLower(permission))
	if permission == "" {
		return false
	}

	for _, item := range authInfo.Permissions {
		if strings.TrimSpace(strings.ToLower(item)) == permission {
			return true
		}
	}

	return false
}

func hasAnyPermission(authInfo *platform.AuthContext, permissions ...string) bool {
	for _, permission := range permissions {
		if hasPermission(authInfo, permission) {
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
		return roleMember
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

func extractQueryParam(route string, key string) string {
	parts := strings.SplitN(strings.TrimSpace(route), "?", 2)
	if len(parts) != 2 {
		return ""
	}

	query := parts[1]
	for _, pair := range strings.Split(query, "&") {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}

		if strings.TrimSpace(strings.ToLower(kv[0])) == strings.TrimSpace(strings.ToLower(key)) {
			return kv[1]
		}
	}

	return ""
}
