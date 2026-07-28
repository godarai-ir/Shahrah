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
	ErrNoHealthyDeployments                = errors.New("no healthy deployments available")
	ErrModelNotFound                       = errors.New("model not found")
	ErrDeploymentNotFound                  = errors.New("deployment not found")
	ErrAllDeploymentsFailed                = errors.New("all deployments failed or are in cooldown")
	ErrNoPlatformCredential                = errors.New("platform provider credential unavailable")
	ErrDependencyUnavailable               = errors.New("routing dependency unavailable")
	ErrDeploymentBillingMetadataUnresolved = errors.New("deployment billing metadata unresolved")
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
	EffectivePricing    *platform.ModelPrice
	PricingSource       string
}

// DeploymentStore allows dynamic deployment lookup from persistence.
type DeploymentStore interface {
	ListDeployments(ctx context.Context, tenantID string) ([]platform.Deployment, error)
}

type byokCredentialResolver interface {
	ResolveProviderAccountAPIKey(ctx context.Context, tenantID string, modelFQN string) (string, error)
}

type PricingResolver func(ctx context.Context, provider, model string) (*platform.ModelPrice, string, error)

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
	lastKnownGoodDeployments      map[string][]platform.Deployment
	lastKnownGoodRoutes           map[string]*Deployment
	deploymentFallbackActive      map[string]bool
	pricingResolver               PricingResolver
	lastKnownGoodPricing          map[string]resolvedPricing
}

