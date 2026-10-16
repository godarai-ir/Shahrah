package credits

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// LotSource is the origin of a credit lot.
type LotSource string

// Lot sources stored in credit_lots.source.
const (
	LotSourcePlan          LotSource = "plan"
	LotSourceTrial         LotSource = "trial"
	LotSourcePromo         LotSource = "promo"
	LotSourcePurchaseAddon LotSource = "purchase_addon"
	LotSourceMigration     LotSource = "migration"
	LotSourceCompensation  LotSource = "compensation"
	LotSourceAdjust        LotSource = "adjust"
	LotSourceOverage       LotSource = "overage"
)

// EntryKind is the kind of a credit ledger entry.
type EntryKind string

// Ledger entry kinds stored in credit_ledger.kind.
const (
	EntryPurchase   EntryKind = "purchase"
	EntryPlanGrant  EntryKind = "plan_grant"
	EntryPromoGrant EntryKind = "promo_grant"
	EntryConsume    EntryKind = "consume"
	EntryRefund     EntryKind = "refund"
	EntryExpire     EntryKind = "expire"
	EntryAdjust     EntryKind = "adjust"
)

// Billing units stored in tenants.billing_unit.
const (
	BillingUnitIRR    = "irr"
	BillingUnitCredit = "credit"
)

// DefaultBYOKFeePercent is the platform fee on externally billed usage when no
// entitlement or override sets one.
var DefaultBYOKFeePercent = decimal.NewFromInt(5)

// SourceRank orders lots that expire at the same time: plan credits are spent
// first, then trial, promo, and compensation credits, then purchased add-ons.
func SourceRank(source LotSource) int16 {
	switch source {
	case LotSourcePlan:
		return 0
	case LotSourceTrial, LotSourcePromo, LotSourceCompensation:
		return 1
	case LotSourcePurchaseAddon:
		return 2
	default:
		return 3
	}
}

// ValidLotSource reports whether source is a known lot source.
func ValidLotSource(source LotSource) bool {
	switch source {
	case LotSourcePlan, LotSourceTrial, LotSourcePromo, LotSourcePurchaseAddon,
		LotSourceMigration, LotSourceCompensation, LotSourceAdjust, LotSourceOverage:
		return true
	default:
		return false
	}
}

// SpendableLot is a lot with a positive remaining balance.
type SpendableLot struct {
	ID          string
	RemainingMC int64
	ExpiresAt   *time.Time
	SourceRank  int16
	CreatedAt   time.Time
}

// Allocation moves AmountMC into or out of a lot. An empty LotID is debt.
type Allocation struct {
	LotID    string
	AmountMC int64
}

// SortForConsumption orders lots by earliest expiry first (never-expiring lots
// last), then by source rank, then by age, which is the PRD consumption order.
func SortForConsumption(lots []SpendableLot) {
	sort.SliceStable(lots, func(i, j int) bool {
		a, b := lots[i], lots[j]
		if !sameExpiry(a.ExpiresAt, b.ExpiresAt) {
			return expiresBefore(a.ExpiresAt, b.ExpiresAt)
		}

		if a.SourceRank != b.SourceRank {
			return a.SourceRank < b.SourceRank
		}

		if !a.CreatedAt.Equal(b.CreatedAt) {
			return a.CreatedAt.Before(b.CreatedAt)
		}

		return a.ID < b.ID
	})
}

func sameExpiry(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	return a.Equal(*b)
}

func expiresBefore(a, b *time.Time) bool {
	if a == nil {
		return false
	}

	if b == nil {
		return true
	}

	return a.Before(*b)
}

// AllocateDebit spends amountMC (a positive amount) from lots, which must
// already be in consumption order. The returned allocations are negative. Any
// shortfall is a single debt allocation with an empty LotID.
func AllocateDebit(lots []SpendableLot, amountMC int64) []Allocation {
	if amountMC <= 0 {
		return nil
	}

	allocations := make([]Allocation, 0, len(lots)+1)
	left := amountMC

	for _, lot := range lots {
		if left == 0 {
			break
		}

		if lot.RemainingMC <= 0 {
			continue
		}

		take := min(lot.RemainingMC, left)
		allocations = append(allocations, Allocation{LotID: lot.ID, AmountMC: -take})
		left -= take
	}

	if left > 0 {
		allocations = append(allocations, Allocation{AmountMC: -left})
	}

	return allocations
}

// DebtEntry is a ledger entry that still owes OutstandingMC (positive).
type DebtEntry struct {
	EntryID       string
	OutstandingMC int64
}

// DebtRepayment is the pair of allocations that moves part of an entry's debt
// onto a newly granted lot without changing the entry's allocation sum.
type DebtRepayment struct {
	EntryID  string
	AmountMC int64
}

// RepayDebt pays the oldest debts first from a grant of grantMC and returns the
// repayments and the amount the grant still has left.
func RepayDebt(debts []DebtEntry, grantMC int64) ([]DebtRepayment, int64) {
	repayments := make([]DebtRepayment, 0, len(debts))
	left := grantMC

	for _, debt := range debts {
		if left == 0 {
			break
		}

		if debt.OutstandingMC <= 0 {
			continue
		}

		pay := min(debt.OutstandingMC, left)
		repayments = append(repayments, DebtRepayment{EntryID: debt.EntryID, AmountMC: pay})
		left -= pay
	}

	return repayments, left
}

// BYOKFeeMC is the balance debit for externally billed usage: the fee percent
// of the credit equivalent, rounded up once to 0.01 credit.
func BYOKFeeMC(equivalentMC int64, feePercent decimal.Decimal) int64 {
	if equivalentMC <= 0 || !feePercent.IsPositive() {
		return 0
	}

	exact := decimal.NewFromInt(equivalentMC).Mul(feePercent).Div(decimal.NewFromInt(100))

	return int64(RoundUp(exact))
}

// ConsumeKey is the idempotency key of the consume entry for one request and product.
func ConsumeKey(tenantID, requestID, product string, spendLogID int64) string {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return fmt.Sprintf("consume:%s:log:%d:%s", tenantID, spendLogID, product)
	}

	return fmt.Sprintf("consume:%s:%s:%s", tenantID, requestID, product)
}

// ExpireKey is the idempotency key of a lot's expiry entry.
func ExpireKey(lotID string) string {
	return "expire:" + lotID
}

// PlanGrantKey is the idempotency key of one monthly plan grant.
func PlanGrantKey(subscriptionID string, sliceStart time.Time) string {
	return fmt.Sprintf("plan_grant:%s:%s", subscriptionID, sliceStart.UTC().Format(time.RFC3339))
}

// TrialGrantKey is the idempotency key of a subscription's trial grant.
func TrialGrantKey(subscriptionID string) string {
	return "trial_grant:" + subscriptionID
}

// PurchaseKey is the idempotency key of a paid add-on order.
func PurchaseKey(orderID string) string {
	return "purchase:" + orderID
}
