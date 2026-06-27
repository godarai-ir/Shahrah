package storage

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrSubscriptionPaymentIntentNotFound = errors.New("subscription payment intent not found")
	ErrPlanPriceUnavailable              = errors.New("plan price is not configured")
)

func (s *Store) PrepareTenantSubscriptionChange(ctx context.Context, tenantID string, req platform.ChangeTenantSubscriptionRequest, callbackURL string, returnURL string) (*platform.SubscriptionChangeResult, error) {
	req.PlanSlug = normalizePlanSlug(req.PlanSlug)
	if req.PlanSlug == "free" || req.PlanSlug == "" {
		req.PlanSlug = "developer"
	}

	req.BillingInterval = normalizeBillingInterval(req.BillingInterval)

	target, err := s.publishedPlanBySlug(ctx, req.PlanSlug)
	if err != nil {
		return nil, err
	}

	current, currentErr := s.GetTenantSubscription(ctx, tenantID)
	if currentErr != nil && !errors.Is(currentErr, ErrSubscriptionNotFound) {
		return nil, currentErr
	}

	if current != nil {
		var currentOrder int

		orderErr := s.pool.QueryRow(ctx, `SELECT sort_order FROM commercial_plans WHERE plan_id = $1::uuid`, current.PlanID).Scan(&currentOrder)
		if orderErr != nil {
			return nil, orderErr
		}

		if current.Status != "trialing" && target.SortOrder < currentOrder {
			subscription, subErr := s.ChangeTenantSubscription(ctx, tenantID, req)
			if subErr != nil {
				return nil, subErr
			}

			return &platform.SubscriptionChangeResult{Mode: "downgrade_scheduled", Subscription: subscription}, nil
		}
	}

	provider := strings.TrimSpace(strings.ToLower(req.Provider))
	if provider == "" {
		provider = "mock"
	}

	price, err := s.planPriceForVersion(ctx, target.Versions[0].ID, req.BillingInterval)
	if err != nil {
		return nil, err
	}

	if price.AmountMinor == nil {
		return nil, ErrPlanPriceUnavailable
	}

	if *price.AmountMinor == 0 {
		subscription, subErr := s.upsertTenantSubscription(ctx, platform.UpsertTenantSubscriptionRequest{
			TenantID: tenantID, PlanSlug: req.PlanSlug, BillingInterval: req.BillingInterval, Status: "active", Metadata: req.Metadata,
		}, "tenant_zero_price")
		if subErr != nil {
			return nil, subErr
		}

		return &platform.SubscriptionChangeResult{Mode: "applied", Subscription: subscription}, nil
	}

	priceSnapshot := map[string]interface{}{
		"billing_interval": price.BillingInterval,
		"currency":         price.Currency,
		"amount_minor":     *price.AmountMinor,
		"discount_percent": price.DiscountPercent,
	}

	intent, err := s.createSubscriptionPaymentIntent(ctx, tenantID, target, req.BillingInterval, provider, callbackURL, returnURL, priceSnapshot, req.Metadata)
	if err != nil {
		return nil, err
	}

	return &platform.SubscriptionChangeResult{Mode: "payment_required", PaymentIntent: intent}, nil
}

