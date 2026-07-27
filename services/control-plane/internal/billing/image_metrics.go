package billing

import (
	"math"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	imageBillingRatingTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shahrah_image_billing_rating_total",
			Help: "Image billing outcomes grouped by provider and rating status.",
		},
		[]string{"provider", "status"},
	)
	imageBillingFallbackTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shahrah_image_billing_fallback_total",
			Help: "Image billing fallback and unresolved outcomes.",
		},
		[]string{"provider", "reason"},
	)
	imageBillingShadowComparisonTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shahrah_image_billing_shadow_comparison_total",
			Help: "Resolution comparison between legacy and dimensional image billing.",
		},
		[]string{"provider", "result"},
	)
	imageBillingShadowCostDeltaUSD = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "shahrah_image_billing_shadow_cost_delta_usd",
			Help:    "Absolute USD cost delta between legacy and dimensional image billing.",
			Buckets: prometheus.ExponentialBuckets(0.000001, 10, 8),
		},
		[]string{"provider"},
	)
)

func observeImageBillingResult(provider string, status string, err error) {
	provider = imageBillingMetricLabel(provider)
	status = strings.ToLower(strings.TrimSpace(status))
	if err != nil {
		observeImageBillingFallback(provider, "unresolved")
		status = "unrated"
	}
	switch status {
	case "estimated_final", "rated", "provider_reported", "unrated":
	case "free":
		status = "rated"
	default:
		status = "unrated"
	}
	imageBillingRatingTotal.WithLabelValues(provider, status).Inc()
}

func observeImageBillingFallback(provider string, reason string) {
	imageBillingFallbackTotal.WithLabelValues(
		imageBillingMetricLabel(provider),
		imageBillingMetricLabel(reason),
	).Inc()
}

func observeImageBillingShadow(
	provider string,
	legacy platform.UsageCostBreakdown,
	legacyErr error,
	dimensional platform.UsageCostBreakdown,
	dimensionalErr error,
) {
	provider = imageBillingMetricLabel(provider)
	result := "both_unresolved"
	switch {
	case legacyErr == nil && dimensionalErr == nil:
		result = "both_resolved"
		imageBillingShadowCostDeltaUSD.WithLabelValues(provider).Observe(
			math.Abs(legacy.TotalCostUSD - dimensional.TotalCostUSD),
		)
	case legacyErr == nil:
		result = "legacy_only"
	case dimensionalErr == nil:
		result = "dimensional_only"
	}
	imageBillingShadowComparisonTotal.WithLabelValues(provider, result).Inc()
	if dimensionalErr != nil {
		observeImageBillingFallback(provider, "shadow_dimensional_unresolved")
	}
}

func imageBillingMetricLabel(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "unknown"
	}
	return value
}
