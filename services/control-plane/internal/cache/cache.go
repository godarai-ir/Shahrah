// Package cache provides local and distributed cache primitives for the control plane.
package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitResult holds the result of a rate-limit check.
type RateLimitResult struct {
	Allowed    bool
	Current    int64
	Limit      int64
	ResetAt    time.Time
	RetryAfter bool
	RuleID     string
	Dimension  string
	Algorithm  string
}

// RateLimitCheck describes one counter in an atomic multi-dimensional rate-limit decision.
type RateLimitCheck struct {
	Key     string
	Limit   int64
	Weight  int64
	Window  time.Duration
	Sliding bool
}

// RateLimitBatchChecker evaluates ordered counters in one backend round trip.
type RateLimitBatchChecker interface {
	CheckRateLimits(ctx context.Context, checks []RateLimitCheck) ([]*RateLimitResult, error)
}

// BudgetWindowSeed describes the active spend allowance window for a tenant.
type BudgetWindowSeed struct {
	WindowKey    string
	AllowanceIRR float64
	ExpiresAt    time.Time
	TTL          time.Duration
	Version      int64
}

// BudgetReservationRequest describes one atomic budget reservation operation.
type BudgetReservationRequest struct {
	WindowKey      string
	ReservationKey string
	ReservationID  string
	EstimatedIRR   float64
	TTL            time.Duration
	Now            time.Time
}

// BudgetCommitRequest describes one atomic budget commit operation.
type BudgetCommitRequest struct {
	WindowKey      string
	ReservationKey string
	ActualIRR      float64
	// Force commits even when the actual amount exceeds the allowance, and
	// records the excess as overage. It is for usage already served.
	Force bool
	Now   time.Time
}

// BudgetSpendRequest adds already-happened spend to a window. The amount's
// unit is the window's unit, fixed by its key prefix.
type BudgetSpendRequest struct {
	WindowKey string
	Amount    float64
	Now       time.Time
}

// BudgetReleaseRequest describes one atomic budget release operation.
type BudgetReleaseRequest struct {
	WindowKey      string
	ReservationKey string
	Now            time.Time
}

// BudgetWindowResult is the result of an allowance operation.
type BudgetWindowResult struct {
	Allowed      bool
	State        string
	AllowanceIRR float64
	ReservedIRR  float64
	CommittedIRR float64
	RemainingIRR float64
	OverageIRR   float64
}

// BudgetWindowSnapshot is the current persisted state of a budget window.
type BudgetWindowSnapshot struct {
	WindowKey     string
	Status        string
	AllowanceIRR  float64
	ReservedIRR   float64
	CommittedIRR  float64
	RemainingIRR  float64
	OverageIRR    float64
	ExpiresAt     time.Time
	Version       int64
	LastUpdatedAt time.Time
}

// TokenSpendDelta is a drained Redis token spend delta ready for persistence.
type TokenSpendDelta struct {
	Key      string
	DeltaUSD float64
}

// TokenSpendDeltaRequest describes one token-spend delta queued for aggregation.
type TokenSpendDeltaRequest struct {
	DeltaKey    string
	DirtySetKey string
	DeltaUSD    float64
	TTL         time.Duration
	Score       int64
}

// KeyLastUsedSnapshot is the newest last-used timestamp drained for one key.
type KeyLastUsedSnapshot struct {
	Key          string
	UsedAtUnixMS int64
}

// TokenUsageCache exposes Redis-backed token usage aggregation primitives.
type TokenUsageCache interface {
	QueueTokenSpendDelta(ctx context.Context, deltaKey string, dirtySetKey string, deltaUSD float64, ttl time.Duration, score int64) error
	DrainTokenSpendDeltas(ctx context.Context, dirtySetKey string, batchSize int) ([]TokenSpendDelta, error)
	QueueKeyLastUsed(ctx context.Context, lastUsedKey string, dirtySetKey string, usedAtUnix int64, ttl time.Duration) (bool, error)
	DrainKeyLastUsed(ctx context.Context, dirtySetKey string, batchSize int) ([]KeyLastUsedSnapshot, error)
}

// TokenUsageBatchCache queues multiple token-spend deltas in one network round trip.
type TokenUsageBatchCache interface {
	QueueTokenSpendDeltas(ctx context.Context, requests []TokenSpendDeltaRequest) error
}

// BudgetWindowInspector exposes live tenant budget window state for UI and settlement flows.
type BudgetWindowInspector interface {
	GetBudgetWindow(ctx context.Context, windowKey string) (*BudgetWindowSnapshot, error)
	ListActiveBudgetWindows(ctx context.Context, batchSize int) ([]BudgetWindowSnapshot, error)
}

// Cache interface defines the contract for cache implementations.
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration)
	Delete(ctx context.Context, key string)
	Increment(ctx context.Context, key string, value int64, ttl time.Duration) (int64, error)
	CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error)
	CheckSlidingWindowRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error)
	CheckSlidingWindowRateLimitWithWeight(ctx context.Context, key string, limit int64, weight int64, window time.Duration) (*RateLimitResult, error)
	Close() error
}

// DualCache implements Cache using both memory and Redis.
type DualCache struct {
	memory *MemoryCache
	redis  redis.UniversalClient
	logger *zap.Logger
}

// NewDualCache creates a new dual cache (memory + Redis).
func NewDualCache(redisHost string, redisPort int, redisPassword string, logger *zap.Logger) (*DualCache, error) {
	return newDualCache(redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	}), logger)
}

