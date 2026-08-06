package support

import "github.com/Delpak/shahrah/packages/contracts/platform"

// CanTransitionTicket reports whether an actor may apply a support-ticket state transition.
func CanTransitionTicket(current, next string, supportActor bool) bool {
	if current == next {
		return true
	}

	switch current {
	case platform.SupportTicketStatusOpen:
		return supportActor && (next == platform.SupportTicketStatusInProgress ||
			next == platform.SupportTicketStatusWaitingForCustomer ||
			next == platform.SupportTicketStatusResolved)
	case platform.SupportTicketStatusInProgress:
		return supportActor && (next == platform.SupportTicketStatusOpen ||
			next == platform.SupportTicketStatusWaitingForCustomer ||
			next == platform.SupportTicketStatusResolved)
	case platform.SupportTicketStatusWaitingForCustomer:
		if next == platform.SupportTicketStatusOpen {
			return true
		}

		return supportActor && (next == platform.SupportTicketStatusInProgress || next == platform.SupportTicketStatusResolved)
	case platform.SupportTicketStatusResolved:
		if next == platform.SupportTicketStatusOpen {
			return true
		}

		return supportActor && next == platform.SupportTicketStatusClosed
	case platform.SupportTicketStatusClosed:
		return false
	default:
		return false
	}
}
