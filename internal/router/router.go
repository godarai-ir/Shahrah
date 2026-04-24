package router

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/BerriAI/go-llm-proxy/internal/cache"
	"github.com/BerriAI/go-llm-proxy/internal/config"
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
	ModelName string
	Provider  string
	Model     string
	Params    config.LiteLLMParams
	Healthy   bool
	Weight    float64
}

// CooldownCache tracks cooldowns for failed deployments.
type CooldownCache struct {
	cache cache.Cache
	ttl   time.Duration
}

// NewCooldownCache creates a new cooldown cache.
func NewCooldownCache(c cache.Cache, ttl time.Duration) *CooldownCache {
	return &CooldownCache{cache: c, ttl: ttl}
}

// Add marks a deployment as cooling down.
func (c *CooldownCache) Add(ctx context.Context, deploymentID string, reason string) error {
	c.cache.Set(ctx, fmt.Sprintf("cooldown:%s", deploymentID), reason, c.ttl)
	c.cache.Set(ctx, fmt.Sprintf("cooldown_expires:%s", deploymentID), time.Now().Add(c.ttl), c.ttl)
	return nil
}

// IsCoolingDown checks if a deployment is currently in cooldown.
func (c *CooldownCache) IsCoolingDown(ctx context.Context, deploymentID string) bool {
	expiresKey := fmt.Sprintf("cooldown_expires:%s", deploymentID)
	if val, found := c.cache.Get(ctx, expiresKey); found {
		if expires, ok := val.(time.Time); ok && expires.After(time.Now()) {
			return true
		}
		c.cache.Delete(ctx, fmt.Sprintf("cooldown:%s", deploymentID))
		c.cache.Delete(ctx, expiresKey)
	}
	return false
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
}

// NewRouter creates a new Router instance.
func NewRouter(modelList []config.ModelDeployment, settings config.RouterSettings, c cache.Cache, logger *zap.Logger) (*Router, error) {
	r := &Router{
		deployments: make(map[string][]*Deployment),
		fallbacks:   make(map[string][]string),
		numRetries:  settings.NumRetries,
		cache:       c,
		logger:      logger,
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
			Params:    md.LiteLLMParams,
			Healthy:   true,
		}

		// Parse provider from model string (e.g., "openai/gpt-3.5-turbo")
		modelParts := md.LiteLLMParams.Model
		if idx := len("openai/"); idx < len(modelParts) && modelParts[:idx] == "openai/" {
			deployment.Provider = "openai"
			deployment.Model = modelParts[idx:]
		} else {
			deployment.Provider = "openai" // default
			deployment.Model = modelParts
		}

		// Use TPM or RPM as weight
		if md.LiteLLMParams.TPM > 0 {
			deployment.Weight = float64(md.LiteLLMParams.TPM)
		} else if md.LiteLLMParams.RPM > 0 {
			deployment.Weight = float64(md.LiteLLMParams.RPM)
		} else if md.LiteLLMParams.Weight > 0 {
			deployment.Weight = md.LiteLLMParams.Weight
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
func (r *Router) Route(ctx context.Context, modelName string) (*Deployment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	deployments, ok := r.deployments[modelName]
	if !ok {
		// Check for wildcard matching
		for key := range r.deployments {
			if key == "*" {
				deployments = r.deployments[key]
				break
			}
		}
		if deployments == nil {
			return nil, ErrModelNotFound
		}
	}

	// Filter out unhealthy/cooldown deployments
	healthyDeployments := r.filterHealthy(ctx, deployments)
	if len(healthyDeployments) == 0 {
		// Try fallbacks
		if fallbacks, ok := r.fallbacks[modelName]; ok {
			for _, fb := range fallbacks {
				if fbDeployments, exists := r.deployments[fb]; exists {
					healthyFB := r.filterHealthy(ctx, fbDeployments)
					if len(healthyFB) > 0 {
						return r.strategy.SelectDeployment(healthyFB, modelName)
					}
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