// NewDualCacheWithSentinel resolves and follows the active Redis master through Sentinel.
func NewDualCacheWithSentinel(sentinelAddrs []string, masterName, redisPassword, sentinelPassword string, logger *zap.Logger) (*DualCache, error) {
	if len(sentinelAddrs) == 0 {
		return nil, errors.New("at least one Redis Sentinel address is required")
	}

	if strings.TrimSpace(masterName) == "" {
		return nil, errors.New("redis Sentinel master name is required")
	}

	return newDualCache(redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:            masterName,
		SentinelAddrs:         sentinelAddrs,
		Password:              redisPassword,
		SentinelPassword:      sentinelPassword,
		DB:                    0,
		DialTimeout:           2 * time.Second,
		ReadTimeout:           2 * time.Second,
		WriteTimeout:          2 * time.Second,
		PoolSize:              32,
		MinIdleConns:          4,
		ContextTimeoutEnabled: true,
	}), logger)
}

func newDualCache(rdb redis.UniversalClient, logger *zap.Logger) (*DualCache, error) {
	memory, err := NewMemoryCache(logger)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		if closeErr := rdb.Close(); closeErr != nil {
			logger.Debug("failed to close Redis after unsuccessful ping", zap.Error(closeErr))
		}

		if closeErr := memory.Close(); closeErr != nil {
			logger.Debug("failed to close memory cache after unsuccessful Redis ping", zap.Error(closeErr))
		}

		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &DualCache{memory: memory, redis: rdb, logger: logger}, nil
}

// Get retrieves a value from the cache (memory first, then Redis).
func (d *DualCache) Get(ctx context.Context, key string) (interface{}, bool) {
	if isDistributedHotPathKey(key) {
		val, err := d.redis.Get(ctx, key).Result()
		if err != nil {
			return nil, false
		}

		return val, true
	}

	// Try memory cache first
	if val, found := d.memory.Get(ctx, key); found {
		return val, true
	}

	// Fallback to Redis
	val, err := d.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, false
	}

	return val, true
}

// Set stores a value in both memory and Redis.
func (d *DualCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	if isDistributedHotPathKey(key) {
		d.memory.Delete(ctx, key)
		d.redis.Set(ctx, key, value, ttl)

		return
	}

	d.memory.Set(ctx, key, value, ttl)
	d.redis.Set(ctx, key, value, ttl)
}

// DistributedGet bypasses the node-local layer for coordination data that
// must remain coherent across gateway instances.
func (d *DualCache) DistributedGet(ctx context.Context, key string) (string, error) {
	return d.redis.Get(ctx, key).Result()
}

// DistributedSet bypasses the node-local layer for coordination data that
// must remain coherent across gateway instances.
func (d *DualCache) DistributedSet(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return d.redis.Set(ctx, key, value, ttl).Err()
}

// Publish sends an invalidation or coordination message through Redis Pub/Sub.
func (d *DualCache) Publish(ctx context.Context, channel string, message string) error {
	return d.redis.Publish(ctx, channel, message).Err()
}

