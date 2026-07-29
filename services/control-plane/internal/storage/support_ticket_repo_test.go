package storage

import (
	"strings"
	"testing"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func TestSupportTicketTransitionMatrix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		current      string
		next         string
		supportActor bool
		allowed      bool
	}{
		{"new support starts work", platform.SupportTicketStatusOpen, platform.SupportTicketStatusInProgress, true, true},
		{"support waits for customer", platform.SupportTicketStatusOpen, platform.SupportTicketStatusWaitingForCustomer, true, true},
		{"tenant cannot triage open ticket", platform.SupportTicketStatusOpen, platform.SupportTicketStatusResolved, false, false},
		{"tenant reply reopens waiting ticket", platform.SupportTicketStatusWaitingForCustomer, platform.SupportTicketStatusOpen, false, true},
		{"tenant reply reopens resolved ticket", platform.SupportTicketStatusResolved, platform.SupportTicketStatusOpen, false, true},
		{"support can resolve active ticket", platform.SupportTicketStatusInProgress, platform.SupportTicketStatusResolved, true, true},
		{"support can close resolved ticket", platform.SupportTicketStatusResolved, platform.SupportTicketStatusClosed, true, true},
		{"closed ticket is immutable", platform.SupportTicketStatusClosed, platform.SupportTicketStatusOpen, true, false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := validSupportTicketTransition(test.current, test.next, test.supportActor); got != test.allowed {
				t.Fatalf("validSupportTicketTransition(%q, %q, %v) = %v want %v", test.current, test.next, test.supportActor, got, test.allowed)
			}
		})
	}
}

func TestNewSupportTicketPublicIDIsOpaque(t *testing.T) {
	t.Parallel()
	first, err := newSupportTicketPublicID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newSupportTicketPublicID()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("generated duplicate public IDs: %q", first)
	}
	if !strings.HasPrefix(first, "TKT-") || len(first) < 12 {
		t.Fatalf("unexpected public ID format: %q", first)
	}
}

func TestSupportMutationRequestHashIsStableAndPayloadSensitive(t *testing.T) {
	t.Parallel()
	request := platform.RateSupportTicketRequest{Score: 5, Comment: "helpful", IdempotencyKey: "key"}
	if supportMutationRequestHash(request) != supportMutationRequestHash(request) {
		t.Fatal("same mutation request produced different hashes")
	}
	changed := request
	changed.Score = 4
	if supportMutationRequestHash(request) == supportMutationRequestHash(changed) {
		t.Fatal("different mutation payloads produced the same hash")
	}
}
