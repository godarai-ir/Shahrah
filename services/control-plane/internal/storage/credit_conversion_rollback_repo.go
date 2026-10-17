package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/Delpak/shahrah/services/control-plane/internal/domain/credits"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RollbackCreditConversion moves a credit tenant back to IRR billing. Debt is
// cleared, every remaining credit is converted back into the IRR wallet at the
// price the tenant was converted at, and what the tenant owes is debited from
// the wallet; anything the wallet cannot cover is recorded as uncollected.
func (s *CreditLedgerRepository) RollbackCreditConversion(ctx context.Context, tenantID, reason string, actor platform.SupportActor) (*platform.CreditConversion, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, billingdomain.ErrDecisionReasonRequired
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	conversion, err := rollbackTenantTx(ctx, tx, tenantID, reason, actor, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.afterConversion(ctx, tenantID, "credit.rolled_back")

	return conversion, nil
}

func rollbackTenantTx(ctx context.Context, tx pgx.Tx, tenantID, reason string, actor platform.SupportActor, now time.Time) (*platform.CreditConversion, error) {
	unit, sequence, err := lockTenantForConversionTx(ctx, tx, tenantID)
	if err != nil {
		return nil, err
	}

	if unit != credits.BillingUnitCredit {
		return nil, billingdomain.ErrCreditBillingNotEnabled
	}

	priceToman, revisionID, err := lastConversionPriceTx(ctx, tx, tenantID)
	if err != nil {
		return nil, err
	}

	state, err := lockCreditTenantTx(ctx, tx, tenantID)
	if err != nil {
		return nil, err
	}

	conversion := &platform.CreditConversion{
		ID: uuid.NewString(), TenantID: tenantID, Sequence: sequence, Direction: credits.ConversionToIRR,
		PriceTomanPerCredit: priceToman, Credits: credits.Format(credits.Millicredits(state.AvailableMC)),
		Debt: credits.Format(credits.Millicredits(state.DebtMC)), Actor: actor.ID, Reason: reason, CreatedAt: now,
	}

	if err := clearCreditsForRollbackTx(ctx, tx, conversion, state, now); err != nil {
		return nil, err
	}

	if err := settleRollbackWalletTx(ctx, tx, conversion, state, now); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `UPDATE tenants SET billing_unit = $2 WHERE tenant_id = $1::uuid`, tenantID, credits.BillingUnitIRR); err != nil {
		return nil, err
	}

	if err := insertCreditConversionTx(ctx, tx, conversion, state.AvailableMC, state.DebtMC, revisionID, ""); err != nil {
		return nil, err
	}

	return conversion, insertFinancialAuditTx(ctx, tx, tenantID, "credit.rolled_back", actor, "credit_conversion", conversion.ID,
		conversion.ID, conversionAuditPayload(conversion))
}

// lastConversionPriceTx is the price of the tenant's latest conversion to
// credits, so a rollback returns credits at the rate they were bought at.
func lastConversionPriceTx(ctx context.Context, tx pgx.Tx, tenantID string) (int64, *string, error) {
	var (
		priceToman int64
		revisionID *string
	)

	err := tx.QueryRow(ctx, `
		SELECT price_toman_per_credit, price_revision_id::text FROM credit_conversions
		WHERE tenant_id = $1::uuid AND direction = $2
		ORDER BY sequence DESC LIMIT 1
	`, tenantID, credits.ConversionToCredit).Scan(&priceToman, &revisionID)
	if errors.Is(err, pgx.ErrNoRows) {
		price, priceErr := activeCreditPrice(ctx, tx)

		return price.Toman, price.RevisionID, priceErr
	}

	return priceToman, revisionID, err
}

// clearCreditsForRollbackTx zeroes the tenant's credits with convert entries:
// one that repays the debt and one that empties the lots.
func clearCreditsForRollbackTx(ctx context.Context, tx pgx.Tx, conversion *platform.CreditConversion, state creditTenantState, now time.Time) error {
	if state.DebtMC > 0 {
		if _, err := grantCreditsTx(ctx, tx, CreditGrant{
			TenantID: conversion.TenantID, Kind: credits.EntryConvert, Source: credits.LotSourceMigration,
			AmountMC: state.DebtMC, IdempotencyKey: credits.ConvertKey(conversion.TenantID, conversion.Sequence, "debt"),
			OccurredAt: now, Reason: conversion.Reason, Actor: conversion.Actor,
			SourceRef: map[string]interface{}{"conversion_id": conversion.ID, "rollback": "debt"},
		}); err != nil {
			return err
		}
	}

	if state.AvailableMC <= 0 {
		return nil
	}

	_, err := debitCreditsTx(ctx, tx, conversion.TenantID, credits.EntryConvert, state.AvailableMC,
		credits.ConvertKey(conversion.TenantID, conversion.Sequence, "lots"), conversion.Reason, conversion.Actor, now)

	return err
}

// settleRollbackWalletTx credits the IRR value of the remaining credits, or
// debits what the debt was worth, to the IRR wallet.
func settleRollbackWalletTx(ctx context.Context, tx pgx.Tx, conversion *platform.CreditConversion, state creditTenantState, now time.Time) error {
	netIRR := credits.CreditsToIRR(state.AvailableMC, conversion.PriceTomanPerCredit) -
		credits.CreditsToIRR(state.DebtMC, conversion.PriceTomanPerCredit)

	params := walletTransactionParams{
		TenantID: conversion.TenantID, SourceType: "credit_conversion", SourceID: conversion.ID,
		IdempotencyKey: credits.ConvertKey(conversion.TenantID, conversion.Sequence, "wallet"),
		Metadata:       map[string]interface{}{"price_toman_per_credit": conversion.PriceTomanPerCredit},
		CreatedBy:      conversion.Actor, CompletedAt: now,
	}

	switch {
	case netIRR > 0:
		params.Type, params.AmountIRR, params.Description = walletTxCreditConversionIn, netIRR, "credits converted back to IRR"
		conversion.ConvertedIRR = netIRR
	case netIRR < 0:
		var walletIRR int64
		if err := tx.QueryRow(ctx, `
			SELECT COALESCE(ROUND(balance_irr)::bigint, 0) FROM wallets WHERE tenant_id = $1::uuid FOR UPDATE
		`, conversion.TenantID).Scan(&walletIRR); err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		owed := -netIRR
		collected := min(owed, max(walletIRR, 0))
		conversion.WalletIRR, conversion.ConvertedIRR, conversion.UncollectedIRR = walletIRR, collected, owed-collected

		if collected == 0 {
			return nil
		}

		params.Type, params.AmountIRR, params.Description = walletTxCreditConversionOut, collected, "credit debt settled in IRR"
	default:
		return nil
	}

	_, err := appendWalletTransactionTx(ctx, tx, params)

	return err
}
