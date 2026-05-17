package admin

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/auth"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func (h *Handlers) ListRoles(c *fiber.Ctx) error {
	if err := requirePermissionAnyOrTenantAdmin(c, "role list permission required", []string{"role:list", "role:manage", "user:manage"}); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.roles == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "role store not configured")
	}

	query := strings.TrimSpace(c.Query("query"))
	limit := 50
	offset := 0
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "limit must be a positive number")
		}
		if parsed > 200 {
			parsed = 200
		}
		limit = parsed
	}
	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "offset must be zero or greater")
		}
		offset = parsed
	}

	items, total, err := h.roles.ListRoles(c.Context(), currentTenant.ID, query, limit, offset)
	if err != nil {
		return err
	}
	var nextOffset *int
	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}

	return c.JSON(platform.RoleListResponse{
		Data:       items,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		NextOffset: nextOffset,
	})
}

// CreateRole creates a custom tenant role.
func (h *Handlers) CreateRole(c *fiber.Ctx) error {
	if err := requireRoleManage(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.roles == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "role store not configured")
	}

	var req platform.CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid request body")
	}
	item, err := h.roles.CreateRole(c.Context(), currentTenant.ID, req)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidRoleName):
			return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid role name")
		case errors.Is(err, storage.ErrInvalidPermission):
			return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid permission set")
		default:
			return err
		}
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// UpdateRole updates one role definition.
func (h *Handlers) UpdateRole(c *fiber.Ctx) error {
	if err := requireRoleManage(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.roles == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "role store not configured")
	}

	roleID := strings.TrimSpace(c.Params("roleID"))
	if roleID == "" {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "roleID is required")
	}

	var req platform.UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid request body")
	}
	item, err := h.roles.UpdateRole(c.Context(), currentTenant.ID, roleID, req)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrRoleNotFound):
			return fiber.NewError(fiber.StatusNotFound, "role not found")
		case errors.Is(err, storage.ErrRoleImmutable):
			return fiber.NewError(fiber.StatusForbidden, "system role cannot be modified")
		case errors.Is(err, storage.ErrInvalidPermission):
			return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid permission set")
		default:
			return err
		}
	}
	return c.JSON(item)
}

// UpdateUserRoles updates tenant role bindings for one user by email.
func (h *Handlers) UpdateUserRoles(c *fiber.Ctx) error {
	if err := requireUserManage(c); err != nil {
		return err
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.roles == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "role store not configured")
	}

	var req platform.UpdateUserRolesRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "invalid request body")
	}
	authInfo := auth.GetAuthInfo(c)
	if authInfo != nil {
		targetUserID, resolveErr := h.resolveTenantUserIDByEmail(c, currentTenant.Slug, req.Email)
		if resolveErr != nil {
			return resolveErr
		}
		if targetUserID != "" && strings.TrimSpace(authInfo.UserID) == targetUserID {
			return fiber.NewError(fiber.StatusForbidden, "you cannot change your own role")
		}
	}
	out, err := h.roles.UpdateUserRoles(c.Context(), currentTenant.ID, req)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidUserRoleUpdate), errors.Is(err, storage.ErrInvalidRoleName), errors.Is(err, storage.ErrInvalidResourceType):
			return fiber.NewError(fiber.StatusUnprocessableEntity, "missing or invalid fields for role update")
		case errors.Is(err, storage.ErrRoleNotFound):
			return fiber.NewError(fiber.StatusUnprocessableEntity, "one or more roles not found")
		case errors.Is(err, storage.ErrUserNotFound):
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		default:
			return err
		}
	}
	return c.JSON(out)
}