// Subscribe returns only message payloads and keeps Redis details out of
// credential/runtime coordination packages. Callers must cancel ctx to close
// the subscription and output channel.
func (d *DualCache) Subscribe(ctx context.Context, channel string) (<-chan string, error) {
	subscription := d.redis.Subscribe(ctx, channel)
	if _, err := subscription.Receive(ctx); err != nil {
		return nil, errors.Join(err, subscription.Close())
	}

	out := make(chan string, 32)
	go func() {
		defer close(out)
		defer func() {
			if err := subscription.Close(); err != nil && d.logger != nil {
				d.logger.Warn("failed to close Redis subscription", zap.Error(err), zap.String("channel", channel))
			}
		}()

		messages := subscription.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case message, ok := <-messages:
				if !ok {
					return
				}

				select {
				case out <- message.Payload:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

// Delete removes a value from both memory and Redis.
func (d *DualCache) Delete(ctx context.Context, key string) {
	d.memory.Delete(ctx, key)
	d.redis.Del(ctx, key)
}

func isDistributedHotPathKey(key string) bool {
	key = strings.TrimSpace(key)

	return strings.HasPrefix(key, "deployments:tenant:") ||
		strings.HasPrefix(key, "wallet:tenant:") ||
		strings.HasPrefix(key, "tenant:host:")
}

// Increment increments a value in Redis, falling back to memory.
func (d *DualCache) Increment(ctx context.Context, key string, value int64, ttl time.Duration) (int64, error) {
	// Increment in Redis
	result, err := d.redis.IncrBy(ctx, key, value).Result()
	if err != nil {
		// Fallback to memory
		return d.memory.Increment(ctx, key, value, ttl)
	}

	// Sync to memory cache
	d.memory.Set(ctx, key, result, ttl)

	return result, nil
}

const fixedWindowRateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local windowMs = tonumber(ARGV[2])

local current = redis.call('GET', key)
if not current then
  redis.call('SET', key, 1, 'PX', windowMs)
  return {1, 1, windowMs}
end

current = tonumber(current)
local pttl = redis.call('PTTL', key)
if pttl < 0 then
  redis.call('PEXPIRE', key, windowMs)
  pttl = windowMs
end

if current >= limit then
  return {0, current, pttl}
end

current = redis.call('INCR', key)
pttl = redis.call('PTTL', key)
if pttl < 0 then
  redis.call('PEXPIRE', key, windowMs)
  pttl = windowMs
end

return {1, current, pttl}
`

const slidingWindowRateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local windowMs = tonumber(ARGV[2])
local nowMs = tonumber(ARGV[3])
local weight = tonumber(ARGV[4])
if weight < 1 then
  weight = 1
end

local windowStart = nowMs - windowMs
redis.call('ZREMRANGEBYSCORE', key, 0, windowStart)

local entries = redis.call('ZRANGE', key, 0, -1)
local total = 0
for _, entry in ipairs(entries) do
  local w = string.match(entry, ":(%d+):")
  if w then
    total = total + tonumber(w)
  else
    total = total + 1
  end
end

if (total + weight) > limit then
  local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
  if oldest and #oldest >= 2 then
    local oldestMs = tonumber(oldest[2])
    local resetIn = (oldestMs + windowMs) - nowMs
    if resetIn < 0 then
      resetIn = 0
    end
    return {0, total, resetIn}
  end
  return {0, total, windowMs}
end

local seqKey = key .. ':seq'
local seq = redis.call('INCR', seqKey)
local member = tostring(nowMs) .. ':' .. tostring(weight) .. ':' .. tostring(seq)
redis.call('ZADD', key, nowMs, member)
redis.call('PEXPIRE', key, windowMs * 2)
redis.call('PEXPIRE', seqKey, windowMs * 2)

return {1, total + weight, windowMs}
`

const combinedRateLimitScript = `
local nowMs = tonumber(ARGV[1])
local out = {}

for index, key in ipairs(KEYS) do
  local offset = 2 + ((index - 1) * 4)
  local algorithm = ARGV[offset]
  local limit = tonumber(ARGV[offset + 1])
  local windowMs = tonumber(ARGV[offset + 2])
  local weight = tonumber(ARGV[offset + 3])
  if weight < 1 then
    weight = 1
  end

  local allowed = 0
  local current = 0
  local resetIn = windowMs

  if algorithm == 'sliding' then
    local windowStart = nowMs - windowMs
    redis.call('ZREMRANGEBYSCORE', key, 0, windowStart)
    local entries = redis.call('ZRANGE', key, 0, -1)
    for _, entry in ipairs(entries) do
      local entryWeight = string.match(entry, ':(%d+):')
      if entryWeight then
        current = current + tonumber(entryWeight)
      else
        current = current + 1
      end
    end

    if (current + weight) <= limit then
      local seqKey = key .. ':seq'
      local seq = redis.call('INCR', seqKey)
      local member = tostring(nowMs) .. ':' .. tostring(weight) .. ':' .. tostring(seq)
      redis.call('ZADD', key, nowMs, member)
      redis.call('PEXPIRE', key, windowMs * 2)
      redis.call('PEXPIRE', seqKey, windowMs * 2)
      current = current + weight
      allowed = 1
    else
      local oldest = redis.call('ZRANGE', key, 0, 0, 'WITHSCORES')
      if oldest and #oldest >= 2 then
        resetIn = (tonumber(oldest[2]) + windowMs) - nowMs
        if resetIn < 0 then
          resetIn = 0
        end
      end
    end
  else
    current = tonumber(redis.call('GET', key) or '0')
    resetIn = redis.call('PTTL', key)
    if current < limit then
      current = redis.call('INCR', key)
      resetIn = redis.call('PTTL', key)
      if resetIn < 0 then
        redis.call('PEXPIRE', key, windowMs)
        resetIn = windowMs
      end
      allowed = 1
    end
  end

  table.insert(out, allowed)
  table.insert(out, current)
  table.insert(out, resetIn)
  if allowed == 0 then
    break
  end
end

return out
`

const queueTokenSpendDeltaScript = `
local deltaKey = KEYS[1]
local dirtySetKey = KEYS[2]
local delta = tonumber(ARGV[1])
local ttlMs = tonumber(ARGV[2])
local score = tonumber(ARGV[3])

if not delta or delta <= 0 then
  return 0
end

redis.call('INCRBYFLOAT', deltaKey, delta)
redis.call('PEXPIRE', deltaKey, ttlMs)
redis.call('ZADD', dirtySetKey, score, deltaKey)
redis.call('PEXPIRE', dirtySetKey, ttlMs)
return 1
`

const drainTokenSpendDeltasScript = `
local dirtySetKey = KEYS[1]
local batchSize = tonumber(ARGV[1])

if not batchSize or batchSize <= 0 then
  batchSize = 100
end

local keys = redis.call('ZRANGE', dirtySetKey, 0, batchSize - 1)
local out = {}
for _, key in ipairs(keys) do
  local value = redis.call('GET', key)
  redis.call('DEL', key)
  redis.call('ZREM', dirtySetKey, key)
  if value then
    table.insert(out, key)
    table.insert(out, value)
  end
end

return out
`

const queueKeyLastUsedScript = `
local lastUsedKey = KEYS[1]
local dirtySetKey = KEYS[2]
local usedAtUnix = tonumber(ARGV[1])
local ttlMs = tonumber(ARGV[2])

if not usedAtUnix or usedAtUnix <= 0 then
  return 0
end

local current = tonumber(redis.call('GET', lastUsedKey) or '0')
if usedAtUnix > current then
  redis.call('SET', lastUsedKey, usedAtUnix, 'PX', ttlMs)
  redis.call('ZADD', dirtySetKey, usedAtUnix, lastUsedKey)
  redis.call('PEXPIRE', dirtySetKey, ttlMs)
  return 1
end

return 0
`

const drainKeyLastUsedScript = `
local dirtySetKey = KEYS[1]
local batchSize = tonumber(ARGV[1])

if not batchSize or batchSize <= 0 then
  batchSize = 100
end

local keys = redis.call('ZRANGE', dirtySetKey, 0, batchSize - 1)
local out = {}
for _, key in ipairs(keys) do
  local value = redis.call('GET', key)
  redis.call('DEL', key)
  redis.call('ZREM', dirtySetKey, key)
  if value then
    table.insert(out, key)
    table.insert(out, value)
  end
end

return out
`

const seedBudgetWindowScript = `
local windowKey = KEYS[1]
local activeSetKey = KEYS[2]
local allowance = tonumber(ARGV[1])
local expiresAt = tonumber(ARGV[2])
local ttlMs = tonumber(ARGV[3])
local version = tonumber(ARGV[4])
local nowUnix = tonumber(ARGV[5])

if not allowance or allowance < 0 then
  return {0, 'invalid_allowance', 0, 0, 0, 0, 0}
end
if not ttlMs or ttlMs <= 0 then
  return {0, 'invalid_ttl', 0, 0, 0, 0, 0}
end

local exists = redis.call('EXISTS', windowKey)
if exists == 0 then
  redis.call('HSET', windowKey,
    'status', 'active',
    'allowance_irr', allowance,
    'reserved_irr', 0,
    'committed_irr', 0,
    'overage_irr', 0,
    'expires_at', expiresAt,
    'version', version,
    'updated_at', nowUnix)
  redis.call('PEXPIRE', windowKey, ttlMs)
else
  redis.call('HSET', windowKey, 'allowance_irr', allowance, 'expires_at', expiresAt, 'version', version, 'updated_at', nowUnix)
  redis.call('PEXPIRE', windowKey, ttlMs)
end

redis.call('ZADD', activeSetKey, expiresAt, windowKey)
redis.call('PEXPIRE', activeSetKey, ttlMs)

local reserved = tonumber(redis.call('HGET', windowKey, 'reserved_irr') or '0')
local committed = tonumber(redis.call('HGET', windowKey, 'committed_irr') or '0')
local overage = tonumber(redis.call('HGET', windowKey, 'overage_irr') or '0')
local remaining = allowance - reserved - committed
return {1, 'active', allowance, reserved, committed, remaining, overage}
`

const reserveBudgetScript = `
local windowKey = KEYS[1]
local reservationKey = KEYS[2]
local activeSetKey = KEYS[3]
local reservationID = ARGV[1]
local estimate = tonumber(ARGV[2])
local ttlMs = tonumber(ARGV[3])
local nowUnix = tonumber(ARGV[4])

if not estimate or estimate < 0 then
  estimate = 0
end
if not ttlMs or ttlMs <= 0 then
  ttlMs = 3600000
end

if redis.call('EXISTS', windowKey) == 0 then
  return {-1, 'window_unavailable', 0, 0, 0, 0, 0}
end

local status = redis.call('HGET', windowKey, 'status') or 'active'
local expiresAt = tonumber(redis.call('HGET', windowKey, 'expires_at') or '0')
if status ~= 'active' or (expiresAt > 0 and nowUnix > expiresAt) then
  return {-1, 'window_unavailable', 0, 0, 0, 0, 0}
end

local existingStatus = redis.call('HGET', reservationKey, 'status')
if existingStatus then
  local allowanceExisting = tonumber(redis.call('HGET', windowKey, 'allowance_irr') or '0')
  local reservedExisting = tonumber(redis.call('HGET', windowKey, 'reserved_irr') or '0')
  local committedExisting = tonumber(redis.call('HGET', windowKey, 'committed_irr') or '0')
  local overageExisting = tonumber(redis.call('HGET', windowKey, 'overage_irr') or '0')
  local remainingExisting = allowanceExisting - reservedExisting - committedExisting
  return {1, existingStatus, allowanceExisting, reservedExisting, committedExisting, remainingExisting, overageExisting}
end

local allowance = tonumber(redis.call('HGET', windowKey, 'allowance_irr') or '0')
local reserved = tonumber(redis.call('HGET', windowKey, 'reserved_irr') or '0')
local committed = tonumber(redis.call('HGET', windowKey, 'committed_irr') or '0')
local overage = tonumber(redis.call('HGET', windowKey, 'overage_irr') or '0')
-- A window with no headroom left refuses even a zero estimate.
local headroom = allowance - reserved - committed
if headroom <= 0 or estimate > headroom then
  return {0, 'exhausted', allowance, reserved, committed, headroom, overage}
end

reserved = tonumber(redis.call('HINCRBYFLOAT', windowKey, 'reserved_irr', estimate))
redis.call('HSET', reservationKey,
  'reservation_id', reservationID,
  'status', 'reserved',
  'estimated_irr', estimate,
  'actual_irr', 0,
  'created_at', nowUnix,
  'updated_at', nowUnix)
redis.call('HSET', windowKey, 'updated_at', nowUnix)
redis.call('PEXPIRE', reservationKey, ttlMs)
redis.call('PEXPIRE', windowKey, ttlMs)
redis.call('ZADD', activeSetKey, expiresAt, windowKey)
redis.call('PEXPIRE', activeSetKey, ttlMs)

local remaining = allowance - reserved - committed
return {1, 'reserved', allowance, reserved, committed, remaining, overage}
`

const commitBudgetScript = `
local windowKey = KEYS[1]
local reservationKey = KEYS[2]
local activeSetKey = KEYS[3]
local actual = tonumber(ARGV[1])
local nowUnix = tonumber(ARGV[2])
local force = ARGV[3] == '1'

if not actual or actual < 0 then
  actual = 0
end

if redis.call('EXISTS', windowKey) == 0 then
  return {-1, 'window_unavailable', 0, 0, 0, 0, 0}
end
if redis.call('EXISTS', reservationKey) == 0 then
  return {-1, 'reservation_unavailable', 0, 0, 0, 0, 0}
end

local currentStatus = redis.call('HGET', reservationKey, 'status')
local allowance = tonumber(redis.call('HGET', windowKey, 'allowance_irr') or '0')
local reserved = tonumber(redis.call('HGET', windowKey, 'reserved_irr') or '0')
local committed = tonumber(redis.call('HGET', windowKey, 'committed_irr') or '0')
local overage = tonumber(redis.call('HGET', windowKey, 'overage_irr') or '0')

if currentStatus == 'committed' then
  local remainingExisting = allowance - reserved - committed
  return {1, 'committed', allowance, reserved, committed, remainingExisting, overage}
end
if currentStatus == 'released' then
  local remainingReleased = allowance - reserved - committed
  return {1, 'released', allowance, reserved, committed, remainingReleased, overage}
end

local estimated = tonumber(redis.call('HGET', reservationKey, 'estimated_irr') or '0')
local projectedReserved = reserved - estimated
local projectedCommitted = committed + actual
if (projectedReserved + projectedCommitted) > allowance and not force then
  local remainingBefore = allowance - reserved - committed
  return {0, 'exhausted', allowance, reserved, committed, remainingBefore, overage}
end
reserved = tonumber(redis.call('HINCRBYFLOAT', windowKey, 'reserved_irr', -estimated))
committed = tonumber(redis.call('HINCRBYFLOAT', windowKey, 'committed_irr', actual))
if committed > allowance then
  overage = committed - allowance
  redis.call('HSET', windowKey, 'overage_irr', overage)
end

redis.call('HSET', reservationKey, 'status', 'committed', 'actual_irr', actual, 'updated_at', nowUnix)
redis.call('HSET', windowKey, 'updated_at', nowUnix)
local expiresAt = tonumber(redis.call('HGET', windowKey, 'expires_at') or '0')
local ttlMs = redis.call('PTTL', windowKey)
if ttlMs and ttlMs > 0 and expiresAt > 0 then
  redis.call('ZADD', activeSetKey, expiresAt, windowKey)
  redis.call('PEXPIRE', activeSetKey, ttlMs)
end
local remaining = allowance - reserved - committed
return {1, 'committed', allowance, reserved, committed, remaining, overage}
`

const recordBudgetSpendScript = `
local windowKey = KEYS[1]
local activeSetKey = KEYS[2]
local amount = tonumber(ARGV[1])
local nowUnix = tonumber(ARGV[2])

if not amount or amount < 0 then
  amount = 0
end

if redis.call('EXISTS', windowKey) == 0 then
  return {-1, 'window_unavailable', 0, 0, 0, 0, 0}
end

local allowance = tonumber(redis.call('HGET', windowKey, 'allowance_irr') or '0')
local reserved = tonumber(redis.call('HGET', windowKey, 'reserved_irr') or '0')
local overage = tonumber(redis.call('HGET', windowKey, 'overage_irr') or '0')
local committed = tonumber(redis.call('HINCRBYFLOAT', windowKey, 'committed_irr', amount))
if committed > allowance then
  overage = committed - allowance
  redis.call('HSET', windowKey, 'overage_irr', overage)
end

redis.call('HSET', windowKey, 'updated_at', nowUnix)
local expiresAt = tonumber(redis.call('HGET', windowKey, 'expires_at') or '0')
local ttlMs = redis.call('PTTL', windowKey)
if ttlMs and ttlMs > 0 and expiresAt > 0 then
  redis.call('ZADD', activeSetKey, expiresAt, windowKey)
  redis.call('PEXPIRE', activeSetKey, ttlMs)
end
local remaining = allowance - reserved - committed
return {1, 'committed', allowance, reserved, committed, remaining, overage}
`

const releaseBudgetScript = `
local windowKey = KEYS[1]
local reservationKey = KEYS[2]
local nowUnix = tonumber(ARGV[1])

if redis.call('EXISTS', windowKey) == 0 then
  return {-1, 'window_unavailable', 0, 0, 0, 0, 0}
end
if redis.call('EXISTS', reservationKey) == 0 then
  return {1, 'missing', tonumber(redis.call('HGET', windowKey, 'allowance_irr') or '0'), tonumber(redis.call('HGET', windowKey, 'reserved_irr') or '0'), tonumber(redis.call('HGET', windowKey, 'committed_irr') or '0'), 0, tonumber(redis.call('HGET', windowKey, 'overage_irr') or '0')}
end

local status = redis.call('HGET', reservationKey, 'status')
local allowance = tonumber(redis.call('HGET', windowKey, 'allowance_irr') or '0')
local reserved = tonumber(redis.call('HGET', windowKey, 'reserved_irr') or '0')
local committed = tonumber(redis.call('HGET', windowKey, 'committed_irr') or '0')
local overage = tonumber(redis.call('HGET', windowKey, 'overage_irr') or '0')

if status == 'reserved' then
  local estimated = tonumber(redis.call('HGET', reservationKey, 'estimated_irr') or '0')
  reserved = tonumber(redis.call('HINCRBYFLOAT', windowKey, 'reserved_irr', -estimated))
  redis.call('HSET', reservationKey, 'status', 'released', 'updated_at', nowUnix)
elseif status ~= 'released' then
  redis.call('HSET', reservationKey, 'updated_at', nowUnix)
end

redis.call('HSET', windowKey, 'updated_at', nowUnix)

local remaining = allowance - reserved - committed
return {1, 'released', allowance, reserved, committed, remaining, overage}
`

const activeBudgetWindowsKey = "budget:active_windows"

// CheckRateLimit atomically checks a fixed-window rate-limit counter in Redis.
func (d *DualCache) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
	if limit <= 0 {
		return &RateLimitResult{
			Allowed: true,
			Current: 0,
			Limit:   limit,
			ResetAt: time.Now().Add(window),
		}, nil
	}

	windowMS := window.Milliseconds()

	result, err := d.redis.Eval(ctx, fixedWindowRateLimitScript, []string{key}, limit, windowMS).Int64Slice()
	if err != nil {
		return d.memory.CheckRateLimit(ctx, key, limit, window)
	}

	if len(result) < 3 {
		return nil, fmt.Errorf("unexpected redis rate-limit result length: %d", len(result))
	}

	resetIn := time.Duration(result[2]) * time.Millisecond

	return &RateLimitResult{
		Allowed: result[0] == 1,
		Current: result[1],
		Limit:   limit,
		ResetAt: time.Now().Add(resetIn),
	}, nil
}

// QueueTokenSpendDelta atomically accumulates one token spend delta and marks it dirty.
func (d *DualCache) QueueTokenSpendDelta(ctx context.Context, deltaKey string, dirtySetKey string, deltaUSD float64, ttl time.Duration, score int64) error {
	if deltaKey == "" || dirtySetKey == "" {
		return fmt.Errorf("token spend delta key and dirty set key are required")
	}

	if ttl <= 0 {
		return fmt.Errorf("token spend delta ttl must be positive")
	}

	return d.redis.Eval(ctx, queueTokenSpendDeltaScript, []string{deltaKey, dirtySetKey}, deltaUSD, ttl.Milliseconds(), score).Err()
}

// QueueTokenSpendDeltas pipelines token-spend deltas into one Redis round trip.
func (d *DualCache) QueueTokenSpendDeltas(ctx context.Context, requests []TokenSpendDeltaRequest) error {
	if len(requests) == 0 {
		return nil
	}

	for _, request := range requests {
		if request.DeltaKey == "" || request.DirtySetKey == "" {
			return fmt.Errorf("token spend delta key and dirty set key are required")
		}

		if request.TTL <= 0 {
			return fmt.Errorf("token spend delta ttl must be positive")
		}
	}

	_, err := d.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, request := range requests {
			pipe.Eval(
				ctx,
				queueTokenSpendDeltaScript,
				[]string{request.DeltaKey, request.DirtySetKey},
				request.DeltaUSD,
				request.TTL.Milliseconds(),
				request.Score,
			)
		}

		return nil
	})

	return err
}