type resolvedPricing struct {
	price  *platform.ModelPrice
	source string
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

func (r *Router) SetPricingResolver(resolver PricingResolver) {
	r.mu.Lock()
	r.pricingResolver = resolver
	r.mu.Unlock()
}

// NewRouter creates a new Router instance.
func NewRouter(modelList []config.ModelDeployment, settings config.RouterSettings, c cache.Cache, store DeploymentStore, logger *zap.Logger) (*Router, error) {
	r := &Router{
		deployments:              make(map[string][]*Deployment),
		fallbacks:                make(map[string][]string),
		numRetries:               settings.NumRetries,
		cache:                    c,
		logger:                   logger,
		store:                    store,
		dynamicTTL:               secondsEnvOrDefault("ROUTER_DEPLOYMENT_CACHE_TTL_SECONDS", 30*time.Second),
		lastKnownGoodDeployments: make(map[string][]platform.Deployment),
		lastKnownGoodRoutes:      make(map[string]*Deployment),
		deploymentFallbackActive: make(map[string]bool),
		lastKnownGoodPricing:     make(map[string]resolvedPricing),
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
	_, modelConfigured := r.deployments[modelName]
	r.mu.RUnlock()

	var dynamicErr error
	var invalidBillingMetadata bool

	dynamicIDs := make(map[string]struct{})

	if r.store != nil && tenantID != "" {
		dynamic, err := r.listCachedDeployments(ctx, tenantID)
		storeFailed := err != nil
		hasResolvedFallback := false

		if err != nil {
			dynamicErr = fmt.Errorf("%w: deployment store: %v", ErrDependencyUnavailable, err)

			if snapshot, ok := r.lastKnownGoodDeploymentSnapshot(tenantID); ok {
				dynamic = snapshot

				if r.markDeploymentFallbackActive(tenantID) {
					r.logger.Warn("Using last-known-good deployment snapshot",
						zap.String("tenant_id", tenantID),
						zap.Error(err),
					)
				}
			}

			if resolved, ok := r.lastKnownGoodRoute(tenantID, modelName); ok {
				deployments = append(deployments, resolved)
				dynamicIDs[resolved.ID] = struct{}{}
				hasResolvedFallback = true
			}
		} else {
			r.rememberDeploymentSnapshot(tenantID, dynamic)
		}

		for _, dep := range dynamic {
			if dep.ModelName != modelName && dep.ModelName != "*" {
				continue
			}

			dynamicIDs[dep.ID] = struct{}{}

			if hasResolvedFallback {
				continue
			}

			if !deploymentBillingMetadataValid(dep) {
				invalidBillingMetadata = true
				continue
			}

			hydrated := r.hydrateDynamicDeployment(dep)
			if hydrated == nil {
				continue
			}

			if storeFailed && dynamicDeploymentNeedsResolvedCredential(hydrated) {
				continue
			}

			deployments = append(deployments, hydrated)
		}
	}

	if !modelConfigured {
		if isModelFQN(modelName) {
			if len(deployments) == 0 {
				if dynamicErr != nil {
					return nil, dynamicErr
				}
				if invalidBillingMetadata {
					return nil, ErrDeploymentBillingMetadataUnresolved
				}

				return nil, ErrModelNotFound
			}
		} else {
			// Check for wildcard matching, preferring provider-scoped patterns like
			// "anthropic/*" before the global "*" fallback.
			deployments = append(deployments, r.matchWildcardDeployments(modelName)...)
			if len(deployments) == 0 {
				if dynamicErr != nil {
					return nil, dynamicErr
				}
				if invalidBillingMetadata {
					return nil, ErrDeploymentBillingMetadataUnresolved
				}

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

					return r.finalizeRoute(ctx, tenantID, modelName, selected, dynamicIDs)
				}
			}
		}

		if dynamicErr != nil {
			return nil, dynamicErr
		}

		return nil, ErrNoHealthyDeployments
	}

	selected, err := r.strategy.SelectDeployment(healthyDeployments, modelName)
	if err != nil {
		return nil, err
	}

	return r.finalizeRoute(ctx, tenantID, modelName, selected, dynamicIDs)
}

// ResolveDeployment returns the exact persisted deployment for an asynchronous
// request. It deliberately bypasses health filtering and load balancing because
// an in-flight provider job cannot be moved to another deployment.
func (r *Router) ResolveDeployment(ctx context.Context, tenantID, modelName, deploymentID string) (*Deployment, error) {
	deploymentID = strings.TrimSpace(deploymentID)
	if deploymentID == "" {
		return nil, ErrDeploymentNotFound
	}

	if deployment := r.configuredDeploymentByID(modelName, deploymentID); deployment != nil {
		return r.resolveExactDeployment(ctx, tenantID, modelName, deployment)
	}

	if r.store == nil || strings.TrimSpace(tenantID) == "" {
		return nil, ErrDeploymentNotFound
	}

	deployments, err := r.listCachedDeployments(ctx, tenantID)
	storeFailed := err != nil
	var dependencyErr error
	if storeFailed {
		dependencyErr = fmt.Errorf("%w: deployment store: %v", ErrDependencyUnavailable, err)
		if snapshot, ok := r.lastKnownGoodDeploymentSnapshot(tenantID); ok {
			deployments = snapshot
		} else if deployment, ok := r.lastKnownGoodRoute(tenantID, modelName); ok && deployment.ID == deploymentID {
			return r.resolveExactDeployment(ctx, tenantID, modelName, deployment)
		} else {
			return nil, dependencyErr
		}
	} else {
		r.rememberDeploymentSnapshot(tenantID, deployments)
	}

	for _, deployment := range deployments {
		if deployment.ID != deploymentID {
			continue
		}
		if deployment.ModelName != modelName && deployment.ModelName != "*" {
			return nil, ErrDeploymentNotFound
		}
		if !deploymentBillingMetadataValid(deployment) {
			return nil, ErrDeploymentBillingMetadataUnresolved
		}

		hydrated := r.hydrateDynamicDeployment(deployment)
		if hydrated == nil {
			return nil, ErrNoPlatformCredential
		}
		if storeFailed && dynamicDeploymentNeedsResolvedCredential(hydrated) {
			if resolved, ok := r.lastKnownGoodRoute(tenantID, modelName); ok && resolved.ID == deploymentID {
				return r.resolveExactDeployment(ctx, tenantID, modelName, resolved)
			}
			return nil, dependencyErr
		}

		return r.finalizeRoute(ctx, tenantID, modelName, hydrated, map[string]struct{}{deploymentID: {}})
	}

	if storeFailed {
		if deployment, ok := r.lastKnownGoodRoute(tenantID, modelName); ok && deployment.ID == deploymentID {
			return r.resolveExactDeployment(ctx, tenantID, modelName, deployment)
		}
		return nil, dependencyErr
	}

	return nil, ErrDeploymentNotFound
}

func (r *Router) resolveExactDeployment(
	ctx context.Context,
	tenantID string,
	modelName string,
	deployment *Deployment,
) (*Deployment, error) {
	resolved, err := r.applyCredentials(ctx, tenantID, modelName, deployment)
	if err != nil || resolved == nil {
		return resolved, err
	}
	return r.applyResolvedPricing(ctx, modelName, resolved), nil
}

func (r *Router) configuredDeploymentByID(modelName, deploymentID string) *Deployment {
	r.mu.RLock()
	defer r.mu.RUnlock()

	find := func(deployments []*Deployment) *Deployment {
		for _, deployment := range deployments {
			if deployment != nil && deployment.ID == deploymentID {
				return cloneDeployment(deployment)
			}
		}
		return nil
	}

	if deployments, configured := r.deployments[modelName]; configured {
		return find(deployments)
	}

	var wildcardDeployments []*Deployment
	for pattern, deployments := range r.deployments {
		prefix, ok := wildcardPrefix(pattern)
		if !ok || !strings.HasPrefix(modelName, prefix) {
			continue
		}
		wildcardDeployments = append(wildcardDeployments, deployments...)
	}
	if len(wildcardDeployments) > 0 {
		return find(wildcardDeployments)
	}

	return find(r.deployments["*"])
}

func (r *Router) rememberDeploymentSnapshot(tenantID string, deployments []platform.Deployment) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return
	}

	r.mu.Lock()
	r.lastKnownGoodDeployments[tenantID] = cloneDeployments(deployments)
	r.deploymentFallbackActive[tenantID] = false

	allowedIDs := make(map[string]struct{}, len(deployments))
	for _, dep := range deployments {
		allowedIDs[dep.ID] = struct{}{}
	}

	for key, route := range r.lastKnownGoodRoutes {
		if !strings.HasPrefix(key, tenantID+"\x00") {
			continue
		}

		if _, ok := allowedIDs[route.ID]; !ok {
			delete(r.lastKnownGoodRoutes, key)
		}
	}
	r.mu.Unlock()
}

