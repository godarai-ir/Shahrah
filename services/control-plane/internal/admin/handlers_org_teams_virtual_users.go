package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/tenant"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

func (h *Handlers) ListOrgUnits(c *fiber.Ctx) error {
	if err := requireOrgUnitRead(c); err != nil {
		return err
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	params, err := parseTenantListParams(c, 200)
	if err != nil {
		return err
	}

	items, total, err := h.orgUnits.ListOrgUnits(c.Context(), currentTenant.ID, params.Query, params.Limit, params.Offset)
	if err != nil {
		return err
	}

	return c.JSON(platform.OrgUnitListResponse{
		Data:       items,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		NextOffset: buildNextOffset(params.Offset, len(items), total),
	})
}

func (h *Handlers) ListTeams(c *fiber.Ctx) error {
	if err := requireTeamRead(c); err != nil {
		return err
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}

	params, err := parseTenantListParams(c, 200)
	if err != nil {
		return err
	}

	items, total, err := h.teams.ListTeams(c.Context(), currentTenant.ID, params.Query, params.Limit, params.Offset)
	if err != nil {
		return err
	}

	return c.JSON(platform.TeamListResponse{
		Data:       items,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		NextOffset: buildNextOffset(params.Offset, len(items), total),
	})
}

func (h *Handlers) CreateTeam(c *fiber.Ctx) error {
	if err := requireTeamCreate(c); err != nil {
		return err
	}

	currentTenant, err := currentTenantOrErr(c)
	if err != nil {
		return err
	}

	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}

	entitlementErr := h.requireEntitlementEnabled(c, currentTenant.ID, platform.FeatureTeamsEnabled)
	if entitlementErr != nil {
		return entitlementErr
	}

	var req platform.CreateTeamRequest

	parseErr := c.BodyParser(&req)
	if parseErr != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Alias = strings.TrimSpace(req.Alias)
	if req.Alias == "" {
		return fiber.NewError(fiber.StatusBadRequest, "alias is required")
	}

	item, err := h.teams.CreateTeam(c.Context(), currentTenant.ID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetTeam(c *fiber.Ctx) error {
	if err := requireTeamRead(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}

	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}

	item, err := h.teams.GetTeam(c.Context(), currentTenant.ID, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team not found")
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpdateTeam(c *fiber.Ctx) error {
	if err := requireTeamManage(c); err != nil {
		return err
	}

	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}

	return handleTenantScopedUpdate(c, "teamID", "team not found", parseBodyInto[platform.UpdateTeamRequest], func(ctx context.Context, tenantID, itemID string, req platform.UpdateTeamRequest) (interface{}, error) {
		return h.teams.UpdateTeam(ctx, tenantID, itemID, req)
	})
}

func (h *Handlers) DeleteTeam(c *fiber.Ctx) error {
	if err := requireTeamDelete(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}

	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}

	if err := h.teams.DeleteTeam(c.Context(), currentTenant.ID, teamID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team not found")
		}

		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) GetTeamPermissions(c *fiber.Ctx) error {
	if err := requireTeamRead(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}

	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}

	item, err := h.teams.GetTeamPermissions(c.Context(), currentTenant.ID, teamID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team not found")
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpdateTeamPermissions(c *fiber.Ctx) error {
	if err := requireTeamManage(c); err != nil {
		return err
	}

	if h.teams == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "team store not configured")
	}

	return handleTenantScopedUpdate(c, "teamID", "team not found", parseBodyInto[platform.UpdateTeamPermissionsRequest], func(ctx context.Context, tenantID, itemID string, req platform.UpdateTeamPermissionsRequest) (interface{}, error) {
		return h.teams.UpdateTeamPermissions(ctx, tenantID, itemID, req)
	})
}