// DrainTokenSpendDeltas atomically drains a bounded batch of token spend deltas.
func (d *DualCache) DrainTokenSpendDeltas(ctx context.Context, dirtySetKey string, batchSize int) ([]TokenSpendDelta, error) {
	values, err := drainFlatPairs(ctx, d.redis, drainTokenSpendDeltasScript, dirtySetKey, batchSize)
	if err != nil {
		return nil, err
	}

	out := make([]TokenSpendDelta, 0, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		key, err := redisString(values[i])
		if err != nil {
			return nil, err
		}

		delta, err := redisFloat(values[i+1])
		if err != nil {
			return nil, err
		}

		out = append(out, TokenSpendDelta{Key: key, DeltaUSD: delta})
	}

	return out, nil
}

// QueueKeyLastUsed records the newest last-used timestamp for a key.
func (d *DualCache) QueueKeyLastUsed(ctx context.Context, lastUsedKey string, dirtySetKey string, usedAtUnix int64, ttl time.Duration) (bool, error) {
	if lastUsedKey == "" || dirtySetKey == "" {
		return false, fmt.Errorf("last-used key and dirty set key are required")
	}

	if ttl <= 0 {
		return false, fmt.Errorf("last-used ttl must be positive")
	}

	updated, err := d.redis.Eval(ctx, queueKeyLastUsedScript, []string{lastUsedKey, dirtySetKey}, usedAtUnix, ttl.Milliseconds()).Int64()
	if err != nil {
		return false, err
	}

	return updated == 1, nil
}

