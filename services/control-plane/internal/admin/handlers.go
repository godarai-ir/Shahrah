package admin

import (
	"context"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/storage"
	"github.com/gofiber/fiber/v2"
)

type keyStore interface {
	CreateKey(ctx context.Context, params platform.CreateKeyParams) (*platform.CreatedKey, error)
	ListKeys(ctx context.Context, tenantID string) ([]platform.CreatedKey, error)
	ListAllowedModelFQNs(ctx context.Context, tenantID string, userID string, teamID string) ([]string, error)
	ListPATs(ctx context.Context, tenantID string, userID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error)
	PATExists(ctx context.Context, tenantID string, userID string, name string) (bool, error)
	RevokePAT(ctx context.Context, tenantID string, userID string, keyID string, revokedByUserID string) (bool, error)
	RevokeAllPATs(ctx context.Context, tenantID string, userID string, revokedByUserID string) (int64, error)
	ListVATTokens(ctx context.Context, tenantID string, virtualUserID string, query string, limit int, offset int) ([]platform.CreatedKey, int, error)
	UpdateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, req platform.UpdateVATTokenRequest) (*platform.CreatedKey, error)
	RevokeVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, revokedByUserID string, reason string) (bool, error)
	RotateVATToken(ctx context.Context, tenantID string, virtualUserID string, keyID string, rotatedByUserID string, gracePeriod string) (*platform.CreatedKey, error)
}

type modelStore interface {
	CreateDeployment(ctx context.Context, dep platform.Deployment) (*platform.Deployment, error)
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
	CreateVirtualModel(ctx context.Context, tenantID string, req platform.CreateVirtualModelRequest, actor string) (*platform.VirtualModel, error)
	ListVirtualModels(ctx context.Context, tenantID string) ([]platform.VirtualModel, error)
	GetVirtualModel(ctx context.Context, tenantID string, id string) (*platform.VirtualModel, error)
	UpdateVirtualModel(ctx context.Context, tenantID string, id string, req platform.UpdateVirtualModelRequest, actor string) (*platform.VirtualModel, error)
	DeleteVirtualModel(ctx context.Context, tenantID string, id string) error
	ListVirtualModelsForPrincipal(ctx context.Context, tenantID string, userID string, teamID string) ([]platform.VirtualModel, error)
	CreateRateLimitRule(ctx context.Context, tenantID string, req platform.CreateRateLimitRuleRequest) (*platform.RateLimitRule, error)
	ListRateLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.RateLimitRule, int, error)
	GetRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) (*platform.RateLimitRule, error)
	UpdateRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string, req platform.UpdateRateLimitRuleRequest) (*platform.RateLimitRule, error)
	UpdateRateLimitRuleStatusByRuleID(ctx context.Context, tenantID string, ruleID string, enabled bool) (*platform.RateLimitRule, error)
	DeleteRateLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) error
	CreateBudgetLimitRule(ctx context.Context, tenantID string, req platform.CreateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error)
	ListBudgetLimitRules(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.BudgetLimitRule, int, error)
	GetBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) (*platform.BudgetLimitRule, error)
	UpdateBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string, req platform.UpdateBudgetLimitRuleRequest) (*platform.BudgetLimitRule, error)
	UpdateBudgetLimitRuleStatusByRuleID(ctx context.Context, tenantID string, ruleID string, enabled bool) (*platform.BudgetLimitRule, error)
	DeleteBudgetLimitRuleByRuleID(ctx context.Context, tenantID string, ruleID string) error
}

