// Package iam owns identity and access management application boundaries.
package iam

import (
	"context"
	"errors"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var (
	errDirectoryRepositoryRequired = errors.New("iam: directory repository is required")
	errAccessRepositoryRequired    = errors.New("iam: access repository is required")
)

type orgUnitRepository interface {
	CreateOrgUnit(context.Context, string, platform.CreateOrgUnitRequest) (*platform.OrgUnit, error)
	ListOrgUnits(context.Context, string, string, int, int) ([]platform.OrgUnit, int, error)
	GetOrgUnit(context.Context, string, string) (*platform.OrgUnit, error)
	UpdateOrgUnit(context.Context, string, string, platform.UpdateOrgUnitRequest) (*platform.OrgUnit, error)
	DeleteOrgUnit(context.Context, string, string) error
	ListOrgUnitMemberships(context.Context, string, string) ([]platform.OrgUnitMembership, error)
	AddOrgUnitMembership(context.Context, string, string, platform.CreateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error)
	UpdateOrgUnitMembership(context.Context, string, string, string, platform.UpdateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error)
	DeleteOrgUnitMembership(context.Context, string, string, string) error
}

type teamRepository interface {
	CreateTeam(context.Context, string, platform.CreateTeamRequest) (*platform.Team, error)
	ListTeams(context.Context, string, string, int, int) ([]platform.Team, int, error)
	GetTeam(context.Context, string, string) (*platform.Team, error)
	UpdateTeam(context.Context, string, string, platform.UpdateTeamRequest) (*platform.Team, error)
	DeleteTeam(context.Context, string, string) error
	GetTeamPermissions(context.Context, string, string) (*platform.TeamPermissionsResponse, error)
	UpdateTeamPermissions(context.Context, string, string, platform.UpdateTeamPermissionsRequest) (*platform.TeamPermissionsResponse, error)
}

type teamMembershipRepository interface {
	ListTeamMemberships(context.Context, string, string) ([]platform.TeamMembership, error)
	AddTeamMembership(context.Context, string, string, platform.CreateTeamMembershipRequest) (*platform.TeamMembership, error)
	UpdateTeamMembership(context.Context, string, string, string, platform.UpdateTeamMembershipRequest) (*platform.TeamMembership, error)
	DeleteTeamMembership(context.Context, string, string, string) error
}

type virtualUserRepository interface {
	CreateVirtualUser(context.Context, string, platform.CreateVirtualUserRequest) (*platform.VirtualUser, error)
	ListVirtualUsers(context.Context, string, string, int, int) ([]platform.VirtualUser, int, error)
	GetVirtualUser(context.Context, string, string) (*platform.VirtualUser, error)
	UpdateVirtualUser(context.Context, string, string, platform.UpdateVirtualUserRequest) (*platform.VirtualUser, error)
	DeleteVirtualUser(context.Context, string, string) error
}

type directoryRepository interface {
	orgUnitRepository
	teamRepository
	teamMembershipRepository
	virtualUserRepository
}

type authenticationRepository interface {
	GetKeyByHash(context.Context, string) (*platform.AuthContext, error)
	ResolveTenantUserAuth(context.Context, string, string) (*platform.AuthContext, error)
	ListAllowedModelFQNs(context.Context, string, string, string) ([]string, error)
}

type apiKeyRepository interface {
	CreateKey(context.Context, platform.CreateKeyParams) (*platform.CreatedKey, error)
	ListKeys(context.Context, string) ([]platform.CreatedKey, error)
	TouchKeyLastUsed(context.Context, string, time.Time) error
}

type personalTokenRepository interface {
	PATExists(context.Context, string, string, string) (bool, error)
	ListPATs(context.Context, string, string, string, int, int) ([]platform.CreatedKey, int, error)
	RevokePAT(context.Context, string, string, string, string) (bool, error)
	RevokeAllPATs(context.Context, string, string, string) (int64, error)
}

type virtualTokenRepository interface {
	ListVATTokens(context.Context, string, string, string, int, int) ([]platform.CreatedKey, int, error)
	UpdateVATToken(context.Context, string, string, string, platform.UpdateVATTokenRequest) (*platform.CreatedKey, error)
	SetVATOnBehalfOf(context.Context, string, string, string, platform.UpdateVATOnBehalfOfRequest, string) (*platform.VATOnBehalfOf, error)
	RevokeVATToken(context.Context, string, string, string, string, string) (bool, error)
	RotateVATToken(context.Context, string, string, string, string, string) (*platform.CreatedKey, error)
}

type roleRepository interface {
	ListRoles(context.Context, string, string, int, int) ([]platform.RoleDefinition, int, error)
	CreateRole(context.Context, string, platform.CreateRoleRequest) (*platform.RoleDefinition, error)
	UpdateRole(context.Context, string, string, platform.UpdateRoleRequest) (*platform.RoleDefinition, error)
	UpdateUserRoles(context.Context, string, platform.UpdateUserRolesRequest) (*platform.UpdateUserRolesResponse, error)
	GetEffectivePermissionsForUser(context.Context, string, string, string) ([]string, []string, error)
}

type accessRepository interface {
	authenticationRepository
	apiKeyRepository
	personalTokenRepository
	virtualTokenRepository
	roleRepository
}

// Service is the IAM application boundary used by transport adapters.
type Service struct {
	orgUnits       orgUnitRepository
	teams          teamRepository
	teamMembers    teamMembershipRepository
	virtualUsers   virtualUserRepository
	authentication authenticationRepository
	apiKeys        apiKeyRepository
	personalTokens personalTokenRepository
	virtualTokens  virtualTokenRepository
	roles          roleRepository
}

// NewService creates an IAM application service over a persistence port.
func NewService(directory directoryRepository, access accessRepository) (*Service, error) {
	if directory == nil {
		return nil, errDirectoryRepositoryRequired
	}

	if access == nil {
		return nil, errAccessRepositoryRequired
	}

	return &Service{
		orgUnits:       directory,
		teams:          directory,
		teamMembers:    directory,
		virtualUsers:   directory,
		authentication: access,
		apiKeys:        access,
		personalTokens: access,
		virtualTokens:  access,
		roles:          access,
	}, nil
}

// CreateOrgUnit creates an organizational unit.
func (s *Service) CreateOrgUnit(ctx context.Context, tenantID string, request platform.CreateOrgUnitRequest) (*platform.OrgUnit, error) {
	return s.orgUnits.CreateOrgUnit(ctx, tenantID, request)
}

// ListOrgUnits lists organizational units.
func (s *Service) ListOrgUnits(ctx context.Context, tenantID, query string, limit, offset int) ([]platform.OrgUnit, int, error) {
	return s.orgUnits.ListOrgUnits(ctx, tenantID, query, limit, offset)
}

// GetOrgUnit gets an organizational unit.
func (s *Service) GetOrgUnit(ctx context.Context, tenantID, orgUnitID string) (*platform.OrgUnit, error) {
	return s.orgUnits.GetOrgUnit(ctx, tenantID, orgUnitID)
}

// UpdateOrgUnit updates an organizational unit.
func (s *Service) UpdateOrgUnit(ctx context.Context, tenantID, orgUnitID string, request platform.UpdateOrgUnitRequest) (*platform.OrgUnit, error) {
	return s.orgUnits.UpdateOrgUnit(ctx, tenantID, orgUnitID, request)
}

// DeleteOrgUnit deletes an organizational unit.
func (s *Service) DeleteOrgUnit(ctx context.Context, tenantID, orgUnitID string) error {
	return s.orgUnits.DeleteOrgUnit(ctx, tenantID, orgUnitID)
}

// ListOrgUnitMemberships lists organizational-unit memberships.
func (s *Service) ListOrgUnitMemberships(ctx context.Context, tenantID, orgUnitID string) ([]platform.OrgUnitMembership, error) {
	return s.orgUnits.ListOrgUnitMemberships(ctx, tenantID, orgUnitID)
}

// AddOrgUnitMembership adds an organizational-unit membership.
func (s *Service) AddOrgUnitMembership(ctx context.Context, tenantID, orgUnitID string, request platform.CreateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error) {
	return s.orgUnits.AddOrgUnitMembership(ctx, tenantID, orgUnitID, request)
}

// UpdateOrgUnitMembership updates an organizational-unit membership.
func (s *Service) UpdateOrgUnitMembership(ctx context.Context, tenantID, orgUnitID, userID string, request platform.UpdateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error) {
	return s.orgUnits.UpdateOrgUnitMembership(ctx, tenantID, orgUnitID, userID, request)
}

// DeleteOrgUnitMembership deletes an organizational-unit membership.
func (s *Service) DeleteOrgUnitMembership(ctx context.Context, tenantID, orgUnitID, userID string) error {
	return s.orgUnits.DeleteOrgUnitMembership(ctx, tenantID, orgUnitID, userID)
}

// CreateTeam creates a team.
func (s *Service) CreateTeam(ctx context.Context, tenantID string, request platform.CreateTeamRequest) (*platform.Team, error) {
	return s.teams.CreateTeam(ctx, tenantID, request)
}

// ListTeams lists teams.
func (s *Service) ListTeams(ctx context.Context, tenantID, query string, limit, offset int) ([]platform.Team, int, error) {
	return s.teams.ListTeams(ctx, tenantID, query, limit, offset)
}

// GetTeam gets a team.
func (s *Service) GetTeam(ctx context.Context, tenantID, teamID string) (*platform.Team, error) {
	return s.teams.GetTeam(ctx, tenantID, teamID)
}

// UpdateTeam updates a team.
func (s *Service) UpdateTeam(ctx context.Context, tenantID, teamID string, request platform.UpdateTeamRequest) (*platform.Team, error) {
	return s.teams.UpdateTeam(ctx, tenantID, teamID, request)
}

// DeleteTeam deletes a team.
func (s *Service) DeleteTeam(ctx context.Context, tenantID, teamID string) error {
	return s.teams.DeleteTeam(ctx, tenantID, teamID)
}

// GetTeamPermissions gets team permissions.
func (s *Service) GetTeamPermissions(ctx context.Context, tenantID, teamID string) (*platform.TeamPermissionsResponse, error) {
	return s.teams.GetTeamPermissions(ctx, tenantID, teamID)
}

// UpdateTeamPermissions updates team permissions.
func (s *Service) UpdateTeamPermissions(ctx context.Context, tenantID, teamID string, request platform.UpdateTeamPermissionsRequest) (*platform.TeamPermissionsResponse, error) {
	return s.teams.UpdateTeamPermissions(ctx, tenantID, teamID, request)
}

// ListTeamMemberships lists team memberships.
func (s *Service) ListTeamMemberships(ctx context.Context, tenantID, teamID string) ([]platform.TeamMembership, error) {
	return s.teamMembers.ListTeamMemberships(ctx, tenantID, teamID)
}

// AddTeamMembership adds a team membership.
func (s *Service) AddTeamMembership(ctx context.Context, tenantID, teamID string, request platform.CreateTeamMembershipRequest) (*platform.TeamMembership, error) {
	return s.teamMembers.AddTeamMembership(ctx, tenantID, teamID, request)
}

// UpdateTeamMembership updates a team membership.
func (s *Service) UpdateTeamMembership(ctx context.Context, tenantID, teamID, userID string, request platform.UpdateTeamMembershipRequest) (*platform.TeamMembership, error) {
	return s.teamMembers.UpdateTeamMembership(ctx, tenantID, teamID, userID, request)
}

// DeleteTeamMembership deletes a team membership.
func (s *Service) DeleteTeamMembership(ctx context.Context, tenantID, teamID, userID string) error {
	return s.teamMembers.DeleteTeamMembership(ctx, tenantID, teamID, userID)
}

// CreateVirtualUser creates a virtual user.
func (s *Service) CreateVirtualUser(ctx context.Context, tenantID string, request platform.CreateVirtualUserRequest) (*platform.VirtualUser, error) {
	return s.virtualUsers.CreateVirtualUser(ctx, tenantID, request)
}

// ListVirtualUsers lists virtual users.
func (s *Service) ListVirtualUsers(ctx context.Context, tenantID, query string, limit, offset int) ([]platform.VirtualUser, int, error) {
	return s.virtualUsers.ListVirtualUsers(ctx, tenantID, query, limit, offset)
}

// GetVirtualUser gets a virtual user.
func (s *Service) GetVirtualUser(ctx context.Context, tenantID, virtualUserID string) (*platform.VirtualUser, error) {
	return s.virtualUsers.GetVirtualUser(ctx, tenantID, virtualUserID)
}

// UpdateVirtualUser updates a virtual user.
func (s *Service) UpdateVirtualUser(ctx context.Context, tenantID, virtualUserID string, request platform.UpdateVirtualUserRequest) (*platform.VirtualUser, error) {
	return s.virtualUsers.UpdateVirtualUser(ctx, tenantID, virtualUserID, request)
}

// DeleteVirtualUser deletes a virtual user.
func (s *Service) DeleteVirtualUser(ctx context.Context, tenantID, virtualUserID string) error {
	return s.virtualUsers.DeleteVirtualUser(ctx, tenantID, virtualUserID)
}

// GetKeyByHash resolves an API key authentication context.
func (s *Service) GetKeyByHash(ctx context.Context, hash string) (*platform.AuthContext, error) {
	return s.authentication.GetKeyByHash(ctx, hash)
}

// ResolveTenantUserAuth resolves a tenant member authentication context.
func (s *Service) ResolveTenantUserAuth(ctx context.Context, tenantID, userID string) (*platform.AuthContext, error) {
	return s.authentication.ResolveTenantUserAuth(ctx, tenantID, userID)
}

// ListAllowedModelFQNs lists models available to a principal.
func (s *Service) ListAllowedModelFQNs(ctx context.Context, tenantID, userID, teamID string) ([]string, error) {
	return s.authentication.ListAllowedModelFQNs(ctx, tenantID, userID, teamID)
}

// CreateKey creates an API key.
func (s *Service) CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error) {
	return s.apiKeys.CreateKey(ctx, params)
}

