package iam

import "errors"

var (
	// ErrRoleNotFound indicates that a role does not exist.
	ErrRoleNotFound = errors.New("role not found")
	// ErrRoleImmutable indicates an attempted mutation of a system role.
	ErrRoleImmutable = errors.New("system role cannot be modified")
	// ErrInvalidRoleName indicates an invalid role name.
	ErrInvalidRoleName = errors.New("invalid role name")
	// ErrInvalidPermission indicates an unknown permission.
	ErrInvalidPermission = errors.New("invalid permission")
	// ErrInvalidResourceType indicates an unsupported IAM resource type.
	ErrInvalidResourceType = errors.New("invalid resource type")
	// ErrInvalidUserRoleUpdate indicates an invalid member role update.
	ErrInvalidUserRoleUpdate = errors.New("invalid user role update payload")
	// ErrUserNotFound indicates that a tenant member does not exist.
	ErrUserNotFound = errors.New("user not found")
)
