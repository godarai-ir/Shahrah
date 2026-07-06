package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"go.uber.org/zap"
)

type usageCounterKeyParts struct {
	TenantID    string
	FeatureKey  string
	WindowStart time.Time
	WindowEnd   time.Time
}

// StartUsageCounterFlusher periodically persists dirty Redis usage counters to PostgreSQL.
func (s *Store) StartUsageCounterFlusher(ctx context.Context) {
	if !s.usageCountersEnabled || s.usageCounterCache == nil || s.usageCounterFlushInterval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(s.usageCounterFlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				flushCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				if err := s.FlushUsageCounters(flushCtx); err != nil && s.logger != nil {
					s.logger.Warn("usage counter shutdown flush failed", zap.Error(err))
				}
				cancel()
				return
			case <-ticker.C:
				if err := s.FlushUsageCounters(ctx); err != nil && s.logger != nil {
					s.logger.Warn("usage counter flush failed", zap.Error(err))
				}
			}
		}
	}()
}

// FlushUsageCounters persists a bounded batch of Redis usage counters to PostgreSQL.
func (s *Store) FlushUsageCounters(ctx context.Context) error {
	if s.usageCounterCache == nil {
		return nil
	}

	snapshots, err := s.usageCounterCache.ListDirtyUsageCounters(ctx, s.usageCounterDirtySetKey(), s.usageCounterFlushBatchSize)
	if err != nil {
		return err
	}
	if len(snapshots) == 0 {
		return nil
	}

	acked := make([]string, 0, len(snapshots))
	for _, snapshot := range snapshots {
		parts, parseErr := s.parseUsageCounterKey(snapshot.Key)
		if parseErr != nil {
			acked = append(acked, snapshot.Key)
			if s.logger != nil {
				s.logger.Warn("dropping malformed usage counter key", zap.String("key", snapshot.Key), zap.Error(parseErr))
			}
			continue
		}
		if snapshot.Used <= 0 {
			acked = append(acked, snapshot.Key)
			continue
		}

		if err := s.upsertUsageCounterSnapshot(ctx, parts, snapshot.Used); err != nil {
			return err
		}
		acked = append(acked, snapshot.Key)
	}

	s.usageCounterCache.AckDirtyUsageCounters(ctx, s.usageCounterDirtySetKey(), acked)
	return nil
}

func (s *Store) upsertUsageCounterSnapshot(ctx context.Context, parts usageCounterKeyParts, used float64) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO tenant_usage_counters (tenant_id, feature_key, window_start, window_end, used, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (tenant_id, feature_key, window_start)
		DO UPDATE SET used = GREATEST(tenant_usage_counters.used, EXCLUDED.used), updated_at = NOW()
	`, parts.TenantID, parts.FeatureKey, parts.WindowStart, parts.WindowEnd, used)
	return err
}

func (s *Store) usageCounterDirtySetKey() string {
	return strings.TrimSuffix(s.usageCounterKeyPrefix, ":") + ":dirty"
}

func (s *Store) usageCounterKey(tenantID string, featureKey string, windowStart time.Time) string {
	return fmt.Sprintf("%s:tenant:%s:feature:%s:window:%d", strings.TrimSuffix(s.usageCounterKeyPrefix, ":"), strings.TrimSpace(tenantID), strings.TrimSpace(featureKey), windowStart.UTC().Unix())
}

func (s *Store) parseUsageCounterKey(key string) (usageCounterKeyParts, error) {
	prefix := strings.TrimSuffix(s.usageCounterKeyPrefix, ":") + ":tenant:"
	if !strings.HasPrefix(key, prefix) {
		return usageCounterKeyParts{}, fmt.Errorf("missing usage counter prefix")
	}

	remainder := strings.TrimPrefix(key, prefix)
	tenantID, remainder, ok := strings.Cut(remainder, ":feature:")
	if !ok {
		return usageCounterKeyParts{}, fmt.Errorf("missing feature segment")
	}
	featureKey, windowRaw, ok := strings.Cut(remainder, ":window:")
	if !ok {
		return usageCounterKeyParts{}, fmt.Errorf("missing window segment")
	}

	windowUnix, err := strconv.ParseInt(windowRaw, 10, 64)
	if err != nil {
		return usageCounterKeyParts{}, fmt.Errorf("parse window unix: %w", err)
	}

	windowStart := time.Unix(windowUnix, 0).UTC()
	return usageCounterKeyParts{
		TenantID:    tenantID,
		FeatureKey:  featureKey,
		WindowStart: windowStart,
		WindowEnd:   windowStart.AddDate(0, 1, 0),
	}, nil
}

var _ cache.UsageCounterCache = (*cache.DualCache)(nil)
