package storage

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/jackc/pgx/v5"
)

// tenantSeatUsage is the user-cap state of one tenant. Every tenant has
// seats.FreeSeatLimit free seats; every paid seat it bought adds one more member.
type tenantSeatUsage struct {
	Used        int
	Active      int
	Pending     int
	Limit       int
	Limited     bool
	FreeLimit   int
	FreeLimited bool
	PaidSeats   int
}

// rowQuerier is satisfied by both a pool and a transaction.
type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// GetTenantSeatUsage reports the tenant's seats against its user cap.
func (s *TenantAdminRepository) GetTenantSeatUsage(ctx context.Context, tenantID string) (*platform.TenantSeatUsage, error) {
	usage, err := tenantSeatUsageTx(ctx, s.pool, tenantID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	out := &platform.TenantSeatUsage{Used: usage.Used, ActiveMembers: usage.Active, PendingInvitations: usage.Pending}
	if usage.Limited {
		limit := usage.Limit
		out.Limit = &limit
	}

	return out, nil
}

// enforceTenantUserCapTx serializes seat changes of one tenant and refuses one
// that would take the tenant over its active-user cap. Seats are active
// members plus pending, unexpired invitations; virtual users and keys do not
// count.
func enforceTenantUserCapTx(ctx context.Context, tx pgx.Tx, tenantID string, adding int) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('users_cap:' || $1))`, tenantID); err != nil {
		return err
	}

	usage, err := tenantSeatUsageTx(ctx, tx, tenantID)
	if err != nil {
		return err
	}

	if usage.Limited && !credits.UserCapAllows(usage.Used, adding, usage.Limit) {
		return billing.ErrPlanUserLimitReached
	}

	return nil
}

// tenantSeatUsageTx reads the tenant's seat count and effective user cap:
// the free seats every tenant gets plus every paid seat it holds.
func tenantSeatUsageTx(ctx context.Context, tx rowQuerier, tenantID string) (tenantSeatUsage, error) {
	var usage tenantSeatUsage

	err := tx.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(1) FROM users u WHERE u.tenant_id = t.tenant_id AND COALESCE(u.active, true)),
			(SELECT COUNT(1) FROM tenant_admin_invitations i
			 WHERE i.tenant_id = t.tenant_id AND i.status = 'pending' AND i.expires_at > NOW()),
			(SELECT COALESCE(SUM(b.quantity), 0)::int FROM tenant_seat_blocks b
			 WHERE b.tenant_id = t.tenant_id AND b.paid_through > NOW())
		FROM tenants t
		WHERE t.tenant_id = $1::uuid
	`, tenantID).Scan(&usage.Active, &usage.Pending, &usage.PaidSeats)
	if err != nil {
		return usage, err
	}

	usage.Used = usage.Active + usage.Pending
	usage.FreeLimit, usage.FreeLimited = seats.FreeSeatLimit, true
	usage.Limit, usage.Limited = usage.FreeLimit+usage.PaidSeats, true

	return usage, nil
}
