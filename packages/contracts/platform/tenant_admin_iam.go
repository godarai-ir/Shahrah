package platform

import "time"

// TenantAdminUser represents a tenant-admin identity record.
type TenantAdminUser struct {
	ID          string     `json:"id"`
	TenantSlug  string     `json:"tenant_slug"`
	Email       string     `json:"email,omitempty"`
	Mobile      string     `json:"mobile,omitempty"`
	FirstName   string     `json:"first_name,omitempty"`
	LastName    string     `json:"last_name,omitempty"`
	Role        string     `json:"role"`
	Roles       []string   `json:"roles,omitempty"`
	Active      bool       `json:"active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TenantConsoleUser = TenantAdminUser

// TenantAdminUserListResponse returns paginated users for one tenant.
type TenantAdminUserListResponse struct {
	Data       []TenantAdminUser `json:"data"`
	Total      int               `json:"total"`
	Limit      int               `json:"limit"`
	Offset     int               `json:"offset"`
	NextOffset *int              `json:"next_offset,omitempty"`
}

type TenantConsoleUserListResponse = TenantAdminUserListResponse

// InviteTenantAdminUserRequest creates an invitation for tenant-admin access.
type InviteTenantAdminUserRequest struct {
	Email                 string `json:"email,omitempty"`
	Mobile                string `json:"mobile,omitempty"`
	FirstName             string `json:"first_name,omitempty"`
	LastName              string `json:"last_name,omitempty"`
	Role                  string `json:"role,omitempty"`
	AcceptInviteClientURL string `json:"accept_invite_client_url,omitempty"`
	// SeatType is the seat the member gets on acceptance: free, pro, max, or
	// unassigned. Empty means the free seat when one is left.
	SeatType string `json:"seat_type,omitempty"`
}

// UpdateTenantAdminUserProfileRequest sets the first and last name of a tenant member.
type UpdateTenantAdminUserProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// TenantAdminInvitation stores invitation metadata.
type TenantAdminInvitation struct {
	InvitationID string     `json:"invitation_id"`
	TenantID     string     `json:"tenant_id"`
	TenantSlug   string     `json:"tenant_slug"`
	Email        string     `json:"email,omitempty"`
	Mobile       string     `json:"mobile,omitempty"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	ExpiresAt    time.Time  `json:"expires_at"`
	AcceptedAt   *time.Time `json:"accepted_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// InviteTenantAdminUserResponse returns the invitation URL for delivery.
type InviteTenantAdminUserResponse struct {
	Invitation TenantAdminInvitation `json:"invitation"`
	Link       string                `json:"link"`
}

// AcceptTenantAdminInvitationRequest accepts an invitation token.
type AcceptTenantAdminInvitationRequest struct {
	Token string `json:"token"`
}

// AcceptTenantAdminInvitationResponse returns accepted invitation and user.
type AcceptTenantAdminInvitationResponse struct {
	Invitation TenantAdminInvitation `json:"invitation"`
	User       TenantAdminUser       `json:"user"`
}