// SeedBudgetWindow creates or refreshes a tenant spend allowance window.
func (d *DualCache) SeedBudgetWindow(ctx context.Context, seed BudgetWindowSeed) (*BudgetWindowResult, error) {
	if seed.WindowKey == "" {
		return nil, fmt.Errorf("budget window key is required")
	}

	if seed.TTL <= 0 {
		return nil, fmt.Errorf("budget window ttl must be positive")
	}

	now := time.Now().UTC()

	values, err := d.redis.Eval(ctx, seedBudgetWindowScript, []string{seed.WindowKey, activeBudgetWindowsKey}, seed.AllowanceIRR, seed.ExpiresAt.Unix(), seed.TTL.Milliseconds(), seed.Version, now.Unix()).Result()
	if err != nil {
		return nil, err
	}

	return parseBudgetWindowResult(values)
}

// ReserveBudget atomically reserves estimated spend from a tenant allowance window.
func (d *DualCache) ReserveBudget(ctx context.Context, req BudgetReservationRequest) (*BudgetWindowResult, error) {
	if req.WindowKey == "" || req.ReservationKey == "" {
		return nil, fmt.Errorf("budget window and reservation keys are required")
	}

	if req.TTL <= 0 {
		return nil, fmt.Errorf("budget reservation ttl must be positive")
	}

	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	values, err := d.redis.Eval(ctx, reserveBudgetScript, []string{req.WindowKey, req.ReservationKey, activeBudgetWindowsKey}, req.ReservationID, req.EstimatedIRR, req.TTL.Milliseconds(), now.Unix()).Result()
	if err != nil {
		return nil, err
	}

	return parseBudgetWindowResult(values)
}

