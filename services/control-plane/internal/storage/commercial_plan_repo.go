package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/seats"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/singleflight"
)

func (s *PlanCatalogRepository) ListFeatureCatalog(ctx context.Context) ([]platform.FeatureDefinition, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT feature_key, value_type, default_value, allow_override, COALESCE(category, ''), COALESCE(description, ''),
		       created_at, updated_at
		FROM feature_catalog
		ORDER BY COALESCE(category, ''), feature_key
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []platform.FeatureDefinition{}

	for rows.Next() {
		item, err := scanFeatureDefinition(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func (s *PlanCatalogRepository) ListPublicCommercialPlans(ctx context.Context) ([]platform.CommercialPlan, error) {
	return s.listCommercialPlans(ctx, true)
}

func (s *PlanCatalogRepository) ListPlatformCommercialPlans(ctx context.Context) ([]platform.CommercialPlan, error) {
	return s.listCommercialPlans(ctx, false)
}

func (s *PlanCatalogRepository) listCommercialPlans(ctx context.Context, publishedOnly bool) ([]platform.CommercialPlan, error) {
	conditions := []string{}
	if publishedOnly {
		conditions = append(conditions, "p.active = true", "pv.status = 'published'")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT p.plan_id::text, p.slug, p.display_name, COALESCE(p.description, ''), p.sort_order, p.active, p.created_at, p.updated_at,
		       pv.version_id::text, pv.version, pv.status, pv.effective_from, pv.published_at, pv.created_at, pv.updated_at
		FROM commercial_plans p
		JOIN plan_versions pv ON pv.plan_id = p.plan_id
		%s
		ORDER BY p.sort_order ASC, p.slug ASC, pv.version DESC
	`, whereClause))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := map[string]*platform.CommercialPlan{}
	order := []string{}

	for rows.Next() {
		var (
			plan    platform.CommercialPlan
			version platform.PlanVersion
		)
		if err := rows.Scan(
			&plan.ID, &plan.Slug, &plan.DisplayName, &plan.Description, &plan.SortOrder, &plan.Active, &plan.CreatedAt, &plan.UpdatedAt,
			&version.ID, &version.Version, &version.Status, &version.EffectiveFrom, &version.PublishedAt, &version.CreatedAt, &version.UpdatedAt,
		); err != nil {
			return nil, err
		}

		version.PlanID = plan.ID

		existing := byID[plan.ID]
		if existing == nil {
			planCopy := plan
			planCopy.Versions = []platform.PlanVersion{}
			byID[plan.ID] = &planCopy
			order = append(order, plan.ID)
			existing = &planCopy
		}

		existing.Versions = append(existing.Versions, version)
		if publishedOnly || version.Status == "published" {
			prices, err := s.listPlanPricesForVersion(ctx, version.ID)
			if err != nil {
				return nil, err
			}

			features, err := s.listPlanFeatureMapForVersion(ctx, version.ID)
			if err != nil {
				return nil, err
			}

			existing.Prices = prices
			existing.Features = features
		}
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, rowsErr
	}

	out := make([]platform.CommercialPlan, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}

	return out, nil
}

func (s *PlanCatalogRepository) CreateCommercialPlan(ctx context.Context, req platform.CreateCommercialPlanRequest) (*platform.CommercialPlan, error) {
	slug := normalizePlanSlug(req.Slug)

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = slug
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	sortOrder := req.SortOrder
	if sortOrder == 0 {
		sortOrder = 100
	}

	var item platform.CommercialPlan
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO commercial_plans (slug, display_name, description, sort_order, active, created_at, updated_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, NOW(), NOW())
		RETURNING plan_id::text, slug, display_name, COALESCE(description, ''), sort_order, active, created_at, updated_at
	`, slug, displayName, strings.TrimSpace(req.Description), sortOrder, active).Scan(
		&item.ID, &item.Slug, &item.DisplayName, &item.Description, &item.SortOrder, &item.Active, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *PlanCatalogRepository) UpdateCommercialPlan(ctx context.Context, planSlug string, req platform.UpdateCommercialPlanRequest) (*platform.CommercialPlan, error) {
	var item platform.CommercialPlan

	err := s.pool.QueryRow(ctx, `
		UPDATE commercial_plans
		SET display_name = COALESCE($2, display_name),
		    description = CASE WHEN $3::boolean THEN NULLIF($4, '') ELSE description END,
		    sort_order = COALESCE($5, sort_order),
		    active = COALESCE($6, active),
		    updated_at = NOW()
		WHERE slug = $1
		RETURNING plan_id::text, slug, display_name, COALESCE(description, ''), sort_order, active, created_at, updated_at
	`, normalizePlanSlug(planSlug), req.DisplayName, req.Description != nil, stringValue(req.Description), req.SortOrder, req.Active).Scan(
		&item.ID, &item.Slug, &item.DisplayName, &item.Description, &item.SortOrder, &item.Active, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, billingdomain.ErrPlanNotFound
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *PlanCatalogRepository) DeactivateCommercialPlan(ctx context.Context, planSlug string) error {
	result, err := s.pool.Exec(ctx, `UPDATE commercial_plans SET active = false, updated_at = NOW() WHERE slug = $1`, normalizePlanSlug(planSlug))
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return billingdomain.ErrPlanNotFound
	}

	return nil
}

func (s *PlanCatalogRepository) CreatePlanVersion(ctx context.Context, planSlug string, req platform.CreatePlanVersionRequest) (*platform.PlanVersion, error) {
	planID, err := s.planIDBySlug(ctx, planSlug)
	if err != nil {
		return nil, err
	}

	version := req.Version
	if version <= 0 {
		queryErr := s.pool.QueryRow(ctx, `SELECT COALESCE(MAX(version), 0) + 1 FROM plan_versions WHERE plan_id = $1::uuid`, planID).Scan(&version)
		if queryErr != nil {
			return nil, queryErr
		}
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var item platform.PlanVersion
	if err := tx.QueryRow(ctx, `
		INSERT INTO plan_versions (plan_id, version, status, effective_from, created_at, updated_at)
		VALUES ($1::uuid, $2, 'draft', $3, NOW(), NOW())
		RETURNING version_id::text, plan_id::text, version, status, effective_from, published_at, created_at, updated_at
	`, planID, version, req.EffectiveFrom).Scan(
		&item.ID, &item.PlanID, &item.Version, &item.Status, &item.EffectiveFrom, &item.PublishedAt, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if req.CopyFromVersion != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO plan_entitlements (version_id, feature_key, value_json)
			SELECT $1::uuid, feature_key, value_json
			FROM plan_entitlements pe
			JOIN plan_versions pv ON pv.version_id = pe.version_id
			WHERE pv.plan_id = $2::uuid AND pv.version = $3
			ON CONFLICT (version_id, feature_key) DO NOTHING
		`, item.ID, planID, *req.CopyFromVersion); err != nil {
			return nil, err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO plan_prices (version_id, billing_interval, currency, amount_minor, discount_percent)
			SELECT $1::uuid, billing_interval, currency, amount_minor, discount_percent
			FROM plan_prices pp
			JOIN plan_versions pv ON pv.version_id = pp.version_id
			WHERE pv.plan_id = $2::uuid AND pv.version = $3
			ON CONFLICT (version_id, billing_interval) DO NOTHING
		`, item.ID, planID, *req.CopyFromVersion); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *PlanCatalogRepository) PublishPlanVersion(ctx context.Context, planSlug string, version int) (*platform.PlanVersion, error) {
	planID, err := s.planIDBySlug(ctx, planSlug)
	if err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	defer rollbackTxBestEffort(ctx, tx)

	if _, err := tx.Exec(ctx, `
		UPDATE plan_versions
		SET status = 'archived', updated_at = NOW()
		WHERE plan_id = $1::uuid AND status = 'published'
	`, planID); err != nil {
		return nil, err
	}

	var item platform.PlanVersion
	if err := tx.QueryRow(ctx, `
		UPDATE plan_versions
		SET status = 'published', published_at = NOW(), updated_at = NOW()
		WHERE plan_id = $1::uuid AND version = $2
		RETURNING version_id::text, plan_id::text, version, status, effective_from, published_at, created_at, updated_at
	`, planID, version).Scan(
		&item.ID, &item.PlanID, &item.Version, &item.Status, &item.EffectiveFrom, &item.PublishedAt, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, billingdomain.ErrPlanVersionNotFound
		}

		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &item, nil
}

func (s *PlanCatalogRepository) UpsertPlanEntitlement(ctx context.Context, planSlug string, version int, req platform.UpsertPlanEntitlementRequest) (*platform.PlanEntitlement, error) {
	versionID, status, err := s.planVersionBySlugAndNumber(ctx, planSlug, version)
	if err != nil {
		return nil, err
	}

	if status != "draft" {
		return nil, billingdomain.ErrPlanVersionPublished
	}

	valueBytes, err := marshalJSONValue(req.Value)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		INSERT INTO plan_entitlements (version_id, feature_key, value_json, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, NOW(), NOW())
		ON CONFLICT (version_id, feature_key)
		DO UPDATE SET value_json = EXCLUDED.value_json, updated_at = NOW()
		RETURNING entitlement_id::text, version_id::text, feature_key, value_json, created_at, updated_at
	`, versionID, strings.TrimSpace(req.FeatureKey), valueBytes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apperrors.ErrNotFound
	}

	item, err := scanPlanEntitlement(rows)
	if err != nil {
		return nil, err
	}

	s.entitlements.invalidateEntitlementsForPlan(ctx, planSlug)

	return item, nil
}

func (s *PlanCatalogRepository) UpsertPlanPrice(ctx context.Context, planSlug string, version int, req platform.UpsertPlanPriceRequest) (*platform.PlanPrice, error) {
	versionID, status, err := s.planVersionBySlugAndNumber(ctx, planSlug, version)
	if err != nil {
		return nil, err
	}

	if status != "draft" {
		return nil, billingdomain.ErrPlanVersionPublished
	}

	currency := strings.TrimSpace(strings.ToUpper(req.Currency))
	if currency == "" {
		currency = "IRR"
	}

	rows, err := s.pool.Query(ctx, `
		INSERT INTO plan_prices (version_id, billing_interval, currency, amount_minor, discount_percent, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (version_id, billing_interval)
		DO UPDATE SET currency = EXCLUDED.currency, amount_minor = EXCLUDED.amount_minor,
		              discount_percent = EXCLUDED.discount_percent, updated_at = NOW()
		RETURNING price_id::text, version_id::text, billing_interval, currency, amount_minor, discount_percent, created_at, updated_at
	`, versionID, normalizeBillingInterval(req.BillingInterval), currency, req.AmountMinor, req.DiscountPercent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apperrors.ErrNotFound
	}

	return scanPlanPrice(rows)
}

func (s *SubscriptionRepository) GetTenantSubscription(ctx context.Context, tenantID string) (*platform.TenantSubscription, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ts.subscription_id::text, ts.tenant_id::text, ts.plan_id::text, p.slug, p.display_name,
		       ts.version_id::text, pv.version, ts.status, ts.billing_interval,
		       ts.current_period_start, ts.current_period_end, ts.cancel_at_period_end,
		       ts.trial_started_at, ts.trial_ends_at, ts.converted_at,
		       COALESCE(ts.pending_version_id::text, ''), ts.price_snapshot_json, ts.metadata,
		       ts.created_at, ts.updated_at
		FROM tenant_subscriptions ts
		JOIN commercial_plans p ON p.plan_id = ts.plan_id
		JOIN plan_versions pv ON pv.version_id = ts.version_id
		WHERE ts.tenant_id = $1::uuid
		ORDER BY CASE WHEN ts.status IN ('trialing', 'active', 'past_due') THEN 0 ELSE 1 END,
		         ts.updated_at DESC
		LIMIT 1
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, billingdomain.ErrSubscriptionNotFound
	}

	return scanTenantSubscription(rows)
}

// ChangeTenantSubscription is retired: tenants are on the free plan until a
// paid seat purchase starts their seat subscription.
func (s *SubscriptionRepository) ChangeTenantSubscription(_ context.Context, _ string, _ platform.ChangeTenantSubscriptionRequest) (*platform.TenantSubscription, error) {
	return nil, billingdomain.ErrPlansRetired
}

func subscriptionStatusIsCurrent(status string) bool {
	switch strings.TrimSpace(status) {
	case "trialing", "active", "past_due":
		return true
	default:
		return false
	}
}

func (s *SubscriptionRepository) UpsertTenantSubscription(ctx context.Context, req platform.UpsertTenantSubscriptionRequest, actor string) (*platform.TenantSubscription, error) {
	return s.upsertTenantSubscription(ctx, req, actor)
}

func (s *SubscriptionRepository) upsertTenantSubscription(ctx context.Context, req platform.UpsertTenantSubscriptionRequest, actor string) (*platform.TenantSubscription, error) {
	req.PlanSlug = normalizePlanSlug(req.PlanSlug)
	if req.PlanSlug == "" {
		req.PlanSlug = seats.FreePlanSlug
	}

	if req.PlanSlug != seats.FreePlanSlug && req.PlanSlug != seats.SeatsPlanSlug {
		return nil, billingdomain.ErrPlansRetired
	}

	interval := normalizeBillingInterval(req.BillingInterval)

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}

	target, err := s.catalog.publishedPlanBySlug(ctx, req.PlanSlug)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	periodStart := now
	if req.PeriodStart != nil {
		periodStart = *req.PeriodStart
	}

	periodEnd := addBillingInterval(periodStart, interval)
	if req.PeriodEnd != nil {
		periodEnd = *req.PeriodEnd
	}

	priceSnapshot, err := s.catalog.priceSnapshotForVersion(ctx, target.Versions[0].ID, interval)
	if err != nil {
		return nil, err
	}

	if req.Metadata == nil {
		req.Metadata = map[string]interface{}{}
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	if _, err := tx.Exec(ctx, `
		UPDATE tenant_subscriptions
		SET status = 'canceled', updated_at = NOW()
		WHERE tenant_id = $1::uuid AND status IN ('trialing', 'active', 'past_due')
	`, req.TenantID); err != nil {
		return nil, err
	}

	subscriptionID := uuid.NewString()
	if _, err := tx.Exec(ctx, `
		INSERT INTO tenant_subscriptions (
			subscription_id, tenant_id, plan_id, version_id, status, billing_interval,
			current_period_start, current_period_end, cancel_at_period_end, price_snapshot_json, metadata, created_at, updated_at
		) VALUES (
			$1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7, $8, false, $9, $10, NOW(), NOW()
		)
	`, subscriptionID, req.TenantID, target.ID, target.Versions[0].ID, status, interval, periodStart, periodEnd, mustJSON(priceSnapshot), mustJSON(req.Metadata)); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.entitlements.invalidateTenantEntitlements(ctx, req.TenantID)

	if eventErr := s.insertSubscriptionEvent(ctx, req.TenantID, subscriptionID, "subscription_changed", actor, map[string]interface{}{
		"plan_slug":        req.PlanSlug,
		"billing_interval": interval,
	}); eventErr != nil {
		return nil, eventErr
	}

	return s.GetTenantSubscription(ctx, req.TenantID)
}

func (s *EntitlementRepository) ResolveEffectiveEntitlements(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error) {
	base, err := s.resolveEffectiveEntitlementsBase(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return base, nil
}

// ResolveEffectiveEntitlementLimits returns subscription and feature limits.
func (s *EntitlementRepository) ResolveEffectiveEntitlementLimits(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementLimits, error) {
	ctx, cancel := withStoreTimeout(ctx, s.hotPathDBTimeout)
	defer cancel()

	base, err := s.resolveEffectiveEntitlementsBase(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if base == nil {
		return nil, fmt.Errorf("empty entitlement result for tenant %s", tenantID)
	}

	return &platform.EffectiveEntitlementLimits{
		TenantID:     base.TenantID,
		Subscription: cloneTenantSubscription(base.Subscription),
		Features:     cloneStringInterfaceMap(base.Features),
		ResolvedAt:   base.ResolvedAt,
	}, nil
}

func (s *EntitlementRepository) resolveEffectiveEntitlementsBase(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error) {
	if cached, ok := s.getCachedEntitlements(ctx, tenantID); ok {
		return cached, nil
	}

	key := entitlementCacheKey(tenantID)
	ch := s.entitlementLoadGroup.DoChan(key, func() (interface{}, error) {
		if cached, ok := s.getCachedEntitlements(ctx, tenantID); ok {
			return cached, nil
		}

		resolved, resolveErr := s.resolveEffectiveEntitlementsFromDB(ctx, tenantID)
		if resolveErr != nil {
			return nil, resolveErr
		}

		s.cacheEntitlements(ctx, resolved)

		return resolved, nil
	})

	var result singleflight.Result
	select {
	case result = <-ch:
	case <-ctx.Done():
		s.entitlementLoadGroup.Forget(key)
		return nil, ctx.Err()
	}

	value, err := result.Val, result.Err
	if err != nil {
		return nil, err
	}

	resolved, ok := value.(*platform.EffectiveEntitlementsResponse)
	if !ok || resolved == nil {
		return nil, fmt.Errorf("unexpected entitlement cache load result for tenant %s", tenantID)
	}

	return cloneEffectiveEntitlements(resolved), nil
}

func (s *EntitlementRepository) resolveEffectiveEntitlementsFromDB(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, error) {
	subscription, err := s.subscriptions.GetTenantSubscription(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		WITH base AS (
			SELECT fc.feature_key, fc.default_value AS value_json, 0 AS source
			FROM feature_catalog fc
		),
		plan_values AS (
			SELECT pe.feature_key, pe.value_json, 1 AS source
			FROM plan_entitlements pe
			WHERE pe.version_id = $2::uuid
		),
		override_values AS (
			SELECT o.feature_key, o.value_json, 2 AS source
			FROM tenant_entitlement_overrides o
			WHERE o.tenant_id = $1::uuid
			  AND o.active = true
			  AND (o.expires_at IS NULL OR o.expires_at > NOW())
		)
		SELECT feature_key, value_json, source FROM base
		UNION ALL
		SELECT feature_key, value_json, source FROM plan_values
		UNION ALL
		SELECT feature_key, value_json, source FROM override_values
		ORDER BY feature_key, source
	`, tenantID, subscription.VersionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make([]billingdomain.EntitlementValue, 0)

	for rows.Next() {
		var (
			key    string
			raw    []byte
			source uint8
		)

		scanErr := rows.Scan(&key, &raw, &source)
		if scanErr != nil {
			return nil, scanErr
		}

		value := parseJSONValue(raw)

		values = append(values, billingdomain.EntitlementValue{
			Key:    key,
			Value:  value,
			Source: billingdomain.EntitlementSource(source),
		})
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, rowsErr
	}

	features, overrides := billingdomain.ResolveEntitlements(values)

	resolved := &platform.EffectiveEntitlementsResponse{
		TenantID:     tenantID,
		Subscription: subscription,
		Features:     features,
		Overrides:    overrides,
		ResolvedAt:   time.Now().UTC(),
	}

	return resolved, nil
}

func (s *EntitlementRepository) getCachedEntitlements(ctx context.Context, tenantID string) (*platform.EffectiveEntitlementsResponse, bool) {
	if s.entitlementCache == nil {
		return nil, false
	}

	value, ok := s.entitlementCache.Get(ctx, entitlementCacheKey(tenantID))
	if !ok {
		return nil, false
	}

	var raw []byte

	switch cached := value.(type) {
	case string:
		raw = []byte(cached)
	case []byte:
		raw = cached
	default:
		return nil, false
	}

	var out platform.EffectiveEntitlementsResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		s.entitlementCache.Delete(ctx, entitlementCacheKey(tenantID))
		return nil, false
	}

	return cloneEffectiveEntitlements(&out), true
}

func (s *EntitlementRepository) cacheEntitlements(ctx context.Context, resolved *platform.EffectiveEntitlementsResponse) {
	if s.entitlementCache == nil || resolved == nil {
		return
	}

	resolvedCopy := cloneEffectiveEntitlements(resolved)

	raw, err := json.Marshal(resolvedCopy)
	if err != nil {
		return
	}

	s.entitlementCache.Set(ctx, entitlementCacheKey(resolved.TenantID), string(raw), s.entitlementCacheTTL)
}

func (s *EntitlementRepository) InvalidateEffectiveEntitlements(ctx context.Context, tenantID string) {
	s.invalidateTenantEntitlements(ctx, tenantID)
}

func (s *EntitlementRepository) invalidateTenantEntitlements(ctx context.Context, tenantID string) {
	if s.entitlementCache != nil && strings.TrimSpace(tenantID) != "" {
		s.entitlementCache.Delete(ctx, entitlementCacheKey(tenantID))
	}

	s.invalidateTenantRuntimeProfile(ctx, tenantID, "entitlements.updated")
}

func (s *EntitlementRepository) invalidateEntitlementsForPlan(ctx context.Context, planSlug string) {
	if s.entitlementCache == nil {
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT DISTINCT ts.tenant_id::text
		FROM tenant_subscriptions ts
		JOIN plan_versions pv ON pv.version_id = ts.version_id
		JOIN commercial_plans p ON p.plan_id = pv.plan_id
		WHERE p.slug = $1
	`, normalizePlanSlug(planSlug))
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tenantID string
		if err := rows.Scan(&tenantID); err == nil {
			s.invalidateTenantEntitlements(ctx, tenantID)
		}
	}
}

func cloneEffectiveEntitlements(in *platform.EffectiveEntitlementsResponse) *platform.EffectiveEntitlementsResponse {
	if in == nil {
		return nil
	}

	out := *in
	if in.Subscription != nil {
		out.Subscription = cloneTenantSubscription(in.Subscription)
	}

	out.Features = cloneStringInterfaceMap(in.Features)

	out.Overrides = cloneStringInterfaceMap(in.Overrides)

	return &out
}

func cloneTenantSubscription(in *platform.TenantSubscription) *platform.TenantSubscription {
	if in == nil {
		return nil
	}

	out := *in
	out.PriceSnapshot = cloneStringInterfaceMap(in.PriceSnapshot)
	out.Metadata = cloneStringInterfaceMap(in.Metadata)

	return &out
}

func cloneStringInterfaceMap(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}

	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		out[key] = cloneJSONValue(value)
	}

	return out
}

func cloneJSONValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[string]interface{}:
		return cloneStringInterfaceMap(typed)
	case []interface{}:
		out := make([]interface{}, len(typed))
		for i, item := range typed {
			out[i] = cloneJSONValue(item)
		}

		return out
	default:
		return typed
	}
}

func entitlementCacheKey(tenantID string) string {
	return "entitlements:tenant:" + strings.TrimSpace(tenantID)
}

func (s *EntitlementRepository) UpsertTenantEntitlementOverride(ctx context.Context, req platform.UpsertTenantEntitlementOverrideRequest) (*platform.TenantEntitlementOverride, error) {
	allowOverride, err := s.featureAllowsOverride(ctx, req.FeatureKey)
	if err != nil {
		return nil, err
	}

	if !allowOverride {
		return nil, billingdomain.ErrFeatureOverrideForbidden
	}

	valueBytes, err := marshalJSONValue(req.Value)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		INSERT INTO tenant_entitlement_overrides (
			tenant_id, feature_key, value_json, reason, active, expires_at, created_by, created_at, updated_at
		) VALUES (
			$1::uuid, $2, $3, NULLIF($4, ''), true, $5, NULLIF($6, ''), NOW(), NOW()
		)
		ON CONFLICT (tenant_id, feature_key) WHERE active = true
		DO UPDATE SET value_json = EXCLUDED.value_json,
		              reason = EXCLUDED.reason,
		              expires_at = EXCLUDED.expires_at,
		              created_by = EXCLUDED.created_by,
		              updated_at = NOW()
		RETURNING override_id::text, tenant_id::text, feature_key, value_json, COALESCE(reason, ''), active, expires_at,
		          COALESCE(created_by, ''), created_at, updated_at
	`, req.TenantID, strings.TrimSpace(req.FeatureKey), valueBytes, strings.TrimSpace(req.Reason), req.ExpiresAt, strings.TrimSpace(req.CreatedBy))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, apperrors.ErrNotFound
	}

	item, err := scanTenantEntitlementOverride(rows)
	if err != nil {
		return nil, err
	}

	s.invalidateTenantEntitlements(ctx, req.TenantID)

	if eventErr := s.subscriptions.insertSubscriptionEvent(ctx, req.TenantID, "", "entitlement_override_upserted", req.CreatedBy, map[string]interface{}{
		"feature_key": req.FeatureKey,
		"reason":      req.Reason,
	}); eventErr != nil {
		return nil, eventErr
	}

	return item, nil
}

func (s *EntitlementRepository) DeleteTenantEntitlementOverride(ctx context.Context, tenantID string, featureKey string, actor string) error {
	result, err := s.pool.Exec(ctx, `
		UPDATE tenant_entitlement_overrides
		SET active = false, updated_at = NOW()
		WHERE tenant_id = $1::uuid AND feature_key = $2 AND active = true
	`, tenantID, strings.TrimSpace(featureKey))
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}

	s.invalidateTenantEntitlements(ctx, tenantID)

	return s.subscriptions.insertSubscriptionEvent(ctx, tenantID, "", "entitlement_override_deleted", actor, map[string]interface{}{
		"feature_key": featureKey,
	})
}

// ProcessSubscriptionTransitions ends seat subscriptions whose period is over.
// A subscription whose seats were renewed moves on to its next period; any
// other one ends and the tenant goes back to the free plan, which never
// expires. Paid seats then lapse on their own (RunSeatLifecycle). Nothing is
// charged here: seats are always prepaid.
func (s *SubscriptionRepository) ProcessSubscriptionTransitions(ctx context.Context, now time.Time, limit int) (int, error) {
	ctx, cancel := withStoreTimeout(ctx, s.subscriptionDBTimeout)
	defer cancel()

	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}

	defer rollbackTxWithFreshTimeout(tx)

	var lockAcquired bool
	if lockErr := tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(hashtext($1))`, "subscription_lifecycle_processor").Scan(&lockAcquired); lockErr != nil {
		return 0, lockErr
	}

	if !lockAcquired {
		return 0, nil
	}

	items, err := endedSeatSubscriptionsTx(ctx, tx, now, limit)
	if err != nil {
		return 0, err
	}

	tenantIDs := make([]string, 0, len(items))
	for _, item := range items {
		if err := endSeatSubscriptionPeriodTx(ctx, tx, item, now); err != nil {
			return 0, err
		}

		tenantIDs = append(tenantIDs, item.tenantID)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}

	for _, tenantID := range tenantIDs {
		s.entitlements.invalidateTenantEntitlements(ctx, tenantID)
	}

	return len(items), nil
}

// endedSeatSubscription is a current, non-free subscription whose period is over.
type endedSeatSubscription struct {
	subscriptionID string
	tenantID       string
	interval       string
	periodEnd      time.Time
	paidThrough    time.Time
}

func endedSeatSubscriptionsTx(ctx context.Context, tx pgx.Tx, now time.Time, limit int) ([]endedSeatSubscription, error) {
	rows, err := tx.Query(ctx, `
		SELECT ts.subscription_id::text, ts.tenant_id::text, ts.billing_interval, ts.current_period_end,
		       COALESCE((SELECT MAX(b.paid_through) FROM tenant_seat_blocks b WHERE b.tenant_id = ts.tenant_id),
		                ts.current_period_end)
		FROM tenant_subscriptions ts
		JOIN commercial_plans p ON p.plan_id = ts.plan_id
		WHERE ts.status IN ('trialing', 'active', 'past_due') AND ts.current_period_end <= $1 AND p.slug <> $3
		ORDER BY ts.current_period_end ASC, ts.subscription_id ASC
		LIMIT $2
		FOR UPDATE OF ts SKIP LOCKED
	`, now, limit, seats.FreePlanSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []endedSeatSubscription{}

	for rows.Next() {
		var item endedSeatSubscription
		if err := rows.Scan(&item.subscriptionID, &item.tenantID, &item.interval, &item.periodEnd, &item.paidThrough); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func endSeatSubscriptionPeriodTx(ctx context.Context, tx pgx.Tx, item endedSeatSubscription, now time.Time) error {
	if next, renewed := seats.NextPeriod(item.periodEnd, item.paidThrough, item.interval); renewed {
		if _, err := tx.Exec(ctx, `
			UPDATE tenant_subscriptions
			SET current_period_start = current_period_end, current_period_end = $2, updated_at = NOW()
			WHERE subscription_id = $1::uuid
		`, item.subscriptionID, next); err != nil {
			return err
		}

		_, err := tx.Exec(ctx, `
			INSERT INTO subscription_events (tenant_id, subscription_id, event_type, actor, payload)
			VALUES ($1::uuid, $2::uuid, 'seat_subscription_renewed', 'system', $3)
		`, item.tenantID, item.subscriptionID, mustJSON(map[string]interface{}{"period_end": next}))

		return err
	}

	if _, err := tx.Exec(ctx, `UPDATE tenant_subscriptions SET status = 'expired', updated_at = NOW() WHERE subscription_id = $1::uuid`, item.subscriptionID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO subscription_events (tenant_id, subscription_id, event_type, actor, payload)
		VALUES ($1::uuid, $2::uuid, 'seat_subscription_ended', 'system', '{}'::jsonb)
	`, item.tenantID, item.subscriptionID); err != nil {
		return err
	}

	return startFreeSubscriptionTx(ctx, tx, item.tenantID, now, "seat_subscription_ended", "system")
}

func (s *PlanCatalogRepository) planIDBySlug(ctx context.Context, slug string) (string, error) {
	var id string
	if err := s.pool.QueryRow(ctx, `SELECT plan_id::text FROM commercial_plans WHERE slug = $1`, normalizePlanSlug(slug)).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", billingdomain.ErrPlanNotFound
		}

		return "", err
	}

	return id, nil
}

