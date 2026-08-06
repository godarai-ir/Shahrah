package storage

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	billingdomain "github.com/Delpak/shahrah/services/control-plane/internal/domain/billing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// GetWallet retrieves or creates the tenant wallet.
func (s *WalletRepository) GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error) {
	wallet, err := s.getWalletLedger(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	return s.decorateWalletWithLiveState(ctx, wallet)
}

func (s *WalletRepository) getWalletLedger(ctx context.Context, tenantID string) (*platform.Wallet, error) {
	ctx, cancel := withHotPathDBTimeout(ctx)
	defer cancel()

	var wallet platform.Wallet

	err := s.pool.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1, 0, 0)
		ON CONFLICT (tenant_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id
		RETURNING wallet_id::text, tenant_id::text, ROUND(balance_irr)::bigint, ROUND(pending_irr)::bigint
	`, tenantID).Scan(&wallet.ID, &wallet.TenantID, &wallet.BalanceIRR, &wallet.PendingIRR)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// TopUpWallet records a completed stub top-up.
func (s *WalletRepository) TopUpWallet(ctx context.Context, tenantID string, amountIRR float64, metadata map[string]interface{}) (*platform.Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	out, err := appendWalletTransactionTx(ctx, tx, walletTransactionParams{
		TenantID:       tenantID,
		Type:           "wallet_topup",
		AmountIRR:      roundedIRR(amountIRR),
		Status:         "successful",
		Description:    "stubbed wallet top-up",
		Metadata:       metadata,
		SourceType:     "legacy_topup",
		SourceID:       uuid.NewString(),
		IdempotencyKey: "legacy_topup:" + uuid.NewString(),
		CompletedAt:    time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}

	out.Metadata = metadata

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, tenantID)

	return out, nil
}

// DebitWallet debits the tenant wallet for inference usage.
func (s *WalletRepository) DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error) {
	ctx, cancel := withHotPathDBTimeout(ctx)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxWithFreshTimeout(tx)

	metadata := map[string]interface{}{
		"request_id": requestID,
		"model":      breakdown.Model,
		"tokens":     breakdown.TotalTokens,
	}

	txn, err := appendWalletTransactionTx(ctx, tx, walletTransactionParams{
		TenantID:       tenantID,
		Type:           "usage_debit",
		AmountIRR:      roundedIRR(breakdown.TotalCostIRR),
		AmountUSD:      &breakdown.TotalCostUSD,
		ExchangeRate:   &breakdown.ExchangeRate,
		Status:         "successful",
		Description:    "inference usage",
		Metadata:       metadata,
		SourceType:     "gateway_request",
		SourceID:       requestID,
		IdempotencyKey: "usage_debit:" + strings.TrimSpace(requestID),
		CompletedAt:    time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, tenantID)

	return txn, nil
}

type walletTransactionParams struct {
	TenantID       string
	Type           string
	AmountIRR      int64
	AmountUSD      *float64
	ExchangeRate   *float64
	Status         string
	Description    string
	Metadata       map[string]interface{}
	SourceType     string
	SourceID       string
	IdempotencyKey string
	CreatedBy      string
	CompletedAt    time.Time
}

func appendWalletTransactionTx(ctx context.Context, tx pgx.Tx, p walletTransactionParams) (*platform.Transaction, error) {
	if p.AmountIRR <= 0 {
		return nil, errors.New("wallet transaction amount must be positive")
	}

	if p.Status == "" {
		p.Status = "successful"
	}

	if p.IdempotencyKey != "" {
		existing, err := scanWalletTransactionRow(tx.QueryRow(ctx, `
			SELECT transaction_id::text, wallet_id::text, tenant_id::text, tx_type,
			       ROUND(amount_irr)::bigint, COALESCE(balance_after_irr, 0)::bigint,
			       amount_usd, exchange_rate, status, COALESCE(description, ''), metadata,
			       COALESCE(source_type, ''), COALESCE(source_id, ''), created_at, completed_at
			FROM wallet_transactions
			WHERE tenant_id::text = $1 AND idempotency_key = $2
			LIMIT 1
		`, p.TenantID, p.IdempotencyKey))
		if err == nil {
			return existing, nil
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	var (
		walletID string
		balance  int64
	)
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1::uuid, 0, 0)
		ON CONFLICT (tenant_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id
		RETURNING wallet_id::text
	`, p.TenantID).Scan(&walletID); err != nil {
		return nil, err
	}

	if err := tx.QueryRow(ctx, `
		SELECT ROUND(balance_irr)::bigint
		FROM wallets
		WHERE wallet_id::text = $1
		FOR UPDATE
	`, walletID).Scan(&balance); err != nil {
		return nil, err
	}

	delta := p.AmountIRR
	if isWalletDebitType(p.Type) {
		delta = -delta
	}

	after := balance + delta
	if after < 0 {
		return nil, billingdomain.ErrInsufficientWalletBalance
	}

	if _, err := tx.Exec(ctx, `
		UPDATE wallets
		SET balance_irr = $2, updated_at = NOW()
		WHERE wallet_id::text = $1
	`, walletID, after); err != nil {
		return nil, err
	}

	metaRaw := marshalJSONBestEffort(p.Metadata)

	completedAt := any(nil)
	if !p.CompletedAt.IsZero() {
		completedAt = p.CompletedAt
	}

	return scanWalletTransactionRow(tx.QueryRow(ctx, `
		INSERT INTO wallet_transactions (
			transaction_id, wallet_id, tenant_id, tx_type, amount_irr, balance_after_irr,
			amount_usd, exchange_rate, status, description, metadata, source_type, source_id,
			idempotency_key, created_by, completed_at, effective_at
		)
		VALUES (
			$1::uuid, $2::uuid, $3::uuid, $4, $5, $6,
			$7, $8, $9, NULLIF($10, ''), $11, NULLIF($12, ''), NULLIF($13, ''),
			NULLIF($14, ''), NULLIF($15, ''), $16, NOW()
		)
		RETURNING transaction_id::text, wallet_id::text, tenant_id::text, tx_type,
		          ROUND(amount_irr)::bigint, COALESCE(balance_after_irr, 0)::bigint,
		          amount_usd, exchange_rate, status, COALESCE(description, ''), metadata,
		          COALESCE(source_type, ''), COALESCE(source_id, ''), created_at, completed_at
	`, uuid.NewString(), walletID, p.TenantID, p.Type, p.AmountIRR, after, p.AmountUSD, p.ExchangeRate,
		p.Status, p.Description, metaRaw, p.SourceType, p.SourceID, p.IdempotencyKey, p.CreatedBy, completedAt))
}

func scanWalletTransactionRow(row pgx.Row) (*platform.Transaction, error) {
	var item platform.Transaction
	if err := row.Scan(
		&item.ID, &item.WalletID, &item.TenantID, &item.Type, &item.AmountIRR, &item.BalanceAfter,
		&item.AmountUSD, &item.ExchangeRate, &item.Status, &item.Description, &item.Metadata,
		&item.SourceType, &item.SourceID, &item.CreatedAt, &item.CompletedAt,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func isWalletDebitType(txType string) bool {
	switch strings.ToLower(strings.TrimSpace(txType)) {
	case "usage_debit", "invoice_adjustment", "manual_debit", "invoice_debit", "consumption", "withdrawal":
		return true
	default:
		return false
	}
}

func roundedIRR(value float64) int64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}

	return int64(math.Round(value))
}
