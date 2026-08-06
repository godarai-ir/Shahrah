package billing

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

const (
	rateLimitKindRPM = "rpm"
	rateLimitKindTPM = "tpm"
)

type rateLimitPolicy struct {
	rpm           int64
	tpm           int64
	window        time.Duration
	algorithm     string
	scopeFields   []string
	retryAfter    bool
	softLimit     bool
	conflict      platform.RateLimitConflictBehavior
	response      platform.RateLimitResponseBehavior
	ruleID        string
	priority      int
	conditionsKey string
}

func (s *RateLimiter) checkRateLimitPolicies(ctx context.Context, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) (*cache.RateLimitResult, error) {
	if s == nil || s.runtimeProfiles == nil || auth == nil || strings.TrimSpace(auth.TenantID) == "" {
		return nil, nil
	}

	profile, err := s.runtimeProfiles.Get(ctx, auth.TenantID)
	if err != nil {
		return nil, fmt.Errorf("load runtime profile for rate limits: %w", err)
	}

	if profile == nil || len(profile.RateLimitRules) == 0 {
		return nil, nil
	}

	rules := append([]platform.RateLimitRule(nil), profile.RateLimitRules...)
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].Priority == rules[j].Priority {
			return rateLimitRuleID(rules[i]) < rateLimitRuleID(rules[j])
		}

		return rules[i].Priority < rules[j].Priority
	})

	now := time.Now().UTC()

	var result *cache.RateLimitResult

	for _, rule := range rules {
		policy, ok, err := s.matchRateLimitRule(rule, auth, breakdown, now)
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

func (s *RateLimiter) matchRateLimitRule(rule platform.RateLimitRule, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown, now time.Time) (rateLimitPolicy, bool, error) {
	if !rule.Enabled {
		return rateLimitPolicy{}, false, nil
	}

	if rule.EffectiveFrom != nil && now.Before(rule.EffectiveFrom.UTC()) {
		return rateLimitPolicy{}, false, nil
	}

	if rule.ExpiresAt != nil && !now.Before(rule.ExpiresAt.UTC()) {
		return rateLimitPolicy{}, false, nil
	}

	matches, err := rateLimitRuleMatches(rule, auth, breakdown)
	if err != nil || !matches {
		return rateLimitPolicy{}, matches, err
	}

	policy, err := parseRateLimitPolicy(rule)
	if err != nil {
		return rateLimitPolicy{}, false, err
	}

	if policy.rpm <= 0 && policy.tpm <= 0 {
		return rateLimitPolicy{}, false, nil
	}

	policy.scopeFields = rateLimitScopeFields(rule)
	policy.conditionsKey = rateLimitConditionsKey(rule)

	return policy, true, nil
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

func (s *RateLimiter) applyRateLimitCounter(ctx context.Context, tenantID string, policy rateLimitPolicy, kind string, scopeHash string, weight int64) (*cache.RateLimitResult, error) {
	key := fmt.Sprintf("ratelimit:tenant:%s:rule:%s:scope:%s:%s", strings.TrimSpace(tenantID), policy.ruleID, scopeHash, kind)
	if kind == rateLimitKindTPM || policy.algorithm == "sliding_window" {
		result, err := s.cache.CheckSlidingWindowRateLimitWithWeight(ctx, key, policyLimit(policy, kind), weight, policy.window)
		return decoratePolicyRateLimitResult(result, policy, kind), err
	}

	result, err := s.cache.CheckRateLimit(ctx, key, policyLimit(policy, kind), policy.window)

	return decoratePolicyRateLimitResult(result, policy, kind), err
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

func rateLimitConditionsKey(rule platform.RateLimitRule) string {
	keys := make([]string, 0, len(rule.Conditions))
	for key := range rule.Conditions {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return strings.Join(keys, ",")
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