func (h *Handlers) CreateOrgUnit(c *fiber.Ctx) error {
	if err := requireOrgUnitCreate(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	if err := h.requireEntitlementEnabled(c, currentTenant.ID, platform.FeatureOrgUnitsEnabled); err != nil {
		return err
	}

	var req platform.CreateOrgUnitRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Alias = strings.TrimSpace(req.Alias)

	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.Alias == "" {
		return fiber.NewError(fiber.StatusBadRequest, "alias is required")
	}

	item, err := h.orgUnits.CreateOrgUnit(c.Context(), currentTenant.ID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetOrgUnit(c *fiber.Ctx) error {
	if err := requireOrgUnitRead(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	if err := h.requireEntitlementEnabled(c, currentTenant.ID, platform.FeatureOrgUnitsEnabled); err != nil {
		return err
	}

	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	if orgUnitID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID is required")
	}

	item, err := h.orgUnits.GetOrgUnit(c.Context(), currentTenant.ID, orgUnitID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "org unit not found")
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpdateOrgUnit(c *fiber.Ctx) error {
	if err := requireOrgUnitManage(c); err != nil {
		return err
	}

	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	return handleTenantScopedUpdate(c, "orgUnitID", "org unit not found", parseBodyInto[platform.UpdateOrgUnitRequest], func(ctx context.Context, tenantID, itemID string, req platform.UpdateOrgUnitRequest) (interface{}, error) {
		return h.orgUnits.UpdateOrgUnit(ctx, tenantID, itemID, req)
	})
}

func (h *Handlers) DeleteOrgUnit(c *fiber.Ctx) error {
	if err := requireOrgUnitDelete(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	if orgUnitID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID is required")
	}

	if err := h.orgUnits.DeleteOrgUnit(c.Context(), currentTenant.ID, orgUnitID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "org unit not found")
		}

		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handlers) ListOrgUnitMembers(c *fiber.Ctx) error {
	if err := requireOrgUnitRead(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	if orgUnitID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID is required")
	}

	items, err := h.orgUnits.ListOrgUnitMemberships(c.Context(), currentTenant.ID, orgUnitID)
	if err != nil {
		return err
	}

	return c.JSON(platform.OrgUnitMembershipListResponse{Data: items})
}

func (h *Handlers) AddOrgUnitMember(c *fiber.Ctx) error {
	if err := requireOrgUnitManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))
	if orgUnitID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID is required")
	}

	var req platform.CreateOrgUnitMembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.UserID = strings.TrimSpace(req.UserID)

	req.Role = strings.TrimSpace(req.Role)
	if req.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id is required")
	}

	if req.Role == "" {
		req.Role = "member"
	}

	item, err := h.orgUnits.AddOrgUnitMembership(c.Context(), currentTenant.ID, orgUnitID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "user not found for this tenant")
		}

		if strings.Contains(strings.ToLower(err.Error()), "invalid uuid") {
			return fiber.NewError(fiber.StatusBadRequest, "user_id must be a UUID platform user id")
		}

		return err
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) UpdateOrgUnitMember(c *fiber.Ctx) error {
	if err := requireOrgUnitManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	orgUnitID := strings.TrimSpace(c.Params("orgUnitID"))

	userID := strings.TrimSpace(c.Params("userID"))
	if orgUnitID == "" || userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "orgUnitID and userID are required")
	}

	var req platform.UpdateOrgUnitMembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Role = strings.TrimSpace(req.Role)
	if req.Role == "" {
		return fiber.NewError(fiber.StatusBadRequest, "role is required")
	}

	item, err := h.orgUnits.UpdateOrgUnitMembership(c.Context(), currentTenant.ID, orgUnitID, userID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "org unit membership not found")
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) DeleteOrgUnitMember(c *fiber.Ctx) error {
	if err := requireOrgUnitManage(c); err != nil {
		return err
	}

	if h.orgUnits == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "org unit store not configured")
	}

	return handleTenantMembershipDelete(c, "orgUnitID", "org unit membership not found", h.orgUnits.DeleteOrgUnitMembership)
}

func (h *Handlers) ListTeamMembers(c *fiber.Ctx) error {
	if err := requireTeamRead(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.memberships == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "membership store not configured")
	}

	if err := h.requireEntitlementEnabled(c, currentTenant.ID, platform.FeatureTeamsEnabled); err != nil {
		return err
	}

	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}

	items, err := h.memberships.ListTeamMemberships(c.Context(), currentTenant.ID, teamID)
	if err != nil {
		return err
	}

	return c.JSON(platform.TeamMembershipListResponse{Data: items})
}