func (s *Store) createSubscriptionPaymentIntent(ctx context.Context, tenantID string, target *platform.CommercialPlan, interval string, provider string, callbackURL string, returnURL string, priceSnapshot map[string]interface{}, metadata map[string]interface{}) (*platform.SubscriptionPaymentIntent, error) {
	if metadata == nil {
		metadata = map[string]interface{}{}
	}

	amount, _ := numericFeatureLimit(priceSnapshot["amount_minor"])
	discount, _ := numericFeatureLimit(priceSnapshot["discount_percent"])

	currency, _ := priceSnapshot["currency"].(string)
	if currency == "" {
		currency = "IRR"
	}

	var publicID string
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO subscription_payment_intents (
			tenant_id, plan_id, version_id, billing_interval, provider, status,
			currency, amount_minor, discount_percent, price_snapshot_json,
			callback_url, return_url, metadata
		) VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, 'PENDING', $6, $7, $8, $9, $10, $11, $12)
		RETURNING public_id::text
	`, tenantID, target.ID, target.Versions[0].ID, interval, provider, currency, amount, discount, mustJSON(priceSnapshot), callbackURL, returnURL, mustJSON(metadata)).Scan(&publicID); err != nil {
		return nil, err
	}

	return s.GetSubscriptionPaymentIntentByPublicID(ctx, publicID)
}

func (s *Store) AttachSubscriptionPaymentGatewayData(ctx context.Context, publicID string, authority string, paymentURL string, raw map[string]interface{}) error {
	payload, _ := json.Marshal(raw)

	result, err := s.pool.Exec(ctx, `
		UPDATE subscription_payment_intents
		SET authority = $2, payment_url = $3,
		    raw_payload = COALESCE(raw_payload, '{}'::jsonb) || $4::jsonb,
		    updated_at = NOW()
		WHERE public_id::text = $1 AND status = 'PENDING'
	`, publicID, authority, paymentURL, payload)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSubscriptionPaymentIntentNotFound
	}

	return nil
}

func (s *Store) GetSubscriptionPaymentIntentByAuthority(ctx context.Context, authority string) (*platform.SubscriptionPaymentIntent, error) {
	row := s.pool.QueryRow(ctx, subscriptionPaymentIntentSelect+` WHERE spi.authority = $1 LIMIT 1`, authority)
	return scanSubscriptionPaymentIntent(row)
}

func (s *Store) GetSubscriptionPaymentIntentByPublicID(ctx context.Context, publicID string) (*platform.SubscriptionPaymentIntent, error) {
	row := s.pool.QueryRow(ctx, subscriptionPaymentIntentSelect+` WHERE spi.public_id::text = $1 LIMIT 1`, publicID)
	return scanSubscriptionPaymentIntent(row)
}

func (s *Store) CompleteSubscriptionPaymentByAuthority(ctx context.Context, authority string, verifyRaw map[string]interface{}, refID string, gatewayTxnID string, cardMask string, payerPhone string) (*platform.SubscriptionPaymentIntent, *platform.TenantSubscription, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	intent, err := scanSubscriptionPaymentIntent(tx.QueryRow(ctx, subscriptionPaymentIntentSelect+` WHERE spi.authority = $1 FOR UPDATE`, authority))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, false, ErrSubscriptionPaymentIntentNotFound
		}

		return nil, nil, false, err
	}

	if intent.Status == "COMPLETED" {
		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			return nil, nil, false, commitErr
		}

		subscription, subErr := s.GetTenantSubscription(ctx, intent.TenantID)

		return intent, subscription, true, subErr
	}

	if intent.Status != "PENDING" {
		return nil, nil, false, errors.New("subscription payment intent is not pending")
	}

	now := time.Now().UTC()

	var (
		currentID, currentPlanID, currentStatus string
		currentEnd                              time.Time
	)

	currentErr := tx.QueryRow(ctx, `
		SELECT subscription_id::text, plan_id::text, status, current_period_end
		FROM tenant_subscriptions
		WHERE tenant_id = $1::uuid AND status IN ('trialing', 'active', 'past_due')
		ORDER BY updated_at DESC LIMIT 1 FOR UPDATE
	`, intent.TenantID).Scan(&currentID, &currentPlanID, &currentStatus, &currentEnd)
	if currentErr != nil && !errors.Is(currentErr, pgx.ErrNoRows) {
		return nil, nil, false, currentErr
	}

	subscriptionID := currentID
	eventType := "subscription_renewed"

	if currentErr == nil && currentPlanID == intent.PlanID {
		if currentStatus == "trialing" {
			eventType = "trial_converted"
		}

		periodBase := currentEnd
		if periodBase.Before(now) {
			periodBase = now
		}

		_, execErr := tx.Exec(ctx, `
			UPDATE tenant_subscriptions
			SET version_id = $2::uuid, status = 'active', billing_interval = $3,
			    current_period_start = $6, current_period_end = $4, converted_at = COALESCE(converted_at, $7),
			    cancel_at_period_end = false, pending_version_id = NULL,
			    price_snapshot_json = $5, updated_at = NOW()
			WHERE subscription_id = $1::uuid
		`, currentID, intent.VersionID, intent.BillingInterval, addBillingInterval(periodBase, intent.BillingInterval), mustJSON(intent.PriceSnapshot), periodBase, now)
		if execErr != nil {
			return nil, nil, false, execErr
		}
	} else {
		if currentErr == nil {
			_, cancelErr := tx.Exec(ctx, `UPDATE tenant_subscriptions SET status = 'canceled', updated_at = NOW() WHERE subscription_id = $1::uuid`, currentID)
			if cancelErr != nil {
				return nil, nil, false, cancelErr
			}
		}

		subscriptionID = uuid.NewString()
		eventType = "subscription_activated"

		_, insertErr := tx.Exec(ctx, `
			INSERT INTO tenant_subscriptions (
				subscription_id, tenant_id, plan_id, version_id, status, billing_interval,
				current_period_start, current_period_end, cancel_at_period_end, price_snapshot_json, metadata
			) VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 'active', $5, $6, $7, false, $8, $9)
		`, subscriptionID, intent.TenantID, intent.PlanID, intent.VersionID, intent.BillingInterval, now, addBillingInterval(now, intent.BillingInterval), mustJSON(intent.PriceSnapshot), mustJSON(intent.Metadata))
		if insertErr != nil {
			return nil, nil, false, insertErr
		}
	}

	verifyPayload, _ := json.Marshal(verifyRaw)

	_, updateIntentErr := tx.Exec(ctx, `
		UPDATE subscription_payment_intents
		SET status = 'COMPLETED', ref_id = $2, gateway_txn_id = $3, card_mask = $4,
		    payer_phone = $5, raw_payload = COALESCE(raw_payload, '{}'::jsonb) || $6::jsonb,
		    completed_subscription_id = $7::uuid, completed_at = NOW(), updated_at = NOW()
		WHERE payment_intent_id = $1::uuid
	`, intent.ID, refID, gatewayTxnID, cardMask, payerPhone, verifyPayload, subscriptionID)
	if updateIntentErr != nil {
		return nil, nil, false, updateIntentErr
	}

	_, eventErr := tx.Exec(ctx, `
		INSERT INTO subscription_events (tenant_id, subscription_id, event_type, actor, payload)
		VALUES ($1::uuid, $2::uuid, $3, 'payment_callback', $4)
	`, intent.TenantID, subscriptionID, eventType, mustJSON(map[string]interface{}{"payment_intent_id": intent.ID, "ref_id": refID}))
	if eventErr != nil {
		return nil, nil, false, eventErr
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return nil, nil, false, commitErr
	}

	s.invalidateTenantEntitlements(ctx, intent.TenantID)

	completed, err := s.GetSubscriptionPaymentIntentByAuthority(ctx, authority)
	if err != nil {
		return nil, nil, false, err
	}

	subscription, err := s.GetTenantSubscription(ctx, intent.TenantID)

	return completed, subscription, false, err
}

func (s *Store) MarkSubscriptionPaymentFailedByAuthority(ctx context.Context, authority string, reason string, raw map[string]interface{}) error {
	payload, _ := json.Marshal(raw)

	result, err := s.pool.Exec(ctx, `
		UPDATE subscription_payment_intents
		SET status = CASE WHEN status = 'COMPLETED' THEN status ELSE 'FAILED' END,
		    failure_reason = $2, raw_payload = COALESCE(raw_payload, '{}'::jsonb) || $3::jsonb,
		    updated_at = NOW()
		WHERE authority = $1
	`, authority, reason, payload)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSubscriptionPaymentIntentNotFound
	}

	return nil
}

func (s *Store) planPriceForVersion(ctx context.Context, versionID string, interval string) (*platform.PlanPrice, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT price_id::text, version_id::text, billing_interval, currency,
		       amount_minor::double precision, discount_percent::double precision, created_at, updated_at
		FROM plan_prices WHERE version_id = $1::uuid AND billing_interval = $2
	`, versionID, normalizeBillingInterval(interval))

	var (
		item   platform.PlanPrice
		amount *float64
	)
	if err := row.Scan(&item.ID, &item.VersionID, &item.BillingInterval, &item.Currency, &amount, &item.DiscountPercent, &item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlanPriceUnavailable
		}

		return nil, err
	}

	item.AmountMinor = amount

	return &item, nil
}