func (r *Router) lastKnownGoodDeploymentSnapshot(tenantID string) ([]platform.Deployment, bool) {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return nil, false
	}

	r.mu.RLock()
	deployments, ok := r.lastKnownGoodDeployments[tenantID]
	r.mu.RUnlock()

	return cloneDeployments(deployments), ok
}

func (r *Router) finalizeRoute(
	ctx context.Context,
	tenantID string,
	modelName string,
	selected *Deployment,
	dynamicIDs map[string]struct{},
) (*Deployment, error) {
	resolved, err := r.applyCredentials(ctx, tenantID, modelName, selected)
	if err != nil || resolved == nil {
		return resolved, err
	}

	resolved = r.applyResolvedPricing(ctx, modelName, resolved)

	if _, ok := dynamicIDs[resolved.ID]; ok {
		r.rememberResolvedRoute(tenantID, modelName, resolved)
	}

	return resolved, nil
}

func (r *Router) applyResolvedPricing(ctx context.Context, modelName string, selected *Deployment) *Deployment {
	if selected == nil ||
		selected.BillingMode != platform.BillingModeWalletBilled ||
		modelPriceResolved(selected.EffectivePricing) {
		return selected
	}

	r.mu.RLock()
	resolver := r.pricingResolver
	r.mu.RUnlock()
	if resolver == nil {
		return selected
	}

	model := pricingModelForRoute(selected, modelName)
	key := strings.ToLower(strings.TrimSpace(selected.Provider) + "\x00" + strings.TrimSpace(model))
	price, source, err := resolver(ctx, selected.Provider, model)
	if err == nil && modelPriceResolved(price) && strings.TrimSpace(source) != "" {
		r.mu.Lock()
		r.lastKnownGoodPricing[key] = resolvedPricing{price: cloneModelPrice(price), source: strings.TrimSpace(source)}
		r.mu.Unlock()

		cloned := cloneDeployment(selected)
		cloned.EffectivePricing = cloneModelPrice(price)
		cloned.PricingSource = strings.TrimSpace(source)
		return cloned
	}

	r.mu.RLock()
	fallback, ok := r.lastKnownGoodPricing[key]
	r.mu.RUnlock()
	if !ok {
		return selected
	}

	cloned := cloneDeployment(selected)
	cloned.EffectivePricing = cloneModelPrice(fallback.price)
	cloned.PricingSource = fallback.source
	return cloned
}

func pricingModelForRoute(deployment *Deployment, modelName string) string {
	if deployment == nil {
		return ""
	}
	model := modelid.CanonicalOfficialModel(deployment.Provider, deployment.Model)
	if model != "" && model != "*" && !strings.HasSuffix(model, "/*") {
		return model
	}
	if provider, _, official, ok := modelid.SplitModelFQN(modelName); ok &&
		strings.EqualFold(provider, deployment.Provider) {
		return official
	}
	return modelid.CanonicalOfficialModel(deployment.Provider, modelName)
}

func modelPriceResolved(price *platform.ModelPrice) bool {
	if price == nil || strings.EqualFold(strings.TrimSpace(price.Status), "unknown") {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(price.Status), "free") || len(price.Profiles) > 0
}

