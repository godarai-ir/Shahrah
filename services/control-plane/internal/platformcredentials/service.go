package platformcredentials

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/providervalidation"
)

var (
	ErrInvalidRequest     = errors.New("invalid_platform_provider_credential_request")
	ErrRevisionConflict   = errors.New("platform_provider_connection_revision_conflict")
	ErrNotFound           = errors.New("platform_provider_connection_not_found")
	ErrRuntimeUnsupported = errors.New("platform_provider_auth_mode_not_runtime_ready")
)

type Actor struct {
	ID               string
	Role             string
	ServiceAccountID string
	RequestID        string
	SourceIP         string
}

type Repository interface {
	CreatePlatformProviderConnection(context.Context, platform.CreatePlatformProviderConnectionRequest, Actor, string) (*platform.PlatformProviderConnection, error)
	UpdatePlatformProviderConnection(context.Context, string, platform.UpdatePlatformProviderConnectionRequest, Actor) (*platform.PlatformProviderConnection, error)
	ListPlatformProviderConnections(context.Context, string, string) ([]platform.PlatformProviderConnection, error)
	GetPlatformProviderConnection(context.Context, string) (*platform.PlatformProviderConnection, error)
	CreatePlatformProviderCredentialVersion(context.Context, string, map[string]string, *time.Time, Actor, string, func(Binding, map[string]string) (EncryptedCredential, error)) (*platform.PlatformProviderCredentialVersion, error)
	GetPlatformProviderCredential(context.Context, string, int64) (EncryptedCredential, error)
	SetPlatformProviderCredentialTestResult(context.Context, string, int64, bool, string, string, Actor) (*platform.PlatformProviderCredentialTestResult, error)
	ActivatePlatformProviderCredential(context.Context, string, int64, int64, Actor, string, string) (*platform.PlatformProviderConnection, error)
	RollbackPlatformProviderCredential(context.Context, string, int64, int64, Actor, string) (*platform.PlatformProviderConnection, error)
	DisablePlatformProviderConnection(context.Context, string, int64, Actor, string) (*platform.PlatformProviderConnection, error)
	RevokePlatformProviderCredential(context.Context, string, int64, int64, Actor, string) (*platform.PlatformProviderConnection, error)
	ActiveCredentialLoader
}

type CredentialProber interface {
	Probe(context.Context, EncryptedCredential, map[string]string) (bool, string, string)
}

type Service struct {
	repo      Repository
	secrets   *EncryptedDBSecretStore
	manager   *Manager
	publisher InvalidationPublisher
	prober    CredentialProber
	catalog   SchemaCatalog
}

func NewService(repo Repository, secrets *EncryptedDBSecretStore, manager *Manager, publisher InvalidationPublisher, catalog SchemaCatalog, prober CredentialProber) *Service {
	if prober == nil {
		prober = NewHTTPProber(nil)
	}

	return &Service{repo: repo, secrets: secrets, manager: manager, publisher: publisher, catalog: catalog, prober: prober}
}

func (s *Service) Schemas(ctx context.Context, providerSlug string) ([]platform.ProviderConnectionSchema, error) {
	if s.catalog == nil {
		return nil, ErrNotFound
	}

	return s.catalog.Schemas(ctx, normalizeProvider(providerSlug))
}

func (s *Service) CreateConnection(ctx context.Context, request platform.CreatePlatformProviderConnectionRequest, actor Actor, idempotencyKey string) (*platform.PlatformProviderConnection, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, ErrInvalidRequest
	}

	request.ProviderSlug = normalizeProvider(request.ProviderSlug)

	request.Environment = strings.ToLower(strings.TrimSpace(request.Environment))
	if request.Environment == "" {
		request.Environment = platform.ProviderCredentialEnvironmentProd
	}

	request.AuthMode = strings.ToLower(strings.TrimSpace(request.AuthMode))

	schema, err := s.schema(ctx, request.ProviderSlug, request.AuthMode, request.SchemaVersion)
	if err != nil || strings.TrimSpace(request.Name) == "" || hasSecretConfig(request.Config) || !validConnectionConfig(schema, request.Config) {
		return nil, ErrInvalidRequest
	}

	if request.Weight == 0 {
		request.Weight = 1
	}

	return s.repo.CreatePlatformProviderConnection(ctx, request, actor, idempotencyKey)
}

