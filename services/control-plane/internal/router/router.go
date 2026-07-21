package router

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
	"github.com/Delpak/shahrah/services/control-plane/internal/platformcredentials"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
)

// Errors.
var (
	ErrNoHealthyDeployments = errors.New("no healthy deployments available")
	ErrModelNotFound        = errors.New("model not found")
	ErrAllDeploymentsFailed = errors.New("all deployments failed or are in cooldown")
	ErrNoPlatformCredential = errors.New("platform provider credential unavailable")
)

// Deployment represents a single model deployment.
type Deployment struct {
	ID                  string
	TenantID            string
	ModelName           string
	Provider            string
	Model               string
	Params              config.ProviderEndpointConfig
	Healthy             bool
	Weight              float64
	CredentialSource    string
	BillingMode         string
	ProviderAccountID   string
	ProviderAccountName string
	ConnectionID        string
	CredentialVersionID string
	ProviderAuth        *platform.ResolvedProviderAuth
}

// DeploymentStore allows dynamic deployment lookup from persistence.
type DeploymentStore interface {
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
}

type byokCredentialResolver interface {
	ResolveProviderAccountAPIKey(ctx context.Context, tenantID string, modelFQN string) (string, error)
}

// Router is the main LLM request router.
type Router struct {
	mu                            sync.RWMutex
	deployments                   map[string][]*Deployment // model_name -> deployments
	cooldowns                     *CooldownCache
	strategy                      RoutingStrategy
	fallbacks                     map[string][]string
	numRetries                    int
	timeout                       time.Duration
	cache                         cache.Cache
	logger                        *zap.Logger
	store                         DeploymentStore
	dynamicTTL                    time.Duration
	loadGroup                     singleflight.Group
	platformCredentials           *platformcredentials.Manager
	platformCredentialMode        string
	platformCredentialEnvironment string
}

func (r *Router) SetPlatformCredentialResolver(manager *platformcredentials.Manager, mode, environment string) {
	r.platformCredentials = manager

	r.platformCredentialMode = strings.TrimSpace(mode)
	if r.platformCredentialMode == "" {
		r.platformCredentialMode = platform.ProviderCredentialModeEnv
	}

	r.platformCredentialEnvironment = strings.TrimSpace(environment)
	if r.platformCredentialEnvironment == "" {
		r.platformCredentialEnvironment = platform.ProviderCredentialEnvironmentProd
	}
}

