package store

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const publicRateWindow = time.Minute

var fixedWindowScript = `
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
local ttl = redis.call('PTTL', KEYS[1])
if current > tonumber(ARGV[1]) then
  return {0, ttl}
end
return {1, ttl}
`

// AllowPublicChatbotRequest enforces independent bot, IP, and session windows in Redis.
func (s *Store) AllowPublicChatbotRequest(
	ctx context.Context,
	publicID, clientIP, sessionID string,
) (bool, time.Duration, error) {
	policies := []struct {
		key   string
		limit int
	}{
		{key: "rag:limit:bot:" + publicID, limit: 300},
		{key: "rag:limit:ip:" + hashRateLimitValue(s.conversationKey, clientIP), limit: 60},
	}

	if strings.TrimSpace(sessionID) != "" {
		policies = append(policies, struct {
			key   string
			limit int
		}{key: "rag:limit:session:" + hashRateLimitValue(s.conversationKey, sessionID), limit: 30})
	}

	for _, policy := range policies {
		allowed, retryAfter, err := s.allowFixedWindow(ctx, policy.key, policy.limit, publicRateWindow)
		if err != nil {
			return false, 0, err
		}

		if !allowed {
			return false, retryAfter, nil
		}
	}

	return true, 0, nil
}

func (s *Store) allowFixedWindow(
	ctx context.Context,
	key string,
	limit int,
	window time.Duration,
) (bool, time.Duration, error) {
	result, err := s.redis.Eval(ctx, fixedWindowScript, []string{key}, limit, window.Milliseconds()).Slice()
	if err != nil {
		return false, 0, fmt.Errorf("enforce public chatbot rate limit: %w", err)
	}

	if len(result) != 2 {
		return false, 0, fmt.Errorf("invalid public chatbot rate limit response")
	}

	allowed, ok := result[0].(int64)
	if !ok {
		return false, 0, fmt.Errorf("invalid public chatbot rate limit decision")
	}

	ttlMilliseconds, ok := result[1].(int64)
	if !ok {
		return false, 0, fmt.Errorf("invalid public chatbot rate limit TTL")
	}

	return allowed == 1, max(time.Duration(ttlMilliseconds)*time.Millisecond, time.Second), nil
}

func hashRateLimitValue(secret []byte, value string) string {
	hash := hmac.New(sha256.New, secret)
	_, _ = hash.Write([]byte(strings.TrimSpace(value)))
	digest := hash.Sum(nil)

	return hex.EncodeToString(digest[:12])
}
