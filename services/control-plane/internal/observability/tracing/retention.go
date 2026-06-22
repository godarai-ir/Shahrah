package tracing

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type TraceRetentionPolicy struct {
	TenantID      string
	RetentionDays *int
}

type RetentionStore interface {
	ListTraceRetentionPolicies(ctx context.Context) ([]TraceRetentionPolicy, error)
	DeleteTraceDataBefore(ctx context.Context, tenantID string, cutoff time.Time) error
}

type RetentionCleaner struct {
	cfg    Config
	store  RetentionStore
	logger *zap.Logger

	now    func() time.Time
	cancel context.CancelFunc
	done   chan struct{}
	once   sync.Once
}

func NewRetentionCleaner(cfg Config, store RetentionStore, logger *zap.Logger) *RetentionCleaner {
	if cfg.RetentionCleanupInterval <= 0 {
		cfg.RetentionCleanupInterval = time.Hour
	}
	if cfg.RetentionCleanupTimeout <= 0 {
		cfg.RetentionCleanupTimeout = 30 * time.Second
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &RetentionCleaner{
		cfg:    cfg,
		store:  store,
		logger: logger,
		now:    time.Now,
	}
}

func (c *RetentionCleaner) Start() {
	if c == nil || !c.cfg.RetentionCleanupEnabled || c.store == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.done = make(chan struct{})
	go c.run(ctx)
}

func (c *RetentionCleaner) Stop() {
	if c == nil || c.cancel == nil || c.done == nil {
		return
	}
	c.once.Do(func() {
		c.cancel()
		<-c.done
	})
}

func (c *RetentionCleaner) run(ctx context.Context) {
	defer close(c.done)
	c.runWithTimeout(ctx)
	ticker := time.NewTicker(c.cfg.RetentionCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.runWithTimeout(ctx)
		}
	}
}

func (c *RetentionCleaner) runWithTimeout(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, c.cfg.RetentionCleanupTimeout)
	defer cancel()
	if err := c.runOnce(ctx); err != nil {
		c.logger.Warn("trace retention cleanup finished with errors", zap.Error(err))
	}
}

func (c *RetentionCleaner) runOnce(ctx context.Context) error {
	policies, err := c.store.ListTraceRetentionPolicies(ctx)
	if err != nil {
		c.logger.Warn("trace retention policy query failed", zap.Error(err))
		return err
	}
	now := c.now().UTC()
	var cleanupErr error
	for _, policy := range policies {
		tenantID := strings.TrimSpace(policy.TenantID)
		if tenantID == "" || policy.RetentionDays == nil {
			continue
		}
		cutoff := now.AddDate(0, 0, -*policy.RetentionDays)
		if err := c.store.DeleteTraceDataBefore(ctx, tenantID, cutoff); err != nil {
			c.logger.Warn(
				"tenant trace retention cleanup failed",
				zap.String("tenant_id", tenantID),
				zap.Int("retention_days", *policy.RetentionDays),
				zap.Time("cutoff", cutoff),
				zap.Error(err),
			)
			cleanupErr = errors.Join(cleanupErr, err)
			continue
		}
		c.logger.Debug(
			"tenant trace retention cleanup scheduled",
			zap.String("tenant_id", tenantID),
			zap.Int("retention_days", *policy.RetentionDays),
			zap.Time("cutoff", cutoff),
		)
	}
	return cleanupErr
}

func ParseTraceRetentionDays(value interface{}) (*int, bool) {
	switch v := value.(type) {
	case nil:
		return nil, true
	case int:
		return positiveRetentionDays(v)
	case int64:
		if v > maxIntValue() {
			return nil, false
		}
		return positiveRetentionDays(int(v))
	case float64:
		if v != math.Trunc(v) || v > float64(maxIntValue()) {
			return nil, false
		}
		return positiveRetentionDays(int(v))
	case json.Number:
		parsed, err := v.Int64()
		if err != nil || parsed > maxIntValue() {
			return nil, false
		}
		return positiveRetentionDays(int(parsed))
	case string:
		raw := strings.TrimSpace(strings.ToLower(v))
		if raw == "" || raw == "null" || raw == "unlimited" {
			return nil, true
		}
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return nil, false
		}
		return positiveRetentionDays(parsed)
	default:
		return nil, false
	}
}

func positiveRetentionDays(value int) (*int, bool) {
	if value <= 0 {
		return nil, false
	}
	return &value, true
}

func maxIntValue() int64 {
	return int64(^uint(0) >> 1)
}
