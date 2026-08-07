package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/runtimeprofile"
)

const (
	rateLimitKindRPM = "rpm"
	rateLimitKindTPM = "tpm"
)

type rateLimitPolicy struct {
	rpm         int64
	tpm         int64
	window      time.Duration
	algorithm   string
	scopeFields []string
	retryAfter  bool
	softLimit   bool
	conflict    platform.RateLimitConflictBehavior
	response    platform.RateLimitResponseBehavior
	ruleID      string
	priority    int
}

type compiledRateLimitRule struct {
	rule     platform.RateLimitRule
	policy   rateLimitPolicy
	parseErr error
}

type compiledRateLimitProfile struct {
	version  int64
	checksum string
	profile  *runtimeprofile.Profile
	rules    []compiledRateLimitRule
}

func (s *RateLimiter) checkRateLimitPolicies(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) (*cache.RateLimitResult, error) {
	if s == nil || s.runtimeProfiles == nil || auth == nil || strings.TrimSpace(auth.TenantID) == "" {
		return nil, nil
	}

	profile, err := s.runtimeProfiles.GetSnapshot(ctx, auth.TenantID)
	if err != nil {
		return nil, fmt.Errorf("load runtime profile for rate limits: %w", err)
	}

	if profile == nil || len(profile.RateLimitRules) == 0 {
		return nil, nil
	}

	rules := s.compiledRateLimitRules(profile)

	now := time.Now().UTC()

	var result *cache.RateLimitResult

	for _, rule := range rules {
		policy, ok, err := matchCompiledRateLimitRule(rule, auth, breakdown, now)
		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		checked, err := s.applyRateLimitPolicy(ctx, auth, breakdown, policy)
		if err != nil {
			if s.rateLimitFailOpen || policy.softLimit {
				continue
			}

			return checked, ErrRateLimitExceeded
		}

		if checked != nil && !checked.Allowed && !policy.softLimit {
			return checked, ErrRateLimitExceeded
		}

		if checked != nil && !checked.Allowed && policy.softLimit {
			checked = &cache.RateLimitResult{
				Allowed:    true,
				Current:    checked.Current,
				Limit:      checked.Limit,
				ResetAt:    checked.ResetAt,
				RetryAfter: false,
			}
		}

		result = mostRestrictiveRateLimitResult(result, checked)
		if policy.conflict == platform.RateLimitConflictFirstMatch {
			return result, nil
		}
	}

	return result, nil
}

func (s *RateLimiter) compiledRateLimitRules(profile *runtimeprofile.Profile) []compiledRateLimitRule {
	tenantID := strings.TrimSpace(profile.TenantID)

	s.policyCacheMu.RLock()
	cached, ok := s.policyCache[tenantID]
	s.policyCacheMu.RUnlock()

	sameChecksum := profile.Checksum != "" && cached.checksum == profile.Checksum
	if ok && cached.version == profile.Version && (sameChecksum || cached.profile == profile) {
		return cached.rules
	}

	rules := make([]compiledRateLimitRule, 0, len(profile.RateLimitRules))
	for _, rule := range profile.RateLimitRules {
		compiled := compiledRateLimitRule{rule: rule}
		if rule.Enabled {
			compiled.policy, compiled.parseErr = parseRateLimitPolicy(rule)
			compiled.policy.scopeFields = rateLimitScopeFields(rule)
		}

		rules = append(rules, compiled)
	}

	s.policyCacheMu.Lock()
	if s.policyCache == nil {
		s.policyCache = make(map[string]compiledRateLimitProfile)
	}

	s.policyCache[tenantID] = compiledRateLimitProfile{
		version:  profile.Version,
		checksum: profile.Checksum,
		profile:  profile,
		rules:    rules,
	}
	s.policyCacheMu.Unlock()

	return rules
}