func cloneModelPrice(price *platform.ModelPrice) *platform.ModelPrice {
	if price == nil {
		return nil
	}
	out := *price
	out.Profiles = append([]platform.PricingProfile(nil), price.Profiles...)
	for i := range out.Profiles {
		out.Profiles[i].Rates = append([]platform.PricingRate(nil), price.Profiles[i].Rates...)
	}
	return &out
}

func deploymentBillingMetadataValid(dep platform.Deployment) bool {
	credentialSource := strings.TrimSpace(dep.CredentialSource)
	billingMode := strings.TrimSpace(dep.BillingMode)
	if credentialSource == "" || billingMode == "" {
		return false
	}
	return billingMode == platform.BillingModeWalletBilled ||
		billingMode == platform.BillingModeExternalBilled
}

func (r *Router) rememberResolvedRoute(tenantID, modelName string, deployment *Deployment) {
	key := resolvedRouteKey(tenantID, modelName)
	if key == "" || deployment == nil {
		return
	}

	r.mu.Lock()
	r.lastKnownGoodRoutes[key] = cloneDeployment(deployment)
	r.mu.Unlock()
}

func (r *Router) lastKnownGoodRoute(tenantID, modelName string) (*Deployment, bool) {
	key := resolvedRouteKey(tenantID, modelName)
	if key == "" {
		return nil, false
	}

	r.mu.RLock()
	deployment, ok := r.lastKnownGoodRoutes[key]
	r.mu.RUnlock()

	return cloneDeployment(deployment), ok
}

func (r *Router) markDeploymentFallbackActive(tenantID string) bool {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return false
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.deploymentFallbackActive[tenantID] {
		return false
	}

	r.deploymentFallbackActive[tenantID] = true

	return true
}

func resolvedRouteKey(tenantID, modelName string) string {
	tenantID = strings.TrimSpace(tenantID)

	modelName = strings.TrimSpace(modelName)
	if tenantID == "" || modelName == "" {
		return ""
	}

	return tenantID + "\x00" + modelName
}

func cloneDeployment(in *Deployment) *Deployment {
	if in == nil {
		return nil
	}

	out := *in

	out.Params = in.Params
	if in.Params.CustomHeaders != nil {
		out.Params.CustomHeaders = make(map[string]string, len(in.Params.CustomHeaders))
		for key, value := range in.Params.CustomHeaders {
			out.Params.CustomHeaders[key] = value
		}
	}

	if in.ProviderAuth != nil {
		providerAuth := *in.ProviderAuth
		out.ProviderAuth = &providerAuth
	}
	out.EffectivePricing = cloneModelPrice(in.EffectivePricing)

	return &out
}

func dynamicDeploymentNeedsResolvedCredential(deployment *Deployment) bool {
	if deployment == nil || !strings.HasPrefix(strings.TrimSpace(deployment.ID), "provider-account:") {
		return false
	}

	noCredential := deployment.ProviderAuth != nil &&
		strings.EqualFold(strings.TrimSpace(deployment.ProviderAuth.Mode), string(platform.ProviderAuthModeNoCredential))

	return !noCredential && strings.TrimSpace(deployment.Params.APIKey) == ""
}

func (r *Router) applyCredentials(ctx context.Context, tenantID, modelName string, selected *Deployment) (*Deployment, error) {
	selected = r.applyBYOK(ctx, tenantID, modelName, selected)

	noCredential := selected != nil &&
		selected.ProviderAuth != nil &&
		strings.EqualFold(strings.TrimSpace(selected.ProviderAuth.Mode), string(platform.ProviderAuthModeNoCredential))
	if selected == nil || noCredential || selected.CredentialSource == platform.CredentialSourceUserBYOK || r.platformCredentialMode == "" || r.platformCredentialMode == platform.ProviderCredentialModeEnv {
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
		return nil
	}

	providerAuth := dep.ProviderAuth
	if providerAuth == nil && hasProviderAccountRuntimeCredential {
		providerAuth = &platform.ResolvedProviderAuth{Mode: "api_key", APIKey: params.APIKey}
	}

	billingMode := strings.TrimSpace(dep.BillingMode)
	if billingMode == "" {
		return nil
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
		EffectivePricing:    dep.EffectivePricing,
		PricingSource:       dep.PricingSource,
	}

	noCredential := hydrated.ProviderAuth != nil &&
		strings.EqualFold(strings.TrimSpace(hydrated.ProviderAuth.Mode), string(platform.ProviderAuthModeNoCredential))
	if isProviderAccountDeploymentWithoutRuntimeCredential(hydrated) && !noCredential && !r.canResolveProviderAccountCredential(dep.ModelName) {
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
		EffectivePricing:    dep.EffectivePricing,
		PricingSource:       dep.PricingSource,
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
