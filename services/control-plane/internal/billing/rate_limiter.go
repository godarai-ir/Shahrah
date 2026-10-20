package billing

import (
	"context"
	"fmt"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
)

// CheckRateLimit validates rate-limit policies and the subscription state.
// It returns the primary rate-limit result to expose in response headers.
func (s *RateLimiter) CheckRateLimit(
	ctx context.Context,
	auth *platform.AuthContext,
	model string,
) (*cache.RateLimitResult, error) {
	policyResult, err := s.checkRateLimitPolicies(ctx, auth, platform.UsageCostBreakdown{
		Model: model,
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

	return policyResult, nil
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