func matchCompiledRateLimitRule(rule compiledRateLimitRule, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, now time.Time) (rateLimitPolicy, bool, error) {
	source := rule.rule
	if !source.Enabled {
		return rateLimitPolicy{}, false, nil
	}

	if source.EffectiveFrom != nil && now.Before(source.EffectiveFrom.UTC()) {
		return rateLimitPolicy{}, false, nil
	}

	if source.ExpiresAt != nil && !now.Before(source.ExpiresAt.UTC()) {
		return rateLimitPolicy{}, false, nil
	}

	matches, err := rateLimitRuleMatches(source, auth, breakdown)
	if err != nil || !matches {
		return rateLimitPolicy{}, matches, err
	}

	if rule.parseErr != nil {
		return rateLimitPolicy{}, false, rule.parseErr
	}

	if rule.policy.rpm <= 0 && rule.policy.tpm <= 0 {
		return rateLimitPolicy{}, false, nil
	}

	return rule.policy, true, nil
}

func parseRateLimitPolicy(rule platform.RateLimitRule) (rateLimitPolicy, error) {
	rpm := int64FromAny(firstPresent(rule.Limit, "rpm", "requests_per_minute"), 0)
	tpm := int64FromAny(firstPresent(rule.Limit, "tpm", "tokens_per_minute"), 0)

	// The tenant-admin rule editor persists the canonical dimension/max shape.
	// Keep accepting the older rpm/tpm keys so existing API clients and stored
	// rules continue to work during the transition.
	if rawMax := firstPresent(rule.Limit, "max"); rawMax != nil {
		maxValue := int64FromAny(rawMax, 0)
		if maxValue <= 0 {
			return rateLimitPolicy{}, fmt.Errorf("rate-limit max must be positive")
		}

		switch strings.ToLower(strings.TrimSpace(asString(firstPresent(rule.Limit, "dimension")))) {
		case "requests", "request", "rpm":
			rpm = maxValue
		case "tokens", "token", "tpm":
			tpm = maxValue
		default:
			return rateLimitPolicy{}, fmt.Errorf("unsupported rate-limit dimension %q", asString(firstPresent(rule.Limit, "dimension")))
		}
	}

	window, err := parseRateLimitWindow(rule.Limit)
	if err != nil {
		return rateLimitPolicy{}, err
	}

	conflict := rule.ConflictBehavior
	if conflict == "" {
		conflict = platform.RateLimitConflictMostRestrictive
	}

	response := rule.ResponseBehavior
	if response == "" {
		response = platform.RateLimitResponseReturn429
	}

	return rateLimitPolicy{
		rpm:        rpm,
		tpm:        tpm,
		window:     window,
		algorithm:  normalizedRateLimitAlgorithm(asString(firstPresent(rule.Limit, "algorithm"))),
		retryAfter: rule.RetryAfterHeader,
		softLimit:  rule.SoftLimit,
		conflict:   conflict,
		response:   response,
		ruleID:     rateLimitRuleID(rule),
		priority:   rule.Priority,
	}, nil
}

func parseRateLimitWindow(limit map[string]interface{}) (time.Duration, error) {
	if rawSeconds := firstPresent(limit, "window_seconds"); rawSeconds != nil {
		seconds := int64FromAny(rawSeconds, 0)
		if seconds <= 0 {
			return 0, fmt.Errorf("rate-limit window_seconds must be positive")
		}

		return time.Duration(seconds) * time.Second, nil
	}

	rawWindow := firstPresent(limit, "window")
	if rawWindow == nil {
		return time.Minute, nil
	}

	windowConfig, ok := rawWindow.(map[string]interface{})
	if !ok {
		seconds := int64FromAny(rawWindow, 0)
		if seconds <= 0 {
			return 0, fmt.Errorf("rate-limit window must be positive")
		}

		return time.Duration(seconds) * time.Second, nil
	}

	value := int64FromAny(windowConfig["value"], 0)
	if value <= 0 {
		return 0, fmt.Errorf("rate-limit window value must be positive")
	}

	var unit time.Duration

	switch strings.ToLower(strings.TrimSpace(asString(windowConfig["unit"]))) {
	case "second", "seconds":
		unit = time.Second
	case "minute", "minutes":
		unit = time.Minute
	case "hour", "hours":
		unit = time.Hour
	case "day", "days":
		unit = 24 * time.Hour
	case "week", "weeks":
		unit = 7 * 24 * time.Hour
	case "month", "months":
		unit = 30 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("unsupported rate-limit window unit %q", asString(windowConfig["unit"]))
	}

	return time.Duration(value) * unit, nil
}

