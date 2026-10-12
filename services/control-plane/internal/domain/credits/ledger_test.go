package credits

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestSortForConsumptionFollowsExpiryThenSourceThenAge(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	monthEnd := base.AddDate(0, 1, 0)
	yearEnd := base.AddDate(1, 0, 0)

	lots := []SpendableLot{
		{ID: "never", RemainingMC: 1, SourceRank: SourceRank(LotSourceAdjust), CreatedAt: base},
		{ID: "addon-year", RemainingMC: 1, ExpiresAt: &yearEnd, SourceRank: SourceRank(LotSourcePurchaseAddon), CreatedAt: base},
		{ID: "addon-month", RemainingMC: 1, ExpiresAt: &monthEnd, SourceRank: SourceRank(LotSourcePurchaseAddon), CreatedAt: base},
		{ID: "promo-month", RemainingMC: 1, ExpiresAt: &monthEnd, SourceRank: SourceRank(LotSourcePromo), CreatedAt: base},
		{ID: "plan-month-new", RemainingMC: 1, ExpiresAt: &monthEnd, SourceRank: SourceRank(LotSourcePlan), CreatedAt: base.Add(time.Hour)},
		{ID: "plan-month-old", RemainingMC: 1, ExpiresAt: &monthEnd, SourceRank: SourceRank(LotSourcePlan), CreatedAt: base},
	}

	SortForConsumption(lots)

	order := make([]string, 0, len(lots))
	for _, lot := range lots {
		order = append(order, lot.ID)
	}

	require.Equal(t, []string{"plan-month-old", "plan-month-new", "promo-month", "addon-month", "addon-year", "never"}, order)
}

func TestSourceRankGroupsTrialPromoAndCompensation(t *testing.T) {
	t.Parallel()

	require.Equal(t, SourceRank(LotSourceTrial), SourceRank(LotSourcePromo))
	require.Equal(t, SourceRank(LotSourceTrial), SourceRank(LotSourceCompensation))
	require.Less(t, SourceRank(LotSourcePlan), SourceRank(LotSourceTrial))
	require.Less(t, SourceRank(LotSourceTrial), SourceRank(LotSourcePurchaseAddon))
}

func TestAllocateDebitSpendsInOrderAndRecordsDebt(t *testing.T) {
	t.Parallel()

	lots := []SpendableLot{{ID: "a", RemainingMC: 300}, {ID: "b", RemainingMC: 0}, {ID: "c", RemainingMC: 500}}

	require.Equal(t, []Allocation{{LotID: "a", AmountMC: -250}}, AllocateDebit(lots, 250))
	require.Equal(t, []Allocation{{LotID: "a", AmountMC: -300}, {LotID: "c", AmountMC: -200}}, AllocateDebit(lots, 500))
	require.Equal(t, []Allocation{{LotID: "a", AmountMC: -300}, {LotID: "c", AmountMC: -500}, {AmountMC: -40}}, AllocateDebit(lots, 840))
	require.Equal(t, []Allocation{{AmountMC: -10}}, AllocateDebit(nil, 10))
	require.Empty(t, AllocateDebit(lots, 0))
}

func TestAllocateDebitConservesTheAmount(t *testing.T) {
	t.Parallel()

	lots := []SpendableLot{{ID: "a", RemainingMC: 70}, {ID: "b", RemainingMC: 30}}

	for _, amount := range []int64{1, 30, 70, 99, 100, 101, 1_000_000} {
		var sum int64
		for _, allocation := range AllocateDebit(lots, amount) {
			sum += allocation.AmountMC
		}

		require.Equal(t, -amount, sum, "amount %d", amount)
	}
}

func TestRepayDebtPaysOldestFirst(t *testing.T) {
	t.Parallel()

	debts := []DebtEntry{{EntryID: "old", OutstandingMC: 40}, {EntryID: "mid", OutstandingMC: 0}, {EntryID: "new", OutstandingMC: 100}}

	repayments, left := RepayDebt(debts, 90)
	require.Equal(t, []DebtRepayment{{EntryID: "old", AmountMC: 40}, {EntryID: "new", AmountMC: 50}}, repayments)
	require.Zero(t, left)

	repayments, left = RepayDebt(debts, 1_000)
	require.Len(t, repayments, 2)
	require.Equal(t, int64(860), left)

	repayments, left = RepayDebt(nil, 5)
	require.Empty(t, repayments)
	require.Equal(t, int64(5), left)
}

func TestBYOKFeeRoundsUpOnce(t *testing.T) {
	t.Parallel()

	require.Equal(t, int64(250), BYOKFeeMC(5_000, DefaultBYOKFeePercent))
	require.Equal(t, int64(10), BYOKFeeMC(10, DefaultBYOKFeePercent), "0.5 mc rounds up to one step")
	require.Equal(t, int64(10), BYOKFeeMC(190, DefaultBYOKFeePercent), "9.5 mc rounds up to one step")
	require.Equal(t, int64(20), BYOKFeeMC(210, DefaultBYOKFeePercent))
	require.Equal(t, int64(80), BYOKFeeMC(1_000, decimal.RequireFromString("7.5")))
	require.Zero(t, BYOKFeeMC(1_000, decimal.Zero))
	require.Zero(t, BYOKFeeMC(0, DefaultBYOKFeePercent))
}

func TestIdempotencyKeysAreDeterministic(t *testing.T) {
	t.Parallel()

	require.Equal(t, "consume:t1:req-1:gateway", ConsumeKey("t1", " req-1 ", "gateway", 9))
	require.Equal(t, "consume:t1:log:9:gateway", ConsumeKey("t1", "", "gateway", 9))
	require.Equal(t, "expire:lot-1", ExpireKey("lot-1"))
	require.Equal(t, "plan_grant:sub-1:2026-10-01T00:00:00Z", PlanGrantKey("sub-1", time.Date(2026, 10, 1, 3, 30, 0, 0, time.FixedZone("IRST", 3*3600+1800))))
	require.Equal(t, "trial_grant:sub-1", TrialGrantKey("sub-1"))
	require.Equal(t, "purchase:order-1", PurchaseKey("order-1"))
}

func TestValidLotSource(t *testing.T) {
	t.Parallel()

	require.True(t, ValidLotSource(LotSourceOverage))
	require.False(t, ValidLotSource("gift"))
}