// NewRouter creates a new Router instance.
func NewRouter(modelList []config.ModelDeployment, settings config.RouterSettings, c cache.Cache, store DeploymentStore, logger *zap.Logger) (*Router, error) {
	r := &Router{
		deployments: make(map[string][]*Deployment),
		fallbacks:   make(map[string][]string),
		numRetries:  settings.NumRetries,
		cache:       c,
		logger:      logger,
		store:       store,
		dynamicTTL:  secondsEnvOrDefault("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", 30*time.Second),
	}

	// Initialize cooldown cache
	cooldownTTL := time.Duration(settings.CooldownTime * float64(time.Second))
	r.cooldowns = NewCooldownCache(c, cooldownTTL)

	// Initialize routing strategy
	r.strategy = NewSimpleShuffle()

	// Parse model list
	for _, md := range modelList {
		deployment := &Deployment{
			ID:               fmt.Sprintf("%v", md.ModelInfo["id"]),
			ModelName:        md.ModelName,
			Params:           md.ProviderEndpoint,
			Healthy:          true,
			CredentialSource: platform.CredentialSourceGodarManagedKey,
			BillingMode:      platform.BillingModeWalletBilled,
		}

		// Parse provider from model string or config (e.g., "openai/gpt-3.5-turbo")
		modelParts := md.ProviderEndpoint.Model

		// First, check if provider is explicitly set in config
		if md.ProviderEndpoint.Provider != "" {
			deployment.Provider = md.ProviderEndpoint.Provider
			deployment.Model = canonicalProviderModel(deployment.Provider, modelParts)
		} else {
			// Parse from the exact provider slugs currently owned by model-discovery.
			providerPrefixes := []string{
				"ai21/", "anthropic/", "aws-bedrock/", "azure-ai-foundry/", "azure-open-ai/",
				"cerebras/", "cohere/", "databricks/", "deepgram/", "deepinfra/", "elevenlabs/",
				"google-gemini/", "google-vertex/", "groq/", "mistral-ai/", "openai/", "openrouter/",
				"perplexity-ai/", "sambanova/", "together-ai/", "xai/",
				// Staged aliases retained for already-deployed configuration.
				"azure/", "mistral/", "togetherai/", "vertexai/", "gemini/", "bedrock/",
			}
			found := false

			for _, prefix := range providerPrefixes {
				if len(modelParts) > len(prefix) && strings.HasPrefix(modelParts, prefix) {
					deployment.Provider = strings.TrimSuffix(prefix, "/")
					deployment.Model = strings.TrimPrefix(modelParts, prefix)
					found = true

					break
				}
			}

			if !found {
				deployment.Provider = "openai" // explicit default - not silent
				deployment.Model = modelParts
			}
		}

		// Use TPM or RPM as weight
		if md.ProviderEndpoint.TPM > 0 {
			deployment.Weight = float64(md.ProviderEndpoint.TPM)
		} else if md.ProviderEndpoint.RPM > 0 {
			deployment.Weight = float64(md.ProviderEndpoint.RPM)
		} else if md.ProviderEndpoint.Weight > 0 {
			deployment.Weight = md.ProviderEndpoint.Weight
		}

		r.deployments[md.ModelName] = append(r.deployments[md.ModelName], deployment)
	}

	// Parse timeout
	if settings.CooldownTime > 0 {
		r.timeout = time.Duration(settings.CooldownTime * float64(time.Second))
	} else {
		r.timeout = 60 * time.Second
	}

	return r, nil
}

// Route selects a deployment for the given model name.
func (r *Router) Route(ctx context.Context, tenantID, modelName string) (*Deployment, error) {
	r.mu.RLock()
	deployments := append([]*Deployment(nil), r.deployments[modelName]...)
	fallbacks := append([]string(nil), r.fallbacks[modelName]...)
	r.mu.RUnlock()

	if r.store != nil && tenantID != "" {
		dynamic, err := r.listCachedDeployments(ctx, tenantID)
		if err == nil {
			for _, dep := range dynamic {
				if dep.ModelName != modelName && dep.ModelName != "*" {
					continue
				}

				hydrated := r.hydrateDynamicDeployment(dep)
				if hydrated == nil {
					continue
				}

				deployments = append(deployments, hydrated)
			}
		}
	}

	_, ok := r.deployments[modelName]
	if !ok {
		if isModelFQN(modelName) {
			if len(deployments) == 0 {
				return nil, ErrModelNotFound
			}
		} else {
			// Check for wildcard matching, preferring provider-scoped patterns like
			// "anthropic/*" before the global "*" fallback.
			deployments = append(deployments, r.matchWildcardDeployments(modelName)...)
			if len(deployments) == 0 {
				return nil, ErrModelNotFound
			}
		}
	}

	// Filter out unhealthy/cooldown deployments
	healthyDeployments := r.filterHealthy(ctx, deployments)
	if len(healthyDeployments) == 0 {
		// Try fallbacks
		for _, fb := range fallbacks {
			if fbDeployments, exists := r.deployments[fb]; exists {
				healthyFB := r.filterHealthy(ctx, fbDeployments)
				if len(healthyFB) > 0 {
					selected, err := r.strategy.SelectDeployment(healthyFB, modelName)
					if err != nil {
						return nil, err
					}

					return r.applyCredentials(ctx, tenantID, modelName, selected)
				}
			}
		}

		return nil, ErrNoHealthyDeployments
	}

	selected, err := r.strategy.SelectDeployment(healthyDeployments, modelName)
	if err != nil {
		return nil, err
	}

	return r.applyCredentials(ctx, tenantID, modelName, selected)
}

