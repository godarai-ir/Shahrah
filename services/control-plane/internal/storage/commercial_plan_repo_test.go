package storage

import "testing"

func TestSubscriptionStatusIsCurrent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status  string
		current bool
	}{
		{status: "active", current: true},
		{status: "trialing", current: true},
		{status: "past_due", current: true},
		{status: "expired", current: false},
		{status: "canceled", current: false},
	}

	for _, test := range tests {
		t.Run(test.status, func(t *testing.T) {
			t.Parallel()

			if got := subscriptionStatusIsCurrent(test.status); got != test.current {
				t.Fatalf("subscriptionStatusIsCurrent(%q) = %v, want %v", test.status, got, test.current)
			}
		})
	}
}
