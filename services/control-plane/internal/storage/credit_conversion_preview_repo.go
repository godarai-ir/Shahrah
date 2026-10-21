package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// creditPlanTarget is a tenant's current subscription and the published
// credit version of the same plan it moves to.
type creditPlanTarget struct {
	SubscriptionID  string
	PlanID          string
	PlanSlug        string
	VersionID       string
	BillingInterval string
	TargetVersionID string
	PriceIRR        *int64
	DiscountPercent float64
	MonthlyCredits  *int64
}

func loadCreditPlanTarget(ctx context.Context, q rowQuerier, tenantID string) (*creditPlanTarget, error) {
	target := &creditPlanTarget{}

	var (
		targetVersion *string
		price         *decimal.Decimal
		discount      *float64
		monthly       *string
	)

	err := q.QueryRow(ctx, `
		SELECT ts.subscription_id::text, ts.plan_id::text, p.slug, ts.version_id::text, ts.billing_interval,
		       target.version_id::text, pp.amount_minor, pp.discount_percent::double precision,
		       grant_entitlement.value_json::text
		FROM tenant_subscriptions ts
		JOIN commercial_plans p ON p.plan_id = ts.plan_id
		LEFT JOIN LATERAL (
			SELECT pv.version_id FROM plan_versions pv
			WHERE pv.plan_id = ts.plan_id AND pv.status = 'published'
			  AND EXISTS (SELECT 1 FROM plan_entitlements pe WHERE pe.version_id = pv.version_id AND pe.feature_key = $2)
			ORDER BY pv.version DESC LIMIT 1
		) target ON true
		LEFT JOIN plan_prices pp ON pp.version_id = target.version_id AND pp.billing_interval = ts.billing_interval
		LEFT JOIN plan_entitlements grant_entitlement
		       ON grant_entitlement.version_id = target.version_id AND grant_entitlement.feature_key = $2
		WHERE ts.tenant_id = $1::uuid AND ts.status IN ('trialing', 'active', 'past_due')
		ORDER BY ts.updated_at DESC
		LIMIT 1
	`, tenantID, credits.FeatureMonthlyGrant).Scan(
		&target.SubscriptionID, &target.PlanID, &target.PlanSlug, &target.VersionID, &target.BillingInterval,
		&targetVersion, &price, &discount, &monthly,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if targetVersion != nil {
		target.TargetVersionID = *targetVersion
	}

	if price != nil {
		amount := price.IntPart()
		target.PriceIRR = &amount
	}

	if discount != nil {
		target.DiscountPercent = *discount
	}

	if monthly != nil {
		if parsed, parseErr := decimal.NewFromString(strings.TrimSpace(*monthly)); parseErr == nil {
			amount := parsed.IntPart() * credits.MillicreditsPerCredit
			target.MonthlyCredits = &amount
		}
	}

	return target, nil
}

// ListCreditConversionPreviews dry-runs the conversion of IRR tenants, oldest
// first.
func (s *CreditLedgerRepository) ListCreditConversionPreviews(ctx context.Context, limit int) ([]platform.CreditConversionPreview, error) {
	if limit <= 0 || limit > maxCreditPreviewList {
		limit = maxCreditPreviewList
	}

	tenantIDs, err := s.irrTenantIDs(ctx, limit)
	if err != nil {
		return nil, err
	}

	previews := make([]platform.CreditConversionPreview, 0, len(tenantIDs))

	for _, tenantID := range tenantIDs {
		preview, previewErr := s.PreviewCreditConversion(ctx, tenantID)
		if previewErr != nil {
			return nil, fmt.Errorf("preview tenant %s: %w", tenantID, previewErr)
		}

		previews = append(previews, *preview)
	}

	return previews, nil
}

func (s *CreditLedgerRepository) irrTenantIDs(ctx context.Context, limit int) ([]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT tenant_id::text FROM tenants WHERE billing_unit = $1 ORDER BY created_at, tenant_id LIMIT $2
	`, credits.BillingUnitIRR, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0, limit)

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, rows.Err()
}

func (s *CreditLedgerRepository) countIRRTenants(ctx context.Context) (int, error) {
	var count int

	err := s.pool.QueryRow(ctx, `SELECT COUNT(1) FROM tenants WHERE billing_unit = $1`, credits.BillingUnitIRR).Scan(&count)

	return count, err
}

// PreviewCreditConversion dry-runs converting one tenant: what its IRR wallet
// becomes, which rules get a credit limit, which plan it moves to, and what
// needs attention first.
func (s *CreditLedgerRepository) PreviewCreditConversion(ctx context.Context, tenantID string) (*platform.CreditConversionPreview, error) {
	preview := &platform.CreditConversionPreview{TenantID: tenantID, Warnings: []string{}}

	var walletIRR int64

	if err := s.pool.QueryRow(ctx, `
		SELECT t.slug, t.billing_unit, COALESCE(ROUND(w.balance_irr)::bigint, 0)
		FROM tenants t LEFT JOIN wallets w ON w.tenant_id = t.tenant_id
		WHERE t.tenant_id = $1::uuid
	`, tenantID).Scan(&preview.TenantSlug, &preview.BillingUnit, &walletIRR); err != nil {
		return nil, notFoundOnNoRows(err)
	}

	price, err := activeCreditPrice(ctx, s.pool)
	if err != nil {
		return nil, err
	}

	preview.PriceTomanPerCredit = price.Toman
	preview.WalletIRR = walletIRR
	preview.UnsettledIRR = s.unsettledIRRWindowSpend(ctx, tenantID)

	creditsMC, residue := credits.ConvertIRRToCreditsMC(walletIRR-preview.UnsettledIRR, price.Toman)
	preview.Credits = credits.Format(credits.Millicredits(creditsMC))
	preview.ResidueIRR = residue

	if preview.UnsettledIRR > 0 {
		preview.Warnings = append(preview.Warnings, platform.CreditConversionWarningUnsettledUsage)
	}

	if preview.BillingUnit == credits.BillingUnitCredit {
		preview.Warnings = append(preview.Warnings, platform.CreditConversionWarningAlreadyConverted)

		return preview, nil
	}

	if countErr := s.pool.QueryRow(ctx, `
		SELECT COUNT(1) FROM budget_limit_rules WHERE tenant_id = $1::uuid AND limit_to > 0 AND limit_mc IS NULL
	`, tenantID).Scan(&preview.RulesToConvert); countErr != nil {
		return nil, countErr
	}

	if warnErr := s.addPendingPaymentWarnings(ctx, preview); warnErr != nil {
		return nil, warnErr
	}

	target, err := loadCreditPlanTarget(ctx, s.pool, tenantID)
	if err != nil {
		return nil, err
	}

	switch {
	case target == nil:
		preview.Warnings = append(preview.Warnings, platform.CreditConversionWarningNoSubscription)
	case target.TargetVersionID == "":
		preview.PlanSlug = target.PlanSlug
		preview.Warnings = append(preview.Warnings, platform.CreditConversionWarningNoCreditPlan)
	default:
		preview.PlanSlug, preview.BillingInterval, preview.NewPlanPriceIRR = target.PlanSlug, target.BillingInterval, target.PriceIRR
		if target.MonthlyCredits != nil {
			preview.NewPlanCredits = credits.Format(credits.Millicredits(*target.MonthlyCredits))
		}
	}

	return preview, nil
}

func (s *CreditLedgerRepository) addPendingPaymentWarnings(ctx context.Context, preview *platform.CreditConversionPreview) error {
	var pendingTopUps, pendingTransfers int

	if err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(1) FROM wallet_topup_intents WHERE tenant_id = $1::uuid AND UPPER(status) = 'PENDING'),
			(SELECT COUNT(1) FROM bank_transfer_requests WHERE tenant_id = $1::uuid AND status = 'pending_review' AND credit_pack_order_id IS NULL AND seat_order_id IS NULL)
	`, preview.TenantID).Scan(&pendingTopUps, &pendingTransfers); err != nil {
		return err
	}

	if pendingTopUps > 0 {
		preview.Warnings = append(preview.Warnings, platform.CreditConversionWarningPendingTopUp)
	}

	if pendingTransfers > 0 {
		preview.Warnings = append(preview.Warnings, platform.CreditConversionWarningPendingTransfer)
	}

	return nil
}

