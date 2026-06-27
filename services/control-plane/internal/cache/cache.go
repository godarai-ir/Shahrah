package cache

import (
	"context"
	"fmt"
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

// Delete removes a value from both memory and Redis.
func (d *DualCache) Delete(ctx context.Context, key string) {
	d.memory.Delete(ctx, key)
	d.redis.Del(ctx, key)
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

// CheckSlidingWindowRateLimit checks a true sliding-window rate-limit counter in Redis.
func (d *DualCache) CheckSlidingWindowRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimitResult, error) {
	return d.CheckSlidingWindowRateLimitWithWeight(ctx, key, limit, 1, window)
}

// CheckSlidingWindowRateLimitWithWeight checks a true sliding-window rate limit with weighted entries.
func (d *DualCache) CheckSlidingWindowRateLimitWithWeight(ctx context.Context, key string, limit int64, weight int64, window time.Duration) (*RateLimitResult, error) {
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
	d.memory.Close()
	return d.redis.Close()
}