func (h *Handlers) AddTeamMember(c *fiber.Ctx) error {
	if err := requireTeamManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.memberships == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "membership store not configured")
	}

	teamID := strings.TrimSpace(c.Params("teamID"))
	if teamID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID is required")
	}

	var req platform.CreateTeamMembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.UserID = strings.TrimSpace(req.UserID)

	req.Role = strings.TrimSpace(req.Role)
	if req.UserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "user_id is required")
	}

	if req.Role == "" {
		req.Role = "team_member"
	}

	item, err := h.memberships.AddTeamMembership(c.Context(), currentTenant.ID, teamID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) UpdateTeamMember(c *fiber.Ctx) error {
	if err := requireTeamManage(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.memberships == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "membership store not configured")
	}

	teamID := strings.TrimSpace(c.Params("teamID"))

	userID := strings.TrimSpace(c.Params("userID"))
	if teamID == "" || userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "teamID and userID are required")
	}

	var req platform.UpdateTeamMembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Role = strings.TrimSpace(req.Role)

	item, err := h.memberships.UpdateTeamMembership(c.Context(), currentTenant.ID, teamID, userID, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "team membership not found")
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) DeleteTeamMember(c *fiber.Ctx) error {
	if err := requireTeamManage(c); err != nil {
		return err
	}

	if h.memberships == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "membership store not configured")
	}

	return handleTenantMembershipDelete(c, "teamID", "team membership not found", h.memberships.DeleteTeamMembership)
}

func (h *Handlers) ListVirtualUsers(c *fiber.Ctx) error {
	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}

	return handleTenantListResponse(c, requireTeamAdminOrHigher, 200, func(currentTenant *platform.Tenant) string {
		return currentTenant.ID
	}, h.virtualUsers.ListVirtualUsers, func(items []platform.VirtualUser, total, limit, offset int) platform.VirtualUserListResponse {
		return platform.VirtualUserListResponse{
			Data:       items,
			Total:      total,
			Limit:      limit,
			Offset:     offset,
			NextOffset: buildNextOffset(offset, len(items), total),
		}
	})
}

func (h *Handlers) CreateVirtualUser(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}

	if err := h.requireVATEnabled(c, currentTenant.ID); err != nil {
		return err
	}

	var req platform.CreateVirtualUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	req.Alias = strings.TrimSpace(req.Alias)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.TeamID = strings.TrimSpace(req.TeamID)

	req.OrgUnitID = strings.TrimSpace(req.OrgUnitID)
	if req.Alias == "" {
		return fiber.NewError(fiber.StatusBadRequest, "alias is required")
	}

	item, err := h.virtualUsers.CreateVirtualUser(c.Context(), currentTenant.ID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handlers) GetVirtualUser(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}

	virtualUserID := strings.TrimSpace(c.Params("virtualUserID"))
	if virtualUserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "virtualUserID is required")
	}

	item, err := h.virtualUsers.GetVirtualUser(c.Context(), currentTenant.ID, virtualUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "virtual user not found")
		}

		return err
	}

	return c.JSON(item)
}

func (h *Handlers) UpdateVirtualUser(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}

	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}

	return handleTenantScopedUpdate(c, "virtualUserID", "virtual user not found", parseBodyInto[platform.UpdateVirtualUserRequest], func(ctx context.Context, tenantID, itemID string, req platform.UpdateVirtualUserRequest) (interface{}, error) {
		return h.virtualUsers.UpdateVirtualUser(ctx, tenantID, itemID, req)
	})
}

func (h *Handlers) DeleteVirtualUser(c *fiber.Ctx) error {
	if err := requireTeamAdminOrHigher(c); err != nil {
		return err
	}

	currentTenant := tenant.FromContext(c)
	if currentTenant == nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenant context missing")
	}

	if h.virtualUsers == nil {
		return fiber.NewError(fiber.StatusNotImplemented, "virtual user store not configured")
	}

	virtualUserID := strings.TrimSpace(c.Params("virtualUserID"))
	if virtualUserID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "virtualUserID is required")
	}

	if err := h.virtualUsers.DeleteVirtualUser(c.Context(), currentTenant.ID, virtualUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "virtual user not found")
		}

		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