func (s *RateLimiter) applyRateLimitPolicy(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, policy rateLimitPolicy) (*cache.RateLimitResult, error) {
	if s.cache == nil {
		if s.rateLimitFailOpen {
			return nil, nil
		}

		return nil, ErrRateLimitExceeded
	}

	scopeHash := rateLimitScopeHash(auth, breakdown, policy.scopeFields)
	if policy.rpm > 0 && policy.tpm > 0 {
		if batchChecker, ok := s.cache.(cache.RateLimitBatchChecker); ok {
			return s.applyCombinedRateLimitPolicy(ctx, batchChecker, auth, breakdown, policy, scopeHash)
		}
	}

	var result *cache.RateLimitResult

	if policy.rpm > 0 {
		checked, err := s.applyRateLimitCounter(ctx, auth.TenantID, policy, rateLimitKindRPM, scopeHash, 1)
		if err != nil {
			return checked, err
		}

		result = mostRestrictiveRateLimitResult(result, checked)
		if checked != nil && !checked.Allowed && !policy.softLimit {
			return checked, nil
		}
	}

	if policy.tpm > 0 {
		weight := int64(breakdown.TotalTokens)
		if weight <= 0 {
			weight = 1
		}

		checked, err := s.applyRateLimitCounter(ctx, auth.TenantID, policy, rateLimitKindTPM, scopeHash, weight)
		if err != nil {
			return checked, err
		}

		result = mostRestrictiveRateLimitResult(result, checked)
	}

	return result, nil
}

func (s *RateLimiter) applyCombinedRateLimitPolicy(
	ctx context.Context,
	checker cache.RateLimitBatchChecker,
	auth *platform.AuthContext,
	breakdown platform.UsageCostBreakdown,
	policy rateLimitPolicy,
	scopeHash string,
) (*cache.RateLimitResult, error) {
	tokenWeight := int64(breakdown.TotalTokens)
	if tokenWeight <= 0 {
		tokenWeight = 1
	}

	checks := []cache.RateLimitCheck{
		{
			Key:     rateLimitCounterKey(auth.TenantID, policy.ruleID, scopeHash, rateLimitKindRPM),
			Limit:   policy.rpm,
			Weight:  1,
			Window:  policy.window,
			Sliding: policy.algorithm == "sliding_window",
		},
		{
			Key:     rateLimitCounterKey(auth.TenantID, policy.ruleID, scopeHash, rateLimitKindTPM),
			Limit:   policy.tpm,
			Weight:  tokenWeight,
			Window:  policy.window,
			Sliding: true,
		},
	}

	results, err := checker.CheckRateLimits(ctx, checks)
	if err != nil {
		return nil, err
	}

	var mostRestrictive *cache.RateLimitResult

	for index, result := range results {
		kind := rateLimitKindRPM
		if index == 1 {
			kind = rateLimitKindTPM
		}

		result = decoratePolicyRateLimitResult(result, policy, kind)
		mostRestrictive = mostRestrictiveRateLimitResult(mostRestrictive, result)

		if result != nil && !result.Allowed {
			return result, nil
		}
	}

	return mostRestrictive, nil
}

func (s *RateLimiter) applyRateLimitCounter(ctx context.Context, tenantID string, policy rateLimitPolicy, kind string, scopeHash string, weight int64) (*cache.RateLimitResult, error) {
	key := rateLimitCounterKey(tenantID, policy.ruleID, scopeHash, kind)
	if kind == rateLimitKindTPM || policy.algorithm == "sliding_window" {
		result, err := s.cache.CheckSlidingWindowRateLimitWithWeight(ctx, key, policyLimit(policy, kind), weight, policy.window)
		return decoratePolicyRateLimitResult(result, policy, kind), err
	}

	result, err := s.cache.CheckRateLimit(ctx, key, policyLimit(policy, kind), policy.window)

	return decoratePolicyRateLimitResult(result, policy, kind), err
}

func rateLimitCounterKey(tenantID, ruleID, scopeHash, kind string) string {
	hashTag := shortHash(strings.Join([]string{
		strings.TrimSpace(tenantID),
		strings.TrimSpace(ruleID),
		strings.TrimSpace(scopeHash),
	}, "|"))

	return fmt.Sprintf("ratelimit:{%s}:%s", hashTag, strings.TrimSpace(kind))
}

