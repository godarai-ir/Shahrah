package router

import (
	"context"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

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