// ListKeys lists tenant API keys.
func (s *Service) ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error) {
	return s.apiKeys.ListKeys(ctx, tenantID)
}

// TouchKeyLastUsed records API-key activity.
func (s *Service) TouchKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error {
	return s.apiKeys.TouchKeyLastUsed(ctx, keyID, usedAt)
}

// PATExists reports whether a personal access token name exists.
func (s *Service) PATExists(ctx context.Context, tenantID, userID, name string) (bool, error) {
	return s.personalTokens.PATExists(ctx, tenantID, userID, name)
}

// ListPATs lists personal access tokens.
func (s *Service) ListPATs(ctx context.Context, tenantID, userID, query string, limit, offset int) ([]platform.CreatedKey, int, error) {
	return s.personalTokens.ListPATs(ctx, tenantID, userID, query, limit, offset)
}

// RevokePAT revokes a personal access token.
func (s *Service) RevokePAT(ctx context.Context, tenantID, userID, keyID, actor string) (bool, error) {
	return s.personalTokens.RevokePAT(ctx, tenantID, userID, keyID, actor)
}

// RevokeAllPATs revokes all personal access tokens for a member.
func (s *Service) RevokeAllPATs(ctx context.Context, tenantID, userID, actor string) (int64, error) {
	return s.personalTokens.RevokeAllPATs(ctx, tenantID, userID, actor)
}

