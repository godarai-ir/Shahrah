package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

// CheckRateLimit validates per-minute request and token limits.
// It returns the primary rate-limit result to expose in response headers.
func (s *RateLimiter) CheckRateLimit(
	ctx context.Context,
	auth *platform.AuthContext,
	model string,
	estimatedTokens int,
) (*cache.RateLimitResult, error) {
	policyResult, err := s.checkRateLimitPolicies(ctx, auth, platform.UsageCostBreakdown{
		Model:       model,
		TotalTokens: estimatedTokens,
	})
	if err != nil {
		return policyResult, err
	}

	if policyResult != nil && !policyResult.Allowed {
		return policyResult, ErrRateLimitExceeded
	}

	if s.entitlements == nil || auth == nil || auth.TenantID == "" {
		return policyResult, nil
	}

	effective, err := s.resolveEffectiveEntitlementLimits(ctx, auth.TenantID)
	if err != nil {
		return nil, ErrSubscriptionInactive
	}

	if effective.Subscription == nil ||
		(effective.Subscription.Status != "active" && effective.Subscription.Status != "trialing") {
		return nil, ErrSubscriptionInactive
	}

	limit, limited := entitlementNumber(effective.Features[platform.FeatureGatewayRequestsMonthly])
	if !limited {
		return policyResult, nil
	}

	counter, allowed, err := s.entitlements.CheckAndIncrementTenantUsageCounter(
		ctx,
		auth.TenantID,
		platform.FeatureGatewayRequestsMonthly,
		limit,
		1,
		time.Now().UTC(),
	)
	if err != nil {
		return nil, err
	}

	result := &cache.RateLimitResult{
		Allowed: allowed,
		Current: int64(counter.Used),
		Limit:   int64(limit),
		ResetAt: counter.WindowEnd,
	}
	if !allowed {
		return result, ErrRateLimitExceeded
	}

	return mostRestrictiveRateLimitResult(policyResult, result), nil
}

func (s *RateLimiter) resolveEffectiveEntitlementLimits(
	ctx context.Context,
	tenantID string,
) (*platform.EffectiveEntitlementLimits, error) {
	if s.runtimeProfiles != nil {
		profile, err := s.runtimeProfiles.GetSnapshot(ctx, tenantID)
		if err == nil && profile != nil && profile.Entitlements != nil {
			out := *profile.Entitlements
			out.Features = cloneInterfaceMap(profile.Entitlements.Features)

			if profile.Entitlements.Subscription != nil {
				subscription := *profile.Entitlements.Subscription
				subscription.PriceSnapshot = cloneInterfaceMap(subscription.PriceSnapshot)
				subscription.Metadata = cloneInterfaceMap(subscription.Metadata)
				out.Subscription = &subscription
			}

			return &out, nil
		}

		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("tenant runtime profile entitlements unavailable")
	}

	return s.entitlements.ResolveEffectiveEntitlementLimits(ctx, tenantID)
}

func entitlementNumber(value interface{}) (float64, bool) {
	switch typed := value.(type) {
	case nil:
		return 0, false
	case int:
		return float64(typed), typed >= 0
	case int64:
		return float64(typed), typed >= 0
	case float64:
		return typed, typed >= 0
	case json.Number:
		out, err := typed.Float64()

		return out, err == nil && out >= 0
	default:
		return 0, false
	}
}