func (r *Router) applyCredentials(ctx context.Context, tenantID, modelName string, selected *Deployment) (*Deployment, error) {
	selected = r.applyBYOK(ctx, tenantID, modelName, selected)
	if selected == nil || selected.CredentialSource == platform.CredentialSourceUserBYOK || r.platformCredentialMode == "" || r.platformCredentialMode == platform.ProviderCredentialModeEnv {
		return selected, nil
	}

	if r.platformCredentials != nil {
		resolved, err := r.platformCredentials.Resolve(selected.Provider, r.platformCredentialEnvironment)
		if err == nil {
			cloned := *selected
			cloned.Params = selected.Params

			cloned.Params.APIKey = resolved.APIKey
			if resolved.BaseURL != "" {
				cloned.Params.APIBase = resolved.BaseURL
			}

			if resolved.APIVersion != "" {
				cloned.Params.APIVersion = resolved.APIVersion
			}

			cloned.Params.CustomHeaders = providerConnectionHeaders(resolved)
			cloned.CredentialSource = platform.CredentialSourcePlatformManaged
			cloned.BillingMode = platform.BillingModeWalletBilled
			cloned.ConnectionID = resolved.ConnectionID
			cloned.CredentialVersionID = resolved.CredentialVersionID
			cloned.ProviderAuth = &platform.ResolvedProviderAuth{Mode: resolved.AuthMode, APIKey: resolved.APIKey}

			return &cloned, nil
		}
	}

	if r.platformCredentialMode == platform.ProviderCredentialModeBreakGlass && strings.TrimSpace(selected.Params.APIKey) != "" {
		cloned := *selected
		cloned.CredentialSource = platform.CredentialSourceEnvBreakGlass
		cloned.BillingMode = platform.BillingModeWalletBilled
		cloned.ProviderAuth = &platform.ResolvedProviderAuth{Mode: "api_key", APIKey: cloned.Params.APIKey}

		return &cloned, nil
	}

	return nil, ErrNoPlatformCredential
}

func providerConnectionHeaders(resolved platformcredentials.ResolvedCredential) map[string]string {
	headers := make(map[string]string)
	if resolved.Organization != "" {
		headers["OpenAI-Organization"] = resolved.Organization
	}

	if resolved.Project != "" {
		headers["OpenAI-Project"] = resolved.Project
	}

	if resolved.HTTPReferer != "" {
		headers["HTTP-Referer"] = resolved.HTTPReferer
	}

	if resolved.AppTitle != "" {
		headers["X-OpenRouter-Title"] = resolved.AppTitle
	}

	if resolved.ClientName != "" {
		headers["X-Client-Name"] = resolved.ClientName
	}

	if len(headers) == 0 {
		return nil
	}

	return headers
}

func (r *Router) listCachedDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error) {
	tenantID = strings.TrimSpace(tenantID)
	if r.cache == nil || r.dynamicTTL <= 0 {
		return r.store.ListDeployments(ctx, tenantID)
	}

	key := deploymentCacheKey(tenantID)
	if cached, ok := r.cache.Get(ctx, key); ok {
		if deployments, ok := deploymentsFromCache(cached); ok {
			return cloneDeployments(deployments), nil
		}
	}

	value, err, _ := r.loadGroup.Do(key, func() (interface{}, error) {
		if cached, ok := r.cache.Get(ctx, key); ok {
			if deployments, ok := deploymentsFromCache(cached); ok {
				return cloneDeployments(deployments), nil
			}
		}

		deployments, err := r.store.ListDeployments(ctx, tenantID)
		if err != nil {
			return nil, err
		}

		deployments = cloneDeployments(deployments)
		if r.canCacheDeploymentSnapshot(deployments) {
			storeDeployments(ctx, r.cache, key, sanitizeDeploymentsForCache(deployments), r.dynamicTTL)
		}

		return deployments, nil
	})
	if err != nil {
		return nil, err
	}

	deployments, ok := value.([]platform.Deployment)
	if !ok {
		return nil, fmt.Errorf("unexpected deployment cache type for tenant %s", tenantID)
	}

	return cloneDeployments(deployments), nil
}

