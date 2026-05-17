package admin

import (
	"errors"
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
)

func requireTenantAdminOrHigher(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTenantAdminOrHigher(authInfo) {
		return nil
	}
	return fiber.NewError(fiber.StatusForbidden, "tenant admin privileges required")
}

func requireTeamAdminOrHigher(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTeamAdminOrHigher(authInfo) {
		return nil
	}
	return fiber.NewError(fiber.StatusForbidden, "team admin privileges required")
}

func requireOrgAdminOrHigher(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsOrgAdminOrHigher(authInfo) {
		return nil
	}
	return fiber.NewError(fiber.StatusForbidden, "organization admin privileges required")
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

func (h *Handlers) requireTenantAdminOrProviderAccountManager(c *fiber.Ctx, store providerAccountStore, tenantID string, accountID string) error {
	authInfo := auth.GetAuthInfo(c)
	if auth.IsTenantAdminOrHigher(authInfo) {
		return nil
	}
	if authInfo == nil {
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions to delete provider account")
	}
	allowed, err := store.IsTenantProviderAccountManager(c.Context(), tenantID, accountID, authInfo.UserID, authInfo.TeamID)
	if err != nil {
		return err
	}
	if !allowed {
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions to delete provider account")
	}
	return nil
}

func mapProviderAccountStoreError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, storage.ErrInvalidAccessPrincipal) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if errors.Is(err, storage.ErrProviderAccountDependencyConflict) {
		return fiber.NewError(fiber.StatusConflict, "provider account is referenced by active keys; revoke or update keys first")
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return fiber.NewError(fiber.StatusConflict, "resource already exists")
		case "23514", "23502", "22P02":
			return fiber.NewError(fiber.StatusBadRequest, "invalid provider account payload")
		}
	}
	return err
}
