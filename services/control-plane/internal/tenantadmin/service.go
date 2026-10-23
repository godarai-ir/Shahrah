package tenantadmin

import (
	"context"
	"errors"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var errRepositoryRequired = errors.New("tenantadmin: repository is required")

type captchaRepository interface {
	CreateTenantAdminCaptcha(context.Context) (string, string, time.Time, error)
	ConsumeTenantAdminCaptcha(context.Context, string, string) (bool, error)
}

type otpRepository interface {
	CreateTenantAdminOTPRequest(context.Context, TenantAdminOTPRequestParams) (*TenantAdminOTPResponse, error)
	VerifyTenantAdminOTPRequest(context.Context, VerifyTenantAdminOTPRequestParams) (*TenantAdminOTPVerification, error)
}

type sessionRepository interface {
	CreateTenantAdminSession(context.Context, CreateTenantAdminSessionParams) (*TenantAdminSessionResponse, error)
	GetTenantAdminSession(context.Context, string) (*TenantAdminSessionResponse, error)
	RevokeTenantAdminSession(context.Context, string) error
}

type authenticationUserRepository interface {
	FindOrCreateTenantAdminUser(context.Context, FindOrCreateTenantAdminUserParams) (*TenantAdminUserResponse, error)
	GetTenantAdminUser(context.Context, string, string) (*TenantAdminUserResponse, error)
	UpdateTenantAdminUserContact(context.Context, UpdateTenantAdminUserContactParams) (*TenantAdminUserResponse, error)
	TouchTenantAdminLastLogin(context.Context, string, string) (*TenantAdminUserResponse, error)
}

type socialRepository interface {
	ProvisionSocialTenant(context.Context, ProvisionSocialTenantParams) (*ProvisionedTenant, error)
	CreateSocialLoginHandoff(context.Context, string, string) (*SocialHandoff, error)
	RedeemSocialLoginHandoff(context.Context, string, string) (*RedeemedSocialHandoff, error)
	CreateSocialLinkTicket(context.Context, string, string) (*SocialHandoff, error)
	CompleteSocialLink(context.Context, string, string) (*SocialHandoff, error)
}

type onboardingRepository interface {
	CheckTenantSlugAvailability(context.Context, string) (*TenantSlugAvailability, error)
	CreateOnboardingOTP(context.Context, CreateOnboardingOTPParams) (*OnboardingOTP, error)
	VerifyOnboardingOTPAndProvision(context.Context, VerifyOnboardingOTPParams) (*ProvisionedTenant, error)
}

type membershipRepository interface {
	ListTenantAdminUsers(context.Context, string, string, int, int) ([]platform.TenantAdminUser, int, error)
	GetTenantAdminUserRecord(context.Context, string, string) (*platform.TenantAdminUser, error)
	InviteTenantAdminUser(context.Context, string, string, platform.InviteTenantAdminUserRequest) (*platform.TenantAdminInvitation, string, error)
	AcceptTenantAdminInvitation(context.Context, string, string, string) (*platform.TenantAdminInvitation, *platform.TenantAdminUser, error)
	SetTenantAdminUserActive(context.Context, string, string, bool) (*platform.TenantAdminUser, error)
	UpdateTenantAdminUserProfile(context.Context, string, string, platform.UpdateTenantAdminUserProfileRequest) (*platform.TenantAdminUser, error)
	GetTenantSeatUsage(context.Context, string) (*platform.TenantSeatUsage, error)
}

type repository interface {
	captchaRepository
	otpRepository
	sessionRepository
	authenticationUserRepository
	socialRepository
	onboardingRepository
	membershipRepository
}

// Service is the tenant-admin authentication and onboarding application boundary.
type Service struct {
	captchas    captchaRepository
	otp         otpRepository
	sessions    sessionRepository
	authUsers   authenticationUserRepository
	social      socialRepository
	onboarding  onboardingRepository
	memberships membershipRepository
}

// NewService creates a tenant-admin application service.
func NewService(repo repository) (*Service, error) {
	if repo == nil {
		return nil, errRepositoryRequired
	}

	return &Service{
		captchas:    repo,
		otp:         repo,
		sessions:    repo,
		authUsers:   repo,
		social:      repo,
		onboarding:  repo,
		memberships: repo,
	}, nil
}

// CreateTenantAdminCaptcha creates a short-lived authentication captcha.
func (s *Service) CreateTenantAdminCaptcha(ctx context.Context) (string, string, time.Time, error) {
	return s.captchas.CreateTenantAdminCaptcha(ctx)
}

// ConsumeTenantAdminCaptcha verifies and consumes a captcha.
func (s *Service) ConsumeTenantAdminCaptcha(ctx context.Context, captchaID, answer string) (bool, error) {
	return s.captchas.ConsumeTenantAdminCaptcha(ctx, captchaID, answer)
}

// CreateTenantAdminOTPRequest creates an OTP request.
func (s *Service) CreateTenantAdminOTPRequest(ctx context.Context, params TenantAdminOTPRequestParams) (*TenantAdminOTPResponse, error) {
	return s.otp.CreateTenantAdminOTPRequest(ctx, params)
}

// VerifyTenantAdminOTPRequest verifies an OTP request.
func (s *Service) VerifyTenantAdminOTPRequest(ctx context.Context, params VerifyTenantAdminOTPRequestParams) (*TenantAdminOTPVerification, error) {
	return s.otp.VerifyTenantAdminOTPRequest(ctx, params)
}

// CreateTenantAdminSession creates an authenticated tenant-admin session.
func (s *Service) CreateTenantAdminSession(ctx context.Context, params CreateTenantAdminSessionParams) (*TenantAdminSessionResponse, error) {
	return s.sessions.CreateTenantAdminSession(ctx, params)
}

// GetTenantAdminSession resolves an authenticated session.
func (s *Service) GetTenantAdminSession(ctx context.Context, sessionID string) (*TenantAdminSessionResponse, error) {
	return s.sessions.GetTenantAdminSession(ctx, sessionID)
}

// RevokeTenantAdminSession revokes an authenticated session.
func (s *Service) RevokeTenantAdminSession(ctx context.Context, sessionID string) error {
	return s.sessions.RevokeTenantAdminSession(ctx, sessionID)
}

// FindOrCreateTenantAdminUser resolves a tenant member for authentication.
func (s *Service) FindOrCreateTenantAdminUser(ctx context.Context, params FindOrCreateTenantAdminUserParams) (*TenantAdminUserResponse, error) {
	return s.authUsers.FindOrCreateTenantAdminUser(ctx, params)
}

// GetTenantAdminUser resolves a tenant member.
func (s *Service) GetTenantAdminUser(ctx context.Context, tenantSlug, userID string) (*TenantAdminUserResponse, error) {
	return s.authUsers.GetTenantAdminUser(ctx, tenantSlug, userID)
}

// UpdateTenantAdminUserContact updates a member contact channel.
func (s *Service) UpdateTenantAdminUserContact(ctx context.Context, params UpdateTenantAdminUserContactParams) (*TenantAdminUserResponse, error) {
	return s.authUsers.UpdateTenantAdminUserContact(ctx, params)
}

// TouchTenantAdminLastLogin records a successful login.
func (s *Service) TouchTenantAdminLastLogin(ctx context.Context, tenantSlug, userID string) (*TenantAdminUserResponse, error) {
	return s.authUsers.TouchTenantAdminLastLogin(ctx, tenantSlug, userID)
}

// ProvisionSocialTenant provisions a tenant from a trusted social identity.
func (s *Service) ProvisionSocialTenant(ctx context.Context, params ProvisionSocialTenantParams) (*ProvisionedTenant, error) {
	return s.social.ProvisionSocialTenant(ctx, params)
}

// CreateSocialLoginHandoff creates a social login handoff.
func (s *Service) CreateSocialLoginHandoff(ctx context.Context, tenantSlug, userID string) (*SocialHandoff, error) {
	return s.social.CreateSocialLoginHandoff(ctx, tenantSlug, userID)
}

// RedeemSocialLoginHandoff redeems a social login handoff.
func (s *Service) RedeemSocialLoginHandoff(ctx context.Context, tenantSlug, code string) (*RedeemedSocialHandoff, error) {
	return s.social.RedeemSocialLoginHandoff(ctx, tenantSlug, code)
}

// CreateSocialLinkTicket creates a social-account link ticket.
func (s *Service) CreateSocialLinkTicket(ctx context.Context, tenantSlug, userID string) (*SocialHandoff, error) {
	return s.social.CreateSocialLinkTicket(ctx, tenantSlug, userID)
}

// CompleteSocialLink completes a social-account link.
func (s *Service) CompleteSocialLink(ctx context.Context, tenantSlug, code string) (*SocialHandoff, error) {
	return s.social.CompleteSocialLink(ctx, tenantSlug, code)
}

// CheckTenantSlugAvailability checks a normalized tenant slug.
func (s *Service) CheckTenantSlugAvailability(ctx context.Context, slug string) (*TenantSlugAvailability, error) {
	return s.onboarding.CheckTenantSlugAvailability(ctx, slug)
}

// CreateOnboardingOTP starts tenant onboarding.
func (s *Service) CreateOnboardingOTP(ctx context.Context, params CreateOnboardingOTPParams) (*OnboardingOTP, error) {
	return s.onboarding.CreateOnboardingOTP(ctx, params)
}

// VerifyOnboardingOTPAndProvision verifies onboarding and provisions a tenant.
func (s *Service) VerifyOnboardingOTPAndProvision(ctx context.Context, params VerifyOnboardingOTPParams) (*ProvisionedTenant, error) {
	return s.onboarding.VerifyOnboardingOTPAndProvision(ctx, params)
}

// ListTenantAdminUsers lists tenant members.
func (s *Service) ListTenantAdminUsers(ctx context.Context, tenantSlug, query string, limit, offset int) ([]platform.TenantAdminUser, int, error) {
	return s.memberships.ListTenantAdminUsers(ctx, tenantSlug, query, limit, offset)
}

// GetTenantAdminUserRecord gets a tenant member record.
func (s *Service) GetTenantAdminUserRecord(ctx context.Context, tenantSlug, userID string) (*platform.TenantAdminUser, error) {
	return s.memberships.GetTenantAdminUserRecord(ctx, tenantSlug, userID)
}

// InviteTenantAdminUser invites a tenant member.
func (s *Service) InviteTenantAdminUser(ctx context.Context, tenantID, tenantSlug string, request platform.InviteTenantAdminUserRequest) (*platform.TenantAdminInvitation, string, error) {
	return s.memberships.InviteTenantAdminUser(ctx, tenantID, tenantSlug, request)
}

// AcceptTenantAdminInvitation accepts a tenant-member invitation.
func (s *Service) AcceptTenantAdminInvitation(ctx context.Context, tenantID, tenantSlug, token string) (*platform.TenantAdminInvitation, *platform.TenantAdminUser, error) {
	return s.memberships.AcceptTenantAdminInvitation(ctx, tenantID, tenantSlug, token)
}

// UpdateTenantAdminUserProfile changes a tenant member's first and last name.
func (s *Service) UpdateTenantAdminUserProfile(ctx context.Context, tenantSlug, userID string, request platform.UpdateTenantAdminUserProfileRequest) (*platform.TenantAdminUser, error) {
	return s.memberships.UpdateTenantAdminUserProfile(ctx, tenantSlug, userID, request)
}

// GetTenantSeatUsage reports the tenant's user seats against its plan cap.
func (s *Service) GetTenantSeatUsage(ctx context.Context, tenantID string) (*platform.TenantSeatUsage, error) {
	return s.memberships.GetTenantSeatUsage(ctx, tenantID)
}

// SetTenantAdminUserActive changes a tenant member status.
func (s *Service) SetTenantAdminUserActive(ctx context.Context, tenantSlug, userID string, active bool) (*platform.TenantAdminUser, error) {
	return s.memberships.SetTenantAdminUserActive(ctx, tenantSlug, userID, active)
}
