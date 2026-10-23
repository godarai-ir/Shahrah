package auth

import (
	"fmt"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

const (
	roleProxyAdmin  = "proxy_admin"
	roleTenantAdmin = "tenant_admin"
	roleMember      = "member"
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

// CheckRouteAuthorization validates whether the current auth context can access the given route.
func CheckRouteAuthorization(method, route string, authInfo *platform.AuthContext) error {
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

	if handled, err := checkAPIV1RouteAuthorization(normalizedMethod, normalizedRoute, authInfo); handled {
		return err
	}

	return requireTenantAdmin(authInfo)
}

func checkAPIV1RouteAuthorization(method, route string, authInfo *platform.AuthContext) (bool, error) {
	if handled, err := checkAPIV1PublicRouteAuthorization(method, route, authInfo); handled {
		return true, err
	}

	if handled, err := checkAPIV1PolicyRouteAuthorization(method, route, authInfo); handled {
		return true, err
	}

	if handled, err := checkAPIV1OrganizationRouteAuthorization(method, route, authInfo); handled {
		return true, err
	}

	if handled, err := checkAPIV1ModelRouteAuthorization(method, route, authInfo); handled {
		return true, err
	}

	return false, nil
}

func checkAPIV1PublicRouteAuthorization(method, route string, authInfo *platform.AuthContext) (bool, error) {
	switch route {
	case "/api/v1/me/permissions", "/api/v1/users/invitations/accept":
		return true, nil
	case "/api/v1/subscription", "/api/v1/entitlements", "/api/v1/subscription/change":
		return true, checkSubscriptionRouteAuthorization(method, authInfo)
	default:
		if strings.HasPrefix(route, "/api/v1/support/") {
			return true, checkSupportTicketRouteAuthorization(authInfo)
		}

		return false, nil
	}
}

func checkSupportTicketRouteAuthorization(authInfo *platform.AuthContext) error {
	if authInfo == nil {
		return fmt.Errorf("support ticket access requires authentication")
	}

	return nil
}

func checkAPIV1PolicyRouteAuthorization(method, route string, authInfo *platform.AuthContext) (bool, error) {
	switch {
	case strings.HasPrefix(route, "/api/v1/gateway-policies/"):
		return true, checkPolicyRouteAuthorization(method, authInfo)
	case strings.HasPrefix(route, "/api/v1/provider-catalog/"):
		return true, checkProviderCatalogRouteAuthorization(method, route, authInfo)
	default:
		return false, nil
	}
}

func checkAPIV1OrganizationRouteAuthorization(method, route string, authInfo *platform.AuthContext) (bool, error) {
	switch {
	case strings.HasPrefix(route, "/api/v1/teams/") && strings.Contains(route, "/members"):
		return true, requireTeamManage(authInfo)
	case strings.HasPrefix(route, "/api/v1/virtual-users"):
		return true, requireTenantAdmin(authInfo)
	default:
		return false, nil
	}
}

func checkAPIV1ModelRouteAuthorization(method, route string, authInfo *platform.AuthContext) (bool, error) {
	switch {
	case strings.HasPrefix(route, "/api/v1/wallet"):
		return true, checkWalletRouteAuthorization(method, route, authInfo)
	case strings.HasPrefix(route, "/api/v1/billing"):
		return true, checkBillingRouteAuthorization(method, route, authInfo)
	case strings.HasPrefix(route, "/api/v1/models"):
		return true, checkModelsRouteAuthorization(method, route, authInfo)
	case strings.HasPrefix(route, "/api/v1/virtual-models"):
		return true, checkVirtualModelsRouteAuthorization(method, authInfo)
	case strings.HasPrefix(route, "/api/v1/tenant-provider-accounts"):
		return true, checkProviderAccountRouteAuthorization(method, route, authInfo)
	case strings.HasPrefix(route, "/api/v1/access-principals"):
		return true, checkAccessPrincipalRouteAuthorization(route, authInfo)
	case strings.HasPrefix(route, "/api/v1/roles"):
		return true, checkRolesRouteAuthorization(method, authInfo)
	case route == "/api/v1/users/roles":
		return true, checkUserManagementRouteAuthorization(method, authInfo)
	case strings.HasPrefix(route, "/api/v1/users"):
		return true, checkUserManagementRouteAuthorization(method, authInfo)
	case strings.HasPrefix(route, "/api/v1/keys"):
		return true, checkKeyRouteAuthorization(method, route, authInfo)
	case strings.HasPrefix(route, "/api/v1/personal-access-tokens"):
		return true, checkPATRouteAuthorization(authInfo)
	default:
		return false, nil
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
	case roleProxyAdmin, roleTenantAdmin, roleMember:
		return nil
	default:
		return fmt.Errorf("insufficient permissions for personal access token route")
	}
}

func checkWalletRouteAuthorization(method, route string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))
	normalizedRoute := strings.TrimSpace(route)

	switch normalizedMethod {
	case "GET", "HEAD", "OPTIONS":
		if hasAnyPermission(authInfo, "billing:read", "billing:manage", "billing:documents_read", "settings:list", "settings:manage") {
			return nil
		}

		return fmt.Errorf("billing read permission required")
	case "PUT", "PATCH":
		if normalizedRoute == "/api/v1/wallet/topup/settings" && hasAnyPermission(authInfo, "billing:manage", "settings:manage") {
			return nil
		}
	case "POST":
		if (normalizedRoute == "/api/v1/wallet/topup" || normalizedRoute == "/api/v1/wallet/topup/intents" || normalizedRoute == "/api/v1/wallet/bank-transfers") && hasAnyPermission(authInfo, "billing:manage", "settings:manage") {
			return nil
		}
	}

	return fmt.Errorf("billing manage permission required")
}

