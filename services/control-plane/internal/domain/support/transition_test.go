package support

import (
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestCanTransitionTicket(t *testing.T) {
	tests := []struct {
		name         string
		current      string
		next         string
		supportActor bool
		want         bool
	}{
		{
			name: "customer reopens waiting ticket", current: platform.SupportTicketStatusWaitingForCustomer,
			next: platform.SupportTicketStatusOpen, want: true,
		},
		{
			name: "customer cannot resolve ticket", current: platform.SupportTicketStatusOpen,
			next: platform.SupportTicketStatusResolved, want: false,
		},
		{
			name: "support resolves ticket", current: platform.SupportTicketStatusInProgress,
			next: platform.SupportTicketStatusResolved, supportActor: true, want: true,
		},
		{
			name: "closed ticket is terminal", current: platform.SupportTicketStatusClosed,
			next: platform.SupportTicketStatusOpen, supportActor: true, want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CanTransitionTicket(test.current, test.next, test.supportActor); got != test.want {
				t.Fatalf("unexpected transition result: got %t want %t", got, test.want)
			}
		})
	}
}
