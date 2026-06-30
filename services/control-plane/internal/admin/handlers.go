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
	CreateTenantGuardrail(ctx context.Context, tenantID string, req platform.CreateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error)
	ListTenantGuardrails(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.TenantGuardrail, int, error)
	GetTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string) (*platform.TenantGuardrail, error)
	UpdateTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string, req platform.UpdateTenantGuardrailRequest, actor string) (*platform.TenantGuardrail, error)
	DeleteTenantGuardrailBySlug(ctx context.Context, tenantID string, slug string) error
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

type logsStore interface {
	ListRequestLogsV2(ctx context.Context, params platform.ListRequestLogsV2Params) ([]platform.RequestLogV2, int, error)
	GetRequestLogV2(ctx context.Context, tenantID string, logID string, includeContent bool) (*platform.RequestLogDetailV2, error)
	GetRequestLogFacetsV2(ctx context.Context, params platform.ListRequestLogsV2Params) (*platform.LogFacetsV2, error)
}

type metricsStore interface {
	QueryModelMetrics(ctx context.Context, tenantID string, req platform.MetricsQueryRequest) ([]map[string]interface{}, error)
	ListModelMetricsFacets(ctx context.Context, tenantID string) (*platform.ModelMetricsFacets, error)
}

type tracePrivacyStore interface {
	GetTenantTracePrivacySettings(ctx context.Context, tenantID string) (*platform.TenantTracePrivacySettings, error)
	UpsertTenantTracePrivacySettings(ctx context.Context, tenantID, capturePolicy, updatedByUserID string) (*platform.TenantTracePrivacySettings, error)
}

type requestLoggingSettingsStore interface {
	GetTenantRequestLoggingSettings(ctx context.Context, tenantID string) (*platform.TenantRequestLoggingSettings, error)
	UpsertTenantRequestLoggingSettings(ctx context.Context, tenantID, loggingMode, updatedByUserID string) (*platform.TenantRequestLoggingSettings, error)
}