func deploymentsFromCache(value interface{}) ([]platform.Deployment, bool) {
	switch typed := value.(type) {
	case []platform.Deployment:
		return cloneDeployments(typed), true
	case string:
		return deploymentsFromJSON([]byte(typed))
	case []byte:
		return deploymentsFromJSON(typed)
	default:
		return nil, false
	}
}

func deploymentsFromJSON(raw []byte) ([]platform.Deployment, bool) {
	var deployments []platform.Deployment
	if err := json.Unmarshal(raw, &deployments); err != nil {
		return nil, false
	}

	return deployments, true
}

func storeDeployments(ctx context.Context, c cache.Cache, key string, deployments []platform.Deployment, ttl time.Duration) {
	raw, err := json.Marshal(deployments)
	if err != nil {
		return
	}

	c.Set(ctx, key, string(raw), ttl)
}

func sanitizeDeploymentsForCache(in []platform.Deployment) []platform.Deployment {
	out := cloneDeployments(in)
	for i := range out {
		out[i].APIKey = ""
	}

	return out
}

func (r *Router) canCacheDeploymentSnapshot(deployments []platform.Deployment) bool {
	if r.hasProviderAccountCredentialResolver() {
		return true
	}

	for _, dep := range deployments {
		if strings.TrimSpace(dep.APIKey) != "" {
			return false
		}
	}

	return true
}

func deploymentCacheKey(tenantID string) string {
	return "deployments:tenant:" + strings.TrimSpace(tenantID)
}

func cloneDeployments(in []platform.Deployment) []platform.Deployment {
	if len(in) == 0 {
		return nil
	}

	out := make([]platform.Deployment, len(in))
	copy(out, in)

	for i := range out {
		if in[i].ModelInfo != nil {
			out[i].ModelInfo = make(map[string]interface{}, len(in[i].ModelInfo))
			for k, v := range in[i].ModelInfo {
				out[i].ModelInfo[k] = v
			}
		}
	}

	return out
}

func (r *Router) matchWildcardDeployments(modelName string) []*Deployment {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []*Deployment

	for pattern, patternDeployments := range r.deployments {
		if pattern == "*" {
			continue
		}

		prefix, ok := wildcardPrefix(pattern)
		if !ok || !strings.HasPrefix(modelName, prefix) {
			continue
		}

		matches = append(matches, patternDeployments...)
	}

	if len(matches) > 0 {
		return matches
	}

	return append(matches, r.deployments["*"]...)
}

func wildcardPrefix(pattern string) (string, bool) {
	if !strings.HasSuffix(pattern, "*") || pattern == "*" {
		return "", false
	}

	return strings.TrimSuffix(pattern, "*"), true
}

func (r *Router) filterHealthy(ctx context.Context, deployments []*Deployment) []*Deployment {
	var healthy []*Deployment

	for _, d := range deployments {
		if !d.Healthy || r.cooldowns.IsCoolingDown(ctx, d.ID) {
			r.logger.Debug("Skipping deployment", zap.String("id", d.ID), zap.Bool("healthy", d.Healthy))
			continue
		}

		healthy = append(healthy, d)
	}

	return healthy
}

// MarkFailed marks a deployment as failed and adds it to cooldown.
func (r *Router) MarkFailed(ctx context.Context, deployment *Deployment, reason string) {
	r.logger.Warn("Marking deployment as failed",
		zap.String("deployment_id", deployment.ID),
		zap.String("reason", reason),
	)

	if addErr := r.cooldowns.Add(ctx, deployment.ID, reason); addErr != nil {
		r.logger.Warn("failed to add deployment cooldown",
			zap.String("deployment_id", deployment.ID),
			zap.String("reason", reason),
			zap.Error(addErr),
		)
	}
}

// weightedRandomPick selects a deployment based on weights.
func weightedRandomPick(deployments []*Deployment) *Deployment {
	totalWeight := 0.0
	for _, d := range deployments {
		totalWeight += d.Weight
	}

	if totalWeight == 0 {
		return deployments[rand.Intn(len(deployments))]
	}

	r := rand.Float64() * totalWeight

	cumulative := 0.0
	for _, d := range deployments {
		cumulative += d.Weight
		if r <= cumulative {
			return d
		}
	}

	return deployments[len(deployments)-1]
}

