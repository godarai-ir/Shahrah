package billing

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
	"github.com/Delpak/shahrah/services/control-plane/internal/testcache"
	"go.uber.org/zap"
)

type batchRateLimitCache struct {
	cache.Cache
	calls  int
	checks []cache.RateLimitCheck
}

func (c *batchRateLimitCache) CheckRateLimits(
	_ context.Context,
	checks []cache.RateLimitCheck,
) ([]*cache.RateLimitResult, error) {
	c.calls++

	c.checks = append([]cache.RateLimitCheck(nil), checks...)

	results := make([]*cache.RateLimitResult, 0, len(checks))
	for _, check := range checks {
		results = append(results, &cache.RateLimitResult{
			Allowed: true,
			Current: check.Weight,
			Limit:   check.Limit,
			ResetAt: time.Now().Add(check.Window),
		})
	}

	return results, nil
}

func TestCheckRateLimitAllowsWhenRuntimeProfileHasNoRules(t *testing.T) {
	service := newRateLimitPolicyTestService(t, nil)

	result, err := service.CheckRateLimit(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, "gpt-test", 10)
	if err != nil {
		t.Fatalf("CheckRateLimit returned error: %v", err)
	}

	if result != nil {
		t.Fatalf("expected no rate-limit result without rules, got %+v", result)
	}
}

func TestCheckRateLimitEnforcesRPMPolicy(t *testing.T) {
	service := newRateLimitPolicyTestService(t, []platform.RateLimitRule{{
		RuleID:           "rpm",
		Enabled:          true,
		RetryAfterHeader: true,
		Limit:            map[string]interface{}{"rpm": 2},
	}})
	auth := &platform.AuthContext{TenantID: "tenant-1"}

	for i := 0; i < 2; i++ {
		if _, err := service.CheckRateLimit(context.Background(), auth, "gpt-test", 10); err != nil {
			t.Fatalf("request %d should be allowed: %v", i+1, err)
		}
	}

	result, err := service.CheckRateLimit(context.Background(), auth, "gpt-test", 10)
	if !errors.Is(err, ErrRateLimitExceeded) {
		t.Fatalf("expected ErrRateLimitExceeded, got %v", err)
	}

	if result == nil || result.Allowed || result.Limit != 2 || !result.RetryAfter {
		t.Fatalf("unexpected exhausted result: %+v", result)
	}
}

func TestCheckRateLimitCombinesRPMAndTPMInOneBackendCall(t *testing.T) {
	memory, err := testcache.New(t, zap.NewNop())
	if err != nil {
		t.Fatalf("create memory cache: %v", err)
	}

	combined := &batchRateLimitCache{Cache: memory}
	service := newRateLimitPolicyTestServiceWithCache(t, combined, []platform.RateLimitRule{{
		RuleID:  "combined",
		Enabled: true,
		Limit: map[string]interface{}{
			"rpm": float64(10),
			"tpm": float64(1000),
		},
	}})

	_, err = service.CheckRateLimit(
		t.Context(),
		&platform.AuthContext{TenantID: "tenant-1"},
		"gpt-test",
		25,
	)
	if err != nil {
		t.Fatalf("check combined rate limit: %v", err)
	}

	if combined.calls != 1 || len(combined.checks) != 2 {
		t.Fatalf("expected one backend call with two checks, got calls=%d checks=%#v", combined.calls, combined.checks)
	}

	firstHashTag := strings.SplitN(combined.checks[0].Key, "}:", 2)[0]
	secondHashTag := strings.SplitN(combined.checks[1].Key, "}:", 2)[0]

	if firstHashTag != secondHashTag {
		t.Fatalf("expected Redis hash-tag co-location, got %q and %q", combined.checks[0].Key, combined.checks[1].Key)
	}
}

func TestCompiledRateLimitRulesRefreshWhenVersionIsReused(t *testing.T) {
	limiter := &RateLimiter{policyCache: make(map[string]compiledRateLimitProfile)}
	first := &runtimeprofile.Profile{
		TenantID: "tenant-1",
		Version:  7,
		Checksum: "checksum-a",
		RateLimitRules: []platform.RateLimitRule{{
			RuleID:  "rpm",
			Enabled: true,
			Limit:   map[string]interface{}{"rpm": 10},
		}},
	}
	second := &runtimeprofile.Profile{
		TenantID: "tenant-1",
		Version:  7,
		Checksum: "checksum-b",
		RateLimitRules: []platform.RateLimitRule{{
			RuleID:  "rpm",
			Enabled: true,
			Limit:   map[string]interface{}{"rpm": 25},
		}},
	}

	firstRules := limiter.compiledRateLimitRules(first)
	secondRules := limiter.compiledRateLimitRules(second)

	if firstRules[0].policy.rpm != 10 || secondRules[0].policy.rpm != 25 {
		t.Fatalf("expected policy cache refresh, got first=%d second=%d", firstRules[0].policy.rpm, secondRules[0].policy.rpm)
	}
}