// ListVATTokens lists virtual access tokens.
func (s *Service) ListVATTokens(ctx context.Context, tenantID, virtualUserID, query string, limit, offset int) ([]platform.CreatedKey, int, error) {
	return s.virtualTokens.ListVATTokens(ctx, tenantID, virtualUserID, query, limit, offset)
}

// UpdateVATToken updates a virtual access token.
func (s *Service) UpdateVATToken(ctx context.Context, tenantID, virtualUserID, keyID string, request platform.UpdateVATTokenRequest) (*platform.CreatedKey, error) {
	return s.virtualTokens.UpdateVATToken(ctx, tenantID, virtualUserID, keyID, request)
}

// SetVATOnBehalfOf enables or disables the on-behalf-of capability of a virtual access token.
func (s *Service) SetVATOnBehalfOf(
	ctx context.Context,
	tenantID, virtualUserID, keyID string,
	request platform.UpdateVATOnBehalfOfRequest,
	actor string,
) (*platform.VATOnBehalfOf, error) {
	return s.virtualTokens.SetVATOnBehalfOf(ctx, tenantID, virtualUserID, keyID, request, actor)
}

// RevokeVATToken revokes a virtual access token.
func (s *Service) RevokeVATToken(ctx context.Context, tenantID, virtualUserID, keyID, actor, reason string) (bool, error) {
	return s.virtualTokens.RevokeVATToken(ctx, tenantID, virtualUserID, keyID, actor, reason)
}

