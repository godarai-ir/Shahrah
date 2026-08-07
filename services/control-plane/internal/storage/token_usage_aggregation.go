package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"go.uber.org/zap"
)

// TokenUsageMetrics records token usage aggregation health without coupling storage to Prometheus.
type TokenUsageMetrics interface {
	IncTokenSpendQueued()
	IncTokenSpendRedisFallback()
	ObserveTokenSpendFlush(batchSize int, duration time.Duration, err error)
	IncKeyLastUsedQueued()
	IncKeyLastUsedSkipped()
	IncKeyLastUsedRedisFallback()
	ObserveKeyLastUsedFlush(batchSize int, duration time.Duration, err error)
}

// StartTokenUsageFlusher periodically persists Redis token usage aggregates to PostgreSQL.
func (s *ObservabilityRepository) StartTokenUsageFlusher(ctx context.Context) {
	startPeriodicFlusher(
		ctx,
		s.logger,
		s.tokenUsageEnabled && s.tokenUsageCache != nil,
		s.tokenUsageFlushInterval,
		"token usage",
		s.FlushTokenUsageAggregates,
	)
}

// FlushTokenUsageAggregates persists one bounded Redis batch for spend and last-used aggregates.
func (s *ObservabilityRepository) FlushTokenUsageAggregates(ctx context.Context) error {
	if s.tokenUsageCache == nil {
		return nil
	}

	if err := s.flushTokenSpendDeltas(ctx); err != nil {
		return err
	}

	return s.flushKeyLastUsed(ctx)
}

func (s *ObservabilityRepository) queueTokenSpendDelta(ctx context.Context, tokenHash string, deltaUSD float64) error {
	tokenHash = strings.TrimSpace(tokenHash)
	if tokenHash == "" || deltaUSD <= 0 {
		return nil
	}

	if !s.tokenUsageEnabled || s.tokenUsageCache == nil {
		return nil
	}

	err := s.tokenUsageCache.QueueTokenSpendDelta(ctx, s.tokenSpendDeltaKey(tokenHash), s.tokenSpendDirtySetKey(), deltaUSD, s.tokenUsageDeltaTTL, time.Now().UTC().Unix())
	if err == nil {
		if s.tokenUsageMetrics != nil {
			s.tokenUsageMetrics.IncTokenSpendQueued()
		}

		return nil
	}

	if s.tokenUsageMetrics != nil {
		s.tokenUsageMetrics.IncTokenSpendRedisFallback()
	}

	if s.logger != nil {
		s.logger.Warn("redis token spend aggregation unavailable; spend log remains source of truth", zap.Error(err))
	}

	if s.tokenUsageFailOpen {
		return nil
	}

	return err
}

func (s *ObservabilityRepository) queueKeyLastUsed(ctx context.Context, keyID string, usedAt time.Time) error {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" || usedAt.IsZero() {
		return nil
	}

	if !s.tokenUsageEnabled || s.tokenUsageCache == nil {
		return s.touchKeyLastUsedFallback(ctx, keyID, usedAt)
	}

	updated, err := s.tokenUsageCache.QueueKeyLastUsed(ctx, s.keyLastUsedKey(keyID), s.keyLastUsedDirtySetKey(), usedAt.UTC().UnixMilli(), s.lastUsedTTL)
	if err == nil {
		if s.tokenUsageMetrics != nil {
			if updated {
				s.tokenUsageMetrics.IncKeyLastUsedQueued()
			} else {
				s.tokenUsageMetrics.IncKeyLastUsedSkipped()
			}
		}

		return nil
	}

	if s.tokenUsageMetrics != nil {
		s.tokenUsageMetrics.IncKeyLastUsedRedisFallback()
	}

	if s.logger != nil {
		s.logger.Warn("redis key last-used aggregation unavailable; using throttled db fallback", zap.Error(err))
	}

	if s.tokenUsageFailOpen {
		return s.touchKeyLastUsedFallback(ctx, keyID, usedAt)
	}

	return err
}