func TestParseRateLimitPolicySupportsTenantAdminRuleShape(t *testing.T) {
	tests := []struct {
		name       string
		limit      map[string]interface{}
		wantRPM    int64
		wantTPM    int64
		wantWindow time.Duration
	}{
		{
			name: "requests per minute",
			limit: map[string]interface{}{
				"dimension": "requests",
				"max":       float64(10),
				"window": map[string]interface{}{
					"value": float64(1),
					"unit":  "minute",
				},
			},
			wantRPM:    10,
			wantWindow: time.Minute,
		},
		{
			name: "tokens per hour",
			limit: map[string]interface{}{
				"dimension": "tokens",
				"max":       float64(5000),
				"window": map[string]interface{}{
					"value": float64(2),
					"unit":  "hour",
				},
			},
			wantTPM:    5000,
			wantWindow: 2 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := parseRateLimitPolicy(platform.RateLimitRule{RuleID: "admin-rule", Limit: tt.limit})
			if err != nil {
				t.Fatalf("parseRateLimitPolicy returned error: %v", err)
			}

			if policy.rpm != tt.wantRPM || policy.tpm != tt.wantTPM || policy.window != tt.wantWindow {
				t.Fatalf("unexpected policy: rpm=%d tpm=%d window=%s", policy.rpm, policy.tpm, policy.window)
			}
		})
	}
}

func TestCheckRateLimitEnforcesTPMPolicyWithEstimatedTokens(t *testing.T) {
	service := newRateLimitPolicyTestService(t, []platform.RateLimitRule{{
		RuleID:  "tpm",
		Enabled: true,
		Limit:   map[string]interface{}{"tpm": 100},
	}})
	auth := &platform.AuthContext{TenantID: "tenant-1"}

	if _, err := service.CheckRateLimit(context.Background(), auth, "gpt-test", 60); err != nil {
		t.Fatalf("first request should be allowed: %v", err)
	}

	result, err := service.CheckRateLimit(context.Background(), auth, "gpt-test", 50)
	if !errors.Is(err, ErrRateLimitExceeded) {
		t.Fatalf("expected ErrRateLimitExceeded, got %v", err)
	}

	if result == nil || result.Allowed || result.Current != 60 || result.Limit != 100 {
		t.Fatalf("unexpected TPM result: %+v", result)
	}
}

func TestCheckRateLimitSkipsNonMatchingPolicy(t *testing.T) {
	service := newRateLimitPolicyTestService(t, []platform.RateLimitRule{{
		RuleID:     "model-rpm",
		Enabled:    true,
		Conditions: map[string]interface{}{"model": "gpt-4"},
		Limit:      map[string]interface{}{"rpm": 1},
	}})
	auth := &platform.AuthContext{TenantID: "tenant-1"}

	result, err := service.CheckRateLimit(context.Background(), auth, "gpt-3.5", 10)
	if err != nil {
		t.Fatalf("nonmatching model should be allowed: %v", err)
	}

	if result != nil {
		t.Fatalf("expected no result for nonmatching policy, got %+v", result)
	}

	if _, err = service.CheckRateLimit(context.Background(), auth, "gpt-4", 10); err != nil {
		t.Fatalf("first matching request should be allowed: %v", err)
	}

	_, err = service.CheckRateLimit(context.Background(), auth, "gpt-4", 10)
	if !errors.Is(err, ErrRateLimitExceeded) {
		t.Fatalf("expected matching model to exhaust policy, got %v", err)
	}
}

func TestCheckRateLimitSkipsDisabledExpiredAndFuturePolicies(t *testing.T) {
	now := time.Now().UTC()
	service := newRateLimitPolicyTestService(t, []platform.RateLimitRule{
		{RuleID: "disabled", Enabled: false, Limit: map[string]interface{}{"rpm": 1}},
		{RuleID: "expired", Enabled: true, ExpiresAt: timePtr(now.Add(-time.Minute)), Limit: map[string]interface{}{"rpm": 1}},
		{RuleID: "future", Enabled: true, EffectiveFrom: timePtr(now.Add(time.Minute)), Limit: map[string]interface{}{"rpm": 1}},
	})

	for i := 0; i < 3; i++ {
		if result, err := service.CheckRateLimit(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, "gpt-test", 10); err != nil || result != nil {
			t.Fatalf("inactive policies should not apply, result=%+v err=%v", result, err)
		}
	}
}