type walletStore interface {
	GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error)
	TopUpWallet(ctx context.Context, tenantID string, amountIRR float64, metadata map[string]interface{}) (*platform.Transaction, error)
	GetLatestExchangeRate(ctx context.Context) (*platform.ExchangeRate, error)
	GetWalletTopUpSettings(ctx context.Context) (*platform.WalletTopUpSettings, error)
	UpdateWalletTopUpSettings(ctx context.Context, minUSD, maxUSD, taxPercent, feePercent float64) (*platform.WalletTopUpSettings, error)
	CreateWalletTopUpIntent(ctx context.Context, p storage.CreateTopUpIntentParams) (*platform.WalletTopUpIntent, error)
	AttachWalletTopUpGatewayData(ctx context.Context, intentPublicID, authority, paymentURL string, raw map[string]interface{}) error
	GetWalletTopUpIntentByAuthority(ctx context.Context, authority string) (*platform.WalletTopUpIntent, error)
	CompleteWalletTopUpByAuthority(ctx context.Context, authority string, verifyRaw map[string]interface{}, refID, gatewayTxnID, cardMask, payerPhone string) (*platform.WalletTopUpIntent, bool, error)
	MarkWalletTopUpFailedByAuthority(ctx context.Context, authority string, reason string, raw map[string]interface{}) error
	ListWalletTransactions(ctx context.Context, tenantID string, limit int) ([]platform.Transaction, error)
}

type providerAccountStore interface {
	CreateTenantProviderAccount(ctx context.Context, tenantID string, req platform.CreateTenantProviderAccountRequest) (*platform.TenantProviderAccount, error)
	ListTenantProviderAccounts(ctx context.Context, tenantID string) ([]platform.TenantProviderAccount, error)
	GetTenantProviderAccountDetails(ctx context.Context, tenantID string, accountID string) (*platform.TenantProviderAccountDetailsResponse, error)
	UpdateTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.UpdateTenantProviderAccountCredentialsRequest) (*platform.TenantProviderAccount, error)
	UpsertTenantProviderAccountModels(ctx context.Context, tenantID string, accountID string, models []platform.TenantProviderAccountModel) ([]platform.TenantProviderAccountModel, error)
	UpdateTenantProviderAccountAccess(ctx context.Context, tenantID string, accountID string, access platform.TenantProviderAccountAccessControl) (*platform.TenantProviderAccountAccessControl, error)
	TestTenantProviderAccountCredentials(ctx context.Context, tenantID string, accountID string, req platform.TestTenantProviderAccountCredentialsRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error)
	ListTenantProviderAccountEffectiveModels(ctx context.Context, tenantID string, accountID string, keyType string, userID string, teamID string) ([]platform.EffectiveModelView, error)
	IsTenantProviderAccountManager(ctx context.Context, tenantID string, accountID string, userID string, teamID string) (bool, error)
	DeleteTenantProviderAccount(ctx context.Context, tenantID string, accountID string) error
	SearchAccessPrincipals(ctx context.Context, tenantID string, kind string, query string, limit int) ([]platform.AccessPrincipalCandidate, error)
}

type orgUnitStore interface {
	CreateOrgUnit(ctx context.Context, tenantID string, req platform.CreateOrgUnitRequest) (*platform.OrgUnit, error)
	ListOrgUnits(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.OrgUnit, int, error)
	GetOrgUnit(ctx context.Context, tenantID string, orgUnitID string) (*platform.OrgUnit, error)
	UpdateOrgUnit(ctx context.Context, tenantID string, orgUnitID string, req platform.UpdateOrgUnitRequest) (*platform.OrgUnit, error)
	DeleteOrgUnit(ctx context.Context, tenantID string, orgUnitID string) error
	ListOrgUnitMemberships(ctx context.Context, tenantID string, orgUnitID string) ([]platform.OrgUnitMembership, error)
	AddOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, req platform.CreateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error)
	UpdateOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, userID string, req platform.UpdateOrgUnitMembershipRequest) (*platform.OrgUnitMembership, error)
	DeleteOrgUnitMembership(ctx context.Context, tenantID string, orgUnitID string, userID string) error
}

type virtualUserStore interface {
	CreateVirtualUser(ctx context.Context, tenantID string, req platform.CreateVirtualUserRequest) (*platform.VirtualUser, error)
	ListVirtualUsers(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.VirtualUser, int, error)
	GetVirtualUser(ctx context.Context, tenantID string, virtualUserID string) (*platform.VirtualUser, error)
	UpdateVirtualUser(ctx context.Context, tenantID string, virtualUserID string, req platform.UpdateVirtualUserRequest) (*platform.VirtualUser, error)
	DeleteVirtualUser(ctx context.Context, tenantID string, virtualUserID string) error
}