func fallbackWeight(dep platform.Deployment) float64 {
	if dep.TPM > 0 {
		return float64(dep.TPM)
	}

	if dep.RPM > 0 {
		return float64(dep.RPM)
	}

	return dep.Weight
}

func isProviderAccountDeploymentWithoutRuntimeCredential(dep platform.Deployment) bool {
	if !strings.HasPrefix(strings.TrimSpace(dep.ID), "provider-account:") {
		return false
	}

	return strings.TrimSpace(dep.APIKey) == "" && strings.TrimSpace(dep.APIBase) == ""
}

func (r *Router) hydrateDynamicDeployment(dep platform.Deployment) *Deployment {
	canonicalModel := canonicalProviderModel(dep.Provider, dep.Model)
	isProviderAccount := strings.HasPrefix(strings.TrimSpace(dep.ID), "provider-account:")
	hasProviderAccountRuntimeCredential := isProviderAccount && strings.TrimSpace(dep.APIKey) != ""
	params := config.ProviderEndpointConfig{
		Model:         canonicalModel,
		APIKey:        dep.APIKey,
		APIBase:       dep.APIBase,
		APIVersion:    dep.APIVersion,
		ProjectID:     dep.ProjectID,
		EndpointID:    dep.EndpointID,
		RecognizerID:  dep.RecognizerID,
		CustomHeaders: dep.ProviderHeaders,
		RPM:           dep.RPM,
		TPM:           dep.TPM,
		Weight:        dep.Weight,
		Timeout:       dep.TimeoutSec,
		RegionName:    dep.RegionName,
	}

	// Provider-account deployments may omit BYOK credentials; in that case, fall back
	// to the shared provider configuration loaded from proxy_config/model_list.
	if isProviderAccount {
		if shared := r.findSharedProviderDeployment(dep.Provider, dep.Model); shared != nil {
			if strings.TrimSpace(params.APIKey) == "" {
				params.APIKey = shared.Params.APIKey
			}

			if strings.TrimSpace(params.APIBase) == "" {
				params.APIBase = shared.Params.APIBase
			}

			if strings.TrimSpace(params.APIVersion) == "" {
				params.APIVersion = shared.Params.APIVersion
			}

			if strings.TrimSpace(params.ProjectID) == "" {
				params.ProjectID = shared.Params.ProjectID
			}

			if strings.TrimSpace(params.EndpointID) == "" {
				params.EndpointID = shared.Params.EndpointID
			}

			if strings.TrimSpace(params.RecognizerID) == "" {
				params.RecognizerID = shared.Params.RecognizerID
			}

			if strings.TrimSpace(params.RegionName) == "" {
				params.RegionName = shared.Params.RegionName
			}

			if params.Timeout == 0 {
				params.Timeout = shared.Params.Timeout
			}
		}
	}

	credentialSource := strings.TrimSpace(dep.CredentialSource)
	if credentialSource == "" {
		credentialSource = platform.CredentialSourceGodarManagedKey
		if hasProviderAccountRuntimeCredential {
			credentialSource = platform.CredentialSourceUserBYOK
		}
	}

	providerAuth := dep.ProviderAuth
	if providerAuth == nil && hasProviderAccountRuntimeCredential {
		providerAuth = &platform.ResolvedProviderAuth{Mode: "api_key", APIKey: params.APIKey}
	}

	billingMode := strings.TrimSpace(dep.BillingMode)
	if billingMode == "" {
		billingMode = platform.BillingModeWalletBilled
		if credentialSource == platform.CredentialSourceUserBYOK {
			billingMode = platform.BillingModeExternalBilled
		}
	}

	hydrated := platform.Deployment{
		ID:                  dep.ID,
		TenantID:            dep.TenantID,
		ModelName:           dep.ModelName,
		Provider:            dep.Provider,
		Model:               canonicalModel,
		APIBase:             params.APIBase,
		APIKey:              params.APIKey,
		APIVersion:          params.APIVersion,
		ProjectID:           params.ProjectID,
		EndpointID:          params.EndpointID,
		RecognizerID:        params.RecognizerID,
		ProviderHeaders:     params.CustomHeaders,
		RegionName:          params.RegionName,
		RPM:                 params.RPM,
		TPM:                 params.TPM,
		Weight:              params.Weight,
		TimeoutSec:          params.Timeout,
		CredentialSource:    credentialSource,
		BillingMode:         billingMode,
		ProviderAccountID:   dep.ProviderAccountID,
		ProviderAccountName: dep.ProviderAccountName,
		ConnectionID:        dep.ConnectionID,
		CredentialVersionID: dep.CredentialVersionID,
		ProviderAuth:        providerAuth,
		ModelInfo:           dep.ModelInfo,
	}
	if isProviderAccountDeploymentWithoutRuntimeCredential(hydrated) && !r.canResolveProviderAccountCredential(dep.ModelName) {
		return nil
	}

	return &Deployment{
		ID:                  dep.ID,
		TenantID:            dep.TenantID,
		ModelName:           dep.ModelName,
		Provider:            dep.Provider,
		Model:               canonicalModel,
		Params:              params,
		Healthy:             true,
		Weight:              fallbackWeight(hydrated),
		CredentialSource:    credentialSource,
		BillingMode:         billingMode,
		ProviderAccountID:   dep.ProviderAccountID,
		ProviderAccountName: dep.ProviderAccountName,
		ConnectionID:        dep.ConnectionID,
		CredentialVersionID: dep.CredentialVersionID,
		ProviderAuth:        providerAuth,
	}
}

