package storage

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Delpak/shahrah/packages/contracts/platform"
	"github.com/jackc/pgx/v5"
)

var (
	ErrBankTransferNotFound  = errors.New("bank transfer request not found")
	ErrFinancialCodeInvalid  = errors.New("financial code is invalid")
	ErrBillingProfileInvalid = errors.New("billing profile is invalid")
)

func (s *Store) CreateBankTransferRequest(ctx context.Context, tenantID string, req platform.CreateBankTransferRequest, actor string) (*platform.BankTransferRequest, error) {
	meta := marshalJSONBestEffort(req.Metadata)

	return scanBankTransferRequest(s.pool.QueryRow(ctx, `
		INSERT INTO bank_transfer_requests (
			tenant_id, amount_irr, tracking_number, bank_transaction_at, receipt_file_url,
			status, metadata, created_by
		)
		VALUES ($1::uuid, $2, $3, $4, NULLIF($5, ''), 'pending_review', $6, NULLIF($7, ''))
		RETURNING bank_transfer_request_id::text, public_id::text, tenant_id::text, amount_irr::bigint,
		          tracking_number, bank_transaction_at, COALESCE(receipt_file_url, ''), status,
		          COALESCE(rejection_reason, ''), COALESCE(wallet_tx_id::text, ''), COALESCE(reviewed_by, ''),
		          reviewed_at, metadata, COALESCE(created_by, ''), created_at, updated_at
	`, tenantID, req.AmountIRR, strings.TrimSpace(req.TrackingNumber), req.BankTransactionAt, strings.TrimSpace(req.ReceiptFileURL), meta, strings.TrimSpace(actor)))
}

func (s *Store) ListBankTransferRequests(ctx context.Context, tenantID string, limit int) ([]platform.BankTransferRequest, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, `
		SELECT bank_transfer_request_id::text, public_id::text, tenant_id::text, amount_irr::bigint,
		       tracking_number, bank_transaction_at, COALESCE(receipt_file_url, ''), status,
		       COALESCE(rejection_reason, ''), COALESCE(wallet_tx_id::text, ''), COALESCE(reviewed_by, ''),
		       reviewed_at, metadata, COALESCE(created_by, ''), created_at, updated_at
		FROM bank_transfer_requests
		WHERE tenant_id::text = $1
		ORDER BY created_at DESC, bank_transfer_request_id DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.BankTransferRequest, 0, limit)

	for rows.Next() {
		item, err := scanBankTransferRequest(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func (s *Store) BankTransferReceiptExists(ctx context.Context, tenantID string, receiptFileURL string) (bool, error) {
	var exists bool

	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM bank_transfer_requests
			WHERE tenant_id::text = $1 AND receipt_file_url = $2
		)
	`, strings.TrimSpace(tenantID), strings.TrimSpace(receiptFileURL)).Scan(&exists)

	return exists, err
}

