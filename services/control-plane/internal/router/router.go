package router

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
	"go.uber.org/zap"
)

// Errors.
var (
	ErrNoHealthyDeployments = errors.New("no healthy deployments available")
	ErrModelNotFound        = errors.New("model not found")
	ErrAllDeploymentsFailed = errors.New("all deployments failed or are in cooldown")
)

// Deployment represents a single model deployment.
type Deployment struct {
	ID        string
	TenantID  string
	ModelName string
	Provider  string
	Model     string
	Params    config.ProviderEndpointConfig
	Healthy   bool
	Weight    float64
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
	mu          sync.RWMutex
	deployments map[string][]*Deployment // model_name -> deployments
	cooldowns   *CooldownCache
	strategy    RoutingStrategy
	fallbacks   map[string][]string
	numRetries  int
	timeout     time.Duration
	cache       cache.Cache
	logger      *zap.Logger
	store       DeploymentStore
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
	}

	// Initialize cooldown cache
	cooldownTTL := time.Duration(settings.CooldownTime * float64(time.Second))
	r.cooldowns = NewCooldownCache(c, cooldownTTL)

	// Initialize routing strategy
	r.strategy = NewSimpleShuffle()

	// Parse model list
	for _, md := range modelList {
		deployment := &Deployment{
			ID:        fmt.Sprintf("%v", md.ModelInfo["id"]),
			ModelName: md.ModelName,
			Params:    md.ProviderEndpoint,
			Healthy:   true,
		}

		// Parse provider from model string or config (e.g., "openai/gpt-3.5-turbo")
		modelParts := md.ProviderEndpoint.Model

		// First, check if provider is explicitly set in config
		if md.ProviderEndpoint.Provider != "" {
			deployment.Provider = md.ProviderEndpoint.Provider
			deployment.Model = canonicalProviderModel(deployment.Provider, modelParts)
		} else {
			// Parse from model string prefix - support all 10 providers
			providerPrefixes := []string{
				"openai/", "azure/", "anthropic/", "groq/", "mistral/",
				"togetherai/", "cohere/", "vertexai/", "gemini/", "bedrock/",
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
		dynamic, err := r.store.ListDeployments(ctx, tenantID)
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

					return r.applyBYOK(ctx, tenantID, modelName, selected), nil
				}
			}
		}

		return nil, ErrNoHealthyDeployments
	}

	selected, err := r.strategy.SelectDeployment(healthyDeployments, modelName)
	if err != nil {
		return nil, err
	}

	return r.applyBYOK(ctx, tenantID, modelName, selected), nil
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
	params := config.ProviderEndpointConfig{
		Model:      canonicalModel,
		APIKey:     dep.APIKey,
		APIBase:    dep.APIBase,
		APIVersion: dep.APIVersion,
		RPM:        dep.RPM,
		TPM:        dep.TPM,
		Weight:     dep.Weight,
		Timeout:    dep.TimeoutSec,
		RegionName: dep.RegionName,
	}

	// Provider-account deployments may omit BYOK credentials; in that case, fall back
	// to the shared provider configuration loaded from proxy_config/model_list.
	if strings.HasPrefix(strings.TrimSpace(dep.ID), "provider-account:") {
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

			if strings.TrimSpace(params.RegionName) == "" {
				params.RegionName = shared.Params.RegionName
			}

			if params.Timeout == 0 {
				params.Timeout = shared.Params.Timeout
			}
		}
	}

	hydrated := platform.Deployment{
		ID:         dep.ID,
		TenantID:   dep.TenantID,
		ModelName:  dep.ModelName,
		Provider:   dep.Provider,
		Model:      canonicalModel,
		APIBase:    params.APIBase,
		APIKey:     params.APIKey,
		APIVersion: params.APIVersion,
		RegionName: params.RegionName,
		RPM:        params.RPM,
		TPM:        params.TPM,
		Weight:     params.Weight,
		TimeoutSec: params.Timeout,
		ModelInfo:  dep.ModelInfo,
	}
	if isProviderAccountDeploymentWithoutRuntimeCredential(hydrated) {
		return nil
	}

	return &Deployment{
		ID:        dep.ID,
		TenantID:  dep.TenantID,
		ModelName: dep.ModelName,
		Provider:  dep.Provider,
		Model:     canonicalModel,
		Params:    params,
		Healthy:   true,
		Weight:    fallbackWeight(hydrated),
	}
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
