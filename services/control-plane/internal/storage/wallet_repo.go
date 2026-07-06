package storage

import (
	"context"
	"errors"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/google/uuid"
)

// ErrInsufficientWalletBalance indicates a debit was attempted with insufficient funds.
var ErrInsufficientWalletBalance = errors.New("insufficient wallet balance")

// GetWallet retrieves or creates the tenant wallet.
func (s *Store) GetWallet(ctx context.Context, tenantID string) (*platform.Wallet, error) {
	var wallet platform.Wallet

	err := s.pool.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1, 0, 0)
		ON CONFLICT (tenant_id) DO UPDATE SET tenant_id = EXCLUDED.tenant_id
		RETURNING wallet_id::text, tenant_id::text, balance_irr, pending_irr
	`, tenantID).Scan(&wallet.ID, &wallet.TenantID, &wallet.BalanceIRR, &wallet.PendingIRR)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// TopUpWallet records a completed stub top-up.
func (s *Store) TopUpWallet(ctx context.Context, tenantID string, amountIRR float64, metadata map[string]interface{}) (*platform.Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var walletID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id, balance_irr, pending_irr)
		VALUES ($1, 0, 0)
		ON CONFLICT (tenant_id) DO UPDATE SET balance_irr = wallets.balance_irr
		RETURNING wallet_id::text
	`, tenantID).Scan(&walletID); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `UPDATE wallets SET balance_irr = balance_irr + $2, updated_at = NOW() WHERE wallet_id::text = $1`, walletID, amountIRR); err != nil {
		return nil, err
	}

	metaRaw := marshalJSONBestEffort(metadata)

	var out platform.Transaction
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallet_transactions (
			transaction_id, wallet_id, tenant_id, tx_type, amount_irr, status, description, metadata
		)
		VALUES ($1, $2, $3, 'DEPOSIT', $4, 'COMPLETED', 'stubbed wallet top-up', $5)
		RETURNING transaction_id::text, wallet_id::text, tenant_id::text, tx_type, amount_irr, status, description, created_at
	`, uuid.NewString(), walletID, tenantID, amountIRR, metaRaw).Scan(
		&out.ID, &out.WalletID, &out.TenantID, &out.Type, &out.AmountIRR, &out.Status, &out.Description, &out.CreatedAt,
	); err != nil {
		return nil, err
	}

	out.Metadata = metadata

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, tenantID)

	return &out, nil
}

// DebitWallet debits the tenant wallet for inference usage.
func (s *Store) DebitWallet(ctx context.Context, tenantID string, breakdown platform.UsageCostBreakdown, requestID string) (*platform.Transaction, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var wallet platform.Wallet
	// Lock the wallet row for the full transaction so concurrent debits serialize.
	if err := tx.QueryRow(ctx, `
		SELECT wallet_id::text, tenant_id::text, balance_irr, pending_irr
		FROM wallets
		WHERE tenant_id = $1
		FOR UPDATE
	`, tenantID).Scan(&wallet.ID, &wallet.TenantID, &wallet.BalanceIRR, &wallet.PendingIRR); err != nil {
		return nil, err
	}

	if wallet.BalanceIRR < breakdown.TotalCostIRR {
		return nil, ErrInsufficientWalletBalance
	}

	if _, err := tx.Exec(ctx, `UPDATE wallets SET balance_irr = balance_irr - $2, updated_at = NOW() WHERE wallet_id::text = $1`, wallet.ID, breakdown.TotalCostIRR); err != nil {
		return nil, err
	}

	metaRaw := marshalJSONBestEffort(map[string]interface{}{
		"request_id": requestID,
		"model":      breakdown.Model,
		"tokens":     breakdown.TotalTokens,
	})

	var txn platform.Transaction
	if err := tx.QueryRow(ctx, `
		INSERT INTO wallet_transactions (
			transaction_id, wallet_id, tenant_id, tx_type, amount_irr, amount_usd, exchange_rate, status, description, metadata
		)
		VALUES ($1, $2, $3, 'CONSUMPTION', $4, $5, $6, 'COMPLETED', 'inference usage', $7)
		RETURNING transaction_id::text, wallet_id::text, tenant_id::text, tx_type, amount_irr, amount_usd, exchange_rate, status, description, created_at
	`, uuid.NewString(), wallet.ID, tenantID, breakdown.TotalCostIRR, breakdown.TotalCostUSD, breakdown.ExchangeRate, metaRaw).Scan(
		&txn.ID, &txn.WalletID, &txn.TenantID, &txn.Type, &txn.AmountIRR, &txn.AmountUSD, &txn.ExchangeRate, &txn.Status, &txn.Description, &txn.CreatedAt,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, tenantID)

	return &txn, nil
}