func checkBillingRouteAuthorization(method, route string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))
	normalizedRoute := strings.TrimSpace(route)

	if strings.Contains(normalizedRoute, "/gift-codes/redeem") || strings.Contains(normalizedRoute, "/discount-codes/redeem") {
		if normalizedMethod == "POST" && hasAnyPermission(authInfo, "billing:codes_redeem", "billing:manage", "settings:manage") {
			return nil
		}

		return fmt.Errorf("billing code redemption permission required")
	}

	if normalizedRoute == "/api/v1/billing/profile" {
		if normalizedMethod == "GET" || normalizedMethod == "HEAD" || normalizedMethod == "OPTIONS" {
			if hasAnyPermission(authInfo, "billing:read", "billing:manage", "billing:profile_manage", "settings:list", "settings:manage") {
				return nil
			}
		}

		if hasAnyPermission(authInfo, "billing:profile_manage", "billing:manage", "settings:manage") {
			return nil
		}

		return fmt.Errorf("billing profile manage permission required")
	}

	if strings.HasPrefix(normalizedRoute, "/api/v1/billing/documents") {
		if hasAnyPermission(authInfo, "billing:documents_read", "billing:read", "billing:manage", "settings:list", "settings:manage") {
			return nil
		}

		return fmt.Errorf("billing documents permission required")
	}

	if normalizedMethod == "GET" || normalizedMethod == "HEAD" || normalizedMethod == "OPTIONS" {
		if hasAnyPermission(authInfo, "billing:read", "billing:manage", "settings:list", "settings:manage") {
			return nil
		}
	}

	return fmt.Errorf("billing read permission required")
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
	if IsTenantAdminOrHigher(authInfo) {
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
	default:
		if hasPermission(authInfo, "user:list") || hasPermission(authInfo, "team:read") {
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

func checkModelsRouteAuthorization(method, route string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	normalizedMethod := strings.ToUpper(strings.TrimSpace(method))
	if normalizedMethod == "GET" || normalizedMethod == "HEAD" || normalizedMethod == "OPTIONS" {
		if hasAnyPermission(authInfo,
			"playground:read", "playground:run",
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

func checkProviderCatalogRouteAuthorization(method, route string, authInfo *platform.AuthContext) error {
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

func checkProviderAccountRouteAuthorization(method, route string, authInfo *platform.AuthContext) error {
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

func checkKeyRouteAuthorization(method, route string, authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) {
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

func requireTeamManage(authInfo *platform.AuthContext) error {
	if IsTenantAdminOrHigher(authInfo) || hasPermission(authInfo, "team:manage") {
		return nil
	}

	return fmt.Errorf("team manage permission required")
}

func hasTeamMemberPermission(authInfo *platform.AuthContext, permission string) bool {
	if authInfo == nil {
		return false
	}

	if IsTenantAdminOrHigher(authInfo) {
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

// normalizeRoute canonicalizes a request path so RBAC rules match the same
// requests the Fiber router does. The router is case-insensitive and ignores
// trailing slashes (CaseSensitive and StrictRouting are off), so the path is
// lowercased and trailing slashes are trimmed. Any query string is preserved.
func normalizeRoute(route string) string {
	path, query, hasQuery := strings.Cut(strings.TrimSpace(route), "?")

	path = strings.ToLower(strings.TrimRight(path, "/"))
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	if hasQuery {
		return path + "?" + query
	}

	return path
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

func extractQueryParam(route, key string) string {
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
