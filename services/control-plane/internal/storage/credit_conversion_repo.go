package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/Delpak/shahrah/services/control-plane/internal/apperrors"
	"github.com/Delpak/shahrah/services/control-plane/internal/cache"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// Wallet transaction types of the credit cutover.
const (
	walletTxCreditConversionOut = "credit_conversion_out"
	walletTxCreditConversionIn  = "credit_conversion_in"
)

const (
	defaultCreditFleetBatch = 20
	maxCreditFleetBatch     = 100
	maxCreditPreviewList    = 500
)

// creditPrice is the active credit list price.
type creditPrice struct {
	Toman      int64
	RevisionID *string
}

func activeCreditPrice(ctx context.Context, q rowQuerier) (creditPrice, error) {
	price := creditPrice{}

	err := q.QueryRow(ctx, `
		SELECT price_toman_per_credit, revision_id::text FROM credit_price_revisions WHERE status = 'active'
	`).Scan(&price.Toman, &price.RevisionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return creditPrice{Toman: credits.ListPriceTomanPerCredit}, nil
	}

	return price, err
}

// settleIRRWindows debits IRR spend that is committed but not yet settled, for
// this and last month, so it is paid before the wallet is converted. After
// the cutover the wallet settlement no longer debits the tenant.
func (s *CreditLedgerRepository) settleIRRWindows(ctx context.Context, tenantID string) error {
	inspector, ok := s.entitlementCache.(cache.BudgetWindowInspector)
	if !ok || s.ledger == nil {
		return nil
	}

	now := time.Now().UTC()
	for _, at := range []time.Time{now.AddDate(0, -1, 0), now} {
		windowID, _ := currentBudgetWindowForTime(at)

		snapshot, err := inspector.GetBudgetWindow(ctx, currentBudgetWindowKey(tenantID, windowID))
		if err != nil {
			return err
		}

		if snapshot == nil {
			continue
		}

		if err := s.ledger.ApplyWalletBudgetSettlement(ctx, *snapshot); err != nil {
			return fmt.Errorf("settle IRR window %s: %w", windowID, err)
		}
	}

	return nil
}

