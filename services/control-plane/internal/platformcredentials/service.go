package platformcredentials

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/providerregistry"
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
	if connection.AuthMode == string(platform.ProviderAuthModeAnthropicWIF) &&
		len(request.Secrets) == 0 {
		tokenSource, _ := connection.Config["token_source_id"].(string) //nolint:errcheck // Optional metadata uses the zero value when absent or mismatched.
		request.Secrets = map[string]string{
			"wif_binding": strings.TrimSpace(tokenSource),
		}
	}

	validSecrets := validCredentialSecrets(schema, request.Secrets)
	if connection.AuthMode == string(platform.ProviderAuthModeAnthropicWIF) {
		validSecrets = strings.TrimSpace(request.Secrets["wif_binding"]) != ""
	}

	if err != nil || !validSecrets || s.secrets == nil {
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

	result := s.prober.Probe(ctx, record, secrets)

	stored, err := s.repo.SetPlatformProviderCredentialTestResult(
		ctx,
		connectionID,
		version,
		result.Ready,
		result.Category,
		result.Summary,
		actor,
	)
	if err != nil || stored == nil {
		return stored, err
	}

	stored.Authenticated = result.Authenticated
	stored.Ready = result.Ready
	stored.Details = result.PlatformDetails()

	return stored, nil
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

type HTTPProber struct{ client *http.Client }

func NewHTTPProber(client *http.Client) *CredentialProberRegistry {
	if client == nil {
		client = &http.Client{Timeout: 6 * time.Second}
	}

	generic := &HTTPProber{client: client}
	registry := NewCredentialProberRegistry(generic)
	registry.Register("xai", &xaiHTTPProber{client: client})

	return registry
}

func (p *HTTPProber) Probe(ctx context.Context, record EncryptedCredential, secrets map[string]string) ProbeResult {
	spec, ok := documentedProbe(record)
	if !ok {
		return unsupportedProbeResult()
	}

	if !safeProbeURL(record.ProviderSlug, spec.url) {
		return ProbeResult{Category: "configuration", Summary: "Provider base URL is not allowed."}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, spec.url, http.NoBody)
	if err != nil {
		return ProbeResult{Category: "configuration", Summary: "Provider validation endpoint is invalid."}
	}

	switch spec.auth {
	case "x-api-key":
		req.Header.Set("x-api-key", secrets["api_key"])

		version := record.APIVersion
		if version == "" {
			version = "2023-06-01"
		}

		req.Header.Set("anthropic-version", version)
	case "api-key":
		req.Header.Set("api-key", secrets["api_key"])
	case "x-goog-api-key":
		req.Header.Set("x-goog-api-key", secrets["api_key"])
	case "bearer-access-token":
		req.Header.Set("Authorization", "Bearer "+secrets["access_token"])
	case "bearer-api-key":
		req.Header.Set("Authorization", "Bearer "+secrets["api_key"])
	}

	applyProbeMetadataHeaders(req, record)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return ProbeResult{Category: "network", Summary: "Provider validation could not reach the upstream service."}
	}

	if closeErr := resp.Body.Close(); closeErr != nil {
		return ProbeResult{Category: "network", Summary: "Provider validation response could not be closed."}
	}

	return genericProbeResult(resp.StatusCode)
}

type xaiHTTPProber struct{ client *http.Client }

func (p *xaiHTTPProber) Probe(
	ctx context.Context,
	record EncryptedCredential,
	secrets map[string]string,
) ProbeResult {
	spec, ok := documentedProbe(record)
	if !ok {
		return unsupportedProbeResult()
	}

	if !safeProbeURL(record.ProviderSlug, spec.url) {
		return ProbeResult{Category: "configuration", Summary: "xAI base URL is not allowed."}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, spec.url, http.NoBody)
	if err != nil {
		return ProbeResult{Category: "configuration", Summary: "xAI validation endpoint is invalid."}
	}

	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(secrets["api_key"]))
	req.Header.Set("Accept", "application/json")

	if strings.TrimSpace(secrets["api_key"]) == "" {
		return unsupportedProbeResult()
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return ProbeResult{Category: "network", Summary: "xAI validation could not reach the upstream service."}
	}

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	closeErr := resp.Body.Close()

	if readErr != nil {
		return ProbeResult{Category: "network", Summary: "xAI validation response could not be read."}
	}

	if closeErr != nil {
		return ProbeResult{Category: "network", Summary: "xAI validation response could not be closed."}
	}

	return classifyXAIAPIKeyProbe(resp.StatusCode, body, record.ModelHint)
}