func (s *Store) ApproveBankTransferRequest(ctx context.Context, tenantID string, requestID string, req platform.ReviewBankTransferRequest, actor platform.SupportActor) (*platform.BankTransferRequest, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	item, err := scanBankTransferRequest(tx.QueryRow(ctx, `
		SELECT bank_transfer_request_id::text, public_id::text, tenant_id::text, amount_irr::bigint,
		       tracking_number, bank_transaction_at, COALESCE(receipt_file_url, ''), status,
		       COALESCE(rejection_reason, ''), COALESCE(wallet_tx_id::text, ''), COALESCE(reviewed_by, ''),
		       reviewed_at, metadata, COALESCE(created_by, ''), created_at, updated_at
		FROM bank_transfer_requests
		WHERE tenant_id::text = $1 AND (bank_transfer_request_id::text = $2 OR public_id::text = $2)
		FOR UPDATE
	`, tenantID, strings.TrimSpace(requestID)))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBankTransferNotFound
		}

		return nil, err
	}

	if item.Status == "approved" {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return nil, commitErr
		}

		return item, nil
	}

	if item.Status != "pending_review" {
		return nil, ErrInvalidTenantStatusChange
	}

	idempotency := strings.TrimSpace(req.IdempotencyKey)
	if idempotency == "" {
		idempotency = "bank_transfer:" + item.ID
	}

	walletTxn, err := appendWalletTransactionTx(ctx, tx, walletTransactionParams{
		TenantID:       tenantID,
		Type:           "bank_transfer_topup",
		AmountIRR:      item.AmountIRR,
		Status:         "successful",
		Description:    "bank transfer top-up",
		Metadata:       map[string]interface{}{"tracking_number": item.TrackingNumber, "bank_transfer_request_id": item.ID},
		SourceType:     "bank_transfer_request",
		SourceID:       item.ID,
		IdempotencyKey: idempotency,
		CreatedBy:      actor.ID,
		CompletedAt:    time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE bank_transfer_requests
		SET status = 'approved', wallet_tx_id = $3::uuid, reviewed_by = NULLIF($4, ''),
		    reviewed_at = NOW(), updated_at = NOW()
		WHERE tenant_id::text = $1 AND bank_transfer_request_id::text = $2
	`, tenantID, item.ID, walletTxn.ID, actor.ID); err != nil {
		return nil, err
	}

	if err := insertFinancialAuditTx(ctx, tx, tenantID, "bank_transfer.approved", actor, "bank_transfer_request", item.ID, idempotency, map[string]interface{}{"wallet_tx_id": walletTxn.ID}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, tenantID)

	return s.GetBankTransferRequest(ctx, tenantID, item.ID)
}

func (s *Store) RejectBankTransferRequest(ctx context.Context, tenantID string, requestID string, req platform.ReviewBankTransferRequest, actor platform.SupportActor) (*platform.BankTransferRequest, error) {
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "rejected by financial review"
	}

	rows, err := s.pool.Query(ctx, `
		UPDATE bank_transfer_requests
		SET status = 'rejected', rejection_reason = $3, reviewed_by = NULLIF($4, ''),
		    reviewed_at = NOW(), updated_at = NOW()
		WHERE tenant_id::text = $1
		  AND (bank_transfer_request_id::text = $2 OR public_id::text = $2)
		  AND status = 'pending_review'
		RETURNING bank_transfer_request_id::text, public_id::text, tenant_id::text, amount_irr::bigint,
		          tracking_number, bank_transaction_at, COALESCE(receipt_file_url, ''), status,
		          COALESCE(rejection_reason, ''), COALESCE(wallet_tx_id::text, ''), COALESCE(reviewed_by, ''),
		          reviewed_at, metadata, COALESCE(created_by, ''), created_at, updated_at
	`, tenantID, strings.TrimSpace(requestID), reason, actor.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, ErrBankTransferNotFound
	}

	item, err := scanBankTransferRequest(rows)
	if err != nil {
		return nil, err
	}

	return item, rows.Err()
}

func (s *Store) GetBankTransferRequest(ctx context.Context, tenantID string, requestID string) (*platform.BankTransferRequest, error) {
	item, err := scanBankTransferRequest(s.pool.QueryRow(ctx, `
		SELECT bank_transfer_request_id::text, public_id::text, tenant_id::text, amount_irr::bigint,
		       tracking_number, bank_transaction_at, COALESCE(receipt_file_url, ''), status,
		       COALESCE(rejection_reason, ''), COALESCE(wallet_tx_id::text, ''), COALESCE(reviewed_by, ''),
		       reviewed_at, metadata, COALESCE(created_by, ''), created_at, updated_at
		FROM bank_transfer_requests
		WHERE tenant_id::text = $1 AND (bank_transfer_request_id::text = $2 OR public_id::text = $2)
	`, tenantID, strings.TrimSpace(requestID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBankTransferNotFound
	}

	return item, err
}

func (s *Store) RedeemGiftCode(ctx context.Context, tenantID string, code string, actor string) (*platform.GiftCodeRedemption, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer rollbackTxBestEffort(ctx, tx)

	var (
		giftID string
		amount int64
	)

	err = tx.QueryRow(ctx, `
		SELECT gift_code_id::text, amount_irr::bigint
		FROM gift_codes
		WHERE LOWER(code) = LOWER($1)
		  AND active = true
		  AND (tenant_id IS NULL OR tenant_id::text = $2)
		  AND (starts_at IS NULL OR starts_at <= NOW())
		  AND (expires_at IS NULL OR expires_at > NOW())
		  AND (max_redemptions IS NULL OR redeemed_count < max_redemptions)
		ORDER BY tenant_id NULLS LAST
		LIMIT 1
		FOR UPDATE
	`, strings.TrimSpace(code), tenantID).Scan(&giftID, &amount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFinancialCodeInvalid
		}

		return nil, err
	}

	existing, err := scanGiftCodeRedemption(tx.QueryRow(ctx, `
		SELECT redemption_id::text, gift_code_id::text, tenant_id::text, COALESCE(wallet_tx_id::text, ''),
		       amount_irr::bigint, status, COALESCE(redeemed_by, ''), created_at
		FROM gift_code_redemptions
		WHERE gift_code_id::text = $1 AND tenant_id::text = $2
	`, giftID, tenantID))
	if err == nil {
		if commitErr := tx.Commit(ctx); commitErr != nil {
			return nil, commitErr
		}

		return existing, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	walletTxn, err := appendWalletTransactionTx(ctx, tx, walletTransactionParams{
		TenantID:       tenantID,
		Type:           "gift_credit",
		AmountIRR:      amount,
		Status:         "successful",
		Description:    "gift code credit",
		Metadata:       map[string]interface{}{"gift_code_id": giftID},
		SourceType:     "gift_code",
		SourceID:       giftID,
		IdempotencyKey: "gift_code:" + giftID + ":" + tenantID,
		CreatedBy:      actor,
		CompletedAt:    time.Now().UTC(),
	})
	if err != nil {
		return nil, err
	}

	var redemption platform.GiftCodeRedemption
	if err := tx.QueryRow(ctx, `
		INSERT INTO gift_code_redemptions (gift_code_id, tenant_id, wallet_tx_id, amount_irr, status, redeemed_by)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, 'used', NULLIF($5, ''))
		RETURNING redemption_id::text, gift_code_id::text, tenant_id::text, COALESCE(wallet_tx_id::text, ''),
		          amount_irr::bigint, status, COALESCE(redeemed_by, ''), created_at
	`, giftID, tenantID, walletTxn.ID, amount, strings.TrimSpace(actor)).Scan(
		&redemption.ID, &redemption.GiftCodeID, &redemption.TenantID, &redemption.WalletTxID,
		&redemption.AmountIRR, &redemption.Status, &redemption.RedeemedBy, &redemption.CreatedAt,
	); err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `UPDATE gift_codes SET redeemed_count = redeemed_count + 1, updated_at = NOW() WHERE gift_code_id::text = $1`, giftID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	s.invalidateWalletCache(ctx, tenantID)

	return &redemption, nil
}

func (s *Store) RedeemDiscountCode(ctx context.Context, tenantID string, code string, actor string) (*platform.DiscountRedemption, error) {
	var (
		item        platform.DiscountRedemption
		wasInserted bool
	)

	err := s.pool.QueryRow(ctx, `
		WITH selected AS (
			SELECT discount_code_id, code, discount_type, value_irr, percent, starts_at, expires_at
			FROM discount_codes
			WHERE LOWER(code) = LOWER($1)
			  AND active = true
			  AND (tenant_id IS NULL OR tenant_id::text = $2)
			  AND (starts_at IS NULL OR starts_at <= NOW())
			  AND (expires_at IS NULL OR expires_at > NOW())
			  AND (max_redemptions IS NULL OR redeemed_count < max_redemptions)
			ORDER BY tenant_id NULLS LAST
			LIMIT 1
		), inserted AS (
			INSERT INTO discount_redemptions (
				discount_code_id, tenant_id, code, discount_type, value_irr, percent,
				starts_at, expires_at, status, redeemed_by
			)
			SELECT discount_code_id, $2::uuid, code, discount_type, value_irr, percent,
			       starts_at, expires_at, 'active', NULLIF($3, '')
			FROM selected
			ON CONFLICT (discount_code_id, tenant_id) DO UPDATE SET updated_at = discount_redemptions.updated_at
			RETURNING redemption_id::text, discount_code_id::text, tenant_id::text, code, discount_type,
			          value_irr::bigint, percent::double precision, starts_at, expires_at, status,
			          used_amount_irr::bigint, COALESCE(redeemed_by, ''), created_at, updated_at,
			          (xmax = 0) AS was_inserted
		)
		SELECT * FROM inserted
	`, strings.TrimSpace(code), tenantID, strings.TrimSpace(actor)).Scan(
		&item.ID, &item.DiscountCodeID, &item.TenantID, &item.Code, &item.DiscountType,
		&item.ValueIRR, &item.Percent, &item.StartsAt, &item.ExpiresAt, &item.Status,
		&item.UsedAmountIRR, &item.RedeemedBy, &item.CreatedAt, &item.UpdatedAt, &wasInserted,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFinancialCodeInvalid
		}

		return nil, err
	}

	if wasInserted {
		_, err = s.pool.Exec(ctx, `
			UPDATE discount_codes
			SET redeemed_count = redeemed_count + 1, updated_at = NOW()
			WHERE discount_code_id::text = $1
		`, item.DiscountCodeID)
		if err != nil {
			return nil, err
		}
	}

	return &item, nil
}

func (s *Store) GetBillingProfile(ctx context.Context, tenantID string) (*platform.BillingProfile, error) {
	item, err := scanBillingProfile(s.pool.QueryRow(ctx, `
		SELECT billing_profile_id::text, tenant_id::text, profile_type, data,
		       COALESCE(created_by, ''), COALESCE(updated_by, ''), created_at, updated_at
		FROM billing_profiles
		WHERE tenant_id::text = $1
	`, tenantID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, pgx.ErrNoRows
	}

	return item, err
}

func (s *Store) UpsertBillingProfile(ctx context.Context, tenantID string, req platform.UpsertBillingProfileRequest, actor string) (*platform.BillingProfile, error) {
	profileType := strings.TrimSpace(strings.ToLower(req.ProfileType))
	if profileType != "individual" && profileType != "company" {
		return nil, ErrBillingProfileInvalid
	}

	if req.Data == nil {
		req.Data = map[string]interface{}{}
	}

	return scanBillingProfile(s.pool.QueryRow(ctx, `
		INSERT INTO billing_profiles (tenant_id, profile_type, data, created_by, updated_by)
		VALUES ($1::uuid, $2, $3, NULLIF($4, ''), NULLIF($4, ''))
		ON CONFLICT (tenant_id) DO UPDATE
		SET profile_type = EXCLUDED.profile_type,
		    data = EXCLUDED.data,
		    updated_by = EXCLUDED.updated_by,
		    updated_at = NOW()
		RETURNING billing_profile_id::text, tenant_id::text, profile_type, data,
		          COALESCE(created_by, ''), COALESCE(updated_by, ''), created_at, updated_at
	`, tenantID, profileType, marshalJSONBestEffort(req.Data), strings.TrimSpace(actor)))
}

func (s *Store) ListInvoices(ctx context.Context, tenantID string, limit int) ([]platform.Invoice, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, invoiceSelect()+`
		WHERE tenant_id::text = $1
		ORDER BY period_end DESC, created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.Invoice, 0, limit)

	for rows.Next() {
		item, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func (s *Store) ListDiscountRedemptions(ctx context.Context, tenantID string, limit int) ([]platform.DiscountRedemption, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	rows, err := s.pool.Query(ctx, `
		SELECT redemption_id::text, discount_code_id::text, tenant_id::text, code, discount_type,
		       value_irr::bigint, percent::double precision, starts_at, expires_at, status,
		       used_amount_irr::bigint, COALESCE(redeemed_by, ''), created_at, updated_at
		FROM discount_redemptions
		WHERE tenant_id::text = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]platform.DiscountRedemption, 0, limit)

	for rows.Next() {
		item, err := scanDiscountRedemption(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, *item)
	}

	return items, rows.Err()
}

func (s *Store) GetBillingUsageSummary(ctx context.Context, tenantID string, rangeKey string) (*platform.BillingUsageSummary, error) {
	start, end, normalized, err := BillingReportWindow(time.Now().UTC(), rangeKey)
	if err != nil {
		return nil, err
	}

	providers, err := s.queryBillingProviderUsage(ctx, tenantID, start, end)
	if err != nil {
		return nil, err
	}

	out := &platform.BillingUsageSummary{Calendar: FinancialCalendar, Preset: normalized, Start: start, End: end, Providers: providers}
	for _, item := range providers {
		out.TotalUsageIRR += item.FinalCostIRR
		out.DiscountIRR += item.DiscountIRR
		out.CacheSavingsIRR += item.CacheSavingsIRR
	}

	out.GiftCreditIRR, err = s.sumGiftCredits(ctx, tenantID, start, end)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) GetMonthlyUsageDetail(ctx context.Context, tenantID string, jalaliYear int, jalaliMonth int) (*platform.MonthlyUsageDetail, error) {
	start, end, err := JalaliMonthWindow(jalaliYear, jalaliMonth)
	if err != nil {
		return nil, err
	}

	providers, err := s.queryBillingProviderUsage(ctx, tenantID, start, end)
	if err != nil {
		return nil, err
	}

	out := &platform.MonthlyUsageDetail{
		Calendar:    FinancialCalendar,
		JalaliYear:  jalaliYear,
		JalaliMonth: jalaliMonth,
		Start:       start,
		End:         end,
		Providers:   providers,
	}
	for _, item := range providers {
		out.TotalUsageIRR += item.FinalCostIRR
		out.DiscountIRR += item.DiscountIRR
		out.CacheSavingsIRR += item.CacheSavingsIRR
	}

	out.GiftCreditIRR, err = s.sumGiftCredits(ctx, tenantID, start, end)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Store) queryBillingProviderUsage(ctx context.Context, tenantID string, start time.Time, end time.Time) ([]platform.BillingProviderUsage, error) {
	if !s.clickHouseMetricsReady() {
		return s.queryBillingUsageFromWallet(ctx, tenantID, start, end)
	}

	rate, err := s.GetLatestExchangeRate(ctx)
	if err != nil || rate == nil || rate.USDtoIRR <= 0 {
		rate = &platform.ExchangeRate{USDtoIRR: 0}
	}

	rows, err := s.clickhouse.QueryContext(ctx, `
		SELECT
			if(provider_name = '', 'unknown', provider_name) AS provider,
			count() AS request_count,
			sum(input_tokens) AS input_tokens,
			sum(output_tokens) AS output_tokens,
			round(sum(cost_usd) * ?) AS raw_cost_irr,
			round(sum(avoided_cost_irr)) AS cache_savings_irr
		FROM gateway_model_events
		WHERE tenant_id = ? AND ts >= ? AND ts < ?
		GROUP BY provider
		ORDER BY raw_cost_irr DESC
		LIMIT 100
	`, rate.USDtoIRR, tenantID, start, end)
	if err != nil {
		return nil, err
	}
	defer closeBestEffort(rows)

	items := make([]platform.BillingProviderUsage, 0)

	var total int64

	for rows.Next() {
		var (
			item                        platform.BillingProviderUsage
			rawCostIRR, cacheSavingsIRR any
		)
		if err := rows.Scan(&item.ProviderName, &item.RequestCount, &item.InputTokens, &item.OutputTokens, &rawCostIRR, &cacheSavingsIRR); err != nil {
			return nil, err
		}

		item.RawCostIRR = billingIRRValue(rawCostIRR)
		item.CacheSavingsIRR = billingIRRValue(cacheSavingsIRR)

		item.FinalCostIRR = item.RawCostIRR - item.DiscountIRR
		if item.FinalCostIRR < 0 {
			item.FinalCostIRR = 0
		}

		total += item.FinalCostIRR
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range items {
		if total > 0 {
			items[i].Share = float64(items[i].FinalCostIRR) / float64(total)
		}
	}

	return items, nil
}

func (s *Store) queryBillingUsageFromWallet(ctx context.Context, tenantID string, start time.Time, end time.Time) ([]platform.BillingProviderUsage, error) {
	var item platform.BillingProviderUsage

	item.ProviderName = "gateway"

	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(COUNT(1), 0)::bigint, COALESCE(SUM(amount_irr), 0)::bigint
		FROM wallet_transactions
		WHERE tenant_id::text = $1
		  AND tx_type IN ('usage_debit', 'CONSUMPTION')
		  AND created_at >= $2 AND created_at < $3
	`, tenantID, start, end).Scan(&item.RequestCount, &item.RawCostIRR)
	if err != nil {
		return nil, err
	}

	item.FinalCostIRR = item.RawCostIRR
	if item.FinalCostIRR > 0 {
		item.Share = 1
		return []platform.BillingProviderUsage{item}, nil
	}

	return []platform.BillingProviderUsage{}, nil
}

func billingIRRValue(value any) int64 {
	switch v := value.(type) {
	case nil:
		return 0
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case uint64:
		if v > math.MaxInt64 {
			return math.MaxInt64
		}

		return int64(v)
	case uint:
		if uint64(v) > math.MaxInt64 {
			return math.MaxInt64
		}

		return int64(v)
	case uint32:
		return int64(v)
	case float64:
		return int64(math.Round(v))
	case float32:
		return int64(math.Round(float64(v)))
	case []byte:
		return billingIRRString(string(v))
	case string:
		return billingIRRString(v)
	default:
		return billingIRRString(fmt.Sprintf("%v", v))
	}
}

func billingIRRString(value string) int64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}

	return int64(math.Round(parsed))
}

