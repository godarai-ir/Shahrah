package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

var ErrTopUpIntentNotFound = errors.New("top-up intent not found")

type CreateTopUpIntentParams struct {
	TenantID     string
	Provider     string
	AmountUSD    float64
	AmountIRR    float64
	ExchangeRate float64
	TaxPercent   float64
	FeePercent   float64
	TaxAmountIRR float64
	FeeAmountIRR float64
	PayableIRR   float64
	CallbackURL  string
	ReturnURL    string
	Metadata     map[string]interface{}
}

func (s *Store) GetWalletTopUpSettings(ctx context.Context) (*platform.WalletTopUpSettings, error) {
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

func (s *Store) UpdateWalletTopUpSettings(ctx context.Context, minUSD, maxUSD, taxPercent, feePercent float64) (*platform.WalletTopUpSettings, error) {
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

func (s *Store) CreateWalletTopUpIntent(ctx context.Context, p CreateTopUpIntentParams) (*platform.WalletTopUpIntent, error) {
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

func (s *Store) AttachWalletTopUpGatewayData(ctx context.Context, intentPublicID, authority, paymentURL string, raw map[string]interface{}) error {
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
		return ErrTopUpIntentNotFound
	}

	return nil
}

func (s *Store) GetWalletTopUpIntentByAuthority(ctx context.Context, authority string) (*platform.WalletTopUpIntent, error) {
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

func (s *Store) CompleteWalletTopUpByAuthority(ctx context.Context, authority string, verifyRaw map[string]interface{}, refID, gatewayTxnID, cardMask, payerPhone string) (*platform.WalletTopUpIntent, bool, error) {
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
	if err := tx.QueryRow(ctx, `
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
	); err != nil {
		return nil, false, err
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
		if err := tx.Commit(ctx); err != nil {
			return nil, false, err
		}

		return &intent, true, nil
	}

	var walletID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1, 0, 0)
		ON CONFLICT (tenant_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id
		RETURNING wallet_id::text
	`, intent.TenantID).Scan(&walletID); err != nil {
		return nil, false, err
	}

	if _, err := tx.Exec(ctx, `UPDATE wallets SET balance_irr = balance_irr + $2, updated_at = NOW() WHERE wallet_id::text = $1`, walletID, intent.PayableIRR); err != nil {
		return nil, false, err
	}

	metaPayload := marshalJSONBestEffort(map[string]interface{}{
		"topup_intent_id": intent.ID,
		"public_id":       intent.PublicID,
		"authority":       authority,
		"provider":        intent.Provider,
		"ref_id":          refID,
		"gateway_txn_id":  gatewayTxnID,
	})

	var walletTxIDOut string
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallet_transactions (
			wallet_id, tenant_id, tx_type, amount_irr, amount_usd, exchange_rate, status, description, metadata
		)
		VALUES ($1, $2, 'DEPOSIT', $3, $4, $5, 'COMPLETED', 'wallet top-up', $6)
		RETURNING transaction_id::text
	`, walletID, intent.TenantID, intent.PayableIRR, intent.AmountUSD, intent.ExchangeRate, metaPayload).Scan(&walletTxIDOut); err != nil {
		return nil, false, err
	}

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

func (s *Store) MarkWalletTopUpFailedByAuthority(ctx context.Context, authority string, reason string, raw map[string]interface{}) error {
	rawPayload := marshalJSONBestEffort(raw)

	res, err := s.pool.Exec(ctx, `
		UPDATE wallet_topup_intents
		SET status = CASE WHEN status = 'COMPLETED' THEN status ELSE 'FAILED' END,
			failure_reason = $2,
			raw_payload = COALESCE(raw_payload, '{}'::jsonb) || $3::jsonb,
			updated_at = NOW()
		WHERE authority = $1
	`, authority, reason, rawPayload)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrTopUpIntentNotFound
	}

	return nil
}

func (s *Store) ListWalletTransactions(ctx context.Context, tenantID string, limit int) ([]platform.Transaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, `
		SELECT transaction_id::text, wallet_id::text, tenant_id::text, tx_type, amount_irr, amount_usd, exchange_rate, status, description, metadata, created_at
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
		if err := rows.Scan(&item.ID, &item.WalletID, &item.TenantID, &item.Type, &item.AmountIRR, &item.AmountUSD, &item.ExchangeRate, &item.Status, &item.Description, &item.Metadata, &item.CreatedAt); err != nil {
			return nil, err
		}

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) BuildTopUpCallbackURL(baseURL, provider string) string {
	return fmt.Sprintf("%s/api/v1/wallet/topup/callback/%s", baseURL, provider)
}