// CommitBudget atomically converts a reservation into committed spend.
func (d *DualCache) CommitBudget(ctx context.Context, req BudgetCommitRequest) (*BudgetWindowResult, error) {
	if req.WindowKey == "" || req.ReservationKey == "" {
		return nil, fmt.Errorf("budget window and reservation keys are required")
	}

	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	force := "0"
	if req.Force {
		force = "1"
	}

	values, err := d.redis.Eval(ctx, commitBudgetScript, []string{req.WindowKey, req.ReservationKey, activeBudgetWindowsKey}, req.ActualIRR, now.Unix(), force).Result()
	if err != nil {
		return nil, err
	}

	return parseBudgetWindowResult(values)
}

// RecordBudgetSpend atomically adds spend to a window without a reservation or
// an admission check, for usage that already happened.
func (d *DualCache) RecordBudgetSpend(ctx context.Context, req BudgetSpendRequest) (*BudgetWindowResult, error) {
	if req.WindowKey == "" {
		return nil, fmt.Errorf("budget window key is required")
	}

	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	values, err := d.redis.Eval(ctx, recordBudgetSpendScript, []string{req.WindowKey, activeBudgetWindowsKey}, req.Amount, now.Unix()).Result()
	if err != nil {
		return nil, err
	}

	return parseBudgetWindowResult(values)
}

