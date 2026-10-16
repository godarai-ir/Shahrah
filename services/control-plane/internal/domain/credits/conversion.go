package credits

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// Conversion directions.
const (
	ConversionToCredit = "to_credit"
	ConversionToIRR    = "to_irr"
)

// ConvertKey is the idempotency key of a tenant's n-th conversion.
func ConvertKey(tenantID string, sequence int, part string) string {
	return fmt.Sprintf("convert:%s:%d:%s", strings.TrimSpace(tenantID), sequence, part)
}

// ConvertIRRToCreditsMC converts an IRR wallet balance to whole credits at a
// list price in toman per credit. It rounds down and returns the IRR residue
// that did not make a whole credit; the residue stays in the IRR wallet.
func ConvertIRRToCreditsMC(amountIRR, listPriceTomanPerCredit int64) (creditsMC, residueIRR int64) {
	if amountIRR <= 0 || listPriceTomanPerCredit <= 0 {
		return 0, max(amountIRR, 0)
	}

	irrPerCredit := listPriceTomanPerCredit * IRRPerToman
	whole := amountIRR / irrPerCredit

	return whole * MillicreditsPerCredit, amountIRR - whole*irrPerCredit
}

// ConvertIRRLimitToCreditsMC converts an IRR budget limit to credits at a list
// price in toman per credit, rounding up to a whole credit so no limit gets
// stricter.
func ConvertIRRLimitToCreditsMC(limitIRR decimal.Decimal, listPriceTomanPerCredit int64) int64 {
	if !limitIRR.IsPositive() || listPriceTomanPerCredit <= 0 {
		return 0
	}

	irrPerCredit := decimal.NewFromInt(listPriceTomanPerCredit * IRRPerToman)

	return limitIRR.Div(irrPerCredit).Ceil().IntPart() * MillicreditsPerCredit
}

// CreditsToIRR converts millicredits back to IRR at the list price a
// conversion used, rounding down to a whole rial.
func CreditsToIRR(amountMC, listPriceTomanPerCredit int64) int64 {
	if amountMC <= 0 || listPriceTomanPerCredit <= 0 {
		return 0
	}

	return decimal.NewFromInt(amountMC).
		Mul(decimal.NewFromInt(listPriceTomanPerCredit * IRRPerToman)).
		Div(decimal.NewFromInt(MillicreditsPerCredit)).
		Floor().IntPart()
}
