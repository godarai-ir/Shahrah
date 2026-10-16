package billing

// PaymentVerification is what a payment gateway confirmed for a payment.
type PaymentVerification struct {
	Raw          map[string]interface{}
	RefID        string
	GatewayTxnID string
	CardMask     string
	PayerPhone   string
}