func applyProbeMetadataHeaders(req *http.Request, record EncryptedCredential) {
	switch normalizeProvider(record.ProviderSlug) {
	case "openai":
		if value := strings.TrimSpace(record.Organization); value != "" {
			req.Header.Set("OpenAI-Organization", value)
		}

		if value := strings.TrimSpace(record.Project); value != "" {
			req.Header.Set("OpenAI-Project", value)
		}
	case "openrouter":
		if value := strings.TrimSpace(record.HTTPReferer); value != "" {
			req.Header.Set("HTTP-Referer", value)
		}

		if value := strings.TrimSpace(record.AppTitle); value != "" {
			req.Header.Set("X-Title", value)
		}
	}
}

type probeSpec struct {
	url  string
	auth string
}

func documentedProbe(record EncryptedCredential) (probeSpec, bool) {
	base := strings.TrimRight(strings.TrimSpace(record.BaseURL), "/")
	switch normalizeProvider(record.ProviderSlug) {
	case "anthropic":
		if base == "" {
			base = "https://api.anthropic.com"
		}

		return probeSpec{url: base + "/v1/models?limit=1", auth: "x-api-key"}, record.AuthMode == "api_key"
	case "azure-open-ai":
		if record.AuthMode != "api_key" || base == "" {
			return probeSpec{}, false
		}

		if strings.Contains(strings.ToLower(base), "/openai/v1") {
			return modelsProbe(base, "", "api-key"), true
		}

		version := strings.TrimSpace(record.APIVersion)
		if version == "" {
			version = "2024-10-21"
		}

		return probeSpec{url: base + "/openai/models?api-version=" + url.QueryEscape(version), auth: "api-key"}, true
	case "cerebras":
		return modelsProbe(base, "https://api.cerebras.ai/v1", "bearer-api-key"), record.AuthMode == "api_key"
	case "cohere":
		if base == "" {
			base = "https://api.cohere.com"
		}

		return probeSpec{url: base + "/v1/models", auth: "bearer-api-key"}, record.AuthMode == "api_key"
	case "databricks":
		if record.AuthMode != "personal_access_token" || base == "" {
			return probeSpec{}, false
		}

		parsed, err := url.Parse(base)
		if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" {
			return probeSpec{}, false
		}

		parsed.Path = "/api/2.0/serving-endpoints"
		parsed.RawPath, parsed.RawQuery, parsed.Fragment = "", "", ""

		return probeSpec{url: parsed.String(), auth: "bearer-access-token"}, true
	case "deepinfra":
		return modelsProbe(base, "https://api.deepinfra.com/v1/openai", "bearer-api-key"), record.AuthMode == "api_key"
	case "deepseek":
		return modelsProbe(base, "https://api.deepseek.com", "bearer-api-key"), record.AuthMode == "api_key"
	case "google-gemini":
		if base == "" {
			base = "https://generativelanguage.googleapis.com"
		}

		lower := strings.ToLower(base)
		for _, suffix := range []string{"/v1beta/openai", "/v1/openai", "/v1beta", "/v1"} {
			if strings.HasSuffix(lower, suffix) {
				base = base[:len(base)-len(suffix)]
				break
			}
		}

		return probeSpec{url: base + "/v1beta/models?pageSize=1", auth: "x-goog-api-key"}, record.AuthMode == "api_key"
	case "groq":
		return modelsProbe(base, "https://api.groq.com/openai/v1", "bearer-api-key"), record.AuthMode == "api_key"
	case "mistral-ai":
		return modelsProbe(base, "https://api.mistral.ai/v1", "bearer-api-key"), record.AuthMode == "api_key"
	case "moonshot":
		return modelsProbe(base, "https://api.moonshot.ai/v1", "bearer-api-key"), record.AuthMode == "api_key"
	case "openai":
		return modelsProbe(base, "https://api.openai.com/v1", "bearer-api-key"), record.AuthMode == "api_key"
	case "openrouter":
		return modelsProbe(base, "https://openrouter.ai/api/v1", "bearer-api-key"), record.AuthMode == "api_key"
	case "sambanova":
		return modelsProbe(base, "", "bearer-api-key"), record.AuthMode == "api_key" && base != ""
	case "together-ai":
		return modelsProbe(base, "https://api.together.xyz/v1", "bearer-api-key"), record.AuthMode == "api_key"
	case "tokenbazaar":
		return modelsProbe(base, "https://zx1.deepwl.net/v1", "bearer-api-key"), record.AuthMode == "api_key"
	case "xai":
		if base == "" {
			base = "https://api.x.ai/v1"
		}

		return probeSpec{url: strings.TrimRight(base, "/") + "/api-key", auth: "bearer-api-key"}, record.AuthMode == "api_key"
	default:
		return probeSpec{}, false
	}
}