// GetCurrentUserPermissions returns effective roles and permissions for the current auth context.
func (h *Handlers) GetCurrentUserPermissions(c *fiber.Ctx) error {
	authInfo := auth.GetAuthInfo(c)
	if authInfo == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	}
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	role := strings.TrimSpace(strings.ToLower(authInfo.Role))
	roles := []string{role}
	permissions := make([]string, 0)

	if h.roles != nil && strings.TrimSpace(authInfo.UserID) != "" {
		resolvedRoles, resolvedPermissions, err := h.roles.GetEffectivePermissionsForUser(c.Context(), currentTenant.ID, authInfo.UserID, role)
		if err != nil {
			return err
		}
		if len(resolvedRoles) > 0 {
			roles = resolvedRoles
		}
		permissions = resolvedPermissions
	}

	if len(permissions) == 0 {
		fallbackPerms := map[string][]string{
			"proxy_admin":     {"policy:list", "policy:manage", "role:list", "role:manage", "user:list", "user:manage", "team:create", "team:read", "team:manage", "team:delete", "org_unit:create", "org_unit:read", "org_unit:manage", "org_unit:delete", "org_unit:list", "virtual_account:create", "virtual_account:read", "virtual_account:manage", "virtual_account:delete", "provider_account:create", "provider_account:read", "provider_account:manage", "provider_account:delete", "provider_account:use_integrations", "settings:list", "settings:manage"},
			"tenant_admin":    {"policy:list", "policy:manage", "role:list", "role:manage", "user:list", "user:manage", "team:create", "team:read", "team:manage", "team:delete", "org_unit:create", "org_unit:read", "org_unit:manage", "org_unit:delete", "org_unit:list", "virtual_account:create", "virtual_account:read", "virtual_account:manage", "virtual_account:delete", "provider_account:create", "provider_account:read", "provider_account:manage", "provider_account:delete", "provider_account:use_integrations", "settings:list", "settings:manage"},
			"org_admin":       {"policy:list", "role:list", "user:list", "user:manage", "org_unit:read", "org_unit:manage", "org_unit:list", "team:create", "team:read", "team:manage", "virtual_account:create", "virtual_account:read", "virtual_account:manage", "provider_account:read", "provider_account:use_integrations", "settings:list"},
			"team_admin":      {"policy:list", "role:list", "user:list", "team:read", "team:manage", "virtual_account:create", "virtual_account:read", "virtual_account:manage", "virtual_account:delete", "provider_account:read", "provider_account:use_integrations", "settings:list"},
			"member":          {"user:list", "settings:list"},
			"readonly_member": {"settings:list"},
		}
		permissions = append(permissions, fallbackPerms[role]...)
	}

	catalogItems := storage.ListPermissionCatalog()
	catalog := make([]platform.PermissionDefinition, 0, len(catalogItems))
	for _, item := range catalogItems {
		catalog = append(catalog, platform.PermissionDefinition{
			Key:         item.Key,
			Group:       item.Group,
			Description: item.Description,
		})
	}

	return c.JSON(platform.MePermissionsResponse{
		UserID:      authInfo.UserID,
		Role:        role,
		Roles:       roles,
		Permissions: permissions,
		Catalog:     catalog,
	})
}

// ListTenantUsers lists tenant-admin users for the current tenant.
func (h *Handlers) ListTenantUsers(c *fiber.Ctx) error {
	if err := requireUserListOrManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.users == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "user store not configured")
	}

	query := strings.TrimSpace(c.Query("query"))
	limit := 50
	offset := 0

	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed <= 0 {
			return fiber.NewError(fiber.StatusBadRequest, "limit must be a positive number")
		}
		if parsed > 200 {
			parsed = 200
		}
		limit = parsed
	}

	if raw := strings.TrimSpace(c.Query("offset")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "offset must be zero or greater")
		}
		offset = parsed
	}

	items, total, err := h.users.ListTenantAdminUsers(c.Context(), currentTenant.Slug, query, limit, offset)
	if err != nil {
		return err
	}

	var nextOffset *int
	if offset+len(items) < total {
		n := offset + len(items)
		nextOffset = &n
	}

	return c.JSON(platform.TenantAdminUserListResponse{
		Data:       items,
		Total:      total,
		Limit:      limit,
		Offset:     offset,
		NextOffset: nextOffset,
	})
}