// ConvertTenantToCredits converts a tenant to credit billing in one
// transaction: its IRR wallet becomes a migration lot that never expires (the
// residue below one credit stays in the wallet), each IRR budget limit gets a
// credit limit rounded up, its subscription moves to the published credit
// version of the same plan, and its billing unit becomes credit.
//
// A tenant already billed in credits only has its wallet converted, which
// picks up an IRR top-up that completed after the cutover.
func (s *CreditLedgerRepository) ConvertTenantToCredits(ctx context.Context, tenantID, reason string, actor platform.SupportActor) (*platform.CreditConversion, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, billingdomain.ErrDecisionReasonRequired
	}

	if err := s.settleIRRWindows(ctx, tenantID); err != nil {
		return nil, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	conversion, err := s.convertTenantTx(ctx, tx, tenantID, reason, actor, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.afterConversion(ctx, tenantID, "credit.converted")

	return conversion, nil
}

func (s *CreditLedgerRepository) convertTenantTx(ctx context.Context, tx pgx.Tx, tenantID, reason string, actor platform.SupportActor, now time.Time) (*platform.CreditConversion, error) {
	unit, sequence, err := lockTenantForConversionTx(ctx, tx, tenantID)
	if err != nil {
		return nil, err
	}

	price, err := activeCreditPrice(ctx, tx)
	if err != nil {
		return nil, err
	}

	var walletIRR int64
	if walletErr := tx.QueryRow(ctx, `
		SELECT COALESCE(ROUND(balance_irr)::bigint, 0) FROM wallets WHERE tenant_id = $1::uuid FOR UPDATE
	`, tenantID).Scan(&walletIRR); walletErr != nil && !errors.Is(walletErr, pgx.ErrNoRows) {
		return nil, walletErr
	}

	creditsMC, residue := credits.ConvertIRRToCreditsMC(walletIRR, price.Toman)
	if unit == credits.BillingUnitCredit && creditsMC == 0 {
		return nil, billingdomain.ErrCreditConversionNothingToDo
	}

	conversion := &platform.CreditConversion{
		ID: uuid.NewString(), TenantID: tenantID, Sequence: sequence, Direction: credits.ConversionToCredit,
		PriceTomanPerCredit: price.Toman, WalletIRR: walletIRR, ConvertedIRR: walletIRR - residue, ResidueIRR: residue,
		Credits: credits.Format(credits.Millicredits(creditsMC)), Debt: credits.Format(0),
		Actor: actor.ID, Reason: reason, CreatedAt: now,
	}

	if walletErr := convertWalletToCreditsTx(ctx, tx, conversion, creditsMC, now); walletErr != nil {
		return nil, walletErr
	}

	var previousSubscriptionID string

	if unit == credits.BillingUnitIRR {
		if conversion.RulesConverted, err = convertBudgetRuleLimitsTx(ctx, tx, tenantID, price.Toman); err != nil {
			return nil, err
		}

		if previousSubscriptionID, err = moveToCreditPlanTx(ctx, tx, tenantID, conversion, now); err != nil {
			return nil, err
		}

		if _, err := tx.Exec(ctx, `UPDATE tenants SET billing_unit = $2 WHERE tenant_id = $1::uuid`, tenantID, credits.BillingUnitCredit); err != nil {
			return nil, err
		}
	}

	if err := insertCreditConversionTx(ctx, tx, conversion, creditsMC, 0, price.RevisionID, previousSubscriptionID); err != nil {
		return nil, err
	}

	return conversion, insertFinancialAuditTx(ctx, tx, tenantID, "credit.converted", actor, "credit_conversion", conversion.ID,
		conversion.ID, conversionAuditPayload(conversion))
}

func lockTenantForConversionTx(ctx context.Context, tx pgx.Tx, tenantID string) (string, int, error) {
	var unit string
	if err := tx.QueryRow(ctx, `SELECT billing_unit FROM tenants WHERE tenant_id = $1::uuid FOR UPDATE`, tenantID).Scan(&unit); err != nil {
		return "", 0, notFoundOnNoRows(err)
	}

	var sequence int
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(sequence), 0) + 1 FROM credit_conversions WHERE tenant_id = $1::uuid
	`, tenantID).Scan(&sequence); err != nil {
		return "", 0, err
	}

	return unit, sequence, nil
}

func convertWalletToCreditsTx(ctx context.Context, tx pgx.Tx, conversion *platform.CreditConversion, creditsMC int64, now time.Time) error {
	if creditsMC <= 0 {
		return nil
	}

	if _, err := grantCreditsTx(ctx, tx, CreditGrant{
		TenantID: conversion.TenantID, Kind: credits.EntryConvert, Source: credits.LotSourceMigration,
		AmountMC: creditsMC, IdempotencyKey: credits.ConvertKey(conversion.TenantID, conversion.Sequence, "lot"),
		OccurredAt: now, Reason: conversion.Reason, Actor: conversion.Actor,
		SourceRef: map[string]interface{}{"conversion_id": conversion.ID, "price_toman_per_credit": conversion.PriceTomanPerCredit},
	}); err != nil {
		return err
	}

	_, err := appendWalletTransactionTx(ctx, tx, walletTransactionParams{
		TenantID: conversion.TenantID, Type: walletTxCreditConversionOut, AmountIRR: conversion.ConvertedIRR,
		Description: "converted to credits", SourceType: "credit_conversion", SourceID: conversion.ID,
		IdempotencyKey: credits.ConvertKey(conversion.TenantID, conversion.Sequence, "wallet"),
		Metadata:       map[string]interface{}{"credits": conversion.Credits, "price_toman_per_credit": conversion.PriceTomanPerCredit},
		CreatedBy:      conversion.Actor, CompletedAt: now,
	})

	return err
}

// convertBudgetRuleLimitsTx gives every IRR-limited rule a credit limit,
// rounded up so no rule gets stricter.
func convertBudgetRuleLimitsTx(ctx context.Context, tx pgx.Tx, tenantID string, priceToman int64) (int, error) {
	rows, err := tx.Query(ctx, `
		SELECT id::text, limit_to::text FROM budget_limit_rules
		WHERE tenant_id = $1::uuid AND limit_to > 0 AND limit_mc IS NULL
		FOR UPDATE
	`, tenantID)
	if err != nil {
		return 0, err
	}

	type ruleLimit struct {
		id      string
		limitMC int64
	}

	var limits []ruleLimit

	for rows.Next() {
		var id, raw string
		if err := rows.Scan(&id, &raw); err != nil {
			rows.Close()

			return 0, err
		}

		limitIRR, parseErr := decimal.NewFromString(raw)
		if parseErr != nil {
			rows.Close()

			return 0, parseErr
		}

		limits = append(limits, ruleLimit{id: id, limitMC: credits.ConvertIRRLimitToCreditsMC(limitIRR, priceToman)})
	}

	rows.Close()

	if err := rows.Err(); err != nil {
		return 0, err
	}

	for _, limit := range limits {
		if _, err := tx.Exec(ctx, `UPDATE budget_limit_rules SET limit_mc = $2, updated_at = NOW() WHERE id = $1::uuid`, limit.id, limit.limitMC); err != nil {
			return 0, err
		}
	}

	return len(limits), nil
}

// moveToCreditPlanTx moves the tenant's subscription to the published credit
// version of its plan (product decision P4): a new active period starts now,
// free, and the new price applies from the next renewal. It returns the
// replaced subscription id, or "" when the tenant has nothing to move.
func moveToCreditPlanTx(ctx context.Context, tx pgx.Tx, tenantID string, conversion *platform.CreditConversion, now time.Time) (string, error) {
	target, err := loadCreditPlanTarget(ctx, tx, tenantID)
	if err != nil || target == nil || target.TargetVersionID == "" {
		return "", err
	}

	conversion.PlanSlug = target.PlanSlug

	if target.TargetVersionID == target.VersionID {
		conversion.SubscriptionID = target.SubscriptionID

		return "", nil
	}

	var amount interface{}
	if target.PriceIRR != nil {
		amount = *target.PriceIRR
	}

	priceSnapshot := map[string]interface{}{
		"billing_interval": target.BillingInterval, "currency": "IRR", "amount_minor": amount, "discount_percent": target.DiscountPercent,
	}

	if _, err := tx.Exec(ctx, `UPDATE tenant_subscriptions SET status = 'canceled', updated_at = NOW() WHERE subscription_id = $1::uuid`, target.SubscriptionID); err != nil {
		return "", err
	}

	conversion.SubscriptionID = uuid.NewString()
	if _, err := tx.Exec(ctx, `
		INSERT INTO tenant_subscriptions (
			subscription_id, tenant_id, plan_id, version_id, status, billing_interval,
			current_period_start, current_period_end, cancel_at_period_end, price_snapshot_json, metadata
		) VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, 'active', $5, $6, $7, false, $8, $9)
	`, conversion.SubscriptionID, tenantID, target.PlanID, target.TargetVersionID, target.BillingInterval,
		now, addBillingInterval(now, target.BillingInterval), mustJSON(priceSnapshot),
		mustJSON(map[string]interface{}{"source": "credit_cutover", "previous_subscription_id": target.SubscriptionID})); err != nil {
		return "", err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO subscription_events (tenant_id, subscription_id, event_type, actor, payload)
		VALUES ($1::uuid, $2::uuid, 'credit_cutover', $3, $4)
	`, tenantID, conversion.SubscriptionID, conversion.Actor, mustJSON(map[string]interface{}{
		"previous_subscription_id": target.SubscriptionID, "conversion_id": conversion.ID,
	})); err != nil {
		return "", err
	}

	return target.SubscriptionID, nil
}

