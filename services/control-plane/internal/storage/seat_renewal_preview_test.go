package storage

import (
	"testing"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/stretchr/testify/require"
)

func holderRow(id string, seat seats.Type, assigned time.Time) seatHolderRow {
	return seatHolderRow{userID: id, email: id + "@example.com", seat: seat, assignedAt: assigned, createdAt: assigned.Add(-24 * time.Hour)}
}

func losingIDs(line platform.SeatRenewalPreviewLine) []string {
	out := make([]string, 0, len(line.Losing))
	for _, member := range line.Losing {
		out = append(out, member.UserID)
	}

	return out
}

func TestRenewalPreviewLineTheLatestAssignedLose(t *testing.T) {
	base := time.Date(2026, time.October, 1, 9, 0, 0, 0, time.UTC)
	periodEnd := base.AddDate(0, 1, 0)
	holders := []seatHolderRow{
		holderRow("late", seats.Pro, base.Add(72*time.Hour)),
		holderRow("early", seats.Pro, base),
		holderRow("mid", seats.Pro, base.Add(24*time.Hour)),
		holderRow("boss", seats.Max, base),
	}

	line, losing := renewalPreviewLine(seats.Pro, holders, periodEnd, 2)

	require.Equal(t, "pro", line.Seat)
	require.Equal(t, 3, line.Assigned, "only members holding this type are ranked")
	require.Equal(t, []string{"late"}, losingIDs(line))
	require.Len(t, losing, 1)
	require.Equal(t, "late@example.com", line.Losing[0].Email)

	everyone, _ := renewalPreviewLine(seats.Pro, holders, periodEnd, 0)
	require.Equal(t, []string{"early", "mid", "late"}, losingIDs(everyone), "ranked first to last")

	none, _ := renewalPreviewLine(seats.Pro, holders, periodEnd, 3)
	require.Empty(t, none.Losing)
	require.NotNil(t, none.Losing, "the list is empty, not null, in the response")
}

func TestRenewalPreviewLineFollowsScheduledChanges(t *testing.T) {
	base := time.Date(2026, time.October, 1, 9, 0, 0, 0, time.UTC)
	periodEnd := base.AddDate(0, 1, 0)

	leaving := holderRow("leaving", seats.Pro, base.Add(96*time.Hour))
	leaving.pendingSeat, leaving.pendingEffectiveAt = "unassigned", &periodEnd

	downgrade := holderRow("downgrade", seats.Max, base)
	downgrade.pendingSeat, downgrade.pendingEffectiveAt = "pro", &periodEnd

	later := holderRow("later", seats.Pro, base.Add(24*time.Hour))
	later.pendingSeat, later.pendingEffectiveAt = "unassigned", ptr(periodEnd.Add(time.Hour))

	holders := []seatHolderRow{leaving, downgrade, later, holderRow("stays", seats.Pro, base)}

	pro, _ := renewalPreviewLine(seats.Pro, holders, periodEnd, 2)
	require.Equal(t, 3, pro.Assigned, "a member already leaving does not take a seat; a move from Max does")
	require.Equal(t, []string{"downgrade"}, losingIDs(pro), "a moved member is assigned at the period end, so ranks last")

	maxLine, _ := renewalPreviewLine(seats.Max, holders, periodEnd, 0)
	require.Zero(t, maxLine.Assigned, "the Max holder moves to Pro at the period end")
}

func TestAttachSeatFallbacks(t *testing.T) {
	base := time.Date(2026, time.October, 1, 9, 0, 0, 0, time.UTC)
	oldest := holderRow("oldest", seats.Pro, base)
	used := holderRow("used", seats.Pro, base.Add(time.Hour))
	used.freeSeatUsed = true
	newest := holderRow("newest", seats.Pro, base.Add(2*time.Hour))

	preview := &platform.SeatRenewalPreview{Lines: []platform.SeatRenewalPreviewLine{{
		Seat: "pro",
		Losing: []platform.SeatRenewalPreviewMember{
			{UserID: "oldest"}, {UserID: "used"}, {UserID: "newest"},
		},
	}}}

	attachSeatFallbacks(preview, []seatHolderRow{oldest, used, newest}, 1)

	got := map[string]string{}
	for _, member := range preview.Lines[0].Losing {
		got[member.UserID] = member.Fallback
	}

	require.Equal(t, map[string]string{"oldest": "free", "used": "unassigned", "newest": "unassigned"}, got)
}

func ptr(t time.Time) *time.Time { return &t }
