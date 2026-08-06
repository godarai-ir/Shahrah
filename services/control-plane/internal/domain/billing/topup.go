package billing

// CreateTopUpIntentParams contains the validated values needed to create a wallet top-up intent.
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