func insertCreditConversionTx(ctx context.Context, tx pgx.Tx, conversion *platform.CreditConversion, creditsMC, debtMC int64, revisionID *string, previousSubscriptionID string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO credit_conversions (
			conversion_id, tenant_id, sequence, direction, price_toman_per_credit, price_revision_id,
			wallet_irr, converted_irr, residue_irr, credits_mc, debt_mc, uncollected_irr, rules_converted,
			previous_subscription_id, subscription_id, plan_slug, actor, reason, created_at
		) VALUES (
			$1::uuid, $2::uuid, $3, $4, $5, $6::uuid, $7, $8, $9, $10, $11, $12, $13,
			NULLIF($14, '')::uuid, NULLIF($15, '')::uuid, NULLIF($16, ''), $17, $18, $19
		)
	`, conversion.ID, conversion.TenantID, conversion.Sequence, conversion.Direction, conversion.PriceTomanPerCredit, revisionID,
		conversion.WalletIRR, conversion.ConvertedIRR, conversion.ResidueIRR, creditsMC, debtMC, conversion.UncollectedIRR,
		conversion.RulesConverted, previousSubscriptionID, conversion.SubscriptionID, conversion.PlanSlug,
		conversion.Actor, conversion.Reason, conversion.CreatedAt)

	return err
}

func conversionAuditPayload(conversion *platform.CreditConversion) map[string]interface{} {
	return map[string]interface{}{
		"direction": conversion.Direction, "sequence": conversion.Sequence, "credits": conversion.Credits, "debt": conversion.Debt,
		"wallet_irr": conversion.WalletIRR, "converted_irr": conversion.ConvertedIRR, "residue_irr": conversion.ResidueIRR,
		"uncollected_irr": conversion.UncollectedIRR, "price_toman_per_credit": conversion.PriceTomanPerCredit,
		"rules_converted": conversion.RulesConverted, "plan_slug": conversion.PlanSlug, "reason": conversion.Reason,
	}
}

// afterConversion refreshes every cache that holds the tenant's billing state.
func (s *CreditLedgerRepository) afterConversion(ctx context.Context, tenantID, reason string) {
	if s.entitlementCache != nil {
		s.entitlementCache.Delete(ctx, entitlementCacheKey(tenantID))
	}

	if s.ledger != nil {
		s.ledger.invalidateWalletCache(ctx, tenantID)
	}

	s.invalidateCreditTenant(ctx, tenantID, reason)
}

// ConvertCreditFleet converts the next IRR tenants, oldest first. Each tenant
// is its own transaction, so one failure stops only that tenant.
func (s *CreditLedgerRepository) ConvertCreditFleet(ctx context.Context, req platform.CreditFleetConversionRequest, actor platform.SupportActor) (*platform.CreditFleetConversionResult, error) {
	if strings.TrimSpace(req.Reason) == "" {
		return nil, billingdomain.ErrDecisionReasonRequired
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultCreditFleetBatch
	}

	limit = min(limit, maxCreditFleetBatch)

	tenantIDs, err := s.irrTenantIDs(ctx, limit)
	if err != nil {
		return nil, err
	}

	result := &platform.CreditFleetConversionResult{Converted: []platform.CreditConversion{}, Failed: []platform.CreditConversionFailure{}}

	for _, tenantID := range tenantIDs {
		conversion, convertErr := s.ConvertTenantToCredits(ctx, tenantID, req.Reason, actor)
		if convertErr != nil {
			if s.logger != nil {
				s.logger.Warn("credit fleet conversion failed", zap.String("tenant_id", tenantID), zap.Error(convertErr))
			}

			result.Failed = append(result.Failed, platform.CreditConversionFailure{
				TenantID: tenantID, TenantSlug: s.tenantSlug(ctx, tenantID), Error: creditConversionFailureCode(convertErr),
			})

			continue
		}

		result.Converted = append(result.Converted, *conversion)
	}

	if result.Remaining, err = s.countIRRTenants(ctx); err != nil {
		return nil, err
	}

	return result, nil
}

// refuseCreditTenantWalletTopUp stops IRR wallet top-ups for credit tenants:
// their wallet is no longer spent, so the money would sit there unused.
func refuseCreditTenantWalletTopUp(ctx context.Context, q rowQuerier, tenantID string) error {
	var unit string
	if err := q.QueryRow(ctx, `SELECT billing_unit FROM tenants WHERE tenant_id::text = $1`, strings.TrimSpace(tenantID)).Scan(&unit); err != nil {
		return notFoundOnNoRows(err)
	}

	if unit == credits.BillingUnitCredit {
		return billingdomain.ErrWalletTopUpForCreditTenant
	}

	return nil
}

func notFoundOnNoRows(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrNotFound
	}

	return err
}

// creditConversionFailureCode is the stable code of a failed fleet conversion;
// the error itself is logged, never returned.
func creditConversionFailureCode(err error) string {
	switch {
	case errors.Is(err, billingdomain.ErrInsufficientWalletBalance):
		return "insufficient_wallet_balance"
	case errors.Is(err, billingdomain.ErrCreditConversionNothingToDo):
		return "nothing_to_convert"
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return "timeout"
	default:
		return "conversion_failed"
	}
}

func (s *CreditLedgerRepository) tenantSlug(ctx context.Context, tenantID string) string {
	var slug string
	if err := s.pool.QueryRow(ctx, `SELECT slug FROM tenants WHERE tenant_id = $1::uuid`, tenantID).Scan(&slug); err != nil {
		return ""
	}

	return slug
}
