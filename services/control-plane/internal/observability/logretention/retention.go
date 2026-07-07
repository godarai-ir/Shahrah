package logretention

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	CleanupEnabled  bool
	CleanupInterval time.Duration
	CleanupTimeout  time.Duration
}

type LogRetentionPolicy struct {
	TenantID      string
	RetentionDays *int
}

type RetentionStore interface {
	ListLogRetentionPolicies(ctx context.Context) ([]LogRetentionPolicy, error)
	DeleteLogDataBefore(ctx context.Context, tenantID string, cutoff time.Time) error
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

func LoadConfigFromEnv() Config {
	cfg := Config{
		CleanupEnabled:  true,
		CleanupInterval: time.Hour,
		CleanupTimeout:  30 * time.Second,
	}

	cfg.CleanupEnabled = boolEnvOrDefault("LOG_RETENTION_CLEANUP_ENABLED", cfg.CleanupEnabled)
	cfg.CleanupInterval = durationMillisEnvOrDefault("LOG_RETENTION_CLEANUP_INTERVAL_MS", cfg.CleanupInterval)
	cfg.CleanupTimeout = durationMillisEnvOrDefault("LOG_RETENTION_CLEANUP_TIMEOUT_MS", cfg.CleanupTimeout)

	return cfg
}

func NewRetentionCleaner(cfg Config, store RetentionStore, logger *zap.Logger) *RetentionCleaner {
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = time.Hour
	}

	if cfg.CleanupTimeout <= 0 {
		cfg.CleanupTimeout = 30 * time.Second
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
	if c == nil || !c.cfg.CleanupEnabled || c.store == nil {
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

	ticker := time.NewTicker(c.cfg.CleanupInterval)
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
	ctx, cancel := context.WithTimeout(parent, c.cfg.CleanupTimeout)
	defer cancel()

	if err := c.runOnce(ctx); err != nil {
		c.logger.Warn("log retention cleanup finished with errors", zap.Error(err))
	}
}

func (c *RetentionCleaner) runOnce(ctx context.Context) error {
	policies, err := c.store.ListLogRetentionPolicies(ctx)
	if err != nil {
		c.logger.Warn("log retention policy query failed", zap.Error(err))
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
		if err := c.store.DeleteLogDataBefore(ctx, tenantID, cutoff); err != nil {
			c.logger.Warn(
				"tenant log retention cleanup failed",
				zap.String("tenant_id", tenantID),
				zap.Int("retention_days", *policy.RetentionDays),
				zap.Time("cutoff", cutoff),
				zap.Error(err),
			)
			cleanupErr = errors.Join(cleanupErr, err)

			continue
		}

		c.logger.Debug(
			"tenant log retention cleanup scheduled",
			zap.String("tenant_id", tenantID),
			zap.Int("retention_days", *policy.RetentionDays),
			zap.Time("cutoff", cutoff),
		)
	}

	return cleanupErr
}

func ParseLogRetentionDays(value interface{}) (*int, bool) {
	switch v := value.(type) {
	case nil:
		return nil, true
	case int:
		return positiveRetentionDays(v)
	case int64:
		return parseRetentionInt64(v)
	case float64:
		return parseRetentionFloat64(v)
	case json.Number:
		return parseRetentionJSONNumber(v)
	case string:
		return parseRetentionString(v)
	default:
		return nil, false
	}
}

func parseRetentionInt64(value int64) (*int, bool) {
	if value > maxIntValue() {
		return nil, false
	}

	return positiveRetentionDays(int(value))
}

func parseRetentionFloat64(value float64) (*int, bool) {
	if value != math.Trunc(value) || value > float64(maxIntValue()) {
		return nil, false
	}

	return positiveRetentionDays(int(value))
}

func parseRetentionJSONNumber(value json.Number) (*int, bool) {
	parsed, err := value.Int64()
	if err != nil {
		return nil, false
	}

	return parseRetentionInt64(parsed)
}

func parseRetentionString(value string) (*int, bool) {
	raw := strings.TrimSpace(strings.ToLower(value))
	if raw == "" || raw == "null" || raw == "unlimited" {
		return nil, true
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return nil, false
	}

	return positiveRetentionDays(parsed)
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

func durationMillisEnvOrDefault(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			return time.Duration(parsed) * time.Millisecond
		}
	}

	return fallback
}

func boolEnvOrDefault(key string, fallback bool) bool {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if parsed, err := strconv.ParseBool(v); err == nil {
			return parsed
		}
	}

	return fallback
}