// ReleaseBudget atomically releases an uncommitted reservation.
func (d *DualCache) ReleaseBudget(ctx context.Context, req BudgetReleaseRequest) (*BudgetWindowResult, error) {
	if req.WindowKey == "" || req.ReservationKey == "" {
		return nil, fmt.Errorf("budget window and reservation keys are required")
	}

	now := req.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	values, err := d.redis.Eval(ctx, releaseBudgetScript, []string{req.WindowKey, req.ReservationKey}, now.Unix()).Result()
	if err != nil {
		return nil, err
	}

	return parseBudgetWindowResult(values)
}

// DrainKeyLastUsed atomically drains newest last-used timestamps for a bounded batch of keys.
func (d *DualCache) DrainKeyLastUsed(ctx context.Context, dirtySetKey string, batchSize int) ([]KeyLastUsedSnapshot, error) {
	values, err := drainFlatPairs(ctx, d.redis, drainKeyLastUsedScript, dirtySetKey, batchSize)
	if err != nil {
		return nil, err
	}

	out := make([]KeyLastUsedSnapshot, 0, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		key, err := redisString(values[i])
		if err != nil {
			return nil, err
		}

		usedAt, err := redisInt(values[i+1])
		if err != nil {
			return nil, err
		}

		out = append(out, KeyLastUsedSnapshot{Key: key, UsedAtUnixMS: usedAt})
	}

	return out, nil
}

func parseBudgetWindowResult(result interface{}) (*BudgetWindowResult, error) {
	values, ok := result.([]interface{})
	if !ok || len(values) < 7 {
		return nil, fmt.Errorf("unexpected redis budget result: %T", result)
	}

	allowed, err := redisInt(values[0])
	if err != nil {
		return nil, err
	}

	state, err := redisString(values[1])
	if err != nil {
		return nil, err
	}

	allowance, err := redisFloat(values[2])
	if err != nil {
		return nil, err
	}

	reserved, err := redisFloat(values[3])
	if err != nil {
		return nil, err
	}

	committed, err := redisFloat(values[4])
	if err != nil {
		return nil, err
	}

	remaining, err := redisFloat(values[5])
	if err != nil {
		return nil, err
	}

	overage, err := redisFloat(values[6])
	if err != nil {
		return nil, err
	}

	return &BudgetWindowResult{
		Allowed:      allowed == 1,
		State:        state,
		AllowanceIRR: allowance,
		ReservedIRR:  reserved,
		CommittedIRR: committed,
		RemainingIRR: remaining,
		OverageIRR:   overage,
	}, nil
}

// GetBudgetWindow returns the current state of one active budget window from Redis.
func (d *DualCache) GetBudgetWindow(ctx context.Context, windowKey string) (*BudgetWindowSnapshot, error) {
	if strings.TrimSpace(windowKey) == "" {
		return nil, fmt.Errorf("budget window key is required")
	}

	values, err := d.redis.HGetAll(ctx, windowKey).Result()
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, nil
	}

	return parseBudgetWindowSnapshot(windowKey, values)
}

// ListActiveBudgetWindows returns a bounded list of active budget windows from Redis.
func (d *DualCache) ListActiveBudgetWindows(ctx context.Context, batchSize int) ([]BudgetWindowSnapshot, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	nowUnix := time.Now().UTC().Unix()
	if err := d.redis.ZRemRangeByScore(ctx, activeBudgetWindowsKey, "-inf", strconv.FormatInt(nowUnix, 10)).Err(); err != nil {
		return nil, err
	}

	keys, err := d.redis.ZRange(ctx, activeBudgetWindowsKey, 0, int64(batchSize-1)).Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, nil
	}

	pipe := d.redis.Pipeline()

	cmds := make([]*redis.MapStringStringCmd, 0, len(keys))
	for _, key := range keys {
		cmds = append(cmds, pipe.HGetAll(ctx, key))
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	out := make([]BudgetWindowSnapshot, 0, len(keys))

	for i, cmd := range cmds {
		values, cmdErr := cmd.Result()
		if cmdErr == redis.Nil || len(values) == 0 {
			if _, removeErr := d.redis.ZRem(ctx, activeBudgetWindowsKey, keys[i]).Result(); removeErr != nil && d.logger != nil {
				d.logger.Warn("failed to remove inactive budget window", zap.Error(removeErr), zap.String("window_key", keys[i]))
			}

			continue
		}

		if cmdErr != nil {
			return nil, cmdErr
		}

		snapshot, parseErr := parseBudgetWindowSnapshot(keys[i], values)
		if parseErr != nil {
			return nil, parseErr
		}

		out = append(out, *snapshot)
	}

	return out, nil
}

func parseBudgetWindowSnapshot(windowKey string, values map[string]string) (*BudgetWindowSnapshot, error) {
	allowance, err := strconv.ParseFloat(values["allowance_irr"], 64)
	if err != nil {
		return nil, err
	}

	reserved, err := strconv.ParseFloat(values["reserved_irr"], 64)
	if err != nil {
		return nil, err
	}

	committed, err := strconv.ParseFloat(values["committed_irr"], 64)
	if err != nil {
		return nil, err
	}

	overage, err := strconv.ParseFloat(values["overage_irr"], 64)
	if err != nil {
		return nil, err
	}

	expiresUnix, err := strconv.ParseInt(values["expires_at"], 10, 64)
	if err != nil {
		return nil, err
	}

	version, err := strconv.ParseInt(values["version"], 10, 64)
	if err != nil {
		return nil, err
	}

	updatedAt := time.Time{}

	if rawUpdated := strings.TrimSpace(values["updated_at"]); rawUpdated != "" {
		updatedUnix, parseUpdatedErr := strconv.ParseInt(rawUpdated, 10, 64)
		if parseUpdatedErr != nil {
			return nil, parseUpdatedErr
		}

		updatedAt = time.Unix(updatedUnix, 0).UTC()
	}

	return &BudgetWindowSnapshot{
		WindowKey:     windowKey,
		Status:        strings.TrimSpace(values["status"]),
		AllowanceIRR:  allowance,
		ReservedIRR:   reserved,
		CommittedIRR:  committed,
		RemainingIRR:  allowance - reserved - committed,
		OverageIRR:    overage,
		ExpiresAt:     time.Unix(expiresUnix, 0).UTC(),
		Version:       version,
		LastUpdatedAt: updatedAt,
	}, nil
}