func decoratePolicyRateLimitResult(result *cache.RateLimitResult, policy rateLimitPolicy, kind string) *cache.RateLimitResult {
	if result == nil {
		return nil
	}

	out := *result
	out.RetryAfter = policy.retryAfter && !policy.softLimit && !out.Allowed
	out.RuleID = policy.ruleID
	out.Dimension = kind
	out.Algorithm = policy.algorithm

	return &out
}

func policyLimit(policy rateLimitPolicy, kind string) int64 {
	if kind == rateLimitKindTPM {
		return policy.tpm
	}

	return policy.rpm
}

func rateLimitRuleMatches(rule platform.RateLimitRule, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) (bool, error) {
	return budgetRuleMatches(platform.BudgetLimitRule{
		Conditions: rule.Conditions,
	}, auth, breakdown)
}

func rateLimitScopeFields(rule platform.RateLimitRule) []string {
	rawScope := firstPresent(rule.Limit, "scope", "scope_by", "applies_per")
	if rawScope != nil {
		values, err := stringSlice(rawScope)
		if err == nil && len(values) > 0 {
			return normalizeRateLimitScopeFields(values)
		}

		value := strings.TrimSpace(asString(rawScope))
		if value != "" {
			return normalizeRateLimitScopeFields([]string{value})
		}
	}

	fields := []string{"tenant_id"}

	for key := range rule.Conditions {
		key = strings.TrimSpace(key)
		if key == "" || key == "conditions" || key == "combinator" {
			continue
		}

		fields = append(fields, key)
	}

	if rawConditions, ok := rule.Conditions["conditions"].([]interface{}); ok {
		for _, raw := range rawConditions {
			condition, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}

			switch strings.TrimSpace(asString(condition["field"])) {
			case "model":
				fields = append(fields, "model")
			case "metadata":
				metadataKey := strings.TrimSpace(asString(condition["metadataKey"]))
				if metadataKey != "" {
					fields = append(fields, "metadata."+metadataKey)
				}
			case "subject":
				switch strings.TrimSpace(asString(condition["subjectKind"])) {
				case "team":
					fields = append(fields, "team_id")
				case "virtual_account":
					fields = append(fields, "key_id")
				case "organization_unit":
					fields = append(fields, "org_unit_id")
				default:
					fields = append(fields, "user_id")
				}
			}
		}
	}

	return normalizeRateLimitScopeFields(fields)
}

func normalizeRateLimitScopeFields(fields []string) []string {
	seen := map[string]bool{}

	out := make([]string, 0, len(fields)+1)
	for _, field := range append([]string{"tenant_id"}, fields...) {
		field = strings.TrimSpace(field)
		if field == "" || seen[field] {
			continue
		}

		seen[field] = true
		out = append(out, field)
	}

	return out
}

func rateLimitScopeHash(auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, fields []string) string {
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		value, _ := budgetContextValue(field, auth, breakdown)
		parts = append(parts, field+"="+strings.TrimSpace(value))
	}

	return shortHash(strings.Join(parts, "|"))
}

func rateLimitRuleID(rule platform.RateLimitRule) string {
	if strings.TrimSpace(rule.RuleID) != "" {
		return strings.TrimSpace(rule.RuleID)
	}

	return strings.TrimSpace(rule.ID)
}

func normalizedRateLimitAlgorithm(value string) string {
	switch strings.TrimSpace(value) {
	case "sliding", "sliding_window":
		return "sliding_window"
	default:
		return "fixed_window"
	}
}

func firstPresent(values map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if values == nil {
			return nil
		}

		if value, ok := values[key]; ok {
			return value
		}
	}

	return nil
}

func int64FromAny(value interface{}, fallback int64) int64 {
	if value == nil {
		return fallback
	}

	parsed, err := parseFloat(asString(value))
	if err != nil {
		return fallback
	}

	return int64(parsed)
}

func mostRestrictiveRateLimitResult(left, right *cache.RateLimitResult) *cache.RateLimitResult {
	if left == nil {
		return right
	}

	if right == nil {
		return left
	}

	leftRemaining := left.Limit - left.Current

	rightRemaining := right.Limit - right.Current
	if rightRemaining < leftRemaining {
		return right
	}

	return left
}