func (s *PlanCatalogRepository) planVersionBySlugAndNumber(ctx context.Context, slug string, version int) (string, string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT pv.version_id::text, pv.status
		FROM plan_versions pv
		JOIN commercial_plans p ON p.plan_id = pv.plan_id
		WHERE p.slug = $1 AND pv.version = $2
		LIMIT 1
	`, normalizePlanSlug(slug), version)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", "", billingdomain.ErrPlanVersionNotFound
	}

	var id, status string
	if err := rows.Scan(&id, &status); err != nil {
		return "", "", err
	}

	return id, status, nil
}

func (s *PlanCatalogRepository) publishedPlanBySlug(ctx context.Context, slug string) (*platform.CommercialPlan, error) {
	plans, err := s.ListPublicCommercialPlans(ctx)
	if err != nil {
		return nil, err
	}

	normalized := normalizePlanSlug(slug)

	for i := range plans {
		plan := &plans[i]
		if plan.Slug == normalized {
			planCopy := *plan
			return &planCopy, nil
		}
	}

	return nil, billingdomain.ErrPlanNotFound
}

func (s *PlanCatalogRepository) listPlanPricesForVersion(ctx context.Context, versionID string) ([]platform.PlanPrice, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT price_id::text, version_id::text, billing_interval, currency, amount_minor, discount_percent, created_at, updated_at
		FROM plan_prices
		WHERE version_id = $1::uuid
		ORDER BY CASE billing_interval WHEN 'monthly' THEN 1 WHEN 'quarterly' THEN 2 WHEN 'yearly' THEN 3 ELSE 4 END
	`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []platform.PlanPrice{}

	for rows.Next() {
		item, err := scanPlanPrice(rows)
		if err != nil {
			return nil, err
		}

		out = append(out, *item)
	}

	return out, rows.Err()
}