const subscriptionPaymentIntentSelect = `
	SELECT spi.payment_intent_id::text, spi.public_id::text, spi.tenant_id::text, spi.plan_id::text,
	       p.slug, spi.version_id::text, pv.version, spi.billing_interval, spi.provider, spi.status,
	       spi.currency, spi.amount_minor::double precision, spi.discount_percent::double precision,
	       spi.price_snapshot_json, COALESCE(spi.authority, ''), COALESCE(spi.payment_url, ''),
	       spi.callback_url, spi.return_url, COALESCE(spi.ref_id, ''), COALESCE(spi.gateway_txn_id, ''),
	       COALESCE(spi.card_mask, ''), COALESCE(spi.payer_phone, ''), COALESCE(spi.failure_reason, ''),
	       spi.metadata, spi.raw_payload, COALESCE(spi.completed_subscription_id::text, ''),
	       spi.created_at, spi.updated_at, spi.completed_at
	FROM subscription_payment_intents spi
	JOIN commercial_plans p ON p.plan_id = spi.plan_id
	JOIN plan_versions pv ON pv.version_id = spi.version_id
`

type subscriptionPaymentRow interface {
	Scan(dest ...any) error
}

func scanSubscriptionPaymentIntent(row subscriptionPaymentRow) (*platform.SubscriptionPaymentIntent, error) {
	var (
		item                              platform.SubscriptionPaymentIntent
		priceRaw, metadataRaw, rawPayload []byte
	)
	if err := row.Scan(
		&item.ID, &item.PublicID, &item.TenantID, &item.PlanID, &item.PlanSlug,
		&item.VersionID, &item.Version, &item.BillingInterval, &item.Provider, &item.Status,
		&item.Currency, &item.AmountMinor, &item.DiscountPercent, &priceRaw,
		&item.Authority, &item.PaymentURL, &item.CallbackURL, &item.ReturnURL,
		&item.RefID, &item.GatewayTxnID, &item.CardMask, &item.PayerPhone, &item.FailureReason,
		&metadataRaw, &rawPayload, &item.CompletedSubscriptionID, &item.CreatedAt, &item.UpdatedAt, &item.CompletedAt,
	); err != nil {
		return nil, err
	}

	item.PriceSnapshot = parseJSONMap(priceRaw)
	item.Metadata = parseJSONMap(metadataRaw)
	item.RawPayload = parseJSONMap(rawPayload)

	return &item, nil
}