func TestCheckRateLimitSoftLimitDoesNotBlock(t *testing.T) {
	service := newRateLimitPolicyTestService(t, []platform.RateLimitRule{{
		RuleID:    "soft",
		Enabled:   true,
		SoftLimit: true,
		Limit:     map[string]interface{}{"rpm": 1},
	}})
	auth := &platform.AuthContext{TenantID: "tenant-1"}

	if _, err := service.CheckRateLimit(context.Background(), auth, "gpt-test", 10); err != nil {
		t.Fatalf("first request should be allowed: %v", err)
	}

	result, err := service.CheckRateLimit(context.Background(), auth, "gpt-test", 10)
	if err != nil {
		t.Fatalf("soft limit should not block: %v", err)
	}

	if result == nil || !result.Allowed || result.Limit != 1 {
		t.Fatalf("unexpected soft-limit result: %+v", result)
	}
}

func TestCheckRateLimitMostRestrictiveResultWins(t *testing.T) {
	service := newRateLimitPolicyTestService(t, []platform.RateLimitRule{
		{RuleID: "loose", Enabled: true, Priority: 1, Limit: map[string]interface{}{"rpm": 10}},
		{RuleID: "strict", Enabled: true, Priority: 2, Limit: map[string]interface{}{"rpm": 2}},
	})
	auth := &platform.AuthContext{TenantID: "tenant-1"}

	result, err := service.CheckRateLimit(context.Background(), auth, "gpt-test", 10)
	if err != nil {
		t.Fatalf("first request should be allowed: %v", err)
	}

	if result == nil || result.Limit != 2 {
		t.Fatalf("expected strict policy result, got %+v", result)
	}
}

func TestCheckRateLimitCacheErrorFailsClosedByDefault(t *testing.T) {
	service := newRateLimitPolicyTestServiceWithCache(t, &errorRateLimitCache{}, []platform.RateLimitRule{{
		RuleID:  "rpm",
		Enabled: true,
		Limit:   map[string]interface{}{"rpm": 1},
	}})

	_, err := service.CheckRateLimit(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, "gpt-test", 10)
	if !errors.Is(err, ErrRateLimitExceeded) {
		t.Fatalf("expected fail-closed ErrRateLimitExceeded, got %v", err)
	}
}

func TestCheckRateLimitCacheErrorCanFailOpen(t *testing.T) {
	service := newRateLimitPolicyTestServiceWithCache(t, &errorRateLimitCache{}, []platform.RateLimitRule{{
		RuleID:  "rpm",
		Enabled: true,
		Limit:   map[string]interface{}{"rpm": 1},
	}})
	service.rateLimitFailOpen = true

	if _, err := service.CheckRateLimit(context.Background(), &platform.AuthContext{TenantID: "tenant-1"}, "gpt-test", 10); err != nil {
		t.Fatalf("expected fail-open request to be allowed, got %v", err)
	}
}

func newRateLimitPolicyTestService(t *testing.T, rules []platform.RateLimitRule) *Service {
	t.Helper()

	mem, err := testcache.New(t, zap.NewNop())
	if err != nil {
		t.Fatalf("NewMemoryCache returned error: %v", err)
	}

	return newRateLimitPolicyTestServiceWithCache(t, mem, rules)
}

func newRateLimitPolicyTestServiceWithCache(t *testing.T, c cache.Cache, rules []platform.RateLimitRule) *Service {
	t.Helper()

	service := NewServiceWithDependencies(Dependencies{
		Cache: c,
		RuntimeProfiles: staticRuntimeProfiles{profile: &runtimeprofile.Profile{
			TenantID:       "tenant-1",
			RateLimitRules: rules,
		}},
	}, 850000, "test")

	return service
}

func timePtr(value time.Time) *time.Time {
	return &value
}

type errorRateLimitCache struct{}

func (c *errorRateLimitCache) Get(context.Context, string) (interface{}, bool)         { return nil, false }
func (c *errorRateLimitCache) Set(context.Context, string, interface{}, time.Duration) {}
func (c *errorRateLimitCache) Delete(context.Context, string)                          {}
func (c *errorRateLimitCache) Increment(context.Context, string, int64, time.Duration) (int64, error) {
	return 0, errors.New("cache unavailable")
}
func (c *errorRateLimitCache) CheckRateLimit(context.Context, string, int64, time.Duration) (*cache.RateLimitResult, error) {
	return nil, errors.New("cache unavailable")
}
func (c *errorRateLimitCache) CheckSlidingWindowRateLimit(context.Context, string, int64, time.Duration) (*cache.RateLimitResult, error) {
	return nil, errors.New("cache unavailable")
}
func (c *errorRateLimitCache) CheckSlidingWindowRateLimitWithWeight(context.Context, string, int64, int64, time.Duration) (*cache.RateLimitResult, error) {
	return nil, errors.New("cache unavailable")
}
func (c *errorRateLimitCache) Close() error { return nil }