func (s *PlanCatalogRepository) listPlanFeatureMapForVersion(ctx context.Context, versionID string) (map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT feature_key, value_json
		FROM plan_entitlements
		WHERE version_id = $1::uuid
		ORDER BY feature_key
	`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]interface{}{}

	for rows.Next() {
		var (
			key string
			raw []byte
		)
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}

		out[key] = parseJSONValue(raw)
	}

	return out, rows.Err()
}

func (s *PlanCatalogRepository) priceSnapshotForVersion(ctx context.Context, versionID, interval string) (map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT billing_interval, currency, amount_minor, discount_percent
		FROM plan_prices
		WHERE version_id = $1::uuid AND billing_interval = $2
		LIMIT 1
	`, versionID, interval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return map[string]interface{}{
			"billing_interval": interval,
			"currency":         "IRR",
			"amount_minor":     0,
			"discount_percent": 0,
		}, nil
	}

	var (
		billingInterval, currency string
		amount                    *float64
		discount                  float64
	)
	if err := rows.Scan(&billingInterval, &currency, &amount, &discount); err != nil {
		return nil, err
	}

	out := map[string]interface{}{
		"billing_interval": billingInterval,
		"currency":         currency,
		"discount_percent": discount,
	}
	if amount != nil {
		out["amount_minor"] = *amount
	}

	return out, nil
}