type promptStore interface {
	CreatePrompt(ctx context.Context, tenantID string, tenantSlug string, req platform.CreatePromptRequest, actor string) (*platform.PromptVersion, error)
	CreatePromptVersion(ctx context.Context, tenantID string, tenantSlug string, promptID string, req platform.CreatePromptVersionRequest, actor string) (*platform.PromptVersion, error)
	ListPrompts(ctx context.Context, tenantID string, query string, limit int, offset int) ([]platform.Prompt, int, error)
	GetPrompt(ctx context.Context, tenantID string, promptID string) (*platform.Prompt, error)
	DeletePrompt(ctx context.Context, tenantID string, promptID string) error
	ListPromptVersions(ctx context.Context, tenantID string, promptID string, limit int, offset int) ([]platform.PromptVersion, int, error)
	GetPromptVersion(ctx context.Context, tenantID string, versionID string) (*platform.PromptVersion, error)
	GetPromptVersionByFQN(ctx context.Context, tenantID string, fqn string) (*platform.PromptVersion, error)
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

type planStore interface {
	GetCommercialSignupSettings(ctx context.Context) (*platform.CommercialSignupSettings, error)
	UpdateCommercialSignupSettings(ctx context.Context, credit int64, actor string) (*platform.CommercialSignupSettings, error)
	ListFeatureCatalog(ctx context.Context) ([]platform.FeatureDefinition, error)
	ListPublicCommercialPlans(ctx context.Context) ([]platform.CommercialPlan, error)
	ListPlatformCommercialPlans(ctx context.Context) ([]platform.CommercialPlan, error)
	CreateCommercialPlan(ctx context.Context, req platform.CreateCommercialPlanRequest) (*platform.CommercialPlan, error)
	UpdateCommercialPlan(ctx context.Context, planSlug string, req platform.UpdateCommercialPlanRequest) (*platform.CommercialPlan, error)
	DeactivateCommercialPlan(ctx context.Context, planSlug string) error
	CreatePlanVersion(ctx context.Context, planSlug string, req platform.CreatePlanVersionRequest) (*platform.PlanVersion, error)
	PublishPlanVersion(ctx context.Context, planSlug string, version int) (*platform.PlanVersion, error)
	UpsertPlanEntitlement(ctx context.Context, planSlug string, version int, req platform.UpsertPlanEntitlementRequest) (*platform.PlanEntitlement, error)
	UpsertPlanPrice(ctx context.Context, planSlug string, version int, req platform.UpsertPlanPriceRequest) (*platform.PlanPrice, error)
	GetTenantSubscription(ctx context.Context, tenantID string) (*platform.TenantSubscription, error)
	ChangeTenantSubscription(ctx context.Context, tenantID string, req platform.ChangeTenantSubscriptionRequest) (*platform.TenantSubscription, error)
	PrepareTenantSubscriptionChange(ctx context.Context, tenantID string, req platform.ChangeTenantSubscriptionRequest, callbackURL string, returnURL string) (*platform.SubscriptionChangeResult, error)
	AttachSubscriptionPaymentGatewayData(ctx context.Context, publicID string, authority string, paymentURL string, raw map[string]interface{}) error
	GetSubscriptionPaymentIntentByAuthority(ctx context.Context, authority string) (*platform.SubscriptionPaymentIntent, error)
	CompleteSubscriptionPaymentByAuthority(ctx context.Context, authority string, verifyRaw map[string]interface{}, refID string, gatewayTxnID string, cardMask string, payerPhone string) (*platform.SubscriptionPaymentIntent, *platform.TenantSubscription, bool, error)
	MarkSubscriptionPaymentFailedByAuthority(ctx context.Context, authority string, reason string, raw map[string]interface{}) error
	ResolveEffectiveEntitlements(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error)
	UpsertTenantSubscription(ctx context.Context, req platform.UpsertTenantSubscriptionRequest, actor string) (*platform.TenantSubscription, error)
	UpsertTenantEntitlementOverride(ctx context.Context, req platform.UpsertTenantEntitlementOverrideRequest) (*platform.TenantEntitlementOverride, error)
	DeleteTenantEntitlementOverride(ctx context.Context, tenantID string, featureKey string, actor string) error
	CountActiveTenantUsers(ctx context.Context, tenantID string) (int, error)
}

type entitlementResolverStore interface {
	ResolveEffectiveEntitlements(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error)
	CountActiveTenantUsers(ctx context.Context, tenantID string) (int, error)
}

type supportTenantOperationsStore interface {
	ListSupportTenants(ctx context.Context, query, status, plan string, limit, offset int) ([]platform.SupportTenantSummary, error)
	GetSupportTenantSummary(ctx context.Context, tenantID string) (*platform.SupportTenantSummary, error)
	GetTenantOwnerContact(ctx context.Context, tenantID string) (*platform.TenantOwnerContact, error)
	ListTenantEvents(ctx context.Context, tenantID string, limit, offset int) ([]platform.TenantEvent, error)
	AdjustTenantWallet(ctx context.Context, tenantID string, req platform.WalletAdjustmentRequest, actor platform.SupportActor) (*platform.WalletAdjustmentResult, error)
	ChangeTenantStatus(ctx context.Context, tenantID, target, reason string, actor platform.SupportActor) (*platform.SupportTenantSummary, error)
	SetTenantSubscriptionCancellation(ctx context.Context, tenantID string, cancel bool, reason string, actor platform.SupportActor) (*platform.TenantSubscription, error)
	ListTenantSubscriptionPaymentIntents(ctx context.Context, tenantID string, limit, offset int) ([]platform.SubscriptionPaymentIntent, error)
}

type Dependencies struct {
	Keys             keyStore
	Models           modelStore
	Wallets          walletStore
	ProviderAccounts providerAccountStore
	OrgUnits         orgUnitStore
	Memberships      membershipStore
	Users            userStore
	VirtualUsers     virtualUserStore
	Teams            teamStore
	Onboarding       onboardingStore
	LoginDiscovery   loginDiscoveryStore
	Traces           traceStore
	Logs             logsStore
	Metrics          metricsStore
	TracePrivacy     tracePrivacyStore
	RequestLogging   requestLoggingSettingsStore
	Prompts          promptStore
	Roles            roleStore
	Plans            planStore
	Entitlements     entitlementResolverStore
	SupportTenants   supportTenantOperationsStore
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
	logs             logsStore
	metrics          metricsStore
	tracePrivacy     tracePrivacyStore
	requestLogging   requestLoggingSettingsStore
	prompts          promptStore
	roles            roleStore
	plans            planStore
	entitlements     entitlementResolverStore
	supportTenants   supportTenantOperationsStore
}

// New creates a management handlers bundle.
func New(deps Dependencies) *Handlers {
	return &Handlers{
		keys:             deps.Keys,
		models:           deps.Models,
		wallets:          deps.Wallets,
		providerAccounts: deps.ProviderAccounts,
		orgUnits:         deps.OrgUnits,
		memberships:      deps.Memberships,
		users:            deps.Users,
		virtualUsers:     deps.VirtualUsers,
		teams:            deps.Teams,
		onboarding:       deps.Onboarding,
		loginDiscovery:   deps.LoginDiscovery,
		traces:           deps.Traces,
		logs:             deps.Logs,
		metrics:          deps.Metrics,
		tracePrivacy:     deps.TracePrivacy,
		requestLogging:   deps.RequestLogging,
		prompts:          deps.Prompts,
		roles:            deps.Roles,
		plans:            deps.Plans,
		entitlements:     deps.Entitlements,
		supportTenants:   deps.SupportTenants,
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
