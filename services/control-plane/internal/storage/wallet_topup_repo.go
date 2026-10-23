package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
)

func (s *WalletPaymentRepository) GetWalletTopUpSettings(ctx context.Context) (*platform.WalletTopUpSettings, error) {
	var out platform.WalletTopUpSettings

	err := s.pool.QueryRow(ctx, `
		INSERT INTO wallet_topup_settings (settings_id)
		VALUES (TRUE)
		ON CONFLICT (settings_id) DO UPDATE SET settings_id = EXCLUDED.settings_id
		RETURNING min_amount_usd, max_amount_usd, tax_percent, fee_percent, updated_at
	`).Scan(&out.MinAmountUSD, &out.MaxAmountUSD, &out.TaxPercent, &out.FeePercent, &out.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (s *WalletPaymentRepository) UpdateWalletTopUpSettings(ctx context.Context, minUSD, maxUSD, taxPercent, feePercent float64) (*platform.WalletTopUpSettings, error) {
	var out platform.WalletTopUpSettings

	err := s.pool.QueryRow(ctx, `
		UPDATE wallet_topup_settings
		SET min_amount_usd = $1,
			max_amount_usd = $2,
			tax_percent = $3,
			fee_percent = $4,
			updated_at = NOW()
		WHERE settings_id = TRUE
		RETURNING min_amount_usd, max_amount_usd, tax_percent, fee_percent, updated_at
	`, minUSD, maxUSD, taxPercent, feePercent).Scan(&out.MinAmountUSD, &out.MaxAmountUSD, &out.TaxPercent, &out.FeePercent, &out.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (s *WalletPaymentRepository) CreateWalletTopUpIntent(ctx context.Context, p billingdomain.CreateTopUpIntentParams) (*platform.WalletTopUpIntent, error) {
	if err := refuseCreditTenantWalletTopUp(ctx, s.pool, p.TenantID); err != nil {
		return nil, err
	}

	metaRaw := marshalJSONBestEffort(p.Metadata)

	var out platform.WalletTopUpIntent

	err := s.pool.QueryRow(ctx, `
		INSERT INTO wallet_topup_intents (
			tenant_id, provider, status,
			amount_usd, amount_irr, exchange_rate,
			tax_percent, fee_percent, tax_amount_irr, fee_amount_irr, payable_irr,
			callback_url, return_url, metadata
		)
		VALUES ($1, $2, 'PENDING', $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING
			topup_intent_id::text,
			public_id::text,
			tenant_id::text,
			provider,
			status,
			amount_usd,
			amount_irr,
			exchange_rate,
			tax_percent,
			fee_percent,
			tax_amount_irr,
			fee_amount_irr,
			payable_irr,
			callback_url,
			return_url,
			metadata,
			raw_payload,
			created_at,
			updated_at
	`, p.TenantID, p.Provider, p.AmountUSD, p.AmountIRR, p.ExchangeRate, p.TaxPercent, p.FeePercent, p.TaxAmountIRR, p.FeeAmountIRR, p.PayableIRR, p.CallbackURL, p.ReturnURL, metaRaw).Scan(
		&out.ID,
		&out.PublicID,
		&out.TenantID,
		&out.Provider,
		&out.Status,
		&out.AmountUSD,
		&out.AmountIRR,
		&out.ExchangeRate,
		&out.TaxPercent,
		&out.FeePercent,
		&out.TaxAmountIRR,
		&out.FeeAmountIRR,
		&out.PayableIRR,
		&out.CallbackURL,
		&out.ReturnURL,
		&out.Metadata,
		&out.RawPayload,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &out, nil
}

func (s *WalletPaymentRepository) AttachWalletTopUpGatewayData(ctx context.Context, intentPublicID, authority, paymentURL string, raw map[string]interface{}) error {
	rawPayload := marshalJSONBestEffort(raw)

	res, err := s.pool.Exec(ctx, `
		UPDATE wallet_topup_intents
		SET authority = $2,
			payment_url = $3,
			raw_payload = COALESCE(raw_payload, '{}'::jsonb) || $4::jsonb,
			updated_at = NOW()
		WHERE public_id::text = $1
	`, intentPublicID, authority, paymentURL, rawPayload)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return billingdomain.ErrTopUpIntentNotFound
	}

	return nil
}

func (s *WalletPaymentRepository) GetWalletTopUpIntentByAuthority(ctx context.Context, authority string) (*platform.WalletTopUpIntent, error) {
	var (
		out             platform.WalletTopUpIntent
		walletTxID      *string
		completedAt     *time.Time
		dbRefID         *string
		dbGatewayTxnID  *string
		dbCardMask      *string
		dbPayerPhone    *string
		dbFailureReason *string
	)

	err := s.pool.QueryRow(ctx, `
		SELECT
			topup_intent_id::text,
			public_id::text,
			tenant_id::text,
			provider,
			status,
			amount_usd,
			amount_irr,
			exchange_rate,
			tax_percent,
			fee_percent,
			tax_amount_irr,
			fee_amount_irr,
			payable_irr,
			authority,
			payment_url,
			ref_id,
			gateway_txn_id,
			card_mask,
			payer_phone,
			callback_url,
			return_url,
			metadata,
			raw_payload,
			failure_reason,
			wallet_tx_id::text,
			completed_at,
			created_at,
			updated_at
		FROM wallet_topup_intents
		WHERE authority = $1
		LIMIT 1
	`, authority).Scan(
		&out.ID,
		&out.PublicID,
		&out.TenantID,
		&out.Provider,
		&out.Status,
		&out.AmountUSD,
		&out.AmountIRR,
		&out.ExchangeRate,
		&out.TaxPercent,
		&out.FeePercent,
		&out.TaxAmountIRR,
		&out.FeeAmountIRR,
		&out.PayableIRR,
		&out.Authority,
		&out.PaymentURL,
		&dbRefID,
		&dbGatewayTxnID,
		&dbCardMask,
		&dbPayerPhone,
		&out.CallbackURL,
		&out.ReturnURL,
		&out.Metadata,
		&out.RawPayload,
		&dbFailureReason,
		&walletTxID,
		&completedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if walletTxID != nil {
		out.WalletTxID = *walletTxID
	}

	if dbRefID != nil {
		out.RefID = *dbRefID
	}

	if dbGatewayTxnID != nil {
		out.GatewayTxnID = *dbGatewayTxnID
	}

	if dbCardMask != nil {
		out.CardMask = *dbCardMask
	}

	if dbPayerPhone != nil {
		out.PayerPhone = *dbPayerPhone
	}

	if dbFailureReason != nil {
		out.FailureReason = *dbFailureReason
	}

	out.CompletedAt = completedAt

	return &out, nil
}

func (s *WalletPaymentRepository) CompleteWalletTopUpByAuthority(ctx context.Context, authority string, verifyRaw map[string]interface{}, refID, gatewayTxnID, cardMask, payerPhone string) (*platform.WalletTopUpIntent, bool, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		intent          platform.WalletTopUpIntent
		meta            map[string]interface{}
		raw             map[string]interface{}
		walletTxID      *string
		completedAt     *time.Time
		dbRefID         *string
		dbGatewayTxnID  *string
		dbCardMask      *string
		dbPayerPhone    *string
		dbFailureReason *string
	)
	if queryErr := tx.QueryRow(ctx, `
		SELECT
			topup_intent_id::text,
			public_id::text,
			tenant_id::text,
			provider,
			status,
			amount_usd,
			amount_irr,
			exchange_rate,
			tax_percent,
			fee_percent,
			tax_amount_irr,
			fee_amount_irr,
			payable_irr,
			authority,
			payment_url,
			ref_id,
			gateway_txn_id,
			card_mask,
			payer_phone,
			callback_url,
			return_url,
			metadata,
			raw_payload,
			failure_reason,
			wallet_tx_id::text,
			completed_at,
			created_at,
			updated_at
		FROM wallet_topup_intents
		WHERE authority = $1
		FOR UPDATE
	`, authority).Scan(
		&intent.ID,
		&intent.PublicID,
		&intent.TenantID,
		&intent.Provider,
		&intent.Status,
		&intent.AmountUSD,
		&intent.AmountIRR,
		&intent.ExchangeRate,
		&intent.TaxPercent,
		&intent.FeePercent,
		&intent.TaxAmountIRR,
		&intent.FeeAmountIRR,
		&intent.PayableIRR,
		&intent.Authority,
		&intent.PaymentURL,
		&dbRefID,
		&dbGatewayTxnID,
		&dbCardMask,
		&dbPayerPhone,
		&intent.CallbackURL,
		&intent.ReturnURL,
		&meta,
		&raw,
		&dbFailureReason,
		&walletTxID,
		&completedAt,
		&intent.CreatedAt,
		&intent.UpdatedAt,
	); queryErr != nil {
		return nil, false, queryErr
	}

	intent.Metadata = meta

	intent.RawPayload = raw
	if walletTxID != nil {
		intent.WalletTxID = *walletTxID
	}

	if dbRefID != nil {
		intent.RefID = *dbRefID
	}

	if dbGatewayTxnID != nil {
		intent.GatewayTxnID = *dbGatewayTxnID
	}

	if dbCardMask != nil {
		intent.CardMask = *dbCardMask
	}

	if dbPayerPhone != nil {
		intent.PayerPhone = *dbPayerPhone
	}

	if dbFailureReason != nil {
		intent.FailureReason = *dbFailureReason
	}

	intent.CompletedAt = completedAt

	if string(intent.Status) == string(platform.WalletTopUpStatusCompleted) {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return nil, false, commitErr
		}

		return &intent, true, nil
	}

	metaPayload := map[string]interface{}{
		"topup_intent_id": intent.ID,
		"public_id":       intent.PublicID,
		"authority":       authority,
		"provider":        intent.Provider,
		"ref_id":          refID,
		"gateway_txn_id":  gatewayTxnID,
	}

	walletTxn, err := appendWalletTransactionTx(ctx, tx, walletTransactionParams{
		TenantID:       intent.TenantID,
		Type:           "wallet_topup",
		AmountIRR:      roundedIRR(intent.AmountIRR),
		AmountUSD:      &intent.AmountUSD,
		ExchangeRate:   &intent.ExchangeRate,
		Status:         "successful",
		Description:    "wallet top-up",
		Metadata:       metaPayload,
		SourceType:     "wallet_topup_intent",
		SourceID:       intent.ID,
		IdempotencyKey: "wallet_topup:" + strings.TrimSpace(authority),
		CompletedAt:    time.Now().UTC(),
	})
	if err != nil {
		return nil, false, err
	}

	walletTxIDOut := walletTxn.ID

	verifyPayload := marshalJSONBestEffort(verifyRaw)
	if _, err := tx.Exec(ctx, `
		UPDATE wallet_topup_intents
		SET status = 'COMPLETED',
			ref_id = $2,
			gateway_txn_id = $3,
			card_mask = $4,
			payer_phone = $5,
			wallet_tx_id = $6,
			raw_payload = COALESCE(raw_payload, '{}'::jsonb) || $7::jsonb,
			completed_at = NOW(),
			updated_at = NOW()
		WHERE topup_intent_id::text = $1
	`, intent.ID, refID, gatewayTxnID, cardMask, payerPhone, walletTxIDOut, verifyPayload); err != nil {
		return nil, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, false, err
	}

	s.ledger.invalidateWalletCache(ctx, intent.TenantID)

	intent.Status = platform.WalletTopUpStatusCompleted
	intent.RefID = refID
	intent.GatewayTxnID = gatewayTxnID
	intent.CardMask = cardMask
	intent.PayerPhone = payerPhone
	intent.WalletTxID = walletTxIDOut
	now := time.Now()
	intent.CompletedAt = &now

	return &intent, false, nil
}

func (s *WalletPaymentRepository) MarkWalletTopUpFailedByAuthority(ctx context.Context, authority string, reason string, raw map[string]interface{}) error {
	rawPayload := marshalJSONBestEffort(raw)

	// A completed top-up is left alone, so a late failure callback cannot change it.
	res, err := s.pool.Exec(ctx, `
		UPDATE wallet_topup_intents
		SET status = 'FAILED',
			failure_reason = $2,
			raw_payload = COALESCE(raw_payload, '{}'::jsonb) || $3::jsonb,
			updated_at = NOW()
		WHERE authority = $1 AND status <> 'COMPLETED'
	`, authority, reason, rawPayload)
	if err != nil {
		return err
	}

	if res.RowsAffected() > 0 {
		return nil
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM wallet_topup_intents WHERE authority = $1)`, authority).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		return billingdomain.ErrTopUpIntentNotFound
	}

	return nil
}

func (s *WalletPaymentRepository) ListWalletTransactions(ctx context.Context, tenantID string, limit int) ([]platform.Transaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, `
		SELECT transaction_id::text, wallet_id::text, tenant_id::text, tx_type,
		       ROUND(amount_irr)::bigint, COALESCE(balance_after_irr, 0)::bigint,
		       amount_usd, exchange_rate, status, COALESCE(description, ''), metadata,
		       COALESCE(source_type, ''), COALESCE(source_id, ''), created_at, completed_at
		FROM wallet_transactions
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]platform.Transaction, 0)

	for rows.Next() {
		var item platform.Transaction
		if err := rows.Scan(
			&item.ID, &item.WalletID, &item.TenantID, &item.Type, &item.AmountIRR, &item.BalanceAfter,
			&item.AmountUSD, &item.ExchangeRate, &item.Status, &item.Description, &item.Metadata,
			&item.SourceType, &item.SourceID, &item.CreatedAt, &item.CompletedAt,
		); err != nil {
			return nil, err
		}

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *WalletPaymentRepository) BuildTopUpCallbackURL(baseURL, provider string) string {
	return fmt.Sprintf("%s/api/v1/wallet/topup/callback/%s", baseURL, provider)
}
