package seats

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func holder(id string, assigned time.Time) Holder {
	return Holder{UserID: id, AssignedAt: assigned}
}

func ids(holders []Holder) []string {
	out := make([]string, 0, len(holders))
	for _, item := range holders {
		out = append(out, item.UserID)
	}

	return out
}

func TestLosingHoldersTheLatestAssignedLoseFirst(t *testing.T) {
	base := time.Date(2026, time.October, 1, 9, 0, 0, 0, time.UTC)
	holders := []Holder{
		holder("c", base.Add(48*time.Hour)),
		holder("a", base),
		holder("b", base.Add(24*time.Hour)),
	}

	require.Equal(t, []string{"c"}, ids(LosingHolders(holders, 2)))
	require.Equal(t, []string{"b", "c"}, ids(LosingHolders(holders, 1)))
	require.Equal(t, []string{"a", "b", "c"}, ids(LosingHolders(holders, 0)))
	require.Equal(t, []string{"a", "b", "c"}, ids(LosingHolders(holders, -3)), "a negative keep keeps nobody")
	require.Empty(t, LosingHolders(holders, 3), "nobody loses a seat when every seat is kept")
	require.Empty(t, LosingHolders(holders, 10))
	require.Equal(t, []string{"c", "a", "b"}, ids(holders), "the input order is left alone")
}

func TestLosingHoldersBreakTiesByUserID(t *testing.T) {
	at := time.Date(2026, time.October, 1, 9, 0, 0, 0, time.UTC)
	holders := []Holder{holder("b", at), holder("c", at), holder("a", at)}

	require.Equal(t, []string{"c"}, ids(LosingHolders(holders, 2)))
	require.Equal(t, []string{"b", "c"}, ids(LosingHolders(holders, 1)))
}

func TestFreeFallbacks(t *testing.T) {
	base := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	candidates := []Candidate{
		{UserID: "new", CreatedAt: base.Add(72 * time.Hour)},
		{UserID: "used", CreatedAt: base, FreeSeatUsed: true},
		{UserID: "old", CreatedAt: base.Add(24 * time.Hour)},
		{UserID: "mid", CreatedAt: base.Add(48 * time.Hour)},
	}

	got := FreeFallbacks(candidates, 2)
	require.Equal(t, FallbackFree, got["old"], "the oldest member gets the free seat first")
	require.Equal(t, FallbackFree, got["mid"])
	require.Equal(t, FallbackUnassigned, got["new"], "no room is left")
	require.Equal(t, FallbackUnassigned, got["used"], "the free seat is given once")

	require.Equal(t, FallbackUnassigned, FreeFallbacks(candidates, 0)["old"])
	require.Equal(t, FallbackUnassigned, FreeFallbacks(candidates, -1)["old"])
	require.Empty(t, FreeFallbacks(nil, 3))
}

func TestRenewalOpensAt(t *testing.T) {
	end := time.Date(2026, time.November, 21, 9, 0, 0, 0, time.UTC)
	opens := RenewalOpensAt(end)

	require.Equal(t, end.Add(-RenewalWindow), opens)
	require.True(t, RenewalOpen(end, opens), "the window opens exactly then")
	require.False(t, RenewalOpen(end, opens.Add(-time.Second)))
}