func drainFlatPairs(ctx context.Context, rdb redis.UniversalClient, script string, dirtySetKey string, batchSize int) ([]interface{}, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	if dirtySetKey == "" {
		return nil, fmt.Errorf("dirty set key is required")
	}

	result, err := rdb.Eval(ctx, script, []string{dirtySetKey}, batchSize).Result()
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	values, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected redis drain result: %T", result)
	}

	return values, nil
}

func redisInt(value interface{}) (int64, error) {
	switch typed := value.(type) {
	case int64:
		return typed, nil
	case string:
		return strconv.ParseInt(typed, 10, 64)
	case []byte:
		return strconv.ParseInt(string(typed), 10, 64)
	default:
		return 0, fmt.Errorf("unexpected redis integer type %T", value)
	}
}

func redisString(value interface{}) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []byte:
		return string(typed), nil
	default:
		return "", fmt.Errorf("unexpected redis string type %T", value)
	}
}

func redisFloat(value interface{}) (float64, error) {
	switch typed := value.(type) {
	case float64:
		return typed, nil
	case int64:
		return float64(typed), nil
	case string:
		return strconv.ParseFloat(typed, 64)
	case []byte:
		return strconv.ParseFloat(string(typed), 64)
	default:
		return 0, fmt.Errorf("unexpected redis float type %T", value)
	}
}

// CheckSlidingWindowRateLimit checks a true sliding-window rate-limit counter in Redis.
func (d *DualCache) CheckSlidingWindowRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
	return d.CheckSlidingWindowRateLimitWithWeight(ctx, key, limit, 1, window)
}

// CheckSlidingWindowRateLimitWithWeight checks a true sliding-window rate limit with weighted entries.
func (d *DualCache) CheckSlidingWindowRateLimitWithWeight(ctx context.Context, key string, limit, weight int64, window time.Duration) (*RateLimitResult, error) {
	if limit <= 0 {
		return &RateLimitResult{
			Allowed: true,
			Current: 0,
			Limit:   limit,
			ResetAt: time.Now().Add(window),
		}, nil
	}

	windowMS := window.Milliseconds()
	nowMS := time.Now().UnixMilli()

	result, err := d.redis.Eval(ctx, slidingWindowRateLimitScript, []string{key}, limit, windowMS, nowMS, weight).Int64Slice()
	if err != nil {
		return d.memory.CheckSlidingWindowRateLimitWithWeight(ctx, key, limit, weight, window)
	}

	if len(result) < 3 {
		return nil, fmt.Errorf("unexpected redis sliding-window result length: %d", len(result))
	}

	resetIn := time.Duration(result[2]) * time.Millisecond

	return &RateLimitResult{
		Allowed: result[0] == 1,
		Current: result[1],
		Limit:   limit,
		ResetAt: time.Now().Add(resetIn),
	}, nil
}

// CheckRateLimits evaluates ordered fixed or sliding counters in one Redis script.
func (d *DualCache) CheckRateLimits(ctx context.Context, checks []RateLimitCheck) ([]*RateLimitResult, error) {
	if len(checks) == 0 {
		return nil, nil
	}

	keys := make([]string, 0, len(checks))
	args := make([]interface{}, 0, 1+len(checks)*4)
	args = append(args, time.Now().UnixMilli())

	for _, check := range checks {
		if check.Key == "" || check.Limit <= 0 || check.Window <= 0 {
			return nil, fmt.Errorf("rate-limit batch check is invalid")
		}

		algorithm := "fixed"
		if check.Sliding {
			algorithm = "sliding"
		}

		keys = append(keys, check.Key)
		args = append(args, algorithm, check.Limit, check.Window.Milliseconds(), maxInt64(check.Weight, 1))
	}

	values, err := d.redis.Eval(ctx, combinedRateLimitScript, keys, args...).Int64Slice()
	if err != nil {
		return d.checkRateLimitsInMemory(ctx, checks)
	}

	if len(values)%3 != 0 || len(values) == 0 {
		return nil, fmt.Errorf("unexpected redis combined rate-limit result length: %d", len(values))
	}

	now := time.Now()

	results := make([]*RateLimitResult, 0, len(values)/3)
	for index := 0; index < len(values); index += 3 {
		check := checks[index/3]
		results = append(results, &RateLimitResult{
			Allowed: values[index] == 1,
			Current: values[index+1],
			Limit:   check.Limit,
			ResetAt: now.Add(time.Duration(values[index+2]) * time.Millisecond),
		})
	}

	return results, nil
}

func (d *DualCache) checkRateLimitsInMemory(ctx context.Context, checks []RateLimitCheck) ([]*RateLimitResult, error) {
	results := make([]*RateLimitResult, 0, len(checks))
	for _, check := range checks {
		var (
			result *RateLimitResult
			err    error
		)

		if check.Sliding {
			result, err = d.memory.CheckSlidingWindowRateLimitWithWeight(
				ctx,
				check.Key,
				check.Limit,
				check.Weight,
				check.Window,
			)
		} else {
			result, err = d.memory.CheckRateLimit(ctx, check.Key, check.Limit, check.Window)
		}

		if err != nil {
			return results, err
		}

		results = append(results, result)

		if result != nil && !result.Allowed {
			break
		}
	}

	return results, nil
}

func maxInt64(value, minimum int64) int64 {
	if value < minimum {
		return minimum
	}

	return value
}

// Close closes both memory and Redis caches.
func (d *DualCache) Close() error {
	if err := d.memory.Close(); err != nil {
		return err
	}

	return d.redis.Close()
}
