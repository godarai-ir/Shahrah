package billing

import (
	"strings"
	"time"
)

// SubscriptionTransitionKind identifies the business action for an ended subscription period.
type SubscriptionTransitionKind uint8

const (
	SubscriptionTransitionExpire SubscriptionTransitionKind = iota
	SubscriptionTransitionDowngrade
)

// SubscriptionTransitionInput contains persistence-neutral lifecycle state.
type SubscriptionTransitionInput struct {
	Status          string
	Interval        string
	CancelAtEnd     bool
	PendingVersion  string
	PendingInterval string
}

// SubscriptionTransitionDecision contains the lifecycle action selected by domain rules.
type SubscriptionTransitionDecision struct {
	Kind      SubscriptionTransitionKind
	Interval  string
	EventType string
}

// DecideSubscriptionTransition selects the lifecycle action for an ended period.
func DecideSubscriptionTransition(input SubscriptionTransitionInput) SubscriptionTransitionDecision {
	if input.CancelAtEnd && strings.TrimSpace(input.PendingVersion) != "" {
		interval := input.PendingInterval
		if strings.TrimSpace(interval) == "" {
			interval = input.Interval
		}

		return SubscriptionTransitionDecision{
			Kind:      SubscriptionTransitionDowngrade,
			Interval:  NormalizeBillingInterval(interval),
			EventType: "downgrade_applied",
		}
	}

	eventType := "subscription_expired"
	if strings.EqualFold(strings.TrimSpace(input.Status), "trialing") {
		eventType = "trial_expired"
	}

	return SubscriptionTransitionDecision{
		Kind:      SubscriptionTransitionExpire,
		Interval:  NormalizeBillingInterval(input.Interval),
		EventType: eventType,
	}
}

// NormalizeBillingInterval converts supported aliases to their canonical value.
func NormalizeBillingInterval(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "quarterly":
		return "quarterly"
	case "yearly", "annual", "annually":
		return "yearly"
	default:
		return "monthly"
	}
}

// AddBillingInterval returns the end of a billing period.
func AddBillingInterval(start time.Time, interval string) time.Time {
	switch NormalizeBillingInterval(interval) {
	case "quarterly":
		return start.AddDate(0, 3, 0)
	case "yearly":
		return start.AddDate(1, 0, 0)
	default:
		return start.AddDate(0, 1, 0)
	}
}
