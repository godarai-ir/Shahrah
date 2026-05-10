package platform

import "time"

// OrgUnit is the canonical org-scoped IAM entity.
type OrgUnit struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Alias       string                 `json:"alias"`
	DisplayName string                 `json:"display_name,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type CreateOrgUnitRequest struct {
	Alias       string                 `json:"alias"`
	DisplayName string                 `json:"display_name,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateOrgUnitRequest struct {
	Alias       string                 `json:"alias,omitempty"`
	DisplayName string                 `json:"display_name,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type OrgUnitListResponse struct {
	Data       []OrgUnit `json:"data"`
	Total      int       `json:"total"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
	NextOffset *int      `json:"next_offset,omitempty"`
}

type OrgUnitMembership struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	OrgUnitID string    `json:"org_unit_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateOrgUnitMembershipRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type UpdateOrgUnitMembershipRequest struct {
	Role string `json:"role"`
}

type OrgUnitMembershipListResponse struct {
	Data []OrgUnitMembership `json:"data"`
}

// Team is the canonical team workspace entity.
type Team struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id"`
	Alias         string                 `json:"alias"`
	DisplayName   string                 `json:"display_name,omitempty"`
	OrgUnitID     string                 `json:"org_unit_id,omitempty"`
	WorkspaceType string                 `json:"workspace_type,omitempty"`
	AllowedModels []string               `json:"allowed_models,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type CreateTeamRequest struct {
	Alias         string                 `json:"alias"`
	DisplayName   string                 `json:"display_name,omitempty"`
	OrgUnitID     string                 `json:"org_unit_id,omitempty"`
	WorkspaceType string                 `json:"workspace_type,omitempty"`
	AllowedModels []string               `json:"allowed_models,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateTeamRequest struct {
	Alias         string                 `json:"alias,omitempty"`
	DisplayName   string                 `json:"display_name,omitempty"`
	OrgUnitID     string                 `json:"org_unit_id,omitempty"`
	WorkspaceType string                 `json:"workspace_type,omitempty"`
	AllowedModels []string               `json:"allowed_models,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type TeamListResponse struct {
	Data       []Team `json:"data"`
	Total      int    `json:"total"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	NextOffset *int   `json:"next_offset,omitempty"`
}

type TeamPermissionsResponse struct {
	TeamID                 string   `json:"team_id"`
	TeamMemberPermissions  []string `json:"team_member_permissions"`
}

type UpdateTeamPermissionsRequest struct {
	TeamMemberPermissions []string `json:"team_member_permissions"`
}

// TeamMembership represents one user's scoped membership in one team.
type TeamMembership struct {
	ID                    string    `json:"id"`
	TenantID              string    `json:"tenant_id"`
	TeamID                string    `json:"team_id"`
	UserID                string    `json:"user_id"`
	Role                  string    `json:"role"`
	TeamMemberPermissions []string  `json:"team_member_permissions,omitempty"`
	Spend                 float64   `json:"spend"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CreateTeamMembershipRequest struct {
	UserID                string   `json:"user_id"`
	Role                  string   `json:"role"`
	TeamMemberPermissions []string `json:"team_member_permissions,omitempty"`
}

type UpdateTeamMembershipRequest struct {
	Role                  string   `json:"role,omitempty"`
	TeamMemberPermissions []string `json:"team_member_permissions,omitempty"`
}

type TeamMembershipListResponse struct {
	Data []TeamMembership `json:"data"`
}

// VirtualUser is the canonical non-human principal for service access.
type VirtualUser struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	OrgUnitID   string                 `json:"org_unit_id,omitempty"`
	TeamID      string                 `json:"team_id,omitempty"`
	Alias       string                 `json:"alias"`
	DisplayName string                 `json:"display_name,omitempty"`
	Active      bool                   `json:"active"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type CreateVirtualUserRequest struct {
	Alias       string                 `json:"alias"`
	DisplayName string                 `json:"display_name,omitempty"`
	OrgUnitID   string                 `json:"org_unit_id,omitempty"`
	TeamID      string                 `json:"team_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateVirtualUserRequest struct {
	Alias       string                 `json:"alias,omitempty"`
	DisplayName string                 `json:"display_name,omitempty"`
	OrgUnitID   string                 `json:"org_unit_id,omitempty"`
	TeamID      string                 `json:"team_id,omitempty"`
	Active      *bool                  `json:"active,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type VirtualUserListResponse struct {
	Data       []VirtualUser `json:"data"`
	Total      int           `json:"total"`
	Limit      int           `json:"limit"`
	Offset     int           `json:"offset"`
	NextOffset *int          `json:"next_offset,omitempty"`
}
