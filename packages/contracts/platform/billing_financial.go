package platform

import "time"

type BankTransferRequest struct {
	ID                string                 `json:"id"`
	PublicID          string                 `json:"public_id"`
	TenantID          string                 `json:"tenant_id"`
	AmountIRR         int64                  `json:"amount_irr"`
	TrackingNumber    string                 `json:"tracking_number"`
	BankTransactionAt time.Time              `json:"bank_transaction_at"`
	ReceiptFileURL    string                 `json:"receipt_file_url,omitempty"`
	Status            string                 `json:"status"`
	RejectionReason   string                 `json:"rejection_reason,omitempty"`
	WalletTxID        string                 `json:"wallet_tx_id,omitempty"`
	ReviewedBy        string                 `json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time             `json:"reviewed_at,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy         string                 `json:"created_by,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type CreateBankTransferRequest struct {
	AmountIRR         int64                  `json:"amount_irr"`
	TrackingNumber    string                 `json:"tracking_number"`
	BankTransactionAt time.Time              `json:"bank_transaction_at"`
	ReceiptFileURL    string                 `json:"receipt_file_url,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

type ReviewBankTransferRequest struct {
	Reason         string `json:"reason,omitempty"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

type GiftCodeRedemption struct {
	ID         string    `json:"id"`
	GiftCodeID string    `json:"gift_code_id"`
	TenantID   string    `json:"tenant_id"`
	WalletTxID string    `json:"wallet_tx_id,omitempty"`
	AmountIRR  int64     `json:"amount_irr"`
	Status     string    `json:"status"`
	RedeemedBy string    `json:"redeemed_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type DiscountRedemption struct {
	ID             string     `json:"id"`
	DiscountCodeID string     `json:"discount_code_id"`
	TenantID       string     `json:"tenant_id"`
	Code           string     `json:"code"`
	DiscountType   string     `json:"discount_type"`
	ValueIRR       *int64     `json:"value_irr,omitempty"`
	Percent        *float64   `json:"percent,omitempty"`
	StartsAt       *time.Time `json:"starts_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	Status         string     `json:"status"`
	UsedAmountIRR  int64      `json:"used_amount_irr"`
	RedeemedBy     string     `json:"redeemed_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type RedeemCodeRequest struct {
	Code string `json:"code"`
}

type BillingProfile struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	ProfileType string                 `json:"profile_type"`
	Data        map[string]interface{} `json:"data"`
	CreatedBy   string                 `json:"created_by,omitempty"`
	UpdatedBy   string                 `json:"updated_by,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type UpsertBillingProfileRequest struct {
	ProfileType string                 `json:"profile_type"`
	Data        map[string]interface{} `json:"data"`
}

type Invoice struct {
	ID                     string                 `json:"id"`
	InvoiceNumber          string                 `json:"invoice_number"`
	TenantID               string                 `json:"tenant_id"`
	SubscriptionID         string                 `json:"subscription_id,omitempty"`
	Status                 string                 `json:"status"`
	IssuedAt               *time.Time             `json:"issued_at,omitempty"`
	PeriodStart            time.Time              `json:"period_start"`
	PeriodEnd              time.Time              `json:"period_end"`
	TenantName             string                 `json:"tenant_name,omitempty"`
	BillingProfileSnapshot map[string]interface{} `json:"billing_profile_snapshot,omitempty"`
	RawUsageIRR            int64                  `json:"raw_usage_irr"`
	DiscountIRR            int64                  `json:"discount_irr"`
	GiftCreditIRR          int64                  `json:"gift_credit_irr"`
	CacheSavingsIRR        int64                  `json:"cache_savings_irr"`
	FinalAmountIRR         int64                  `json:"final_amount_irr"`
	NetAdjustmentIRR       int64                  `json:"net_adjustment_irr"`
	WalletTxID             string                 `json:"wallet_tx_id,omitempty"`
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
	LineItems              []InvoiceLineItem      `json:"line_items,omitempty"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

type InvoiceLineItem struct {
	ID           string                 `json:"id"`
	InvoiceID    string                 `json:"invoice_id"`
	ItemType     string                 `json:"item_type"`
	ProviderName string                 `json:"provider_name,omitempty"`
	Description  string                 `json:"description"`
	Quantity     float64                `json:"quantity"`
	AmountIRR    int64                  `json:"amount_irr"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

type BillingProviderUsage struct {
	ProviderName    string  `json:"provider_name"`
	RequestCount    int64   `json:"request_count"`
	InputTokens     int64   `json:"input_tokens"`
	OutputTokens    int64   `json:"output_tokens"`
	RawCostIRR      int64   `json:"raw_cost_irr"`
	DiscountIRR     int64   `json:"discount_irr"`
	CacheSavingsIRR int64   `json:"cache_savings_irr"`
	FinalCostIRR    int64   `json:"final_cost_irr"`
	Share           float64 `json:"share"`
}

type BillingUsageSummary struct {
	Range           string                 `json:"range"`
	Start           time.Time              `json:"start"`
	End             time.Time              `json:"end"`
	TotalUsageIRR   int64                  `json:"total_usage_irr"`
	DiscountIRR     int64                  `json:"discount_irr"`
	CacheSavingsIRR int64                  `json:"cache_savings_irr"`
	GiftCreditIRR   int64                  `json:"gift_credit_irr"`
	Providers       []BillingProviderUsage `json:"providers"`
}

type MonthlyUsageDetail struct {
	Month           string                 `json:"month"`
	Start           time.Time              `json:"start"`
	End             time.Time              `json:"end"`
	TotalUsageIRR   int64                  `json:"total_usage_irr"`
	GiftCreditIRR   int64                  `json:"gift_credit_irr"`
	DiscountIRR     int64                  `json:"discount_irr"`
	CacheSavingsIRR int64                  `json:"cache_savings_irr"`
	Providers       []BillingProviderUsage `json:"providers"`
}

type BankTransferListResponse struct {
	Data []BankTransferRequest `json:"data"`
}

type InvoiceListResponse struct {
	Data []Invoice `json:"data"`
}

type TransactionListResponse struct {
	Data []Transaction `json:"data"`
}

type DiscountRedemptionListResponse struct {
	Data []DiscountRedemption `json:"data"`
}