func (r *Router) canResolveProviderAccountCredential(modelName string) bool {
	if !isModelFQN(modelName) {
		return false
	}

	return r.hasProviderAccountCredentialResolver()
}

func (r *Router) hasProviderAccountCredentialResolver() bool {
	if r == nil || r.store == nil {
		return false
	}

	_, ok := r.store.(byokCredentialResolver)

	return ok
}

func (r *Router) findSharedProviderDeployment(provider string, model string) *Deployment {
	provider = strings.TrimSpace(strings.ToLower(provider))
	model = strings.TrimSpace(model)

	if provider == "" {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var wildcard *Deployment

	for _, deployments := range r.deployments {
		for _, dep := range deployments {
			if dep == nil || strings.TrimSpace(strings.ToLower(dep.Provider)) != provider {
				continue
			}

			if strings.EqualFold(strings.TrimSpace(dep.Model), model) {
				return dep
			}

			if strings.TrimSpace(dep.Model) == "*" && wildcard == nil {
				wildcard = dep
			}
		}
	}

	return wildcard
}

func (r *Router) applyBYOK(ctx context.Context, tenantID, modelName string, dep *Deployment) *Deployment {
	if dep == nil || r.store == nil || strings.TrimSpace(tenantID) == "" {
		return dep
	}

	if !isModelFQN(modelName) {
		return dep
	}

	resolver, ok := r.store.(byokCredentialResolver)
	if !ok {
		return dep
	}

	apiKey, err := resolver.ResolveProviderAccountAPIKey(ctx, tenantID, modelName)
	if err != nil || strings.TrimSpace(apiKey) == "" {
		return dep
	}

	cloned := *dep
	cloned.Params = dep.Params
	cloned.Params.APIKey = apiKey
	cloned.CredentialSource = platform.CredentialSourceUserBYOK
	cloned.BillingMode = platform.BillingModeExternalBilled
	cloned.ProviderAuth = &platform.ResolvedProviderAuth{Mode: "api_key", APIKey: apiKey}

	return &cloned
}

func isModelFQN(model string) bool {
	parts := strings.Split(strings.TrimSpace(model), ":")
	if len(parts) != 3 {
		return false
	}

	return strings.TrimSpace(parts[0]) != "" && strings.TrimSpace(parts[1]) != "" && strings.TrimSpace(parts[2]) != ""
}

func canonicalProviderModel(provider string, model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}

	return modelid.CanonicalOfficialModel(provider, model)
}

func secondsEnvOrDefault(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds < 0 {
		return fallback
	}

	return time.Duration(seconds) * time.Second
}