func (s *Service) UpdateConnection(ctx context.Context, id string, request platform.UpdatePlatformProviderConnectionRequest, actor Actor) (*platform.PlatformProviderConnection, error) {
	if request.Revision <= 0 || hasSecretConfig(request.Config) {
		return nil, ErrInvalidRequest
	}

	if request.Config != nil {
		connection, err := s.repo.GetPlatformProviderConnection(ctx, id)
		if err != nil {
			return nil, err
		}

		schema, err := s.schema(ctx, connection.ProviderSlug, connection.AuthMode, connection.SchemaVersion)
		if err != nil || !validConnectionConfig(schema, request.Config) {
			return nil, ErrInvalidRequest
		}
	}

	return s.repo.UpdatePlatformProviderConnection(ctx, id, request, actor)
}

func (s *Service) ListConnections(ctx context.Context, provider, environment string) ([]platform.PlatformProviderConnection, error) {
	return s.repo.ListPlatformProviderConnections(ctx, normalizeProvider(provider), strings.ToLower(strings.TrimSpace(environment)))
}

func (s *Service) GetConnection(ctx context.Context, id string) (*platform.PlatformProviderConnection, error) {
	return s.repo.GetPlatformProviderConnection(ctx, id)
}

func (s *Service) CreateCredentialVersion(ctx context.Context, connectionID string, request platform.CreatePlatformProviderCredentialVersionRequest, actor Actor, idempotencyKey string) (*platform.PlatformProviderCredentialVersion, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, ErrInvalidRequest
	}

	connection, err := s.repo.GetPlatformProviderConnection(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	schema, err := s.schema(ctx, connection.ProviderSlug, connection.AuthMode, connection.SchemaVersion)
	if err != nil || !validCredentialSecrets(schema, request.Secrets) || s.secrets == nil {
		return nil, ErrInvalidRequest
	}

	return s.repo.CreatePlatformProviderCredentialVersion(ctx, connectionID, request.Secrets, request.ExpiresAt, actor, idempotencyKey, s.secrets.Encrypt)
}

func (s *Service) TestCredential(ctx context.Context, connectionID string, version int64, actor Actor) (*platform.PlatformProviderCredentialTestResult, error) {
	record, err := s.repo.GetPlatformProviderCredential(ctx, connectionID, version)
	if err != nil {
		return nil, err
	}

	if schema, schemaErr := s.schema(ctx, record.ProviderSlug, record.AuthMode, record.SchemaVersion); schemaErr != nil || !schema.RuntimeReady {
		return nil, ErrRuntimeUnsupported
	}

	secrets, err := s.secrets.Decrypt(record)
	if err != nil {
		return nil, ErrDecryptFailed
	}
	defer clearSecretMap(secrets)

	success, category, summary := s.prober.Probe(ctx, record, secrets)

	return s.repo.SetPlatformProviderCredentialTestResult(ctx, connectionID, version, success, category, summary, actor)
}

func (s *Service) Activate(ctx context.Context, connectionID string, version, revision int64, actor Actor, idempotencyKey, reason string) (*platform.PlatformProviderConnection, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, ErrInvalidRequest
	}

	connectionBefore, err := s.repo.GetPlatformProviderConnection(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	if schema, schemaErr := s.schema(ctx, connectionBefore.ProviderSlug, connectionBefore.AuthMode, connectionBefore.SchemaVersion); schemaErr != nil || !schema.RuntimeReady {
		return nil, ErrRuntimeUnsupported
	}

	connection, err := s.repo.ActivatePlatformProviderCredential(ctx, connectionID, version, revision, actor, idempotencyKey, sanitizeReason(reason))
	if err != nil {
		return connection, err
	}

	return connection, s.afterLifecycleMutation(ctx, connection, "activate")
}

