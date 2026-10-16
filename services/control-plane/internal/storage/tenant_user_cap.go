package storage

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/jackc/pgx/v5"
)

// tenantSeatCountSQL counts a tenant's seats: active members plus pending,
// unexpired invitations.
func tenantSeatCountSQL(tenant string) string {
	return `((SELECT COUNT(1) FROM users u WHERE u.tenant_id = ` + tenant + ` AND COALESCE(u.active, true))
		+ (SELECT COUNT(1) FROM tenant_admin_invitations i
		   WHERE i.tenant_id = ` + tenant + ` AND i.status = 'pending' AND i.expires_at > NOW()))`
}

// tenantSeatUsage is the user-cap state of one tenant.
type tenantSeatUsage struct {
	Used    int
	Active  int
	Pending int
	Limit   int
	Limited bool
}

// rowQuerier is satisfied by both a pool and a transaction.
type rowQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// GetTenantSeatUsage reports the tenant's seats against its plan's user cap.
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
// that would take the tenant over its plan's active-user cap. Seats are active
// members plus pending, unexpired invitations; virtual users and keys do not
// count. A tenant without a current subscription is not capped here: the
// handlers already require one where it applies.
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

// tenantSeatUsageTx reads the tenant's seat count and effective user cap.
func tenantSeatUsageTx(ctx context.Context, tx rowQuerier, tenantID string) (tenantSeatUsage, error) {
	var (
		usage          tenantSeatUsage
		hasPlan        bool
		trialingCredit bool
		raw            string
		trialMaxUsers  int
	)

	err := tx.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(1) FROM users u WHERE u.tenant_id = t.tenant_id AND COALESCE(u.active, true)),
			(SELECT COUNT(1) FROM tenant_admin_invitations i
			 WHERE i.tenant_id = t.tenant_id AND i.status = 'pending' AND i.expires_at > NOW()),
			ts.version_id IS NOT NULL,
			COALESCE(ts.status = 'trialing' AND t.billing_unit = $3, false),
			COALESCE(
				(SELECT o.value_json::text FROM tenant_entitlement_overrides o
				 WHERE o.tenant_id = t.tenant_id AND o.feature_key = $2 AND o.active
				   AND (o.expires_at IS NULL OR o.expires_at > NOW())
				 LIMIT 1),
				(SELECT pe.value_json::text FROM plan_entitlements pe
				 WHERE pe.version_id = ts.version_id AND pe.feature_key = $2),
				(SELECT fc.default_value::text FROM feature_catalog fc WHERE fc.feature_key = $2),
				'null'),
			COALESCE((SELECT s.trial_max_users FROM commercial_signup_settings s WHERE s.settings_id = true), $4)
		FROM tenants t
		LEFT JOIN LATERAL (
			SELECT status, version_id FROM tenant_subscriptions
			WHERE tenant_id = t.tenant_id AND status IN ('trialing', 'active', 'past_due')
			ORDER BY updated_at DESC
			LIMIT 1
		) ts ON true
		WHERE t.tenant_id = $1::uuid
	`, tenantID, credits.FeatureUsersMax, credits.BillingUnitCredit, credits.TrialMaxUsers).Scan(
		&usage.Active, &usage.Pending, &hasPlan, &trialingCredit, &raw, &trialMaxUsers,
	)
	if err != nil {
		return usage, err
	}

	usage.Used = usage.Active + usage.Pending

	if hasPlan {
		usage.Limit, usage.Limited = credits.UserLimit(raw, trialingCredit, trialMaxUsers)
	}

	return usage, nil
}

// ensurePlanFitsSeats refuses moving a tenant to a plan version whose user cap
// is below the seats the tenant already uses.
func (s *SubscriptionRepository) ensurePlanFitsSeats(ctx context.Context, tenantID, versionID string) error {
	var (
		used int
		raw  string
	)

	err := s.pool.QueryRow(ctx, `
		SELECT
			`+tenantSeatCountSQL("$1::uuid")+`,
			COALESCE(
				(SELECT o.value_json::text FROM tenant_entitlement_overrides o
				 WHERE o.tenant_id = $1::uuid AND o.feature_key = $3 AND o.active
				   AND (o.expires_at IS NULL OR o.expires_at > NOW())
				 LIMIT 1),
				(SELECT pe.value_json::text FROM plan_entitlements pe
				 WHERE pe.version_id = $2::uuid AND pe.feature_key = $3),
				'null')
	`, tenantID, versionID, credits.FeatureUsersMax).Scan(&used, &raw)
	if err != nil {
		return err
	}

	limit, limited := credits.UserLimit(raw, false, 0)
	if limited && used > limit {
		return billing.ErrPlanUserLimitReached
	}

	return nil
}
