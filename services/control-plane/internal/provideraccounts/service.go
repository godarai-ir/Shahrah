package provideraccounts

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var errRepositoryRequired = errors.New("provider accounts: repository is required")

type accountRepository interface {
	CreateTenantProviderAccount(context.Context, string, platform.CreateTenantProviderAccountRequest) (*platform.TenantProviderAccount, error)
	ListTenantProviderAccounts(context.Context, string) ([]platform.TenantProviderAccount, error)
	ListTenantProviderAccountCatalogRecords(context.Context, string) ([]platform.TenantProviderAccountCatalogRecord, error)
	GetTenantProviderAccountDetails(context.Context, string, string) (*platform.TenantProviderAccountDetailsResponse, error)
	UpdateTenantProviderAccountCredentials(context.Context, string, string, platform.UpdateTenantProviderAccountCredentialsRequest) (*platform.TenantProviderAccount, error)
	DeleteTenantProviderAccount(context.Context, string, string) error
}

type configurationRepository interface {
	UpsertTenantProviderAccountModels(context.Context, string, string, []platform.TenantProviderAccountModel) ([]platform.TenantProviderAccountModel, error)
	UpdateTenantProviderAccountAccess(context.Context, string, string, platform.TenantProviderAccountAccessControl) (*platform.TenantProviderAccountAccessControl, error)
}

type connectionRepository interface {
	TestTenantProviderAccountCredentials(context.Context, string, string, platform.TestTenantProviderAccountCredentialsRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error)
	TestProviderConnection(context.Context, platform.TestProviderConnectionRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error)
}

type authorizationRepository interface {
	ListTenantProviderAccountEffectiveModels(context.Context, string, string, string, string, string) ([]platform.EffectiveModelView, error)
	IsTenantProviderAccountManager(context.Context, string, string, string, string) (bool, error)
	SearchAccessPrincipals(context.Context, string, string, string, int) ([]platform.AccessPrincipalCandidate, error)
}

type repository interface {
	accountRepository
	configurationRepository
	connectionRepository
	authorizationRepository
}

type errorTranslator interface {
	TranslateProviderAccountError(error) error
}

// Service owns tenant provider-account management use cases.
type Service struct {
	accounts      accountRepository
	configuration configurationRepository
	connections   connectionRepository
	authorization authorizationRepository
	translate     func(error) error
}

// NewService creates a provider-account application service.
func NewService(repo repository) (*Service, error) {
	if repo == nil {
		return nil, errRepositoryRequired
	}

	translate := func(err error) error { return err }
	if translator, ok := repo.(errorTranslator); ok {
		translate = translator.TranslateProviderAccountError
	}

	return &Service{
		accounts:      repo,
		configuration: repo,
		connections:   repo,
		authorization: repo,
		translate:     translate,
	}, nil
}

// CreateTenantProviderAccount creates a tenant provider account.
func (s *Service) CreateTenantProviderAccount(ctx context.Context, tenantID string, request platform.CreateTenantProviderAccountRequest) (*platform.TenantProviderAccount, error) {
	result, err := s.accounts.CreateTenantProviderAccount(ctx, tenantID, request)

	return result, s.translate(err)
}

// ListTenantProviderAccounts lists tenant provider accounts.
func (s *Service) ListTenantProviderAccounts(ctx context.Context, tenantID string) ([]platform.TenantProviderAccount, error) {
	result, err := s.accounts.ListTenantProviderAccounts(ctx, tenantID)

	return result, s.translate(err)
}

// ListTenantProviderAccountCatalogRecords lists provider-account catalog records.
func (s *Service) ListTenantProviderAccountCatalogRecords(ctx context.Context, tenantID string) ([]platform.TenantProviderAccountCatalogRecord, error) {
	result, err := s.accounts.ListTenantProviderAccountCatalogRecords(ctx, tenantID)

	return result, s.translate(err)
}

// GetTenantProviderAccountDetails gets provider-account details.
func (s *Service) GetTenantProviderAccountDetails(ctx context.Context, tenantID, accountID string) (*platform.TenantProviderAccountDetailsResponse, error) {
	result, err := s.accounts.GetTenantProviderAccountDetails(ctx, tenantID, accountID)

	return result, s.translate(err)
}

// UpdateTenantProviderAccountCredentials updates provider-account credentials.
func (s *Service) UpdateTenantProviderAccountCredentials(ctx context.Context, tenantID, accountID string, request platform.UpdateTenantProviderAccountCredentialsRequest) (*platform.TenantProviderAccount, error) {
	result, err := s.accounts.UpdateTenantProviderAccountCredentials(ctx, tenantID, accountID, request)

	return result, s.translate(err)
}

// UpsertTenantProviderAccountModels replaces provider-account model settings.
func (s *Service) UpsertTenantProviderAccountModels(ctx context.Context, tenantID, accountID string, models []platform.TenantProviderAccountModel) ([]platform.TenantProviderAccountModel, error) {
	result, err := s.configuration.UpsertTenantProviderAccountModels(ctx, tenantID, accountID, models)

	return result, s.translate(err)
}

// UpdateTenantProviderAccountAccess updates provider-account access controls.
func (s *Service) UpdateTenantProviderAccountAccess(ctx context.Context, tenantID, accountID string, access platform.TenantProviderAccountAccessControl) (*platform.TenantProviderAccountAccessControl, error) {
	result, err := s.configuration.UpdateTenantProviderAccountAccess(ctx, tenantID, accountID, access)

	return result, s.translate(err)
}

// TestTenantProviderAccountCredentials tests stored or replacement credentials.
func (s *Service) TestTenantProviderAccountCredentials(ctx context.Context, tenantID, accountID string, request platform.TestTenantProviderAccountCredentialsRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error) {
	result, err := s.connections.TestTenantProviderAccountCredentials(ctx, tenantID, accountID, request)

	return result, s.translate(err)
}

// TestProviderConnection tests an ad-hoc provider connection.
func (s *Service) TestProviderConnection(ctx context.Context, request platform.TestProviderConnectionRequest) (*platform.TestTenantProviderAccountCredentialsResponse, error) {
	result, err := s.connections.TestProviderConnection(ctx, request)

	return result, s.translate(err)
}

// ListTenantProviderAccountEffectiveModels lists effective provider-account models.
func (s *Service) ListTenantProviderAccountEffectiveModels(ctx context.Context, tenantID, accountID, keyType, userID, teamID string) ([]platform.EffectiveModelView, error) {
	result, err := s.authorization.ListTenantProviderAccountEffectiveModels(ctx, tenantID, accountID, keyType, userID, teamID)

	return result, s.translate(err)
}

// IsTenantProviderAccountManager checks provider-account management access.
func (s *Service) IsTenantProviderAccountManager(ctx context.Context, tenantID, accountID, userID, teamID string) (bool, error) {
	result, err := s.authorization.IsTenantProviderAccountManager(ctx, tenantID, accountID, userID, teamID)

	return result, s.translate(err)
}

// DeleteTenantProviderAccount deletes a provider account.
func (s *Service) DeleteTenantProviderAccount(ctx context.Context, tenantID, accountID string) error {
	return s.translate(s.accounts.DeleteTenantProviderAccount(ctx, tenantID, accountID))
}

// SearchAccessPrincipals searches provider-account access principals.
func (s *Service) SearchAccessPrincipals(ctx context.Context, tenantID, kind, query string, limit int) ([]platform.AccessPrincipalCandidate, error) {
	result, err := s.authorization.SearchAccessPrincipals(ctx, tenantID, kind, query, limit)

	return result, s.translate(err)
}