func modelsProbe(base, fallback, auth string) probeSpec {
	if base == "" {
		base = fallback
	}

	return probeSpec{url: strings.TrimRight(base, "/") + "/models", auth: auth}
}

func safeProbeURL(providerSlug, raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	if net.ParseIP(host) != nil || host == "localhost" || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return false
	}

	exact := map[string]string{
		"ai21": "api.ai21.com", "anthropic": "api.anthropic.com", "cerebras": "api.cerebras.ai",
		"cohere": "api.cohere.com", "deepinfra": "api.deepinfra.com", "deepseek": "api.deepseek.com", "google-gemini": "generativelanguage.googleapis.com",
		"groq": "api.groq.com", "mistral-ai": "api.mistral.ai", "moonshot": "api.moonshot.ai", "openai": "api.openai.com",
		"openrouter": "openrouter.ai", "perplexity-ai": "api.perplexity.ai", "together-ai": "api.together.xyz",
		"tokenbazaar": "zx1.deepwl.net", "xai": "api.x.ai",
	}
	if expected := exact[normalizeProvider(providerSlug)]; expected != "" {
		return host == expected
	}

	switch normalizeProvider(providerSlug) {
	case "azure-ai-foundry", "azure-open-ai":
		return strings.HasSuffix(host, ".openai.azure.com") || strings.HasSuffix(host, ".services.ai.azure.com")
	case "databricks":
		return strings.HasSuffix(host, ".cloud.databricks.com") || strings.HasSuffix(host, ".azuredatabricks.net") || strings.HasSuffix(host, ".gcp.databricks.com")
	case "sambanova":
		return true // SambaStack is customer-hosted; public HTTPS and local/IP blocks above still apply.
	default:
		return false
	}
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
	value = strings.ToLower(strings.TrimSpace(value))
	if definition, ok := providerregistry.Lookup(value); ok {
		return definition.Slug
	}

	switch value {
	case "azure":
		return "azure-open-ai"
	case "bedrock":
		return "aws-bedrock"
	case "vertexai":
		return "google-vertex"
	case "togetherai":
		return "together-ai"
	case "perplexity", "perplexityai":
		return "perplexity-ai"
	default:
		return value
	}
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

		if field.Type == "multi_select" {
			values, ok := value.([]string)
			if !ok {
				raw, rawOK := value.([]interface{})
				if !rawOK {
					return false
				}

				for _, item := range raw {
					text, textOK := item.(string)
					if !textOK || strings.TrimSpace(text) == "" {
						return false
					}
				}
			} else {
				for _, text := range values {
					if strings.TrimSpace(text) == "" {
						return false
					}
				}
			}
		} else {
			text, ok := value.(string)
			if !ok || strings.TrimSpace(text) == "" {
				return false
			}
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