func (s *Service) Rollback(ctx context.Context, connectionID string, version, revision int64, actor Actor, reason string) (*platform.PlatformProviderConnection, error) {
	connectionBefore, err := s.repo.GetPlatformProviderConnection(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	if schema, schemaErr := s.schema(ctx, connectionBefore.ProviderSlug, connectionBefore.AuthMode, connectionBefore.SchemaVersion); schemaErr != nil || !schema.RuntimeReady {
		return nil, ErrRuntimeUnsupported
	}

	connection, err := s.repo.RollbackPlatformProviderCredential(ctx, connectionID, version, revision, actor, sanitizeReason(reason))
	if err != nil {
		return connection, err
	}

	return connection, s.afterLifecycleMutation(ctx, connection, "rollback")
}

func (s *Service) Disable(ctx context.Context, connectionID string, revision int64, actor Actor, reason string) (*platform.PlatformProviderConnection, error) {
	connection, err := s.repo.DisablePlatformProviderConnection(ctx, connectionID, revision, actor, sanitizeReason(reason))
	if err != nil {
		return connection, err
	}

	return connection, s.afterLifecycleMutation(ctx, connection, "disable")
}

func (s *Service) Revoke(ctx context.Context, connectionID string, version, revision int64, actor Actor, reason string) (*platform.PlatformProviderConnection, error) {
	connection, err := s.repo.RevokePlatformProviderCredential(ctx, connectionID, version, revision, actor, sanitizeReason(reason))
	if err != nil {
		return connection, err
	}

	return connection, s.afterLifecycleMutation(ctx, connection, "revoke")
}

func (s *Service) afterLifecycleMutation(ctx context.Context, connection *platform.PlatformProviderConnection, action string) error {
	var mutationErrors []error

	if s.manager != nil && connection != nil {
		s.manager.Invalidate(connection.ProviderSlug, connection.Environment)

		if err := s.manager.Reload(ctx); err != nil {
			mutationErrors = append(mutationErrors, err)
		}
	}

	if s.publisher != nil && connection != nil {
		if err := s.publisher.PublishPlatformCredentialInvalidation(ctx, InvalidationEvent{ProviderSlug: connection.ProviderSlug, Environment: connection.Environment, ConnectionID: connection.ID, CredentialVersionID: connection.ActiveCredentialVersionID, Action: action}); err != nil {
			mutationErrors = append(mutationErrors, err)
		}
	}

	return errors.Join(mutationErrors...)
}

type HTTPProber struct {
	prober *providervalidation.Prober
}

func NewHTTPProber(client *http.Client) *HTTPProber {
	return &HTTPProber{prober: providervalidation.NewProber(client)}
}

func (p *HTTPProber) Probe(ctx context.Context, record EncryptedCredential, secrets map[string]string) (bool, string, string) {
	config := map[string]interface{}{
		"base_url":     record.BaseURL,
		"api_version":  record.APIVersion,
		"organization": record.Organization,
		"project":      record.Project,
		"http_referer": record.HTTPReferer,
		"app_title":    record.AppTitle,
		"client_name":  record.ClientName,
	}
	result := p.prober.Probe(ctx, providervalidation.Request{
		Provider: record.ProviderSlug,
		AuthMode: record.AuthMode,
		Config:   config,
		Secrets:  secrets,
	})

	return result.Success, result.Category, result.Message
}

func hasSecretConfig(config map[string]interface{}) bool {
	for key, value := range config {
		normalized := strings.ToLower(strings.NewReplacer("_", "", "-", "", ".", "").Replace(key))
		switch normalized {
		case "apikey", "clientsecret", "secretaccesskey", "sessiontoken", "serviceaccountjson", "credential", "credentials", "token", "password":
			return true
		}

		if nested, ok := value.(map[string]interface{}); ok && hasSecretConfig(nested) {
			return true
		}
	}

	return false
}

func normalizeProvider(value string) string {
	return providervalidation.NormalizeProvider(value)
}
func sanitizeReason(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 255 {
		value = value[:255]
	}

	return value
}
func clearSecretMap(secrets map[string]string) {
	for key := range secrets {
		secrets[key] = ""
		delete(secrets, key)
	}
}

func (s *Service) schema(ctx context.Context, providerSlug, authMode string, version int) (platform.ProviderConnectionSchema, error) {
	items, err := s.Schemas(ctx, providerSlug)
	if err != nil {
		return platform.ProviderConnectionSchema{}, err
	}

	authMode = strings.ToLower(strings.TrimSpace(authMode))
	for _, item := range items {
		if item.AuthMode == authMode && item.Version == version {
			return item, nil
		}
	}

	return platform.ProviderConnectionSchema{}, ErrNotFound
}

func validConnectionConfig(schema platform.ProviderConnectionSchema, config map[string]interface{}) bool {
	allowed := make(map[string]platform.ProviderConnectionField)

	for _, field := range schema.Fields {
		if field.Scope == "connection" {
			allowed[field.Key] = field
		}
	}

	for key, value := range config {
		field, ok := allowed[key]
		if !ok || field.Secret {
			return false
		}

		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return false
		}
	}

	for key, field := range allowed {
		if !field.Required {
			continue
		}

		value, ok := config[key].(string)
		if !ok || strings.TrimSpace(value) == "" {
			return false
		}
	}

	return true
}

func validCredentialSecrets(schema platform.ProviderConnectionSchema, secrets map[string]string) bool {
	allowed := make(map[string]platform.ProviderConnectionField)

	for _, field := range schema.Fields {
		if field.Scope == "credential" && field.Secret {
			allowed[field.Key] = field
		}
	}

	for key, value := range secrets {
		if _, ok := allowed[key]; !ok || strings.TrimSpace(value) == "" {
			return false
		}
	}

	for key, field := range allowed {
		if field.Required && strings.TrimSpace(secrets[key]) == "" {
			return false
		}
	}

	return len(secrets) > 0
}
