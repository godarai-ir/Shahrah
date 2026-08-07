package admin

import (
	"errors"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/apiheaders"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/provideraccounts"
	"github.com/gofiber/fiber/v2"
)

func requireTenantAdminOrHigher(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	return fiber.NewError(fiber.StatusForbidden, "tenant admin privileges required")
}

func requireWalletRead(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "billing read permission required", []string{"billing:read", "billing:manage", "settings:list", "settings:manage"})
}

func requireWalletManage(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "billing manage permission required", []string{"billing:manage", "settings:manage"})
}

func requireBillingRead(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "billing read permission required", []string{"billing:read", "billing:manage", "billing:documents_read", "settings:list", "settings:manage"})
}

func requireBillingManage(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "billing manage permission required", []string{"billing:manage", "settings:manage"})
}

func requireBillingProfileManage(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "billing profile manage permission required", []string{"billing:profile_manage", "billing:manage", "settings:manage"})
}

func requireBillingDocumentsRead(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "billing documents permission required", []string{"billing:documents_read", "billing:read", "billing:manage", "settings:list", "settings:manage"})
}

func requireBillingCodesRedeem(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "billing code redemption permission required", []string{"billing:codes_redeem", "billing:manage", "settings:manage"})
}

func requireTeamAdminOrHigher(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTeamAdminOrHigher(authInfo) {
		return nil
	}

	return fiber.NewError(fiber.StatusForbidden, "team admin privileges required")
}

func delegatedUserID(c *fiber.Ctx) string {
	return apiheaders.GetUserID(c)
}

func requireOrgUnitCreate(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsOrgAdminOrHigher(authInfo) {
		return nil
	}

	return requirePermissionAnyOrTenantAdmin(c, "org unit create permission required", []string{"org_unit:create", "org_unit:manage"})
}

func requireOrgUnitRead(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsOrgAdminOrHigher(authInfo) {
		return nil
	}

	return requirePermissionAnyOrTenantAdmin(c, "org unit read permission required", []string{"org_unit:list", "org_unit:read", "org_unit:manage"})
}

func requireOrgUnitManage(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsOrgAdminOrHigher(authInfo) {
		return nil
	}

	return requirePermissionAnyOrTenantAdmin(c, "org unit manage permission required", []string{"org_unit:manage"})
}

func requireOrgUnitDelete(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsOrgAdminOrHigher(authInfo) {
		return nil
	}

	return requirePermissionAnyOrTenantAdmin(c, "org unit delete permission required", []string{"org_unit:delete", "org_unit:manage"})
}

func requireTeamRead(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsOrgAdminOrHigher(authInfo) {
		return nil
	}

	return requirePermissionAnyOrTenantAdmin(c, "team read permission required", []string{"team:read", "team:manage"})
}

func requireTeamCreate(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsOrgAdminOrHigher(authInfo) {
		return nil
	}

	return requirePermissionAnyOrTenantAdmin(c, "team create permission required", []string{"team:create", "team:manage"})
}

func requireTeamManage(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTeamAdminOrHigher(authInfo) {
		return nil
	}

	return requirePermissionAnyOrTenantAdmin(c, "team manage permission required", []string{"team:manage"})
}

func requireTeamDelete(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsOrgAdminOrHigher(authInfo) {
		return nil
	}

	return requirePermissionAnyOrTenantAdmin(c, "team delete permission required", []string{"team:delete", "team:manage"})
}

func requirePolicyListOrManage(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "policy list permission required", []string{"policy:list", "policy:manage"})
}

func requirePolicyManage(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "policy manage permission required", []string{"policy:manage"})
}

func requireLogsMetadataRead(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "logs metadata permission required", []string{"logs:read_metadata", "logs:read_content"})
}

func requireLogsContentRead(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "logs content permission required", []string{"logs:read_content"})
}

func requireLogsExport(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "logs export permission required", []string{"logs:export"})
}

func requireLogsManageViews(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "logs saved view permission required", []string{"logs:manage_views"})
}

func requireLogsReplay(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "logs replay permission required", []string{"logs:replay"})
}

func requirePermissionAnyOrTenantAdmin(c *fiber.Ctx, errMsg string, permissions []string) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTenantAdminOrHigher(authInfo) {
		return nil
	}

	if authInfo == nil {
		return fiber.NewError(fiber.StatusForbidden, errMsg)
	}

	permissionSet := make(map[string]struct{}, len(authInfo.Permissions))
	for _, permission := range authInfo.Permissions {
		permissionSet[strings.TrimSpace(strings.ToLower(permission))] = struct{}{}
	}

	for _, permission := range permissions {
		if _, ok := permissionSet[strings.TrimSpace(strings.ToLower(permission))]; ok {
			return nil
		}
	}

	return fiber.NewError(fiber.StatusForbidden, errMsg)
}

func requireAccessPrincipalSearchPermission(c *fiber.Ctx, kind string) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTeamAdminOrHigher(authInfo) {
		return nil
	}

	if authInfo == nil {
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions for principal search")
	}

	permissions := make(map[string]struct{}, len(authInfo.Permissions))
	for _, permission := range authInfo.Permissions {
		permissions[strings.TrimSpace(strings.ToLower(permission))] = struct{}{}
	}

	has := func(permission string) bool {
		_, ok := permissions[strings.TrimSpace(strings.ToLower(permission))]
		return ok
	}

	switch strings.TrimSpace(strings.ToLower(kind)) {
	case "user":
		if has("user:list") || has("user:manage") {
			return nil
		}
	case "team":
		if has("team:read") || has("team:manage") {
			return nil
		}
	case "org_unit":
		if has("org_unit:read") || has("org_unit:list") || has("org_unit:manage") {
			return nil
		}
	}

	return fiber.NewError(fiber.StatusForbidden, "insufficient permissions for principal search")
}

func requireUserManage(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "user manage permission required", []string{"user:manage"})
}

func requireRoleManage(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "role manage permission required", []string{"role:manage"})
}

func requireUserListOrManage(c *fiber.Ctx) error {
	return requirePermissionAnyOrTenantAdmin(c, "user list permission required", []string{"user:list", "user:manage"})
}

func (h *Handlers) resolveTenantUserIDByEmail(c *fiber.Ctx, tenantSlug string, email string) (string, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail == "" {
		return "", nil
	}

	if h.users == nil {
		return "", fiber.NewError(fiber.StatusNotImplemented, "user store not configured")
	}

	items, _, err := h.users.ListTenantAdminUsers(c.Context(), tenantSlug, normalizedEmail, 200, 0)
	if err != nil {
		return "", err
	}

	for _, item := range items {
		if strings.TrimSpace(strings.ToLower(item.Email)) == normalizedEmail {
			return strings.TrimSpace(item.ID), nil
		}
	}

	return "", nil
}

func mapProviderAccountStoreError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, provideraccounts.ErrInvalidAccessPrincipal) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if errors.Is(err, provideraccounts.ErrInvalidAPIBase) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if errors.Is(err, provideraccounts.ErrConfigContainsSecret) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if errors.Is(err, provideraccounts.ErrPrivatePricingRequiresCredential) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if errors.Is(err, provideraccounts.ErrDependencyConflict) {
		return fiber.NewError(fiber.StatusConflict, "provider account is referenced by active keys; revoke or update keys first")
	}

	if errors.Is(err, provideraccounts.ErrAlreadyExists) {
		return fiber.NewError(fiber.StatusConflict, "resource already exists")
	}

	if errors.Is(err, provideraccounts.ErrInvalidPayload) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid provider account payload")
	}

	return err
}