type membershipStore interface {
	ListTeamMemberships(ctx context.Context, tenantID string, teamID string) ([]platform.TeamMembership, error)
	AddTeamMembership(ctx context.Context, tenantID string, teamID string, req platform.CreateTeamMembershipRequest) (*platform.TeamMembership, error)
	UpdateTeamMembership(ctx context.Context, tenantID string, teamID string, userID string, req platform.UpdateTeamMembershipRequest) (*platform.TeamMembership, error)
	DeleteTeamMembership(ctx context.Context, tenantID string, teamID string, userID string) error
}

type teamStore interface {
	CreateTeam(ctx context.Context, tenantID string, req platform.CreateTeamRequest) (*platform.Team, error)
	ListTeams(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.Team, int, error)
	GetTeam(ctx context.Context, tenantID string, teamID string) (*platform.Team, error)
	UpdateTeam(ctx context.Context, tenantID string, teamID string, req platform.UpdateTeamRequest) (*platform.Team, error)
	DeleteTeam(ctx context.Context, tenantID string, teamID string) error
	GetTeamPermissions(ctx context.Context, tenantID string, teamID string) (*platform.TeamPermissionsResponse, error)
	UpdateTeamPermissions(ctx context.Context, tenantID string, teamID string, req platform.UpdateTeamPermissionsRequest) (*platform.TeamPermissionsResponse, error)
}

type userStore interface {
	ListTenantAdminUsers(ctx context.Context, tenantSlug string, query string, limit int, offset int) ([]platform.TenantAdminUser, int, error)
	GetTenantAdminUserRecord(ctx context.Context, tenantSlug string, userID string) (*platform.TenantAdminUser, error)
	InviteTenantAdminUser(ctx context.Context, tenantID string, tenantSlug string, req platform.InviteTenantAdminUserRequest) (*platform.TenantAdminInvitation, string, error)
	AcceptTenantAdminInvitation(ctx context.Context, tenantID string, tenantSlug string, token string) (*platform.TenantAdminInvitation, *platform.TenantAdminUser, error)
	SetTenantAdminUserActive(ctx context.Context, tenantSlug string, userID string, active bool) (*platform.TenantAdminUser, error)
}

type roleStore interface {
	ListRoles(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.RoleDefinition, int, error)
	CreateRole(ctx context.Context, tenantID string, req platform.CreateRoleRequest) (*platform.RoleDefinition, error)
	UpdateRole(ctx context.Context, tenantID string, roleID string, req platform.UpdateRoleRequest) (*platform.RoleDefinition, error)
	UpdateUserRoles(ctx context.Context, tenantID string, req platform.UpdateUserRolesRequest) (*platform.UpdateUserRolesResponse, error)
	GetEffectivePermissionsForUser(ctx context.Context, tenantID string, userID string, fallbackRole string) ([]string, []string, error)
}

type traceStore interface {
	ListRequestTraces(ctx context.Context, params platform.ListRequestTracesParams) ([]platform.RequestTrace, int, error)
	GetRequestTrace(ctx context.Context, tenantID string, traceID string) (*platform.RequestTrace, error)
	ListTraceSpans(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error)
	ListRequestTracesV2(ctx context.Context, params platform.ListRequestTracesV2Params) ([]platform.RequestTraceV2, int, error)
	GetRequestTraceV2(ctx context.Context, tenantID string, traceID string) (*platform.RequestTraceDetailV2, error)
	ListTraceTimelineV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceSpan, error)
	ListTraceAttemptsV2(ctx context.Context, tenantID string, traceID string) ([]platform.TraceAttemptV2, error)
	GetTraceFacetsV2(ctx context.Context, params platform.ListRequestTracesV2Params) (*platform.TraceFacetsV2, error)
}

type metricsStore interface {
	QueryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error)
	ListModelMetricsFacets(ctx context.Context, tenantID string) (*platform.ModelMetricsFacets, error)
}

