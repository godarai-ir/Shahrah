package seats

import (
	"slices"
	"strings"
	"time"
)

// Holder is a member holding a paid seat of one type.
type Holder struct {
	UserID     string
	AssignedAt time.Time
}

// LosingHolders returns the holders who lose their seat when only keep seats of
// their type stay paid for. Holders are ranked by when they were assigned, then
// by user id, and the last ones lose first. RunSeatLifecycle ranks lapsing seats
// the same way (ORDER BY assigned_at, user_id), so the renewal preview and the
// worker never disagree about who loses a seat.
func LosingHolders(holders []Holder, keep int) []Holder {
	ranked := slices.Clone(holders)
	slices.SortStableFunc(ranked, func(a, b Holder) int {
		if byTime := a.AssignedAt.Compare(b.AssignedAt); byTime != 0 {
			return byTime
		}

		return strings.Compare(a.UserID, b.UserID)
	})

	keep = max(keep, 0)
	if keep >= len(ranked) {
		return []Holder{}
	}

	return ranked[keep:]
}

// Fallback is what a member holds after losing a paid seat.
type Fallback string

const (
	// FallbackFree means the worker gives the member a free seat.
	FallbackFree Fallback = "free"
	// FallbackUnassigned means the member is left without a seat.
	FallbackUnassigned Fallback = "unassigned"
)

// Candidate is a member who just lost a paid seat.
type Candidate struct {
	UserID       string
	CreatedAt    time.Time
	FreeSeatUsed bool
}

// FreeFallbacks says what each candidate holds after losing a paid seat. The
// worker gives the free seat to members without a seat who never had one, oldest
// member first, while the tenant has room; everyone else stays unassigned.
func FreeFallbacks(candidates []Candidate, room int) map[string]Fallback {
	out := make(map[string]Fallback, len(candidates))

	eligible := make([]Candidate, 0, len(candidates))

	for _, candidate := range candidates {
		out[candidate.UserID] = FallbackUnassigned

		if !candidate.FreeSeatUsed {
			eligible = append(eligible, candidate)
		}
	}

	slices.SortStableFunc(eligible, func(a, b Candidate) int {
		if byTime := a.CreatedAt.Compare(b.CreatedAt); byTime != 0 {
			return byTime
		}

		return strings.Compare(a.UserID, b.UserID)
	})

	for i := 0; i < len(eligible) && i < max(room, 0); i++ {
		out[eligible[i].UserID] = FallbackFree
	}

	return out
}

// RenewalOpensAt is when renewal of a seat subscription ending at periodEnd opens.
func RenewalOpensAt(periodEnd time.Time) time.Time {
	return periodEnd.Add(-RenewalWindow)
}
