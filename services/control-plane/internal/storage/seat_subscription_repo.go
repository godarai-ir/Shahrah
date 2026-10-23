package storage

import (
	"context"
	"errors"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/jackc/pgx/v5"
)

// seatPlanSubscription is a new free or seats subscription row.
type seatPlanSubscription struct {
	TenantID    string
	PlanSlug    string
	Interval    string
	PeriodStart time.Time
	PeriodEnd   time.Time
	Price       map[string]interface{}
	Source      string
	Actor       string
}

// insertSeatPlanSubscriptionTx starts an active free or seats subscription and
// records a subscription_started event. The caller ends the tenant's current
// subscription first.
func insertSeatPlanSubscriptionTx(ctx context.Context, tx pgx.Tx, sub seatPlanSubscription) (string, error) {
	price := sub.Price
	if price == nil {
		price = map[string]interface{}{"currency": "IRR", "amount_minor": 0}
	}

	var subscriptionID string

	err := tx.QueryRow(ctx, `
		INSERT INTO tenant_subscriptions (
			tenant_id, plan_id, version_id, status, billing_interval,
			current_period_start, current_period_end, price_snapshot_json, metadata
		)
		SELECT $1::uuid, p.plan_id, pv.version_id, 'active', $3, $4, $5, $6, jsonb_build_object('source', $7::text)
		FROM commercial_plans p
		JOIN plan_versions pv ON pv.plan_id = p.plan_id AND pv.status = 'published'
		WHERE p.slug = $2
		RETURNING subscription_id::text
	`, sub.TenantID, sub.PlanSlug, seats.NormalizeInterval(sub.Interval), sub.PeriodStart, sub.PeriodEnd,
		mustJSON(price), sub.Source).Scan(&subscriptionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", billingdomain.ErrPlanNotFound
	}

	if err != nil {
		return "", err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO subscription_events (tenant_id, subscription_id, event_type, actor, payload)
		VALUES ($1::uuid, $2::uuid, 'subscription_started', $3, $4)
	`, sub.TenantID, subscriptionID, sub.Actor, mustJSON(map[string]interface{}{
		"plan_slug": sub.PlanSlug, "source": sub.Source, "period_end": sub.PeriodEnd,
	})); err != nil {
		return "", err
	}

	return subscriptionID, nil
}

// startFreeSubscriptionTx puts a tenant on the free plan from start on.
func startFreeSubscriptionTx(ctx context.Context, tx pgx.Tx, tenantID string, start time.Time, source, actor string) error {
	_, err := insertSeatPlanSubscriptionTx(ctx, tx, seatPlanSubscription{
		TenantID: tenantID, PlanSlug: seats.FreePlanSlug, Interval: "monthly",
		PeriodStart: start, PeriodEnd: seats.FreePeriodEnd, Source: source, Actor: actor,
	})

	return err
}

// startSeatSubscriptionForOrderTx starts the seat subscription when a paid
// seat purchase completes on a tenant that has none running: the purchase is
// the subscription, and its period is the order's.
func startSeatSubscriptionForOrderTx(ctx context.Context, tx pgx.Tx, order *platform.SeatOrder, now time.Time) error {
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('seat_subscription:' || $1))`, order.TenantID); err != nil {
		return err
	}

	sub, err := currentSeatSubscription(ctx, tx, order.TenantID)
	if err != nil {
		return err
	}

	if sub.running(now) {
		return nil
	}

	if sub.Found {
		if _, err = tx.Exec(ctx, `
			UPDATE tenant_subscriptions SET status = 'canceled', updated_at = NOW() WHERE subscription_id = $1::uuid
		`, sub.ID); err != nil {
			return err
		}
	}

	_, err = insertSeatPlanSubscriptionTx(ctx, tx, seatPlanSubscription{
		TenantID: order.TenantID, PlanSlug: seats.SeatsPlanSlug, Interval: order.BillingInterval,
		PeriodStart: order.PeriodStart, PeriodEnd: order.PeriodEnd,
		Price: map[string]interface{}{
			"currency": "IRR", "amount_minor": order.PriceIRR, "billing_interval": order.BillingInterval,
			"seat_order_id": order.ID,
		},
		Source: "seat_order", Actor: "seat_order",
	})

	return err
}
