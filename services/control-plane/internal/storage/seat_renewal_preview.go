package storage

import (
	"context"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/jackc/pgx/v5"
)

// seatHolderRow is a member holding a paid seat, with what decides their fate
// when a block lapses.
type seatHolderRow struct {
	userID, email, name string
	seat                seats.Type
	assignedAt          time.Time
	createdAt           time.Time
	freeSeatUsed        bool
	pendingSeat         string
	pendingEffectiveAt  *time.Time
}

// PreviewSeatRenewal prices renewing keep seats of each paid type and says which
// members lose their seat at the period end, without changing anything. A type
// missing from keep uses the count the admin saved last.
func (s *SeatRepository) PreviewSeatRenewal(ctx context.Context, tenantID string, keep map[seats.Type]int, now time.Time) (*platform.SeatRenewalPreview, error) {
	now = now.UTC()

	sub, err := currentSeatSubscription(ctx, s.pool, tenantID)
	if err != nil {
		return nil, err
	}

	if !sub.running(now) {
		return nil, seats.ErrRenewalNotDue
	}

	nextEnd, err := seats.PeriodEnd(sub.PeriodEnd, sub.Interval)
	if err != nil {
		return nil, err
	}

	stock, err := loadSeatStock(ctx, s.pool, tenantID, now)
	if err != nil {
		return nil, err
	}

	renewed, err := loadRenewedSeats(ctx, s.pool, tenantID, now, sub.PeriodEnd)
	if err != nil {
		return nil, err
	}

	holders, err := loadSeatHolders(ctx, s.pool, tenantID)
	if err != nil {
		return nil, err
	}

	freeRoom, err := freeSeatRoom(ctx, s.pool, tenantID)
	if err != nil {
		return nil, err
	}

	taxPercent, err := creditTaxPercent(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	preview := &platform.SeatRenewalPreview{
		PeriodEnd: sub.PeriodEnd, NextPeriodEnd: nextEnd, Lines: []platform.SeatRenewalPreviewLine{},
		TaxPercent: taxPercent.StringFixed(2),
	}

	var allLosing []seatHolderRow

	for _, seatType := range seats.PaidTypes {
		held := stock[seatType]

		wanted, ok := keep[seatType]
		if !ok {
			wanted = held.Renewing
		}

		if wanted < 0 || wanted > held.Purchased {
			return nil, seats.ErrInvalidRenewal
		}

		unitPrice, _, priceErr := seatUnitPeriodPrice(ctx, s.pool, seatType, sub.Interval)
		if priceErr != nil {
			return nil, priceErr
		}

		// Seats already paid for the next period stay whatever is asked, and
		// only the rest of the kept seats are paid for now.
		stays := max(wanted, renewed[seatType])
		preview.PriceIRR += unitPrice * int64(max(wanted-renewed[seatType], 0))

		line, losing := renewalPreviewLine(seatType, holders, sub.PeriodEnd, stays)
		line.Held, line.Renewed, line.Keep, line.UnitPriceIRR = held.Purchased, renewed[seatType], wanted, unitPrice
		preview.Lines = append(preview.Lines, line)
		allLosing = append(allLosing, losing...)
	}

	attachSeatFallbacks(preview, allLosing, freeRoom)

	preview.TaxIRR = seatTaxIRR(preview.PriceIRR, taxPercent)
	preview.PayableIRR = preview.PriceIRR + preview.TaxIRR

	return preview, nil
}

// renewalPreviewLine ranks the members holding a seat type as they will be at
// the period end and returns those who lose it when stays seats stay paid.
func renewalPreviewLine(seatType seats.Type, holders []seatHolderRow, periodEnd time.Time, stays int) (platform.SeatRenewalPreviewLine, []seatHolderRow) {
	line := platform.SeatRenewalPreviewLine{Seat: string(seatType), Losing: []platform.SeatRenewalPreviewMember{}}
	byID := map[string]seatHolderRow{}
	ranked := []seats.Holder{}

	for _, holder := range holders {
		effective, assignedAt, leaves := holder.atPeriodEnd(periodEnd)
		if leaves || effective != seatType {
			continue
		}

		byID[holder.userID] = holder
		ranked = append(ranked, seats.Holder{UserID: holder.userID, AssignedAt: assignedAt})
	}

	line.Assigned = len(ranked)

	losing := make([]seatHolderRow, 0)

	for _, loser := range seats.LosingHolders(ranked, stays) {
		row := byID[loser.UserID]
		losing = append(losing, row)

		assignedAt := loser.AssignedAt
		line.Losing = append(line.Losing, platform.SeatRenewalPreviewMember{
			UserID: row.userID, Email: row.email, Name: row.name, AssignedAt: &assignedAt,
		})
	}

	return line, losing
}

// atPeriodEnd is the seat a holder has once the scheduled changes due by the
// period end are applied: the type, when it was assigned, and whether the
// member is already leaving the paid seats.
func (h seatHolderRow) atPeriodEnd(periodEnd time.Time) (seats.Type, time.Time, bool) {
	if h.pendingEffectiveAt != nil && !h.pendingEffectiveAt.After(periodEnd) {
		switch seats.Type(h.pendingSeat) {
		case seats.Unassigned:
			return h.seat, h.assignedAt, true
		case seats.Pro:
			// A scheduled move is assigned at the moment it applies.
			return seats.Pro, periodEnd, false
		}
	}

	return h.seat, h.assignedAt, false
}

// attachSeatFallbacks tells each losing member what they hold afterwards.
func attachSeatFallbacks(preview *platform.SeatRenewalPreview, losing []seatHolderRow, freeRoom int) {
	candidates := make([]seats.Candidate, 0, len(losing))
	for _, row := range losing {
		candidates = append(candidates, seats.Candidate{UserID: row.userID, CreatedAt: row.createdAt, FreeSeatUsed: row.freeSeatUsed})
	}

	fallbacks := seats.FreeFallbacks(candidates, freeRoom)

	for i := range preview.Lines {
		for j := range preview.Lines[i].Losing {
			member := &preview.Lines[i].Losing[j]
			member.Fallback = string(fallbacks[member.UserID])
		}
	}
}

// loadRenewedSeats counts, per paid type, the seats already paid for the next
// period: blocks paid past the current period end.
func loadRenewedSeats(ctx context.Context, q rowsQuerier, tenantID string, now, periodEnd time.Time) (map[seats.Type]int, error) {
	rows, err := q.Query(ctx, `
		SELECT seat_type, COALESCE(SUM(quantity), 0)::int
		FROM tenant_seat_blocks
		WHERE tenant_id = $1::uuid AND paid_through > $2 AND paid_through > $3
		GROUP BY seat_type
	`, tenantID, now, periodEnd)
	if err != nil {
		return nil, err
	}

	out := map[seats.Type]int{}

	_, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (struct{}, error) {
		var (
			seatType string
			quantity int
		)

		if scanErr := row.Scan(&seatType, &quantity); scanErr != nil {
			return struct{}{}, scanErr
		}

		out[seats.Type(seatType)] = quantity

		return struct{}{}, nil
	})

	return out, err
}