// RotateVATToken rotates a virtual access token.
func (s *Service) RotateVATToken(ctx context.Context, tenantID, virtualUserID, keyID, actor, gracePeriod string) (*platform.CreatedKey, error) {
	return s.virtualTokens.RotateVATToken(ctx, tenantID, virtualUserID, keyID, actor, gracePeriod)
}

// ListRoles lists tenant roles.
func (s *Service) ListRoles(ctx context.Context, tenantID, query string, limit, offset int) ([]platform.RoleDefinition, int, error) {
	return s.roles.ListRoles(ctx, tenantID, query, limit, offset)
}

// CreateRole creates a tenant role.
func (s *Service) CreateRole(ctx context.Context, tenantID string, request platform.CreateRoleRequest) (*platform.RoleDefinition, error) {
	return s.roles.CreateRole(ctx, tenantID, request)
}

// UpdateRole updates a tenant role.
func (s *Service) UpdateRole(ctx context.Context, tenantID, roleID string, request platform.UpdateRoleRequest) (*platform.RoleDefinition, error) {
	return s.roles.UpdateRole(ctx, tenantID, roleID, request)
}

// UpdateUserRoles updates a member's role bindings.
func (s *Service) UpdateUserRoles(ctx context.Context, tenantID string, request platform.UpdateUserRolesRequest) (*platform.UpdateUserRolesResponse, error) {
	return s.roles.UpdateUserRoles(ctx, tenantID, request)
}

// GetEffectivePermissionsForUser resolves effective member permissions.
func (s *Service) GetEffectivePermissionsForUser(ctx context.Context, tenantID, userID, fallbackRole string) ([]string, []string, error) {
	return s.roles.GetEffectivePermissionsForUser(ctx, tenantID, userID, fallbackRole)
}
