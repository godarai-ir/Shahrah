package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	supportdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/support"
	"github.com/jackc/pgx/v5"
)

func (s *SupportRepository) ListSupportTenants(ctx context.Context, query, status, plan string, limit, offset int) ([]platform.SupportTenantSummary, error) {
	query = strings.TrimSpace(query)
	status = strings.TrimSpace(strings.ToLower(status))
	plan = strings.TrimSpace(strings.ToLower(plan))

	rows, err := s.pool.Query(ctx, `
		SELECT t.tenant_id::text, t.slug, t.domain, t.status, t.created_at,
		       COALESCE(cp.slug, ''), COALESCE(cp.display_name, ''), COALESCE(ts.status, ''), ts.current_period_end,
		       ROUND(COALESCE(w.balance_irr, 0))::bigint, ROUND(COALESCE(w.pending_irr, 0))::bigint,
		       COALESCE(tuc.used, 0)::double precision
		FROM tenants t
		LEFT JOIN tenant_subscriptions ts ON ts.tenant_id = t.tenant_id AND ts.status IN ('trialing', 'active', 'past_due')
		LEFT JOIN commercial_plans cp ON cp.plan_id = ts.plan_id
		LEFT JOIN wallets w ON w.tenant_id = t.tenant_id
		LEFT JOIN LATERAL (
			SELECT used FROM tenant_usage_counters
			WHERE tenant_id = t.tenant_id AND feature_key = $1 AND window_end > NOW()
			ORDER BY window_start DESC LIMIT 1
		) tuc ON true
		WHERE ($2 = '' OR t.slug ILIKE '%' || $2 || '%' OR t.domain ILIKE '%' || $2 || '%' OR t.tenant_id::text = $2)
		  AND ($3 = '' OR LOWER(t.status) = $3)
		  AND ($4 = '' OR LOWER(COALESCE(cp.slug, '')) = $4)
		ORDER BY t.created_at DESC, t.tenant_id DESC
		LIMIT $5 OFFSET $6
	`, platform.FeatureGatewayRequestsMonthly, query, status, plan, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.SupportTenantSummary, 0, limit)

	for rows.Next() {
		var item platform.SupportTenantSummary
		if err := rows.Scan(&item.ID, &item.Slug, &item.Domain, &item.Status, &item.CreatedAt,
			&item.PlanSlug, &item.PlanDisplayName, &item.SubscriptionStatus, &item.PeriodEnd,
			&item.WalletBalanceMinor, &item.WalletPendingMinor, &item.UsageUsed); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *SupportRepository) GetSupportTenantSummary(ctx context.Context, tenantID string) (*platform.SupportTenantSummary, error) {
	var item platform.SupportTenantSummary

	err := s.pool.QueryRow(ctx, `
		SELECT t.tenant_id::text, t.slug, t.domain, t.status, t.created_at,
		       COALESCE(cp.slug, ''), COALESCE(cp.display_name, ''), COALESCE(ts.status, ''), ts.current_period_end,
		       ROUND(COALESCE(w.balance_irr, 0))::bigint, ROUND(COALESCE(w.pending_irr, 0))::bigint,
		       COALESCE(tuc.used, 0)::double precision
		FROM tenants t
		LEFT JOIN tenant_subscriptions ts ON ts.tenant_id = t.tenant_id AND ts.status IN ('trialing', 'active', 'past_due')
		LEFT JOIN commercial_plans cp ON cp.plan_id = ts.plan_id
		LEFT JOIN wallets w ON w.tenant_id = t.tenant_id
		LEFT JOIN LATERAL (
			SELECT used FROM tenant_usage_counters
			WHERE tenant_id = t.tenant_id AND feature_key = $1 AND window_end > NOW()
			ORDER BY window_start DESC LIMIT 1
		) tuc ON true
		WHERE t.tenant_id::text = $2
		LIMIT 1
	`, platform.FeatureGatewayRequestsMonthly, tenantID).Scan(
		&item.ID, &item.Slug, &item.Domain, &item.Status, &item.CreatedAt,
		&item.PlanSlug, &item.PlanDisplayName, &item.SubscriptionStatus, &item.PeriodEnd,
		&item.WalletBalanceMinor, &item.WalletPendingMinor, &item.UsageUsed,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, supportdomain.ErrTenantNotFound
	}

	return &item, err
}

func (s *SupportRepository) GetTenantOwnerContact(ctx context.Context, tenantID string) (*platform.TenantOwnerContact, error) {
	var owner platform.TenantOwnerContact

	err := s.pool.QueryRow(ctx, `
		SELECT u.user_id::text, COALESCE(u.user_alias, ''), COALESCE(u.email, ''), COALESCE(u.mobile, '')
		FROM users u
		WHERE u.tenant_id::text = $1
		  AND COALESCE(u.active, true) = true
		  AND (
			LOWER(COALESCE(u.role, '')) = 'tenant_admin'
			OR EXISTS (
				SELECT 1 FROM user_role_bindings urb
				JOIN roles r ON r.role_id = urb.role_id AND r.tenant_id = urb.tenant_id
				WHERE urb.tenant_id = u.tenant_id AND urb.user_id = u.user_id
				  AND urb.is_active = true AND LOWER(r.name) = 'tenant_admin'
			)
		  )
		ORDER BY CASE WHEN LOWER(COALESCE(u.role, '')) = 'tenant_admin' THEN 0 ELSE 1 END,
		         u.created_at ASC, u.user_id ASC
		LIMIT 1
	`, tenantID).Scan(&owner.UserID, &owner.Name, &owner.Email, &owner.Phone)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return &owner, err
}

func (s *SupportRepository) ListTenantEvents(ctx context.Context, tenantID string, limit, offset int) ([]platform.TenantEvent, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT event_id::text, tenant_id::text, event_type, severity,
		       COALESCE(actor_id, ''), COALESCE(actor_role, ''), COALESCE(reason, ''),
		       COALESCE(request_id, ''), payload, created_at
		FROM tenant_events WHERE tenant_id::text = $1
		ORDER BY created_at DESC, event_id DESC LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.TenantEvent, 0, limit)

	for rows.Next() {
		var (
			item platform.TenantEvent
			raw  []byte
		)
		if err := rows.Scan(&item.ID, &item.TenantID, &item.Type, &item.Severity, &item.ActorID,
			&item.ActorRole, &item.Reason, &item.RequestID, &raw, &item.CreatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(raw, &item.Payload); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *SupportRepository) AdjustTenantWallet(ctx context.Context, tenantID string, req platform.WalletAdjustmentRequest, actor platform.SupportActor) (*platform.WalletAdjustmentResult, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		existing                          platform.WalletAdjustmentResult
		existingDirection, existingReason string
		existingAmount                    int64
	)

	err = tx.QueryRow(ctx, `
		SELECT wa.direction, wa.amount_minor::bigint, wa.reason, wa.balance_before::bigint, wa.balance_after::bigint,
		       wt.transaction_id::text, wt.wallet_id::text, wt.tenant_id::text, wt.tx_type, wt.amount_irr,
		       wt.status, COALESCE(wt.description, ''), wt.created_at
		FROM wallet_manual_adjustments wa
		JOIN wallet_transactions wt ON wt.transaction_id = wa.wallet_transaction_id
		WHERE wa.tenant_id::text = $1 AND wa.idempotency_key = $2
	`, tenantID, req.IdempotencyKey).Scan(&existingDirection, &existingAmount, &existingReason,
		&existing.BalanceBefore, &existing.BalanceAfter, &existing.Transaction.ID, &existing.Transaction.WalletID,
		&existing.Transaction.TenantID, &existing.Transaction.Type, &existing.Transaction.AmountIRR,
		&existing.Transaction.Status, &existing.Transaction.Description, &existing.Transaction.CreatedAt)
	if err == nil {
		if existingDirection != req.Direction || existingAmount != req.AmountMinor || existingReason != req.Reason {
			return nil, supportdomain.ErrIdempotencyPayloadConflict
		}

		existing.Idempotent = true

		return &existing, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	metadata := map[string]interface{}{"reason": req.Reason, "external_reference": req.ExternalReference, "actor_id": actor.ID, "request_id": actor.RequestID}

	var result platform.WalletAdjustmentResult

	txType := "manual_adjustment"
	if req.Direction == "debit" {
		txType = "manual_debit"
	}

	walletTxn, err := appendWalletTransactionTx(ctx, tx, walletTransactionParams{
		TenantID:       tenantID,
		Type:           txType,
		AmountIRR:      req.AmountMinor,
		Status:         "successful",
		Description:    req.Reason,
		Metadata:       metadata,
		SourceType:     "manual_adjustment",
		SourceID:       req.ExternalReference,
		IdempotencyKey: "manual_adjustment:" + req.IdempotencyKey,
		CreatedBy:      actor.ID,
		CompletedAt:    time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}

	result.Transaction = *walletTxn
	result.Transaction.Metadata = metadata

	result.BalanceAfter = walletTxn.BalanceAfter
	if req.Direction == "debit" {
		result.BalanceBefore = walletTxn.BalanceAfter + req.AmountMinor
	} else {
		result.BalanceBefore = walletTxn.BalanceAfter - req.AmountMinor
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO wallet_manual_adjustments (
			tenant_id, wallet_transaction_id, idempotency_key, direction, amount_minor, reason,
			external_reference, actor_id, actor_role, request_id, balance_before, balance_after
		) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, NULLIF($9, ''), NULLIF($10, ''), $11, $12)
	`, tenantID, result.Transaction.ID, req.IdempotencyKey, req.Direction, req.AmountMinor, req.Reason,
		req.ExternalReference, actor.ID, actor.Role, actor.RequestID, result.BalanceBefore, result.BalanceAfter); err != nil {
		return nil, err
	}

	if err := insertTenantEventTx(ctx, tx, tenantID, "wallet.manual_"+req.Direction, "warning", actor, req.Reason, map[string]interface{}{
		"transaction_id": result.Transaction.ID, "amount_minor": req.AmountMinor, "balance_before": result.BalanceBefore, "balance_after": result.BalanceAfter,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, tenantID)

	return &result, nil
}

func (s *SupportRepository) ChangeTenantStatus(ctx context.Context, tenantID, target, reason string, actor platform.SupportActor) (*platform.SupportTenantSummary, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var current string
	if err := tx.QueryRow(ctx, `SELECT status FROM tenants WHERE tenant_id::text = $1 FOR UPDATE`, tenantID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, supportdomain.ErrTenantNotFound
		}

		return nil, err
	}

	valid := (current == "active" && target == "suspended") || (current == "suspended" && target == "active")
	if !valid {
		return nil, fmt.Errorf("%w: %s to %s", supportdomain.ErrInvalidTenantStatusChange, current, target)
	}

	if _, err := tx.Exec(ctx, `UPDATE tenants SET status = $2, updated_at = NOW() WHERE tenant_id::text = $1`, tenantID, target); err != nil {
		return nil, err
	}

	eventType := "tenant.suspended"

	severity := "critical"
	if target == "active" {
		eventType, severity = "tenant.reactivated", "warning"
	}

	if err := insertTenantEventTx(ctx, tx, tenantID, eventType, severity, actor, reason, map[string]interface{}{"status_before": current, "status_after": target}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateTenantRuntimeProfile(ctx, tenantID, eventType)

	return s.GetSupportTenantSummary(ctx, tenantID)
}

func (s *SupportRepository) SetTenantSubscriptionCancellation(ctx context.Context, tenantID string, cancel bool, reason string, actor platform.SupportActor) (*platform.TenantSubscription, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		subscriptionID string
		current        bool
	)
	if err := tx.QueryRow(ctx, `
		SELECT subscription_id::text, cancel_at_period_end
		FROM tenant_subscriptions
		WHERE tenant_id::text = $1 AND status IN ('trialing', 'active', 'past_due')
		FOR UPDATE
	`, tenantID).Scan(&subscriptionID, &current); err != nil {
		return nil, err
	}

	if current == cancel {
		return nil, supportdomain.ErrInvalidTenantStatusChange
	}

	if _, err := tx.Exec(ctx, `UPDATE tenant_subscriptions SET cancel_at_period_end = $2, updated_at = NOW() WHERE subscription_id::text = $1`, subscriptionID, cancel); err != nil {
		return nil, err
	}

	eventType := "subscription.cancellation_scheduled"

	legacyType := "cancellation_scheduled"
	if !cancel {
		eventType, legacyType = "subscription.cancellation_reversed", "cancellation_reversed"
	}

	payload := map[string]interface{}{"cancel_at_period_end": cancel, "reason": reason}

	raw := marshalJSONBestEffort(payload)
	if _, err := tx.Exec(ctx, `
		INSERT INTO subscription_events (tenant_id, subscription_id, event_type, actor, payload)
		VALUES ($1, $2, $3, $4, $5)
	`, tenantID, subscriptionID, legacyType, actor.ID, raw); err != nil {
		return nil, err
	}

	if err := insertTenantEventTx(ctx, tx, tenantID, eventType, "warning", actor, reason, payload); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateTenantEntitlements(ctx, tenantID)

	return s.GetTenantSubscription(ctx, tenantID)
}

func (s *SupportRepository) ListTenantSubscriptionPaymentIntents(ctx context.Context, tenantID string, limit, offset int) ([]platform.SubscriptionPaymentIntent, error) {
	rows, err := s.pool.Query(ctx, subscriptionPaymentIntentSelect+`
		WHERE spi.tenant_id::text = $1
		ORDER BY spi.created_at DESC, spi.payment_intent_id DESC
		LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.SubscriptionPaymentIntent, 0, limit)

	for rows.Next() {
		item, err := scanSubscriptionPaymentIntent(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func insertTenantEventTx(ctx context.Context, tx pgx.Tx, tenantID, eventType, severity string, actor platform.SupportActor, reason string, payload map[string]interface{}) error {
	raw := marshalJSONBestEffort(payload)
	_, err := tx.Exec(ctx, `
		INSERT INTO tenant_events (tenant_id, event_type, severity, actor_id, actor_role, reason, request_id, payload)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8)
	`, tenantID, eventType, severity, actor.ID, actor.Role, reason, actor.RequestID, raw)

	return err
}
