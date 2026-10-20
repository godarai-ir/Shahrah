package seats

import "errors"

var (
	// ErrSeatsUnavailable is returned when the tenant has no active paid
	// subscription to attach paid seats to.
	ErrSeatsUnavailable = errors.New("paid seats need an active subscription")
	// ErrSeatPurchaseRequired is returned when no seat of the type is free in
	// the tenant's stock.
	ErrSeatPurchaseRequired = errors.New("no seat of this type is available; buy one first")
	// ErrFreeSeatsFull is returned when every free seat of the plan is taken.
	ErrFreeSeatsFull = errors.New("every free seat of the plan is taken")
	// ErrMemberNotFound is returned for a user who is not a member of the tenant.
	ErrMemberNotFound = errors.New("member not found")
	// ErrNoPendingChange is returned when a member has no scheduled seat change.
	ErrNoPendingChange = errors.New("member has no scheduled seat change")
	// ErrRequestNotFound is returned for an unknown upgrade request.
	ErrRequestNotFound = errors.New("seat upgrade request not found")
	// ErrRequestsDisabled is returned when the tenant turned upgrade requests off.
	ErrRequestsDisabled = errors.New("seat upgrade requests are turned off")
	// ErrRequestOpen is returned when the member already has an open request.
	ErrRequestOpen = errors.New("member already has an open seat request")
	// ErrRequestNotUpgrade is returned when the requested seat is not better
	// than the member's seat.
	ErrRequestNotUpgrade = errors.New("requested seat is not an upgrade")
	// ErrRequestClosed is returned when a request was already decided.
	ErrRequestClosed = errors.New("seat upgrade request was already decided")
	// ErrOrderNotFound is returned for an unknown seat order.
	ErrOrderNotFound = errors.New("seat order not found")
	// ErrOrderNotPayable is returned for a seat order that is not pending.
	ErrOrderNotPayable = errors.New("seat order is not pending")
	// ErrRenewalNotDue is returned when no paid seat lapses before the period ends.
	ErrRenewalNotDue = errors.New("no paid seats need renewal")
	// ErrInvalidRenewal is returned for a renewal quantity above the seats held.
	ErrInvalidRenewal = errors.New("renewal quantity must be between 0 and the seats held")
)
