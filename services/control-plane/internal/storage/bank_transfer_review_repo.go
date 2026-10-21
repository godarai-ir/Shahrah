package storage

import (
	"context"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

// ListPlatformBankTransferRequests returns the cross-tenant review queue, oldest first so the
// longest-waiting transfers are reviewed first. An empty status lists every status.
func (s *FinancialRepository) ListPlatformBankTransferRequests(ctx context.Context, status, query string, limit, offset int) (*platform.PlatformBankTransferListResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	filter := `WHERE ($1 = '' OR b.status = $1) AND ($2 = '' OR t.slug ILIKE '%' || $2 || '%' OR b.tracking_number ILIKE '%' || $2 || '%')`
	result := &platform.PlatformBankTransferListResponse{Data: make([]platform.PlatformBankTransferItem, 0, limit)}

	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM bank_transfer_requests b JOIN tenants t ON t.tenant_id = b.tenant_id `+filter,
		status, query).Scan(&result.Total); err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT b.bank_transfer_request_id::text, b.public_id::text, b.tenant_id::text, b.amount_irr::bigint,
		       b.tracking_number, b.bank_transaction_at, COALESCE(b.receipt_file_url, ''), b.status,
		       COALESCE(b.rejection_reason, ''), COALESCE(b.wallet_tx_id::text, ''), COALESCE(b.reviewed_by, ''),
		       b.reviewed_at, b.metadata, COALESCE(b.created_by, ''), b.created_at, b.updated_at,
		       t.slug,
		       CASE WHEN b.credit_pack_order_id IS NOT NULL THEN 'credit_pack'
		            WHEN b.seat_order_id IS NOT NULL THEN 'seat'
		            ELSE 'wallet_topup' END
		FROM bank_transfer_requests b
		JOIN tenants t ON t.tenant_id = b.tenant_id
		`+filter+`
		ORDER BY b.created_at ASC, b.bank_transfer_request_id ASC
		LIMIT $3 OFFSET $4
	`, status, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item platform.PlatformBankTransferItem

		if err := rows.Scan(
			&item.ID, &item.PublicID, &item.TenantID, &item.AmountIRR, &item.TrackingNumber, &item.BankTransactionAt,
			&item.ReceiptFileURL, &item.Status, &item.RejectionReason, &item.WalletTxID, &item.ReviewedBy,
			&item.ReviewedAt, &item.Metadata, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
			&item.TenantSlug, &item.Kind,
		); err != nil {
			return nil, err
		}

		result.Data = append(result.Data, item)
	}

	return result, rows.Err()
}
