// Package tenantadmin owns tenant administration authentication and onboarding use cases.
package tenantadmin

import (
	"errors"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var (
	// ErrTenantAdminUserNotFound indicates that an active tenant member could not be found.
	ErrTenantAdminUserNotFound = errors.New("tenant member not found")
	// ErrTenantSlugInvalid indicates an invalid or reserved tenant slug.
	ErrTenantSlugInvalid = errors.New("tenant slug is invalid or reserved")
	// ErrTenantSlugTaken indicates an existing tenant slug.
	ErrTenantSlugTaken = errors.New("tenant slug is already taken")
	// ErrOnboardingNotFound indicates an unknown onboarding request.
	ErrOnboardingNotFound = errors.New("onboarding request not found")
	// ErrOnboardingExpired indicates an expired onboarding request.
	ErrOnboardingExpired = errors.New("onboarding request expired")
	// ErrOnboardingOTPFailed indicates an invalid onboarding OTP.
	ErrOnboardingOTPFailed = errors.New("invalid onboarding otp")
	// ErrSocialAccountInvalid indicates incomplete trusted social identity data.
	ErrSocialAccountInvalid = errors.New("social account is missing or its email is not verified")
	// ErrSocialAccessDenied indicates that a social account is not linked to a tenant.
	ErrSocialAccessDenied = errors.New("social account is not linked to this tenant")
	// ErrSocialHandoffInvalid indicates an invalid or consumed handoff.
	ErrSocialHandoffInvalid = errors.New("social handoff is invalid, expired, or already consumed")
	// ErrSocialLinkConflict indicates a conflicting social identity link.
	ErrSocialLinkConflict = errors.New("user or social account is already linked to another account")
	// ErrInvalidInvitationToken indicates an invalid tenant-member invitation token.
	ErrInvalidInvitationToken = errors.New("invalid tenant admin invitation token")
	// ErrInvitationExpired indicates an expired tenant-member invitation.
	ErrInvitationExpired = errors.New("tenant admin invitation expired")
	// ErrInvitationAlreadyUsed indicates a consumed tenant-member invitation.
	ErrInvitationAlreadyUsed = errors.New("tenant admin invitation already used")
	// ErrInvitationContactTaken indicates an invitation contact used by another member.
	ErrInvitationContactTaken = errors.New("tenant admin invitation contact already used")
	// ErrInvitationAlreadyPending indicates an existing pending invitation.
	ErrInvitationAlreadyPending = errors.New("tenant admin invitation already pending")
	// ErrTenantMemberAlreadyExists indicates an existing tenant member.
	ErrTenantMemberAlreadyExists = errors.New("tenant admin user already member of tenant")
	// ErrTenantMemberDeactivated indicates a deactivated tenant member.
	ErrTenantMemberDeactivated = errors.New("tenant admin user is deactivated")
	// ErrInvalidInvitationEmail indicates an invalid invitation email address.
	ErrInvalidInvitationEmail = errors.New("invalid tenant admin invitation email")
)

// TenantAdminOTPResponse is the result of creating an OTP request.
type TenantAdminOTPResponse struct {
	RequestID         string `json:"request_id"`
	Channel           string `json:"channel"`
	DestinationMasked string `json:"destination_masked"`
	ExpiresInSeconds  int    `json:"expires_in_seconds"`
	OTPCode           string `json:"-"`
	DebugOTP          string `json:"debug_otp,omitempty"`
}

// TenantAdminOTPVerification is the typed result of verifying an OTP request.
type TenantAdminOTPVerification struct {
	OK     bool                          `json:"ok"`
	Error  string                        `json:"error,omitempty"`
	Record *TenantAdminOTPRecordResponse `json:"record,omitempty"`
}

// TenantAdminOTPRecordResponse is the public representation of a verified OTP record.
type TenantAdminOTPRecordResponse struct {
	ID           string  `json:"id"`
	TenantSlug   string  `json:"tenant_slug"`
	Purpose      string  `json:"purpose"`
	Channel      string  `json:"channel"`
	Identifier   string  `json:"identifier"`
	UserID       *string `json:"user_id"`
	PendingValue *string `json:"pending_value"`
	OTPCode      string  `json:"otp_code"`
	Attempts     int     `json:"attempts"`
	ExpiresAtISO string  `json:"expires_at_iso"`
	CreatedAtISO string  `json:"created_at_iso"`
}

// TenantAdminSessionResponse is the public representation of a tenant-admin session.
type TenantAdminSessionResponse struct {
	ID           string `json:"id"`
	TenantSlug   string `json:"tenant_slug"`
	UserID       string `json:"user_id"`
	CreatedAtISO string `json:"created_at_iso"`
	ExpiresAtISO string `json:"expires_at_iso"`
}

// TenantAdminUserResponse is the public representation of a tenant-admin user.
type TenantAdminUserResponse struct {
	ID              string   `json:"id"`
	TenantSlug      string   `json:"tenant_slug"`
	Email           string   `json:"email"`
	Mobile          string   `json:"mobile"`
	Role            string   `json:"role"`
	Active          bool     `json:"active"`
	SocialProviders []string `json:"social_providers,omitempty"`
	CreatedAtISO    string   `json:"created_at_iso"`
	UpdatedAtISO    string   `json:"updated_at_iso"`
	LastLoginAtISO  string   `json:"last_login_at_iso,omitempty"`
}

// TenantAdminOTPRequestParams holds parameters for creating an OTP request.
type TenantAdminOTPRequestParams struct {
	TenantSlug   string
	Purpose      string
	Channel      string
	Identifier   string
	UserID       *string
	PendingValue *string
}

// VerifyTenantAdminOTPRequestParams holds parameters for verifying an OTP.
type VerifyTenantAdminOTPRequestParams struct {
	TenantSlug string
	RequestID  string
	OTPCode    string
	Purpose    string
	UserID     *string
}

// CreateTenantAdminSessionParams holds parameters for creating a session.
type CreateTenantAdminSessionParams struct {
	TenantSlug string
	UserID     string
}

// FindOrCreateTenantAdminUserParams holds parameters for finding a user.
type FindOrCreateTenantAdminUserParams struct {
	TenantSlug string
	Channel    string
	Identifier string
}

// UpdateTenantAdminUserContactParams holds parameters for updating user contact.
type UpdateTenantAdminUserContactParams struct {
	TenantSlug string
	UserID     string
	Field      string
	Value      string
}

// ProvisionSocialTenantParams contains trusted social provisioning input.
type ProvisionSocialTenantParams struct {
	AccountID    string `json:"account_id"`
	TenantSlug   string `json:"tenant_slug"`
	TenantDomain string `json:"tenant_domain"`
	PlanSlug     string `json:"plan_slug"`
}

// SocialHandoff is a short-lived social login or link handoff.
type SocialHandoff struct {
	Code       string    `json:"code"`
	TenantSlug string    `json:"tenant_slug"`
	UserID     string    `json:"user_id"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// RedeemedSocialHandoff contains the resulting session and user.
type RedeemedSocialHandoff struct {
	Session *TenantAdminSessionResponse `json:"session"`
	User    *TenantAdminUserResponse    `json:"user"`
}

// TenantSlugAvailability describes whether a normalized tenant slug is available.
type TenantSlugAvailability struct {
	Slug      string `json:"slug"`
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"`
}

// CreateOnboardingOTPParams contains tenant onboarding input.
type CreateOnboardingOTPParams struct {
	TenantSlug   string                 `json:"tenant_slug"`
	TenantDomain string                 `json:"tenant_domain"`
	PlanSlug     string                 `json:"plan_slug"`
	Email        string                 `json:"email"`
	Mobile       string                 `json:"mobile"`
	CustomerName string                 `json:"customer_name"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// OnboardingOTP is the result of starting tenant onboarding.
type OnboardingOTP struct {
	RequestID         string `json:"request_id"`
	TenantSlug        string `json:"tenant_slug"`
	Channel           string `json:"channel"`
	DestinationMasked string `json:"destination_masked"`
	ExpiresInSeconds  int    `json:"expires_in_seconds"`
	AlreadyPending    bool   `json:"already_pending,omitempty"`
	DebugOTP          string `json:"debug_otp,omitempty"`
}

// VerifyOnboardingOTPParams contains onboarding verification input.
type VerifyOnboardingOTPParams struct {
	RequestID string `json:"request_id"`
	OTPCode   string `json:"otp_code"`
}

// ProvisionedTenant is the result of successful tenant provisioning.
type ProvisionedTenant struct {
	Tenant   platform.Tenant `json:"tenant"`
	LoginURL string          `json:"login_url"`
}