// loadSeatHolders lists the active members holding a paid seat.
func loadSeatHolders(ctx context.Context, q rowsQuerier, tenantID string) ([]seatHolderRow, error) {
	rows, err := q.Query(ctx, `
		SELECT a.seat_type, COALESCE(a.pending_seat_type, ''), a.pending_effective_at, a.assigned_at,
		       u.user_id::text, COALESCE(u.email, ''), TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')),
		       u.created_at,
		       EXISTS (SELECT 1 FROM tenant_free_seat_grants g WHERE g.tenant_id = u.tenant_id AND g.user_id = u.user_id)
		FROM tenant_seat_assignments a
		JOIN users u ON u.user_id = a.user_id AND u.tenant_id = a.tenant_id AND COALESCE(u.active, true)
		WHERE a.tenant_id = $1::uuid AND a.seat_type IN ('pro', 'max')
	`, tenantID)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (seatHolderRow, error) {
		var (
			item       seatHolderRow
			seatType   string
			assignedAt *time.Time
		)

		if scanErr := row.Scan(&seatType, &item.pendingSeat, &item.pendingEffectiveAt, &assignedAt,
			&item.userID, &item.email, &item.name, &item.createdAt, &item.freeSeatUsed); scanErr != nil {
			return item, scanErr
		}

		item.seat = seats.Type(seatType)
		item.name = strings.TrimSpace(item.name)

		// A seat without a recorded time ranks last, as it does in the worker's SQL.
		item.assignedAt = seats.FreePeriodEnd
		if assignedAt != nil {
			item.assignedAt = assignedAt.UTC()
		}

		return item, nil
	})
}

// freeSeatRoom is how many free seats the tenant can still give.
func freeSeatRoom(ctx context.Context, q rowQuerier, tenantID string) (int, error) {
	var held int

	err := q.QueryRow(ctx, `
		SELECT COUNT(1) FROM tenant_seat_assignments a
		JOIN users u ON u.user_id = a.user_id AND u.tenant_id = a.tenant_id AND COALESCE(u.active, true)
		WHERE a.tenant_id = $1::uuid AND a.seat_type = 'free'
	`, tenantID).Scan(&held)
	if err != nil {
		return 0, err
	}

	return max(seats.FreeSeatLimit-held, 0), nil
}