type tracePrivacyStore interface {
	GetTenantTracePrivacySettings(ctx context.Context, tenantID string) (*platform.TenantTracePrivacySettings, error)
	UpsertTenantTracePrivacySettings(ctx context.Context, tenantID, capturePolicy, updatedByUserID string) (*platform.TenantTracePrivacySettings, error)
}

type onboardingStore interface {
	CheckTenantSlugAvailability(ctx context.Context, slug string) (*storage.TenantSlugAvailability, error)
	CreateOnboardingOTP(ctx context.Context, req storage.CreateOnboardingOTPParams) (*storage.OnboardingOTP, error)
	VerifyOnboardingOTPAndProvision(ctx context.Context, req storage.VerifyOnboardingOTPParams) (*storage.ProvisionedTenant, error)
}

type loginDiscoveryStore interface {
	GetTenantBySlug(ctx context.Context, slug string) (*platform.Tenant, error)
	ListTenantsByAdminIdentifier(ctx context.Context, channel string, identifier string) ([]platform.Tenant, error)
}

// Handlers serves bootstrap management APIs.
type Handlers struct {
	keys             keyStore
	models           modelStore
	wallets          walletStore
	providerAccounts providerAccountStore
	orgUnits         orgUnitStore
	memberships      membershipStore
	users            userStore
	virtualUsers     virtualUserStore
	teams            teamStore
	onboarding       onboardingStore
	loginDiscovery   loginDiscoveryStore
	traces           traceStore
	metrics          metricsStore
	tracePrivacy     tracePrivacyStore
	roles            roleStore
}

// New creates a management handlers bundle.
func New(keys keyStore, models modelStore, wallets walletStore) *Handlers {
	accounts, _ := models.(providerAccountStore)
	orgUnits, _ := keys.(orgUnitStore)
	memberships, _ := keys.(membershipStore)
	users, _ := models.(userStore)
	virtualUsers, _ := keys.(virtualUserStore)
	teams, _ := keys.(teamStore)
	onboarding, _ := keys.(onboardingStore)
	loginDiscovery, _ := keys.(loginDiscoveryStore)
	traces, _ := keys.(traceStore)
	metrics, _ := keys.(metricsStore)
	tracePrivacy, _ := keys.(tracePrivacyStore)
	roles, _ := keys.(roleStore)
	if tracePrivacy == nil {
		tracePrivacy, _ = wallets.(tracePrivacyStore)
	}
	return &Handlers{
		keys:             keys,
		models:           models,
		wallets:          wallets,
		providerAccounts: accounts,
		orgUnits:         orgUnits,
		memberships:      memberships,
		users:            users,
		virtualUsers:     virtualUsers,
		teams:            teams,
		onboarding:       onboarding,
		loginDiscovery:   loginDiscovery,
		traces:           traces,
		metrics:          metrics,
		tracePrivacy:     tracePrivacy,
		roles:            roles,
	}
}

// ListRoles lists tenant-scoped role definitions.
func isTenantAdminUser(user *platform.TenantAdminUser) bool {
	if user == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(user.Role), "tenant_admin") {
		return true
	}
	for _, role := range user.Roles {
		if strings.EqualFold(strings.TrimSpace(role), "tenant_admin") {
			return true
		}
	}
	return false
}

func (h *Handlers) requireProviderAccountsStore() (providerAccountStore, error) {
	if h.providerAccounts == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "provider account store not configured")
	}
	return h.providerAccounts, nil
}

func (h *Handlers) requireOnboardingStore() (onboardingStore, error) {
	if h.onboarding == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "onboarding store not configured")
	}
	return h.onboarding, nil
}

func (h *Handlers) requireLoginDiscoveryStore() (loginDiscoveryStore, error) {
	if h.loginDiscovery == nil {
		return nil, fiber.NewError(fiber.StatusNotImplemented, "login discovery store not configured")
	}
	return h.loginDiscovery, nil
}

func appendUnique(values []string, candidate string) []string {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return values
	}
	for _, existing := range values {
		if existing == candidate {
			return values
		}
	}
	return append(values, candidate)
}