func (s *Store) sumGiftCredits(ctx context.Context, tenantID string, start time.Time, end time.Time) (int64, error) {
	var out int64

	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_irr), 0)::bigint
		FROM gift_code_redemptions
		WHERE tenant_id::text = $1 AND created_at >= $2 AND created_at < $3
	`, tenantID, start, end).Scan(&out)

	return out, err
}

func scanBankTransferRequest(row interface{ Scan(dest ...any) error }) (*platform.BankTransferRequest, error) {
	var item platform.BankTransferRequest
	if err := row.Scan(
		&item.ID, &item.PublicID, &item.TenantID, &item.AmountIRR, &item.TrackingNumber, &item.BankTransactionAt,
		&item.ReceiptFileURL, &item.Status, &item.RejectionReason, &item.WalletTxID, &item.ReviewedBy,
		&item.ReviewedAt, &item.Metadata, &item.CreatedBy, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func scanGiftCodeRedemption(row interface{ Scan(dest ...any) error }) (*platform.GiftCodeRedemption, error) {
	var item platform.GiftCodeRedemption
	if err := row.Scan(&item.ID, &item.GiftCodeID, &item.TenantID, &item.WalletTxID, &item.AmountIRR, &item.Status, &item.RedeemedBy, &item.CreatedAt); err != nil {
		return nil, err
	}

	return &item, nil
}

func scanDiscountRedemption(row interface{ Scan(dest ...any) error }) (*platform.DiscountRedemption, error) {
	var item platform.DiscountRedemption
	if err := row.Scan(
		&item.ID, &item.DiscountCodeID, &item.TenantID, &item.Code, &item.DiscountType,
		&item.ValueIRR, &item.Percent, &item.StartsAt, &item.ExpiresAt, &item.Status,
		&item.UsedAmountIRR, &item.RedeemedBy, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func scanBillingProfile(row interface{ Scan(dest ...any) error }) (*platform.BillingProfile, error) {
	var item platform.BillingProfile
	if err := row.Scan(&item.ID, &item.TenantID, &item.ProfileType, &item.Data, &item.CreatedBy, &item.UpdatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}

	return &item, nil
}

func invoiceSelect() string {
	return `
		SELECT invoice_id::text, invoice_number, tenant_id::text, COALESCE(subscription_id::text, ''),
		       status, issued_at, period_start, period_end, COALESCE(tenant_name, ''),
		       billing_profile_snapshot, raw_usage_irr::bigint, discount_irr::bigint,
		       gift_credit_irr::bigint, cache_savings_irr::bigint, final_amount_irr::bigint,
		       net_adjustment_irr::bigint, COALESCE(wallet_tx_id::text, ''), metadata, created_at, updated_at
		FROM invoices
	`
}

func scanInvoice(row interface{ Scan(dest ...any) error }) (*platform.Invoice, error) {
	var item platform.Invoice
	if err := row.Scan(
		&item.ID, &item.InvoiceNumber, &item.TenantID, &item.SubscriptionID, &item.Status, &item.IssuedAt,
		&item.PeriodStart, &item.PeriodEnd, &item.TenantName, &item.BillingProfileSnapshot,
		&item.RawUsageIRR, &item.DiscountIRR, &item.GiftCreditIRR, &item.CacheSavingsIRR,
		&item.FinalAmountIRR, &item.NetAdjustmentIRR, &item.WalletTxID, &item.Metadata,
		&item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

func insertFinancialAuditTx(ctx context.Context, tx pgx.Tx, tenantID, eventType string, actor platform.SupportActor, sourceType, sourceID, idempotencyKey string, payload map[string]interface{}) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO financial_audit_events (
			tenant_id, event_type, actor_id, actor_role, source_type, source_id, idempotency_key, payload
		)
		VALUES ($1::uuid, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8)
	`, tenantID, eventType, actor.ID, actor.Role, sourceType, sourceID, idempotencyKey, marshalJSONBestEffort(payload))

	return err
}
