package billing

import "testing"

func TestDecideSubscriptionTransition(t *testing.T) {
	tests := []struct {
		name  string
		input SubscriptionTransitionInput
		want  SubscriptionTransitionDecision
	}{
		{
			name:  "trial expires",
			input: SubscriptionTransitionInput{Status: "trialing", Interval: "monthly"},
			want: SubscriptionTransitionDecision{
				Kind: SubscriptionTransitionExpire, Interval: "monthly", EventType: "trial_expired",
			},
		},
		{
			name: "pending downgrade uses requested interval",
			input: SubscriptionTransitionInput{
				Status: "active", Interval: "monthly", CancelAtEnd: true,
				PendingVersion: "version-2", PendingInterval: "annual",
			},
			want: SubscriptionTransitionDecision{
				Kind: SubscriptionTransitionDowngrade, Interval: "yearly", EventType: "downgrade_applied",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := DecideSubscriptionTransition(test.input); got != test.want {
				t.Fatalf("unexpected decision: got %+v want %+v", got, test.want)
			}
		})
	}
}
