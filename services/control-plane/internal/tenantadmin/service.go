package tenantadmin

import (
	"context"
	"errors"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var errRepositoryRequired = errors.New("tenantadmin: repository is required")

type repository interface {
	CreateTenantAdminCaptcha(context.Context) (string, string, time.Time, error)
	ConsumeTenantAdminCaptcha(context.Context, string, string) (bool, error)
	CreateTenantAdminOTPRequest(context.Context, TenantAdminOTPRequestParams) (*TenantAdminOTPResponse, error)
	VerifyTenantAdminOTPRequest(context.Context, VerifyTenantAdminOTPRequestParams) (*TenantAdminOTPVerification, error)
	CreateTenantAdminSession(context.Context, CreateTenantAdminSessionParams) (*TenantAdminSessionResponse, error)
	GetTenantAdminSession(context.Context, string) (*TenantAdminSessionResponse, error)
	RevokeTenantAdminSession(context.Context, string) error
	FindOrCreateTenantAdminUser(context.Context, FindOrCreateTenantAdminUserParams) (*TenantAdminUserResponse, error)
	GetTenantAdminUser(context.Context, string, string) (*TenantAdminUserResponse, error)
	UpdateTenantAdminUserContact(context.Context, UpdateTenantAdminUserContactParams) (*TenantAdminUserResponse, error)
	TouchTenantAdminLastLogin(context.Context, string, string) (*TenantAdminUserResponse, error)
	ProvisionSocialTenant(context.Context, ProvisionSocialTenantParams) (*ProvisionedTenant, error)
	CreateSocialLoginHandoff(context.Context, string, string) (*SocialHandoff, error)
	RedeemSocialLoginHandoff(context.Context, string, string) (*RedeemedSocialHandoff, error)
	CreateSocialLinkTicket(context.Context, string, string) (*SocialHandoff, error)
	CompleteSocialLink(context.Context, string, string) (*SocialHandoff, error)
	CheckTenantSlugAvailability(context.Context, string) (*TenantSlugAvailability, error)
	CreateOnboardingOTP(context.Context, CreateOnboardingOTPParams) (*OnboardingOTP, error)
	VerifyOnboardingOTPAndProvision(context.Context, VerifyOnboardingOTPParams) (*ProvisionedTenant, error)
	ListTenantAdminUsers(context.Context, string, string, int, int) ([]platform.TenantAdminUser, int, error)
	GetTenantAdminUserRecord(context.Context, string, string) (*platform.TenantAdminUser, error)
	InviteTenantAdminUser(context.Context, string, string, platform.InviteTenantAdminUserRequest) (*platform.TenantAdminInvitation, string, error)
	AcceptTenantAdminInvitation(context.Context, string, string, string) (*platform.TenantAdminInvitation, *platform.TenantAdminUser, error)
	SetTenantAdminUserActive(context.Context, string, string, bool) (*platform.TenantAdminUser, error)
}

// Service is the tenant-admin authentication and onboarding application boundary.
type Service struct {
	repository
}

// NewService creates a tenant-admin application service.
func NewService(repo repository) (*Service, error) {
	if repo == nil {
		return nil, errRepositoryRequired
	}

	return &Service{repository: repo}, nil
}