func (s *ObservabilityRepository) touchKeyLastUsedFallback(ctx context.Context, keyID string, usedAt time.Time) error {
	interval := s.lastUsedFlushInterval
	if interval <= 0 {
		interval = time.Minute
	}

	s.lastUsedFallbackMu.Lock()
	if s.lastUsedFallback == nil {
		s.lastUsedFallback = make(map[string]time.Time)
	}

	lastWrite := s.lastUsedFallback[keyID]
	if !lastWrite.IsZero() && usedAt.Sub(lastWrite) < interval {
		s.lastUsedFallbackMu.Unlock()

		if s.tokenUsageMetrics != nil {
			s.tokenUsageMetrics.IncKeyLastUsedSkipped()
		}

		return nil
	}

	s.lastUsedFallback[keyID] = usedAt
	s.lastUsedFallbackMu.Unlock()

	return s.updateKeyLastUsedBatch(ctx, []keyLastUsedUpdate{{KeyID: keyID, UsedAt: usedAt.UTC()}})
}

func (s *ObservabilityRepository) flushTokenSpendDeltas(ctx context.Context) error {
	start := time.Now()

	deltas, err := s.tokenUsageCache.DrainTokenSpendDeltas(ctx, s.tokenSpendDirtySetKey(), s.tokenUsageFlushBatchSize)
	if err != nil {
		if s.tokenUsageMetrics != nil {
			s.tokenUsageMetrics.ObserveTokenSpendFlush(0, time.Since(start), err)
		}

		return err
	}

	if len(deltas) == 0 {
		return nil
	}

	updates := make([]tokenSpendUpdate, 0, len(deltas))
	for _, delta := range deltas {
		tokenHash, parseErr := s.parseTokenSpendDeltaKey(delta.Key)
		if parseErr != nil {
			if s.logger != nil {
				s.logger.Warn("dropping malformed token spend delta key", zap.String("key", delta.Key), zap.Error(parseErr))
			}

			continue
		}

		if delta.DeltaUSD > 0 {
			updates = append(updates, tokenSpendUpdate{TokenHash: tokenHash, DeltaUSD: delta.DeltaUSD})
		}
	}

	if len(updates) > 0 {
		err = s.updateTokenSpendBatch(ctx, updates)
		if err != nil {
			s.restoreTokenSpendDeltas(ctx, updates)
		}
	}

	if s.tokenUsageMetrics != nil {
		s.tokenUsageMetrics.ObserveTokenSpendFlush(len(updates), time.Since(start), err)
	}

	return err
}

func (s *ObservabilityRepository) flushKeyLastUsed(ctx context.Context) error {
	start := time.Now()

	snapshots, err := s.tokenUsageCache.DrainKeyLastUsed(ctx, s.keyLastUsedDirtySetKey(), s.tokenUsageFlushBatchSize)
	if err != nil {
		if s.tokenUsageMetrics != nil {
			s.tokenUsageMetrics.ObserveKeyLastUsedFlush(0, time.Since(start), err)
		}

		return err
	}

	if len(snapshots) == 0 {
		return nil
	}

	updates := make([]keyLastUsedUpdate, 0, len(snapshots))
	for _, snapshot := range snapshots {
		keyID, parseErr := s.parseKeyLastUsedKey(snapshot.Key)
		if parseErr != nil {
			if s.logger != nil {
				s.logger.Warn("dropping malformed key last-used key", zap.String("key", snapshot.Key), zap.Error(parseErr))
			}

			continue
		}

		if snapshot.UsedAtUnixMS > 0 {
			updates = append(updates, keyLastUsedUpdate{KeyID: keyID, UsedAt: time.UnixMilli(snapshot.UsedAtUnixMS).UTC()})
		}
	}

	if len(updates) > 0 {
		err = s.updateKeyLastUsedBatch(ctx, updates)
		if err != nil {
			s.restoreKeyLastUsed(ctx, updates)
		}
	}

	if s.tokenUsageMetrics != nil {
		s.tokenUsageMetrics.ObserveKeyLastUsedFlush(len(updates), time.Since(start), err)
	}

	return err
}

type tokenSpendUpdate struct {
	TokenHash string
	DeltaUSD  float64
}

type keyLastUsedUpdate struct {
	KeyID  string
	UsedAt time.Time
}

func (s *ObservabilityRepository) updateTokenSpendBatch(ctx context.Context, updates []tokenSpendUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	tokenHashes := make([]string, 0, len(updates))

	deltas := make([]float64, 0, len(updates))
	for _, update := range updates {
		tokenHashes = append(tokenHashes, update.TokenHash)
		deltas = append(deltas, update.DeltaUSD)
	}

	_, err := s.pool.Exec(ctx, `
		WITH delta(token_hash, spend_delta) AS (
			SELECT v.token_hash, SUM(v.spend_delta)
			FROM UNNEST($1::text[], $2::double precision[]) AS v(token_hash, spend_delta)
			GROUP BY v.token_hash
		)
		UPDATE verification_tokens vt
		SET spend = COALESCE(vt.spend, 0) + delta.spend_delta,
		    updated_at = NOW()
		FROM delta
		WHERE vt.token_hash = delta.token_hash
	`, tokenHashes, deltas)

	return err
}

