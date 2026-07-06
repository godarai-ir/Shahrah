// Package cache provides local and distributed cache primitives for the control plane.
package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitResult holds the result of a rate-limit check.
type RateLimitResult struct {
	Allowed bool
	Current int64
	Limit   int64
	ResetAt time.Time
}

// UsageCounterRequest describes an atomic fixed-window usage counter update.
type UsageCounterRequest struct {
	Key         string
	DirtySetKey string
	Limit       float64
	Increment   float64
	TTL         time.Duration
	WindowEnd   time.Time
}

// UsageCounterResult holds the result of an atomic usage counter update.
type UsageCounterResult struct {
	Allowed bool
	Used    float64
}

// UsageCounterSnapshot is a Redis-backed counter value selected for persistence.
type UsageCounterSnapshot struct {
	Key  string
	Used float64
}

// TokenSpendDelta is a drained Redis token spend delta ready for persistence.
type TokenSpendDelta struct {
	Key      string
	DeltaUSD float64
}

// KeyLastUsedSnapshot is the newest last-used timestamp drained for one key.
type KeyLastUsedSnapshot struct {
	Key          string
	UsedAtUnixMS int64
}

// UsageCounterCache exposes distributed usage-counter operations.
type UsageCounterCache interface {
	CheckUsageCounter(ctx context.Context, req UsageCounterRequest) (*UsageCounterResult, error)
	ListDirtyUsageCounters(ctx context.Context, dirtySetKey string, batchSize int) ([]UsageCounterSnapshot, error)
	AckDirtyUsageCounters(ctx context.Context, dirtySetKey string, keys []string)
}

// TokenUsageCache exposes Redis-backed token usage aggregation primitives.
type TokenUsageCache interface {
	QueueTokenSpendDelta(ctx context.Context, deltaKey string, dirtySetKey string, deltaUSD float64, ttl time.Duration, score int64) error
	DrainTokenSpendDeltas(ctx context.Context, dirtySetKey string, batchSize int) ([]TokenSpendDelta, error)
	QueueKeyLastUsed(ctx context.Context, lastUsedKey string, dirtySetKey string, usedAtUnix int64, ttl time.Duration) (bool, error)
	DrainKeyLastUsed(ctx context.Context, dirtySetKey string, batchSize int) ([]KeyLastUsedSnapshot, error)
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
	redis  *redis.Client
	logger *zap.Logger
}

// NewDualCache creates a new dual cache (memory + Redis).
func NewDualCache(redisHost string, redisPort int, redisPassword string, logger *zap.Logger) (*DualCache, error) {
	memory, err := NewMemoryCache(logger)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
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

const usageCounterScript = `
local counterKey = KEYS[1]
local dirtySetKey = KEYS[2]
local limit = tonumber(ARGV[1])
local increment = tonumber(ARGV[2])
local ttlMs = tonumber(ARGV[3])
local windowEndUnix = tonumber(ARGV[4])

if not increment or increment <= 0 then
  increment = 1
end

local current = tonumber(redis.call('GET', counterKey) or '0')
if limit and limit >= 0 and (current + increment) > limit then
  return {0, current}
end

local updated = tonumber(redis.call('INCRBYFLOAT', counterKey, increment))
local pttl = redis.call('PTTL', counterKey)
if pttl < 0 then
  redis.call('PEXPIRE', counterKey, ttlMs)
end
redis.call('ZADD', dirtySetKey, windowEndUnix, counterKey)
redis.call('PEXPIRE', dirtySetKey, ttlMs)

return {1, updated}
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

// CheckUsageCounter atomically increments and checks a fixed-window usage counter in Redis.
func (d *DualCache) CheckUsageCounter(ctx context.Context, req UsageCounterRequest) (*UsageCounterResult, error) {
	if req.Key == "" || req.DirtySetKey == "" {
		return nil, fmt.Errorf("usage counter key and dirty set key are required")
	}
	if req.TTL <= 0 {
		return nil, fmt.Errorf("usage counter ttl must be positive")
	}

	result, err := d.redis.Eval(ctx, usageCounterScript, []string{req.Key, req.DirtySetKey}, req.Limit, req.Increment, req.TTL.Milliseconds(), req.WindowEnd.Unix()).Result()
	if err != nil {
		return nil, err
	}

	values, ok := result.([]interface{})
	if !ok || len(values) < 2 {
		return nil, fmt.Errorf("unexpected redis usage counter result: %T", result)
	}

	allowed, err := redisInt(values[0])
	if err != nil {
		return nil, err
	}
	used, err := redisFloat(values[1])
	if err != nil {
		return nil, err
	}

	return &UsageCounterResult{
		Allowed: allowed == 1,
		Used:    used,
	}, nil
}

// ListDirtyUsageCounters returns a bounded batch of dirty Redis usage counters.
func (d *DualCache) ListDirtyUsageCounters(ctx context.Context, dirtySetKey string, batchSize int) ([]UsageCounterSnapshot, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	keys, err := d.redis.ZRange(ctx, dirtySetKey, 0, int64(batchSize-1)).Result()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}

	pipe := d.redis.Pipeline()
	cmds := make([]*redis.StringCmd, 0, len(keys))
	for _, key := range keys {
		cmds = append(cmds, pipe.Get(ctx, key))
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, err
	}

	snapshots := make([]UsageCounterSnapshot, 0, len(keys))
	for i, cmd := range cmds {
		value, err := cmd.Float64()
		if err == redis.Nil {
			snapshots = append(snapshots, UsageCounterSnapshot{Key: keys[i]})
			continue
		}
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, UsageCounterSnapshot{Key: keys[i], Used: value})
	}

	return snapshots, nil
}

// AckDirtyUsageCounters removes flushed usage counters from the dirty index.
func (d *DualCache) AckDirtyUsageCounters(ctx context.Context, dirtySetKey string, keys []string) {
	if len(keys) == 0 {
		return
	}
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		args[i] = key
	}
	d.redis.ZRem(ctx, dirtySetKey, args...)
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

func drainFlatPairs(ctx context.Context, rdb *redis.Client, script string, dirtySetKey string, batchSize int) ([]interface{}, error) {
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

// Close closes both memory and Redis caches.
func (d *DualCache) Close() error {
	if err := d.memory.Close(); err != nil {
		return err
	}

	return d.redis.Close()
}