func (s *EntitlementRepository) featureAllowsOverride(ctx context.Context, key string) (bool, error) {
	var allow bool
	if err := s.pool.QueryRow(ctx, `SELECT allow_override FROM feature_catalog WHERE feature_key = $1`, strings.TrimSpace(key)).Scan(&allow); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, billingdomain.ErrFeatureNotFound
		}

		return false, err
	}

	return allow, nil
}

func (s *SubscriptionRepository) insertSubscriptionEvent(ctx context.Context, tenantID, subscriptionID, eventType, actor string, payload map[string]interface{}) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO subscription_events (tenant_id, subscription_id, event_type, actor, payload, created_at)
		VALUES ($1::uuid, $2::uuid, $3, NULLIF($4, ''), $5, NOW())
	`, nullUUID(tenantID), nullUUID(subscriptionID), strings.TrimSpace(eventType), strings.TrimSpace(actor), mustJSON(payload))

	return err
}

func scanFeatureDefinition(rows pgx.Rows) (*platform.FeatureDefinition, error) {
	var (
		item platform.FeatureDefinition
		raw  []byte
	)
	if err := rows.Scan(&item.Key, &item.ValueType, &raw, &item.AllowOverride, &item.Category, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}

	item.DefaultValue = parseJSONValue(raw)

	return &item, nil
}

func scanPlanEntitlement(rows pgx.Rows) (*platform.PlanEntitlement, error) {
	var (
		item platform.PlanEntitlement
		raw  []byte
	)
	if err := rows.Scan(&item.ID, &item.VersionID, &item.FeatureKey, &raw, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}

	item.Value = parseJSONValue(raw)

	return &item, nil
}

func scanPlanPrice(rows pgx.Rows) (*platform.PlanPrice, error) {
	var (
		item   platform.PlanPrice
		amount *float64
	)
	if err := rows.Scan(&item.ID, &item.VersionID, &item.BillingInterval, &item.Currency, &amount, &item.DiscountPercent, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}

	item.AmountMinor = amount

	return &item, nil
}

func scanTenantSubscription(rows pgx.Rows) (*platform.TenantSubscription, error) {
	var (
		item        platform.TenantSubscription
		pending     string
		priceRaw    []byte
		metadataRaw []byte
	)
	if err := rows.Scan(
		&item.ID, &item.TenantID, &item.PlanID, &item.PlanSlug, &item.PlanDisplayName,
		&item.VersionID, &item.Version, &item.Status, &item.BillingInterval,
		&item.CurrentPeriodStart, &item.CurrentPeriodEnd, &item.CancelAtPeriodEnd,
		&item.TrialStartedAt, &item.TrialEndsAt, &item.ConvertedAt,
		&pending, &priceRaw, &metadataRaw, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.PendingVersionID = pending
	item.PriceSnapshot = parseJSONMap(priceRaw)
	item.Metadata = parseJSONMap(metadataRaw)

	return &item, nil
}

func scanTenantEntitlementOverride(rows pgx.Rows) (*platform.TenantEntitlementOverride, error) {
	var (
		item platform.TenantEntitlementOverride
		raw  []byte
	)
	if err := rows.Scan(
		&item.ID, &item.TenantID, &item.FeatureKey, &raw, &item.Reason, &item.Active,
		&item.ExpiresAt, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	item.Value = parseJSONValue(raw)

	return &item, nil
}

func normalizePlanSlug(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}

func normalizeBillingInterval(value string) string {
	return billingdomain.NormalizeBillingInterval(value)
}

func addBillingInterval(start time.Time, interval string) time.Time {
	return billingdomain.AddBillingInterval(start, interval)
}

func marshalJSONValue(value interface{}) ([]byte, error) {
	if value == nil {
		return []byte("null"), nil
	}

	return json.Marshal(value)
}

func parseJSONValue(raw []byte) interface{} {
	if len(raw) == 0 {
		return nil
	}

	var out interface{}

	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()

	if err := decoder.Decode(&out); err != nil {
		return nil
	}

	return normalizeJSONNumbers(out)
}

func parseJSONMap(raw []byte) map[string]interface{} {
	parsed := parseJSONValue(raw)
	if out, ok := parsed.(map[string]interface{}); ok {
		return out
	}

	return map[string]interface{}{}
}

func normalizeJSONNumbers(value interface{}) interface{} {
	switch v := value.(type) {
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}

		if f, err := v.Float64(); err == nil {
			return f
		}

		return string(v)
	case map[string]interface{}:
		for key, item := range v {
			v[key] = normalizeJSONNumbers(item)
		}

		return v
	case []interface{}:
		for i, item := range v {
			v[i] = normalizeJSONNumbers(item)
		}

		return v
	default:
		return v
	}
}

func mustJSON(value map[string]interface{}) []byte {
	if value == nil {
		value = map[string]interface{}{}
	}

	return marshalJSONBestEffort(value)
}
