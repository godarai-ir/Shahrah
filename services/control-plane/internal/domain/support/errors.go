// Package support contains support-domain rules and errors.
package support

import "errors"

var (
	ErrIdempotencyPayloadConflict     = errors.New("idempotency key was already used with a different payload")
	ErrInvalidTenantStatusChange      = errors.New("invalid tenant status change")
	ErrSupportTicketClosed            = errors.New("support ticket is closed")
	ErrSupportTicketForbidden         = errors.New("support ticket forbidden")
	ErrSupportTicketIdempotency       = errors.New("support ticket idempotency conflict")
	ErrSupportTicketInvalidTransition = errors.New("invalid support ticket status transition")
	ErrSupportTicketNotFound          = errors.New("support ticket not found")
	ErrSupportTicketRateLimited       = errors.New("support ticket rate limit exceeded")
	ErrSupportTicketRatingNotAllowed  = errors.New("support ticket rating not allowed")
	ErrSupportTicketStaleVersion      = errors.New("support ticket stale version")
	ErrTenantNotFound                 = errors.New("tenant not found")
)
