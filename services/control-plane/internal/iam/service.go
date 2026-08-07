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

type directoryRepository interface {
	CreateOrgUnit(context.Context, string, platform.CreateOrgUnitRequest) (*platform.OrgUnit, error)
	ListOrgUnits(context.Context, string, string, int, int) ([]platform.OrgUnit, int, error)
	GetOrgUnit(context.Context, string, string) (*platform.OrgUnit, error)
	UpdateOrgUnit(context.Context, string, string, platform.UpdateOrgUnitRequest) (*platform.OrgUnit, error)
	DeleteOrgUnit(context.Context, string, string) error
	ListOrgUnitMemberships(context.Context, string, string) ([]platform.OrgUnitMembership, error)
	AddOrgUnitMembership(context.Context, string, string, platform.CreateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error)
	UpdateOrgUnitMembership(context.Context, string, string, string, platform.UpdateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error)
	DeleteOrgUnitMembership(context.Context, string, string, string) error
	CreateTeam(context.Context, string, platform.CreateTeamRequest) (*platform.Team, error)
	ListTeams(context.Context, string, string, int, int) ([]platform.Team, int, error)
	GetTeam(context.Context, string, string) (*platform.Team, error)
	UpdateTeam(context.Context, string, string, platform.UpdateTeamRequest) (*platform.Team, error)
	DeleteTeam(context.Context, string, string) error
	GetTeamPermissions(context.Context, string, string) (*platform.TeamPermissionsResponse, error)
	UpdateTeamPermissions(context.Context, string, string, platform.UpdateTeamPermissionsRequest) (*platform.TeamPermissionsResponse, error)
	ListTeamMemberships(context.Context, string, string) ([]platform.TeamMembership, error)
	AddTeamMembership(context.Context, string, string, platform.CreateTeamMembershipRequest) (*platform.TeamMembership, error)
	UpdateTeamMembership(context.Context, string, string, string, platform.UpdateTeamMembershipRequest) (*platform.TeamMembership, error)
	DeleteTeamMembership(context.Context, string, string, string) error
	CreateVirtualUser(context.Context, string, platform.CreateVirtualUserRequest) (*platform.VirtualUser, error)
	ListVirtualUsers(context.Context, string, string, int, int) ([]platform.VirtualUser, int, error)
	GetVirtualUser(context.Context, string, string) (*platform.VirtualUser, error)
	UpdateVirtualUser(context.Context, string, string, platform.UpdateVirtualUserRequest) (*platform.VirtualUser, error)
	DeleteVirtualUser(context.Context, string, string) error
}

type accessRepository interface {
	GetKeyByHash(context.Context, string) (*platform.AuthContext, error)
	ResolveTenantUserAuth(context.Context, string, string) (*platform.AuthContext, error)
	ListAllowedModelFQNs(context.Context, string, string, string) ([]string, error)
	CreateKey(context.Context, platform.CreateKeyParams) (*platform.CreatedKey, error)
	ListKeys(context.Context, string) ([]platform.CreatedKey, error)
	TouchKeyLastUsed(context.Context, string, time.Time) error
	PATExists(context.Context, string, string, string) (bool, error)
	ListPATs(context.Context, string, string, string, int, int) ([]platform.CreatedKey, int, error)
	RevokePAT(context.Context, string, string, string, string) (bool, error)
	RevokeAllPATs(context.Context, string, string, string) (int64, error)
	ListVATTokens(context.Context, string, string, string, int, int) ([]platform.CreatedKey, int, error)
	UpdateVATToken(context.Context, string, string, string, platform.UpdateVATTokenRequest) (*platform.CreatedKey, error)
	RevokeVATToken(context.Context, string, string, string, string, string) (bool, error)
	RotateVATToken(context.Context, string, string, string, string, string) (*platform.CreatedKey, error)
	ListRoles(context.Context, string, string, int, int) ([]platform.RoleDefinition, int, error)
	CreateRole(context.Context, string, platform.CreateRoleRequest) (*platform.RoleDefinition, error)
	UpdateRole(context.Context, string, string, platform.UpdateRoleRequest) (*platform.RoleDefinition, error)
	UpdateUserRoles(context.Context, string, platform.UpdateUserRolesRequest) (*platform.UpdateUserRolesResponse, error)
	GetEffectivePermissionsForUser(context.Context, string, string, string) ([]string, []string, error)
}

// Service is the IAM application boundary used by transport adapters.
// Repository methods are promoted intentionally until IAM use cases gain
// orchestration that belongs above persistence.
type Service struct {
	directoryRepository
	accessRepository
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
		directoryRepository: directory,
		accessRepository:    access,
	}, nil
}
