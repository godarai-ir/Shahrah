package router

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/config"
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

		// Parse provider from model string (e.g., "openai/gpt-3.5-turbo")
		modelParts := md.ProviderEndpoint.Model
		if idx := len("openai/"); idx < len(modelParts) && modelParts[:idx] == "openai/" {
			deployment.Provider = "openai"
			deployment.Model = modelParts[idx:]
		} else {
			deployment.Provider = "openai" // default
			deployment.Model = modelParts
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
				deployments = append(deployments, &Deployment{
					ID:        dep.ID,
					TenantID:  dep.TenantID,
					ModelName: dep.ModelName,
					Provider:  dep.Provider,
					Model:     dep.Model,
					Params: config.ProviderEndpointConfig{
						Model:      dep.Model,
						APIKey:     dep.APIKey,
						APIBase:    dep.APIBase,
						APIVersion: dep.APIVersion,
						RPM:        dep.RPM,
						TPM:        dep.TPM,
						Weight:     dep.Weight,
						Timeout:    dep.TimeoutSec,
						RegionName: dep.RegionName,
					},
					Healthy: true,
					Weight:  fallbackWeight(dep),
				})
			}
		}
	}

	_, ok := r.deployments[modelName]
	if !ok {
		// Check for wildcard matching
		for _, dep := range r.deployments["*"] {
			deployments = append(deployments, dep)
		}
		if deployments == nil {
			return nil, ErrModelNotFound
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
					return r.strategy.SelectDeployment(healthyFB, modelName)
				}
			}
		}
		return nil, ErrNoHealthyDeployments
	}

	return r.strategy.SelectDeployment(healthyDeployments, modelName)
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
	r.cooldowns.Add(ctx, deployment.ID, reason)
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
