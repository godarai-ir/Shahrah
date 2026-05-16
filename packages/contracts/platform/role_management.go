package platform

import "time"

// RoleDefinition stores one tenant-scoped role and its permission set.
type RoleDefinition struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description,omitempty"`
	IsSystem    bool      `json:"is_system"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRoleRequest struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleRequest struct {
	DisplayName string   `json:"display_name,omitempty"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type RoleListResponse struct {
	Data       []RoleDefinition `json:"data"`
	Total      int              `json:"total"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
	NextOffset *int             `json:"next_offset,omitempty"`
}

type UpdateUserRolesRequest struct {
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	ResourceType string   `json:"resourceType,omitempty"`
}

type UpdateUserRolesResponse struct {
	UserID       string   `json:"user_id"`
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	ResourceType string   `json:"resourceType,omitempty"`
}

type PermissionDefinition struct {
	Key         string `json:"key"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

type MePermissionsResponse struct {
	UserID       string                 `json:"user_id,omitempty"`
	Role         string                 `json:"role"`
	Roles        []string               `json:"roles"`
	Permissions  []string               `json:"permissions"`
	Catalog      []PermissionDefinition `json:"catalog,omitempty"`
	ResourceType string                 `json:"resource_type,omitempty"`
}