func (s *ObservabilityRepository) updateKeyLastUsedBatch(ctx context.Context, updates []keyLastUsedUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	keyIDs := make([]string, 0, len(updates))

	usedAt := make([]time.Time, 0, len(updates))
	for _, update := range updates {
		keyIDs = append(keyIDs, update.KeyID)
		usedAt = append(usedAt, update.UsedAt)
	}

	_, err := s.pool.Exec(ctx, `
		WITH latest(id, used_at) AS (
			SELECT v.id::uuid, MAX(v.used_at)
			FROM UNNEST($1::text[], $2::timestamptz[]) AS v(id, used_at)
			GROUP BY v.id
		)
		UPDATE verification_tokens vt
		SET last_used_at = GREATEST(COALESCE(vt.last_used_at, latest.used_at), latest.used_at),
		    updated_at = NOW()
		FROM latest
		WHERE vt.id = latest.id
		  AND (vt.last_used_at IS NULL OR vt.last_used_at < latest.used_at)
	`, keyIDs, usedAt)

	return err
}

func (s *ObservabilityRepository) restoreTokenSpendDeltas(ctx context.Context, updates []tokenSpendUpdate) {
	for _, update := range updates {
		if err := s.tokenUsageCache.QueueTokenSpendDelta(ctx, s.tokenSpendDeltaKey(update.TokenHash), s.tokenSpendDirtySetKey(), update.DeltaUSD, s.tokenUsageDeltaTTL, time.Now().UTC().Unix()); err != nil && s.logger != nil {
			s.logger.Error("failed to restore drained token spend delta", zap.Error(err))
		}
	}
}

func (s *ObservabilityRepository) restoreKeyLastUsed(ctx context.Context, updates []keyLastUsedUpdate) {
	for _, update := range updates {
		if _, err := s.tokenUsageCache.QueueKeyLastUsed(ctx, s.keyLastUsedKey(update.KeyID), s.keyLastUsedDirtySetKey(), update.UsedAt.UTC().UnixMilli(), s.lastUsedTTL); err != nil && s.logger != nil {
			s.logger.Error("failed to restore drained key last-used timestamp", zap.Error(err))
		}
	}
}

func (s *ObservabilityRepository) tokenSpendDirtySetKey() string {
	return s.tokenUsagePrefix() + ":token_spend_delta:dirty"
}

func (s *ObservabilityRepository) tokenSpendDeltaKey(tokenHash string) string {
	return s.tokenUsagePrefix() + ":token_spend_delta:" + strings.TrimSpace(tokenHash)
}

func (s *ObservabilityRepository) parseTokenSpendDeltaKey(key string) (string, error) {
	prefix := s.tokenUsagePrefix() + ":token_spend_delta:"
	if !strings.HasPrefix(key, prefix) {
		return "", fmt.Errorf("missing token spend prefix")
	}

	tokenHash := strings.TrimPrefix(key, prefix)
	if tokenHash == "" || tokenHash == "dirty" {
		return "", fmt.Errorf("missing token hash")
	}

	return tokenHash, nil
}

func (s *ObservabilityRepository) keyLastUsedDirtySetKey() string {
	return s.tokenUsagePrefix() + ":key_last_used:dirty"
}

func (s *ObservabilityRepository) keyLastUsedKey(keyID string) string {
	return s.tokenUsagePrefix() + ":key_last_used:" + strings.TrimSpace(keyID)
}

func (s *ObservabilityRepository) parseKeyLastUsedKey(key string) (string, error) {
	prefix := s.tokenUsagePrefix() + ":key_last_used:"
	if !strings.HasPrefix(key, prefix) {
		return "", fmt.Errorf("missing key last-used prefix")
	}

	keyID := strings.TrimPrefix(key, prefix)
	if keyID == "" || keyID == "dirty" {
		return "", fmt.Errorf("missing key id")
	}

	return keyID, nil
}

func (s *ObservabilityRepository) tokenUsagePrefix() string {
	return strings.TrimSuffix(s.tokenUsageKeyPrefix, ":")
}

var _ cache.TokenUsageCache = (*cache.DualCache)(nil)
