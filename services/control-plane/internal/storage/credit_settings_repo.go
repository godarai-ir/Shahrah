package storage

import (
	"context"
	"errors"
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// resolveTenantFeatureJSON returns the effective entitlement JSON of one
// feature: an active tenant override, then the current plan version, then the
// catalog default.
func resolveTenantFeatureJSON(ctx context.Context, q rowQuerier, tenantID, featureKey string) (string, error) {
	var raw string

	err := q.QueryRow(ctx, `
		SELECT COALESCE(
			(SELECT o.value_json::text FROM tenant_entitlement_overrides o
			 WHERE o.tenant_id = $1::uuid AND o.feature_key = $2 AND o.active
			   AND (o.expires_at IS NULL OR o.expires_at > NOW())
			 LIMIT 1),
			(SELECT pe.value_json::text FROM tenant_subscriptions ts
			 JOIN plan_entitlements pe ON pe.version_id = ts.version_id AND pe.feature_key = $2
			 WHERE ts.tenant_id = $1::uuid AND ts.status IN ('trialing', 'active', 'past_due')
			 ORDER BY ts.updated_at DESC LIMIT 1),
			(SELECT fc.default_value::text FROM feature_catalog fc WHERE fc.feature_key = $2),
			'null')
	`, tenantID, featureKey).Scan(&raw)

	return raw, err
}

// GetCreditSettings returns the tenant's credit billing preferences.
func (s *CreditPackRepository) GetCreditSettings(ctx context.Context, tenantID string) (*platform.CreditSettings, error) {
	settings := &platform.CreditSettings{}

	var capMC *int64

	err := s.pool.QueryRow(ctx, `
		SELECT t.billing_unit, COALESCE(cs.overage_enabled, false), cs.overage_cap_mc
		FROM tenants t
		LEFT JOIN credit_tenant_settings cs ON cs.tenant_id = t.tenant_id
		WHERE t.tenant_id = $1::uuid
	`, tenantID).Scan(&settings.BillingUnit, &settings.OverageEnabled, &capMC)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, billingdomain.ErrCreditBillingNotEnabled
	}

	if err != nil {
		return nil, err
	}

	raw, err := resolveTenantFeatureJSON(ctx, s.pool, tenantID, credits.FeatureOverageAllowed)
	if err != nil {
		return nil, err
	}

	settings.OverageAllowed = strings.TrimSpace(raw) == "true"

	if capMC != nil {
		formatted := credits.Format(credits.Millicredits(*capMC))
		settings.OverageCapCredits = &formatted
	}

	return settings, nil
}

// UpdateCreditSettings changes the tenant's overage preferences. Overage can
// only be turned on for credit tenants whose plan allows it.
func (s *CreditPackRepository) UpdateCreditSettings(ctx context.Context, tenantID string, req platform.UpdateCreditSettingsRequest, actor platform.SupportActor) (*platform.CreditSettings, error) {
	current, err := s.GetCreditSettings(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	if current.BillingUnit != credits.BillingUnitCredit {
		return nil, billingdomain.ErrCreditBillingNotEnabled
	}

	enabled := current.OverageEnabled
	if req.OverageEnabled != nil {
		enabled = *req.OverageEnabled
	}

	if enabled && !current.OverageAllowed {
		return nil, billingdomain.ErrOverageNotAllowed
	}

	capValue, hasCap, err := overageCapMC(current.OverageCapCredits, req.OverageCapCredits)
	if err != nil {
		return nil, err
	}

	var capMC *int64
	if hasCap {
		capMC = &capValue
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO credit_tenant_settings (tenant_id, overage_enabled, overage_cap_mc, updated_by)
		VALUES ($1::uuid, $2, $3, NULLIF($4, ''))
		ON CONFLICT (tenant_id) DO UPDATE
		SET overage_enabled = EXCLUDED.overage_enabled, overage_cap_mc = EXCLUDED.overage_cap_mc,
		    updated_by = EXCLUDED.updated_by, updated_at = NOW()
	`, tenantID, enabled, capMC, actor.ID); err != nil {
		return nil, err
	}

	if err := insertFinancialAuditTx(ctx, tx, tenantID, "credit_settings.updated", actor, "credit_tenant_settings", tenantID, "",
		map[string]interface{}{"overage_enabled": enabled, "overage_cap_mc": capMC}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetCreditSettings(ctx, tenantID)
}

// overageCapMC applies a requested cap change: nil keeps the current cap, an
// empty string removes it, and anything else must be a positive credit amount
// with at most two decimals.
func overageCapMC(current, requested *string) (int64, bool, error) {
	value := current
	if requested != nil {
		value = requested
	}

	if value == nil || strings.TrimSpace(*value) == "" {
		return 0, false, nil
	}

	parsed, err := decimal.NewFromString(strings.TrimSpace(*value))
	if err != nil || !parsed.IsPositive() || !parsed.Equal(parsed.Round(2)) {
		return 0, false, billingdomain.ErrInvalidOverageCap
	}

	return parsed.Mul(decimal.NewFromInt(credits.MillicreditsPerCredit)).IntPart(), true, nil
}