// unsettledIRRWindowSpend is IRR spend committed in the current month's wallet
// window and not yet debited from the wallet.
func (s *CreditLedgerRepository) unsettledIRRWindowSpend(ctx context.Context, tenantID string) int64 {
	inspector, ok := s.entitlementCache.(cache.BudgetWindowInspector)
	if !ok || s.ledger == nil {
		return 0
	}

	windowID, _ := currentBudgetWindowForTime(time.Now().UTC())

	snapshot, err := inspector.GetBudgetWindow(ctx, currentBudgetWindowKey(tenantID, windowID))
	if err != nil || snapshot == nil {
		return 0
	}

	settlement, err := s.ledger.getBudgetWindowSettlement(ctx, tenantID, windowID)
	if err != nil {
		return 0
	}

	return max(0, roundedIRR(snapshot.CommittedIRR)-settlement.SettledIRR)
}

// ListCreditConversions returns a tenant's conversions, newest first.
func (s *CreditLedgerRepository) ListCreditConversions(ctx context.Context, tenantID string) ([]platform.CreditConversion, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT conversion_id::text, tenant_id::text, sequence, direction, price_toman_per_credit, wallet_irr,
		       converted_irr, residue_irr, credits_mc, debt_mc, uncollected_irr, rules_converted,
		       COALESCE(plan_slug, ''), COALESCE(subscription_id::text, ''), actor, reason, created_at
		FROM credit_conversions WHERE tenant_id = $1::uuid ORDER BY sequence DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []platform.CreditConversion{}

	for rows.Next() {
		var (
			item              platform.CreditConversion
			creditsMC, debtMC int64
		)

		if err := rows.Scan(&item.ID, &item.TenantID, &item.Sequence, &item.Direction, &item.PriceTomanPerCredit,
			&item.WalletIRR, &item.ConvertedIRR, &item.ResidueIRR, &creditsMC, &debtMC, &item.UncollectedIRR,
			&item.RulesConverted, &item.PlanSlug, &item.SubscriptionID, &item.Actor, &item.Reason, &item.CreatedAt); err != nil {
			return nil, err
		}

		item.Credits = credits.Format(credits.Millicredits(creditsMC))
		item.Debt = credits.Format(credits.Millicredits(debtMC))
		out = append(out, item)
	}

	return out, rows.Err()
}
