package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

const (
	budgetTargetTenantCredit = "tenant_credit"
	budgetTargetPolicy       = "policy"
)

type budgetReservationTarget struct {
	platform.BudgetReservationTarget
	tenantID     string
	allowanceIRR float64
	expiresAt    time.Time
	ttl          time.Duration
	// seedCreditPool seeds the window from the tenant's credit balance.
	seedCreditPool bool
}

func (s *BudgetService) reserveBudgetTarget(ctx context.Context, budgetCache BudgetWindowCache, target budgetReservationTarget, reservationID string, now time.Time) (*cache.BudgetWindowResult, error) {
	// Rule windows are re-seeded on every reservation so a changed limit
	// applies at once; the pool window is seeded on first use and refreshed
	// when the balance changes.
	if !target.seedCreditPool {
		if seedErr := s.seedBudgetTarget(ctx, budgetCache, target); seedErr != nil {
			return nil, seedErr
		}
	}

	result, err := budgetCache.ReserveBudget(ctx, cache.BudgetReservationRequest{
		WindowKey:      target.WindowKey,
		ReservationKey: target.ReservationKey,
		ReservationID:  reservationID,
		EstimatedIRR:   target.EstimatedIRR,
		TTL:            target.ttl,
		Now:            now,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: reserve failed: %v", ErrBudgetUnavailable, err)
	}

	if result != nil && result.State != "window_unavailable" {
		return result, nil
	}

	if seedErr := s.seedBudgetTarget(ctx, budgetCache, target); seedErr != nil {
		return nil, seedErr
	}

	result, err = budgetCache.ReserveBudget(ctx, cache.BudgetReservationRequest{
		WindowKey:      target.WindowKey,
		ReservationKey: target.ReservationKey,
		ReservationID:  reservationID,
		EstimatedIRR:   target.EstimatedIRR,
		TTL:            target.ttl,
		Now:            now,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: reserve failed: %v", ErrBudgetUnavailable, err)
	}

	return result, nil
}

func (s *BudgetService) seedBudgetTarget(ctx context.Context, budgetCache BudgetWindowCache, target budgetReservationTarget) error {
	if target.seedCreditPool {
		return s.seedCreditPoolWindow(ctx, budgetCache, target)
	}

	result, err := budgetCache.SeedBudgetWindow(ctx, cache.BudgetWindowSeed{
		WindowKey:    target.WindowKey,
		AllowanceIRR: target.allowanceIRR,
		ExpiresAt:    target.expiresAt,
		TTL:          target.ttl,
		Version:      target.expiresAt.Unix(),
	})
	if err != nil {
		return fmt.Errorf("%w: seed failed: %v", ErrBudgetUnavailable, err)
	}

	if result == nil || !result.Allowed {
		return ErrBudgetUnavailable
	}

	return nil
}

func (s *BudgetService) releaseBudgetTargetsBestEffort(ctx context.Context, budgetCache BudgetWindowCache, targets []budgetReservationTarget, now time.Time) error {
	var releaseErrors []error

	for i := len(targets) - 1; i >= 0; i-- {
		_, err := budgetCache.ReleaseBudget(ctx, cache.BudgetReleaseRequest{
			WindowKey:      targets[i].WindowKey,
			ReservationKey: targets[i].ReservationKey,
			Now:            now,
		})
		if err != nil {
			releaseErrors = append(releaseErrors, err)
		}
	}

	return errors.Join(releaseErrors...)
}

func budgetRuleMatches(rule platform.BudgetLimitRule, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) (bool, error) {
	if len(rule.Conditions) == 0 {
		return true, nil
	}

	if rawConditions, ok := rule.Conditions["conditions"]; ok {
		for key, expected := range rule.Conditions {
			if key == "conditions" || key == "combinator" {
				continue
			}

			actual, known := budgetContextValue(key, auth, breakdown)
			if !known {
				return false, fmt.Errorf("unknown condition %q", key)
			}

			if !conditionValueMatches(actual, expected) {
				return false, nil
			}
		}

		conditions, ok := rawConditions.([]interface{})
		if !ok {
			return false, fmt.Errorf("conditions must be an array")
		}

		for _, raw := range conditions {
			condition, ok := raw.(map[string]interface{})
			if !ok {
				return false, fmt.Errorf("condition must be an object")
			}

			matches, err := budgetStructuredConditionMatches(condition, auth, breakdown)
			if err != nil || !matches {
				return matches, err
			}
		}

		return true, nil
	}

	keys := make([]string, 0, len(rule.Conditions))
	for key := range rule.Conditions {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		actual, known := budgetContextValue(key, auth, breakdown)
		if !known {
			return false, fmt.Errorf("unknown condition %q", key)
		}

		if !conditionValueMatches(actual, rule.Conditions[key]) {
			return false, nil
		}
	}

	return true, nil
}

func budgetStructuredConditionMatches(condition map[string]interface{}, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) (bool, error) {
	field := strings.TrimSpace(asString(condition["field"]))

	operator := strings.TrimSpace(asString(condition["operator"]))
	if operator == "" {
		operator = "in"
	}

	switch field {
	case "subject":
		key := "user_id"

		switch strings.TrimSpace(asString(condition["subjectKind"])) {
		case "team":
			key = "team_id"
		case "virtual_account":
			key = "virtual_account_id"
		}

		actual, _ := budgetContextValue(key, auth, breakdown)

		return stringListOperatorMatches(actual, operator, condition["values"])
	case "model":
		actual, _ := budgetContextValue("model", auth, breakdown)
		return stringListOperatorMatches(actual, operator, condition["values"])
	case "metadata":
		metadataKey := strings.TrimSpace(asString(condition["metadataKey"]))
		if metadataKey == "" {
			return false, fmt.Errorf("metadata key is required")
		}

		actual, known := budgetContextValue("metadata."+metadataKey, auth, breakdown)

		switch operator {
		case "exists":
			return known && strings.TrimSpace(actual) != "", nil
		case "not_exists":
			return !known || strings.TrimSpace(actual) == "", nil
		case "equals":
			return actual == strings.TrimSpace(asString(condition["value"])), nil
		case "not_equals":
			return actual != strings.TrimSpace(asString(condition["value"])), nil
		case "contains":
			return strings.Contains(actual, strings.TrimSpace(asString(condition["value"]))), nil
		case "starts_with":
			return strings.HasPrefix(actual, strings.TrimSpace(asString(condition["value"]))), nil
		case "ends_with":
			return strings.HasSuffix(actual, strings.TrimSpace(asString(condition["value"]))), nil
		case "gt", "gte", "lt", "lte":
			return numericOperatorMatches(actual, operator, condition["value"])
		default:
			return stringListOperatorMatches(actual, operator, condition["values"])
		}
	default:
		return false, fmt.Errorf("unknown condition field %q", field)
	}
}

func stringListOperatorMatches(actual string, operator string, rawValues interface{}) (bool, error) {
	values, err := stringSlice(rawValues)
	if err != nil {
		return false, err
	}

	switch operator {
	case "in":
		for _, value := range values {
			if actual == value {
				return true, nil
			}
		}

		return false, nil
	case "not_in":
		for _, value := range values {
			if actual == value {
				return false, nil
			}
		}

		return true, nil
	default:
		return false, fmt.Errorf("unsupported operator %q", operator)
	}
}

func conditionValueMatches(actual string, expected interface{}) bool {
	values, err := stringSlice(expected)
	if err != nil {
		return actual == strings.TrimSpace(fmt.Sprint(expected))
	}

	for _, value := range values {
		if actual == value {
			return true
		}
	}

	return false
}

func budgetRuleScope(rule platform.BudgetLimitRule, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) (string, string, error) {
	scope := string(platform.BudgetAppliesPerTenant)
	if rule.BudgetAppliesPer != nil && strings.TrimSpace(*rule.BudgetAppliesPer) != "" {
		scope = strings.TrimSpace(*rule.BudgetAppliesPer)
	}

	value, known := budgetContextValue(scope, auth, breakdown)
	if !known {
		return "", "", fmt.Errorf("unknown applies-per %q", scope)
	}

	if strings.TrimSpace(value) == "" {
		return "", "", fmt.Errorf("missing value for applies-per %q", scope)
	}

	return scope, value, nil
}

func budgetContextValue(key string, auth *platform.AuthContext, breakdown platform.UsageCostBreakdown) (string, bool) {
	key = strings.TrimSpace(key)
	switch key {
	case string(platform.BudgetAppliesPerTenant), "tenant_id":
		return strings.TrimSpace(auth.TenantID), true
	case string(platform.BudgetAppliesPerUser), "user_id":
		return strings.TrimSpace(auth.UserID), true
	case string(platform.BudgetAppliesPerTeam), "team_id":
		return strings.TrimSpace(auth.TeamID), true
	case "organization_id":
		return strings.TrimSpace(auth.OrganizationID), true
	case string(platform.BudgetAppliesPerVirtualAccount), "virtual_account", "virtual_account_id", "key_id":
		return strings.TrimSpace(auth.KeyID), true
	case "key_alias":
		return strings.TrimSpace(auth.KeyAlias), true
	case "token_hash":
		return strings.TrimSpace(auth.TokenHash), true
	case string(platform.BudgetAppliesPerModel):
		return strings.TrimSpace(breakdown.Model), true
	case "key_type":
		return strings.TrimSpace(auth.KeyType), true
	case "role":
		return strings.TrimSpace(auth.Role), true
	case "virtual_user_id":
		return strings.TrimSpace(auth.VirtualUserID), true
	case "request_type":
		return strings.TrimSpace(breakdown.RequestType), true
	case "endpoint":
		return strings.TrimSpace(breakdown.Endpoint), true
	default:
		if strings.HasPrefix(key, "metadata.") {
			if breakdown.Metadata == nil {
				return "", false
			}

			value, ok := breakdown.Metadata[strings.TrimPrefix(key, "metadata.")]
			if !ok {
				return "", false
			}

			return strings.TrimSpace(asString(value)), true
		}

		return "", false
	}
}

func numericOperatorMatches(actual string, operator string, expected interface{}) (bool, error) {
	actualNumber, err := parseFloat(actual)
	if err != nil {
		return false, err
	}

	expectedNumber, err := parseFloat(asString(expected))
	if err != nil {
		return false, err
	}

	switch operator {
	case "gt":
		return actualNumber > expectedNumber, nil
	case "gte":
		return actualNumber >= expectedNumber, nil
	case "lt":
		return actualNumber < expectedNumber, nil
	case "lte":
		return actualNumber <= expectedNumber, nil
	default:
		return false, fmt.Errorf("unsupported numeric operator %q", operator)
	}
}

func budgetPolicyWindowForTime(unit platform.BudgetLimitUnit, now time.Time) (string, time.Time, error) {
	utc := now.UTC()

	switch unit {
	case platform.BudgetLimitCostPerDay:
		start := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
		return start.Format("20060102"), start.AddDate(0, 0, 1), nil
	case platform.BudgetLimitCostPerWeek:
		weekday := int(utc.Weekday())
		if weekday == 0 {
			weekday = 7
		}

		start := time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1-weekday)
		year, week := start.ISOWeek()

		return fmt.Sprintf("%04dW%02d", year, week), start.AddDate(0, 0, 7), nil
	case platform.BudgetLimitCostPerMonth:
		windowID, windowEnd := budgetWindowForTime(utc)
		return windowID, windowEnd, nil
	default:
		return "", time.Time{}, fmt.Errorf("unsupported budget unit %q", unit)
	}
}

func ruleIdentifier(rule platform.BudgetLimitRule) string {
	if strings.TrimSpace(rule.RuleID) != "" {
		return strings.TrimSpace(rule.RuleID)
	}

	return strings.TrimSpace(rule.ID)
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:16]
}

func asString(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(value)
	}
}

func parseFloat(value string) (float64, error) {
	out, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric condition value %q", value)
	}

	return out, nil
}

func stringSlice(value interface{}) ([]string, error) {
	switch typed := value.(type) {
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if trimmed := strings.TrimSpace(item); trimmed != "" {
				out = append(out, trimmed)
			}
		}

		return out, nil
	case []interface{}:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if trimmed := strings.TrimSpace(asString(item)); trimmed != "" {
				out = append(out, trimmed)
			}
		}

		return out, nil
	default:
		trimmed := strings.TrimSpace(asString(value))
		if trimmed == "" {
			return nil, fmt.Errorf("empty condition value")
		}

		return []string{trimmed}, nil
	}
}