// InviteTenantUser creates a tenant-admin invitation link.
func (h *Handlers) InviteTenantUser(c *fiber.Ctx) error {
	if err := requireUserManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.users == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "user store not configured")
	}

	var req platform.InviteTenantAdminUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Mobile = strings.TrimSpace(req.Mobile)
	req.Role = strings.TrimSpace(strings.ToLower(req.Role))
	req.AcceptInviteClientURL = strings.TrimSpace(req.AcceptInviteClientURL)

	if req.Email == "" && req.Mobile == "" {
		return fiber.NewError(fiber.StatusBadRequest, "either email or mobile is required")
	}
	if req.Email != "" && !storage.IsValidEmailAddress(req.Email) {
		return fiber.NewError(fiber.StatusBadRequest, "Please enter a valid email address.")
	}
	if req.Role == "" {
		req.Role = "member"
	}

	invitation, link, err := h.users.InviteTenantAdminUser(c.Context(), currentTenant.ID, currentTenant.Slug, req)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidInvitationEmail):
			return fiber.NewError(fiber.StatusBadRequest, "Please enter a valid email address.")
		case errors.Is(err, storage.ErrTenantMemberAlreadyExists):
			return fiber.NewError(fiber.StatusConflict, "This user is already a member of this tenant.")
		case errors.Is(err, storage.ErrTenantMemberDeactivated):
			return fiber.NewError(fiber.StatusConflict, "This user is deactivated and cannot receive an invitation.")
		case errors.Is(err, storage.ErrInvitationAlreadyPending):
			return fiber.NewError(fiber.StatusConflict, "An invitation has already been sent to this email.")
		default:
			return err
		}
	}
	return c.Status(fiber.StatusCreated).JSON(platform.InviteTenantAdminUserResponse{
		Invitation: *invitation,
		Link:       link,
	})
}

// AcceptTenantUserInvitation accepts invite token and activates/creates user.
func (h *Handlers) AcceptTenantUserInvitation(c *fiber.Ctx) error {
	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.users == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "user store not configured")
	}

	var req platform.AcceptTenantAdminInvitationRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "token is required")
	}

	invitation, user, err := h.users.AcceptTenantAdminInvitation(c.Context(), currentTenant.ID, currentTenant.Slug, req.Token)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInvalidInvitationToken):
			return fiber.NewError(fiber.StatusBadRequest, "invalid invitation token")
		case errors.Is(err, storage.ErrInvitationExpired):
			return fiber.NewError(fiber.StatusGone, "invitation expired")
		case errors.Is(err, storage.ErrInvitationAlreadyUsed):
			return fiber.NewError(fiber.StatusConflict, "invitation already used")
		case errors.Is(err, storage.ErrInvitationContactTaken):
			return fiber.NewError(fiber.StatusConflict, "invitation contact already belongs to another account")
		case errors.Is(err, pgx.ErrNoRows):
			return fiber.NewError(fiber.StatusNotFound, "invitation not found")
		default:
			return err
		}
	}

	return c.JSON(platform.AcceptTenantAdminInvitationResponse{
		Invitation: *invitation,
		User:       *user,
	})
}

// ActivateTenantUser marks one tenant-admin user as active.
func (h *Handlers) ActivateTenantUser(c *fiber.Ctx) error {
	return h.setTenantUserActive(c, true)
}

// DeactivateTenantUser marks one tenant-admin user as deactivated.
func (h *Handlers) DeactivateTenantUser(c *fiber.Ctx) error {
	return h.setTenantUserActive(c, false)
}

func (h *Handlers) setTenantUserActive(c *fiber.Ctx, active bool) error {
	if err := requireUserManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}
	if h.users == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "user store not configured")
	}

	userID := strings.TrimSpace(c.Params("userID"))
	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "userID is required")
	}
	authInfo := auth.GetAuthInfo(c)
	if authInfo != nil && strings.TrimSpace(authInfo.UserID) == userID {
		return fiber.NewError(fiber.StatusForbidden, "you cannot change your own account status")
	}
	target, err := h.users.GetTenantAdminUserRecord(c.Context(), currentTenant.Slug, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return err
	}
	if !active && isTenantAdminUser(target) {
		return fiber.NewError(fiber.StatusForbidden, "tenant admin user cannot be deactivated")
	}
	user, err := h.users.SetTenantAdminUserActive(c.Context(), currentTenant.Slug, userID, active)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}
		return err
	}
	return c.JSON(user)
}
